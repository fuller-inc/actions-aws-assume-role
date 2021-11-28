package assumerole

import (
	"context"
	"errors"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/aws-sdk-go-v2/service/sts/types"
	"github.com/aws/smithy-go"
	"github.com/fuller-inc/actions-aws-assume-role/provider/assume-role/github"
)

var errAccessDenied = &awsAccessDeniedError{}

type awsAccessDeniedError struct{}

func (err *awsAccessDeniedError) Error() string                 { return "AccessDenied" }
func (err *awsAccessDeniedError) ErrorCode() string             { return "AccessDenied" }
func (err *awsAccessDeniedError) ErrorMessage() string          { return "AccessDenied" }
func (err *awsAccessDeniedError) ErrorFault() smithy.ErrorFault { return smithy.FaultUnknown }

type githubClientDummy struct{}

func (c *githubClientDummy) CreateStatus(ctx context.Context, token, owner, repo, ref string, status *github.CreateStatusRequest) (*github.CreateStatusResponse, error) {
	if token != "ghs_dummyGitHubToken" || owner != "shogo82148" || repo != "actions-aws-assume-role" || ref != "e3a45c6c16c1464826b36a598ff39e6cc98c4da4" {
		return nil, &github.UnexpectedStatusCodeError{StatusCode: http.StatusBadRequest}
	}
	return &github.CreateStatusResponse{
		Creator: &github.CreateStatusResponseCreator{
			Login: "github-actions[bot]",
			ID:    41898282,
			Type:  "Bot",
		},
	}, nil
}

func (c *githubClientDummy) GetRepo(ctx context.Context, nextIDFormat bool, token, owner, repo string) (*github.GetRepoResponse, error) {
	return &github.GetRepoResponse{
		NodeID: "MDEwOlJlcG9zaXRvcnkzNDg4NDkwMzk=",
	}, nil
}

func (c *githubClientDummy) GetUser(ctx context.Context, nextIDFormat bool, token, user string) (*github.GetUserResponse, error) {
	return &github.GetUserResponse{
		NodeID: "MDQ6VXNlcjExNTczNDQ=",
	}, nil
}

func (c *githubClientDummy) ParseIDToken(ctx context.Context, idToken string) (*github.ActionsIDToken, error) {
	return nil, errors.New("invalid jwt")
}

func (c *githubClientDummy) ValidateAPIURL(url string) error {
	return nil
}

type stsClientDummy struct{}

func (c *stsClientDummy) AssumeRole(ctx context.Context, params *sts.AssumeRoleInput, optFns ...func(*sts.Options)) (*sts.AssumeRoleOutput, error) {
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
}

func NewDummyHandler() *Handler {
	return &Handler{
		github: &githubClientDummy{},
		sts:    &stsClientDummy{},
	}
}
