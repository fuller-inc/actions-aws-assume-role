package assumerole

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/shogo82148/actions-aws-assume-role/provider/assume-role/github"
)

type githubClientMock struct {
	CreateStatusFunc func(ctx context.Context, token, owner, repo, ref string, status *github.CreateStatusRequest) (*github.CreateStatusResponse, error)
}

func (c *githubClientMock) CreateStatus(ctx context.Context, token, owner, repo, ref string, status *github.CreateStatusRequest) (*github.CreateStatusResponse, error) {
	return c.CreateStatusFunc(ctx, token, owner, repo, ref, status)
}

func TestValidateGitHubToken(t *testing.T) {
	h := &Handler{
		github: &githubClientMock{
			CreateStatusFunc: func(ctx context.Context, token, owner, repo, ref string, status *github.CreateStatusRequest) (*github.CreateStatusResponse, error) {
				if token != "v1.dummyGitHubToken" {
					t.Errorf("unexpected GitHub Token: want %q, got %q", "v1.dummyGitHubToken", token)
				}
				if owner != "shogo82148" {
					t.Errorf("unexpected owner: want %q, got %q", "shogo82148", owner)
				}
				if repo != "actions-aws-assume-role" {
					t.Errorf("unexpected repo: want %q, got %q", "actions-aws-assume-role", repo)
				}
				if status.State != github.CommitStatePending {
					t.Errorf("unexpected commit status state: want %s, got %s", github.CommitStatePending, status.State)
				}
				if status.Context != commitStatusContext {
					t.Errorf("unexpected commit status context: want %q, got %q", commitStatusContext, status.Context)
				}
				return &github.CreateStatusResponse{
					Creator: &github.CreateStatusResponseCreator{
						Login: creatorLogin,
						ID:    creatorID,
						Type:  creatorType,
					},
				}, nil
			},
		},
	}
	err := h.validateGitHubToken(context.Background(), &requestBody{
		GitHubToken: "v1.dummyGitHubToken",
		Repository:  "shogo82148/actions-aws-assume-role",
		SHA:         "e3a45c6c16c1464826b36a598ff39e6cc98c4da4",
	})
	if err != nil {
		t.Error(err)
	}
}

func TestValidateGitHubToken_PermissionError(t *testing.T) {
	h := &Handler{
		github: &githubClientMock{
			CreateStatusFunc: func(ctx context.Context, token, owner, repo, ref string, status *github.CreateStatusRequest) (*github.CreateStatusResponse, error) {
				return nil, &github.UnexpectedStatusCodeError{
					StatusCode: http.StatusBadRequest,
				}
			},
		},
	}
	err := h.validateGitHubToken(context.Background(), &requestBody{
		GitHubToken: "v1.dummyGitHubToken",
		Repository:  "shogo82148/actions-aws-assume-role",
		SHA:         "e3a45c6c16c1464826b36a598ff39e6cc98c4da4",
	})
	if err == nil {
		t.Error("want error, but not")
	}

	var validate *validationError
	if !errors.As(err, &validate) {
		t.Errorf("want validation error, got %T", err)
	}
}

func TestValidateGitHubToken_InvalidCreator(t *testing.T) {
	h := &Handler{
		github: &githubClientMock{
			CreateStatusFunc: func(ctx context.Context, token, owner, repo, ref string, status *github.CreateStatusRequest) (*github.CreateStatusResponse, error) {
				return &github.CreateStatusResponse{
					Creator: &github.CreateStatusResponseCreator{
						Login: "shogo82148",
						ID:    1157344,
						Type:  "User",
					},
				}, nil
			},
		},
	}
	err := h.validateGitHubToken(context.Background(), &requestBody{
		GitHubToken: "v1.dummyGitHubToken",
		Repository:  "shogo82148/actions-aws-assume-role",
		SHA:         "e3a45c6c16c1464826b36a598ff39e6cc98c4da4",
	})
	if err == nil {
		t.Error("want error, but not")
	}

	var validate *validationError
	if !errors.As(err, &validate) {
		t.Errorf("want validation error, got %T", err)
	}
}
