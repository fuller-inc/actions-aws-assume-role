package assumerole

import (
	"context"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/aws-sdk-go-v2/service/sts/types"
	"github.com/aws/smithy-go"
	"github.com/shogo82148/actions-aws-assume-role/provider/assume-role/github"
)

var errAccessDenied = &awsAccessDeniedError{}

type awsAccessDeniedError struct{}

func (err *awsAccessDeniedError) Error() string                 { return "AccessDenied" }
func (err *awsAccessDeniedError) ErrorCode() string             { return "AccessDenied" }
func (err *awsAccessDeniedError) ErrorMessage() string          { return "AccessDenied" }
func (err *awsAccessDeniedError) ErrorFault() smithy.ErrorFault { return smithy.FaultUnknown }

type githubClientDummy struct{}

func (c *githubClientDummy) CreateStatus(ctx context.Context, token, owner, repo, ref string, status *github.CreateStatusRequest) (*github.CreateStatusResponse, error) {
	if token != "v1.dummyGitHubToken" || owner != "shogo82148" || repo != "actions-aws-assume-role" || ref != "e3a45c6c16c1464826b36a598ff39e6cc98c4da4" {
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
