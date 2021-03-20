package assumerole

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/smithy-go"
	"github.com/shogo82148/actions-aws-assume-role/provider/assume-role/github"
)

const (
	commitStatusContext = "aws-assume-role"
	creatorLogin        = "github-actions[bot]"
	creatorID           = 41898282
	creatorType         = "Bot"
)

type githubClient interface {
	CreateStatus(ctx context.Context, token, owner, repo, ref string, status *github.CreateStatusRequest) (*github.CreateStatusResponse, error)
}

type stsClient interface {
	AssumeRole(ctx context.Context, params *sts.AssumeRoleInput, optFns ...func(*sts.Options)) (*sts.AssumeRoleOutput, error)
}

type validationError struct {
	message string
}

func (err *validationError) Error() string {
	return err.message
}

type Handler struct {
	github githubClient
	sts    stsClient
}

func NewHandler() *Handler {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	return &Handler{
		github: github.NewClient(nil),
		sts:    sts.NewFromConfig(cfg),
	}
}

type requestBody struct {
	GitHubToken     string `json:"github_token"`
	RoleToAssume    string `json:"role_to_assume"`
	RoleSessionName string `json:"role_session_name"`
	Repository      string `json:"repository"`
	SHA             string `json:"sha"`
}

type responseBody struct {
	AccessKeyId     string    `json:"access_key_id"`
	SecretAccessKey string    `json:"secret_access_key"`
	SessionToken    string    `json:"session_token"`
	Expiration      time.Time `json:"expiration"`
}

type errorResponseBody struct {
	Message string `json:"message"`
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	data, err := io.ReadAll(r.Body)
	if err != nil {
		h.handleError(w, r, fmt.Errorf("failed to read the request body: %w", err))
		return
	}
	var payload *requestBody
	if err := json.Unmarshal(data, &payload); err != nil {
		h.handleError(w, r, &validationError{
			message: fmt.Sprintf("failed to unmarshal the request body: %v", err),
		})
		return
	}

	resp, err := h.handle(ctx, payload)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("failed to write the response: %v", err)
	}
}

func (h *Handler) handle(ctx context.Context, req *requestBody) (*responseBody, error) {
	if err := h.validateGitHubToken(ctx, req); err != nil {
		return nil, err
	}

	resp, err := h.assumeRole(ctx, req)
	if err != nil {
		var validation *validationError
		if errors.As(err, &validation) {
			h.updateCommitStatus(ctx, req, &github.CreateStatusRequest{
				State:       github.CommitStateFailure,
				Description: "validation error",
				Context:     commitStatusContext,
			})
		} else {
			h.updateCommitStatus(ctx, req, &github.CreateStatusRequest{
				State:       github.CommitStateError,
				Description: "internal error",
				Context:     commitStatusContext,
			})
		}
		return nil, err
	}
	h.updateCommitStatus(ctx, req, &github.CreateStatusRequest{
		State:       github.CommitStateSuccess,
		Description: "AWS credentials are configured",
		Context:     commitStatusContext,
	})
	return resp, nil
}

func (h *Handler) handleError(w http.ResponseWriter, r *http.Request, err error) {
	log.Println(err)
	status := http.StatusInternalServerError
	var body *errorResponseBody

	var validation *validationError
	if errors.As(err, &validation) {
		status = http.StatusBadRequest
		body = &errorResponseBody{
			Message: validation.message,
		}
	}

	if body == nil {
		body = &errorResponseBody{
			Message: "Internal Server Error",
		}
	}
	data, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.WriteHeader(status)
	w.Write(data)
}

func (h *Handler) validateGitHubToken(ctx context.Context, req *requestBody) error {
	resp, err := h.updateCommitStatus(ctx, req, &github.CreateStatusRequest{
		State:       github.CommitStatePending,
		Description: "configuring AWS credentials",
		Context:     commitStatusContext,
	})
	if err != nil {
		var githubErr *github.UnexpectedStatusCodeError
		if errors.As(err, &githubErr) {
			if 400 <= githubErr.StatusCode && githubErr.StatusCode < 500 {
				return &validationError{
					message: "Your GITHUB_TOKEN doesn't have enough permission. Write-Permission is required.",
				}
			}
		}
		return err
	}
	if resp.Creator.Login != creatorLogin || resp.Creator.ID != creatorID || resp.Creator.Type != creatorType {
		h.updateCommitStatus(ctx, req, &github.CreateStatusRequest{
			State:       github.CommitStateFailure,
			Description: "validation error",
			Context:     commitStatusContext,
		})
		return &validationError{
			message: "The owner of GITHUB_TOKEN is invalid. GITHUB_TOKEN must be generated by @github-actions[bot].",
		}
	}
	return nil
}

func (h *Handler) updateCommitStatus(ctx context.Context, req *requestBody, status *github.CreateStatusRequest) (*github.CreateStatusResponse, error) {
	idx := strings.IndexByte(req.Repository, '/')
	if idx < 0 {
		return nil, &validationError{
			message: fmt.Sprintf("invalid repository name: %s", req.Repository),
		}
	}
	owner := req.Repository[:idx]
	repo := req.Repository[idx+1:]
	return h.github.CreateStatus(ctx, req.GitHubToken, owner, repo, req.SHA, status)
}

func (h *Handler) assumeRole(ctx context.Context, req *requestBody) (*responseBody, error) {
	// validate IAM Role
	// https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_create_for-user_externalid.html#external-id-use
	// > In addition when a customer gives you a role ARN, test whether you can assume the role both with and without the correct external ID.
	validationInput := &sts.AssumeRoleInput{
		RoleArn:         aws.String(req.RoleToAssume),
		RoleSessionName: aws.String(req.RoleSessionName),

		// set shortest duration seconds. because we don't use this credential actually.
		DurationSeconds: aws.Int32(900),

		// request without the correct external ID
	}
	_, err := h.sts.AssumeRole(ctx, validationInput)
	if err == nil {
		return nil, &validationError{
			message: "The AssumeRolePolicy of your IAM Role is too open. Please configure ExternalId conditions.",
		}
	}
	var ae smithy.APIError
	if !errors.As(err, &ae) || ae.ErrorCode() != "AccessDenied" {
		// We expected AccessDenied error, but got another error. (maybe network error etc.)
		// We can't continue this process.
		return nil, err
	}

	// assume role with the correct external ID
	input := *validationInput
	input.ExternalId = aws.String(req.Repository)
	input.DurationSeconds = nil
	resp, err := h.sts.AssumeRole(ctx, &input)
	if err != nil {
		var ae smithy.APIError
		if errors.As(err, &ae) && ae.ErrorCode() == "AccessDenied" {
			return nil, &validationError{
				message: ae.ErrorMessage(),
			}
		}
		return nil, err
	}
	return &responseBody{
		AccessKeyId:     aws.ToString(resp.Credentials.AccessKeyId),
		SecretAccessKey: aws.ToString(resp.Credentials.SecretAccessKey),
		SessionToken:    aws.ToString(resp.Credentials.SessionToken),
		Expiration:      aws.ToTime(resp.Credentials.Expiration),
	}, nil
}
