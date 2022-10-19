package github

import (
	"errors"
	"fmt"
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
}

func (err *UnexpectedStatusCodeError) Error() string {
	return fmt.Sprintf("unexpected status code: %d", err.StatusCode)
}

// Client is a very light weight GitHub API Client.
type Client struct {
	baseURL    string
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
	return &Client{
		baseURL:    apiBaseURL,
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
