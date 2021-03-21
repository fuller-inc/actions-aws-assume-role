package assumerole

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/aws-sdk-go-v2/service/sts/types"
	"github.com/shogo82148/actions-aws-assume-role/provider/assume-role/github"
)

type githubClientMock struct {
	CreateStatusFunc func(ctx context.Context, token, owner, repo, ref string, status *github.CreateStatusRequest) (*github.CreateStatusResponse, error)
}

func (c *githubClientMock) CreateStatus(ctx context.Context, token, owner, repo, ref string, status *github.CreateStatusRequest) (*github.CreateStatusResponse, error) {
	return c.CreateStatusFunc(ctx, token, owner, repo, ref, status)
}

type stsClientMock struct {
	AssumeRoleFunc func(ctx context.Context, params *sts.AssumeRoleInput, optFns ...func(*sts.Options)) (*sts.AssumeRoleOutput, error)
}

func (c *stsClientMock) AssumeRole(ctx context.Context, params *sts.AssumeRoleInput, optFns ...func(*sts.Options)) (*sts.AssumeRoleOutput, error) {
	return c.AssumeRoleFunc(ctx, params, optFns...)
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
				if ref != "e3a45c6c16c1464826b36a598ff39e6cc98c4da4" {
					t.Errorf("unexpected ref: want %q, got %q", "e3a45c6c16c1464826b36a598ff39e6cc98c4da4", ref)
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

func TestAssumeRole_AssumeRolePolicyTooOpen(t *testing.T) {
	h := &Handler{
		sts: &stsClientMock{
			AssumeRoleFunc: func(ctx context.Context, params *sts.AssumeRoleInput, optFns ...func(*sts.Options)) (*sts.AssumeRoleOutput, error) {
				return &sts.AssumeRoleOutput{}, nil
			},
		},
	}
	_, err := h.assumeRole(context.Background(), &requestBody{
		RoleToAssume:    "arn:aws:iam::123456789012:role/assume-role-test",
		RoleSessionName: "GitHubActions",
		Repository:      "shogo82148/actions-aws-assume-role",
	})
	var validate *validationError
	if !errors.As(err, &validate) {
		t.Errorf("want validation error, got %T", err)
	}
}

func TestAssumeRole(t *testing.T) {
	h := &Handler{
		sts: &stsClientMock{
			AssumeRoleFunc: func(ctx context.Context, params *sts.AssumeRoleInput, optFns ...func(*sts.Options)) (*sts.AssumeRoleOutput, error) {
				if params.ExternalId == nil {
					return nil, errAccessDenied
				}
				return &sts.AssumeRoleOutput{
					Credentials: &types.Credentials{
						AccessKeyId:     aws.String("AKIAIOSFODNN7EXAMPLE"),
						SecretAccessKey: aws.String("wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"),
						SessionToken:    aws.String("session-token"),
					},
				}, nil
			},
		},
	}
	resp, err := h.assumeRole(context.Background(), &requestBody{
		RoleToAssume:    "arn:aws:iam::123456789012:role/assume-role-test",
		RoleSessionName: "GitHubActions",
		Repository:      "shogo82148/actions-aws-assume-role",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.AccessKeyId != "AKIAIOSFODNN7EXAMPLE" {
		t.Errorf("want %q, got %q", "AKIAIOSFODNN7EXAMPLE", resp.AccessKeyId)
	}
	if resp.SecretAccessKey != "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY" {
		t.Errorf("want %q, got %q", "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY", resp.SecretAccessKey)
	}
	if resp.SessionToken != "session-token" {
		t.Errorf("want %q, got %q", "session-token", resp.SessionToken)
	}
}

func TestSanitizeTagValue(t *testing.T) {
	cases := []struct {
		input  string
		output string
	}{
		{
			input:  "abcdefghijklmnopqrstuvwxyz",
			output: "abcdefghijklmnopqrstuvwxyz",
		},
		{
			input:  "0123456789",
			output: "0123456789",
		},
		{
			input:  " !\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~",
			output: " __________+_-./:__=__@__________",
		},
		{
			input:  "😀",
			output: "_",
		},
		{
			input:  strings.Repeat("a", 500),
			output: strings.Repeat("a", 256),
		},
	}
	for _, tc := range cases {
		got := sanitizeTagValue(tc.input)
		if got != tc.output {
			t.Errorf("want %q, got %q", tc.output, got)
		}
	}
}
