package github

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

type GetUserResponse struct {
	NodeID string `json:"node_id"`

	// omit other fields, we don't use them.
}

func (c *Client) GetUser(ctx context.Context, token, user string) (*GetUserResponse, error) {
	// validate the parameters
	if err := validateUserName(user); err != nil {
		return nil, err
	}

	// build the request
	u := c.baseURL.JoinPath("users", url.PathEscape(user))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", githubUserAgent)
	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("X-GitHub-Api-Version", githubAPIVersion)
	// It forces the value for all id fields in my query to return the next ID format.
	// https://github.blog/2021-11-16-graphql-global-id-migration-update/#how-do-i-migrate-my-service
	req.Header.Set("X-Github-Next-Global-ID", "1")

	// send the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// parse the response
	if err := handleUnexpectedStatusCode(resp); err != nil {
		return nil, err
	}

	var ret *GetUserResponse
	if err := json.NewDecoder(resp.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
}
