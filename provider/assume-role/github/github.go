package github

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/shogo82148/goat/oidc"
)

const (
	// The value of User-Agent header
	githubUserAgent = "actions-aws-assume-role/1.0"

	// The default url of Github API
	defaultAPIBaseURL = "https://api.github.com"

	oidcIssuer = "https://token.actions.githubusercontent.com"
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
	Body       string
}

func (err *UnexpectedStatusCodeError) Error() string {
	return fmt.Sprintf("unexpected status code: %d, body: %q", err.StatusCode, err.Body)
}

func handleUnexpectedStatusCode(resp *http.Response) error {
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			body = []byte("error: " + err.Error())
		}
		return &UnexpectedStatusCodeError{
			StatusCode: resp.StatusCode,
			Body:       string(body),
		}
	}
	return nil
}

// Client is a very light weight GitHub API Client.
type Client struct {
	baseURL    *url.URL
	httpClient *http.Client

	// configure for OpenID Connect
	oidcClient *oidc.Client
}

func NewClient(httpClient *http.Client) (*Client, error) {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	oidcClient, err := oidc.NewClient(&oidc.ClientConfig{
		Doer:      httpClient,
		Issuer:    oidcIssuer,
		UserAgent: githubUserAgent,
	})
	if err != nil {
		return nil, err
	}
	u, err := url.Parse(apiBaseURL)
	if err != nil {
		return nil, err
	}
	return &Client{
		baseURL:    u,
		httpClient: httpClient,
		oidcClient: oidcClient,
	}, nil
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
	if u != c.baseURL.String() {
		if c.baseURL.String() == defaultAPIBaseURL {
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
		return "", fmt.Errorf("unknown scheme: %q", u.Scheme)
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

func validateUserName(s string) error {
	for _, r := range s {
		// normal user name
		var ok bool
		ok = ok || 'a' <= r && r <= 'z'
		ok = ok || 'A' <= r && r <= 'Z'
		ok = ok || '0' <= r && r <= '9'
		ok = ok || r == '-'

		// Enterprise Managed Users contains '_'.
		// https://docs.github.com/en/enterprise-cloud@latest/admin/identity-and-access-management/using-enterprise-managed-users-for-iam/about-enterprise-managed-users
		ok = ok || r == '_'

		// GitHub Apps contains '[' and ']'.
		ok = ok || r == '[' || r == ']'

		if !ok {
			return fmt.Errorf("github: username contains invalid character: %q", r)
		}
	}
	return nil
}

func validateRepoName(s string) error {
	for _, r := range s {
		var ok bool
		ok = ok || 'a' <= r && r <= 'z'
		ok = ok || 'A' <= r && r <= 'Z'
		ok = ok || '0' <= r && r <= '9'
		ok = ok || r == '-' || r == '_' || r == '.'
		if !ok {
			return fmt.Errorf("github: repo name contains invalid character: %q", r)
		}
	}
	return nil
}

func validateRef(s string) error {
	for _, r := range s {
		if (r < 'a' || 'f' < r) && (r < 'A' || 'F' < r) && (r < '0' || '9' < r) {
			return fmt.Errorf("github: ref contains invalid character: %q", r)
		}
	}
	return nil
}
