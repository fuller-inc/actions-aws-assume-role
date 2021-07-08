package github

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const (
	githubUserAgent   = "actions-aws-assume-role/1.0"
	defaultAPIBaseURL = "https://api.github.com"
)

var apiBaseURL string

func init() {
	u := os.Getenv("GITHUB_API_URL")
	if u == "" {
		u = defaultAPIBaseURL
	}

	var err error
	apiBaseURL, err = canonicalURL(u)
	if err != nil {
		panic(err)
	}
}

type UnexpectedStatusCodeError struct {
	StatusCode int
}

func (err *UnexpectedStatusCodeError) Error() string {
	return fmt.Sprintf("unexpected status code: %d", err.StatusCode)
}

// Client is a very light weight GitHub API Client.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{
		baseURL:    apiBaseURL,
		httpClient: httpClient,
	}
}

type CommitState string

const (
	CommitStateError   CommitState = "error"
	CommitStateFailure CommitState = "failure"
	CommitStatePending CommitState = "pending"
	CommitStateSuccess CommitState = "success"
)

type CreateStatusRequest struct {
	State       CommitState `json:"state"`
	TargetURL   string      `json:"target_url,omitempty"`
	Description string      `json:"description,omitempty"`
	Context     string      `json:"context,omitempty"`
}

type CreateStatusResponse struct {
	Creator *CreateStatusResponseCreator `json:"creator"`
	// omit other fields, we don't use them.
}

type CreateStatusResponseCreator struct {
	Login string `json:"login"`
	ID    int64  `json:"id"`
	Type  string `json:"type"`
	// omit other fields, we don't use them.
}

type GetRepoResponse struct {
	NodeID string `json:"node_id"`

	// omit other fields, we don't use them.
}

// CreateStatus creates a commit status.
// https://docs.github.com/en/rest/reference/repos#create-a-commit-status
func (c *Client) CreateStatus(ctx context.Context, token, owner, repo, ref string, status *CreateStatusRequest) (*CreateStatusResponse, error) {
	// build the request
	u := fmt.Sprintf("%s/repos/%s/%s/statuses/%s", c.baseURL, owner, repo, ref)
	body, err := json.Marshal(status)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", githubUserAgent)
	req.Header.Set("Authorization", "token "+token)

	// send the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// parse the response
	if resp.StatusCode != http.StatusCreated {
		return nil, &UnexpectedStatusCodeError{StatusCode: resp.StatusCode}
	}

	var ret *CreateStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
}

func (c *Client) GetRepo(ctx context.Context, token, owner, repo string) (*GetRepoResponse, error) {
	// build the request
	u := fmt.Sprintf("%s/repos/%s/%s", c.baseURL, owner, repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", githubUserAgent)
	req.Header.Set("Authorization", "token "+token)

	// send the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// parse the response
	if resp.StatusCode != http.StatusCreated {
		return nil, &UnexpectedStatusCodeError{StatusCode: resp.StatusCode}
	}

	var ret *GetRepoResponse
	if err := json.NewDecoder(resp.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
}

func (c *Client) ValidateAPIURL(url string) error {
	// for backward compatibility, treat zero string as defaultAPIBaseURL
	if url == "" {
		url = defaultAPIBaseURL
	}

	u, err := canonicalURL(url)
	if err != nil {
		return err
	}
	if u != c.baseURL {
		if c.baseURL == defaultAPIBaseURL {
			return errors.New(
				"it looks that you use GitHub Enterprise Server, " +
					"but the credential provider doesn't support it. " +
					"I recommend you to build your own credential provider",
			)
		}
		return errors.New("your api server is not verified by the credential provider")
	}
	return nil
}

func canonicalURL(rawurl string) (string, error) {
	u, err := url.Parse(rawurl)
	if err != nil {
		return "", err
	}

	host := u.Hostname()
	port := u.Port()

	// host is case insensitive.
	host = strings.ToLower(host)

	// remove trailing slashes.
	u.Path = strings.TrimRight(u.Path, "/")

	// omit the default port number.
	defaultPort := "80"
	switch u.Scheme {
	case "http":
	case "https":
		defaultPort = "443"
	case "":
		u.Scheme = "http"
	default:
		return "", fmt.Errorf("unknown scheme: %s", u.Scheme)
	}
	if port == defaultPort {
		port = ""
	}

	if port == "" {
		u.Host = host
	} else {
		u.Host = net.JoinHostPort(host, port)
	}

	// we don't use query and fragment, so drop them.
	u.RawFragment = ""
	u.RawQuery = ""

	return u.String(), nil
}
