package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type GetRepoResponse struct {
	NodeID string `json:"node_id"`

	// omit other fields, we don't use them.
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
	if resp.StatusCode != http.StatusOK {
		return nil, &UnexpectedStatusCodeError{StatusCode: resp.StatusCode}
	}

	var ret *GetRepoResponse
	if err := json.NewDecoder(resp.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
}
