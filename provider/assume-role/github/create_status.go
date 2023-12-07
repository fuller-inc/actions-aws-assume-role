package github

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

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

// CreateStatus creates a commit status.
// https://docs.github.com/en/rest/reference/repos#create-a-commit-status
func (c *Client) CreateStatus(ctx context.Context, token, owner, repo, ref string, status *CreateStatusRequest) (*CreateStatusResponse, error) {
	// validate the parameters
	if err := validateUserName(owner); err != nil {
		return nil, err
	}
	if err := validateRepoName(repo); err != nil {
		return nil, err
	}
	if err := validateRef(ref); err != nil {
		return nil, err
	}

	// build the request
	u := c.baseURL.JoinPath("repos", url.PathEscape(owner), url.PathEscape(repo), "statuses", url.PathEscape(ref))
	body, err := json.Marshal(status)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewReader(body))
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
	if err := handleUnexpectedStatusCode(resp); err != nil {
		return nil, err
	}

	var ret *CreateStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
}
