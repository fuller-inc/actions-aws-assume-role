package assumerole

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/aws-sdk-go-v2/service/sts/types"
	"github.com/aws/smithy-go"
	"github.com/fuller-inc/actions-aws-assume-role/provider/assume-role/github"
	"github.com/shogo82148/aws-xray-yasdk-go/xrayaws-v2"
	"github.com/shogo82148/aws-xray-yasdk-go/xrayhttp"
)

const (
	commitStatusContext = "aws-assume-role"
	creatorLogin        = "github-actions[bot]"
	creatorID           = 41898282
	creatorType         = "Bot"
)

type githubClient interface {
	CreateStatus(ctx context.Context, token, owner, repo, ref string, status *github.CreateStatusRequest) (*github.CreateStatusResponse, error)
	GetRepo(ctx context.Context, nextIDFormat bool, token, owner, repo string) (*github.GetRepoResponse, error)
	GetUser(ctx context.Context, nextIDFormat bool, token, user string) (*github.GetUserResponse, error)
	ValidateAPIURL(url string) error
	ParseIDToken(ctx context.Context, idToken string) (*github.ActionsIDToken, error)
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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cfg, err := config.LoadDefaultConfig(ctx, xrayaws.WithXRay())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	client := xrayhttp.Client(nil)
	githubClient, err := github.NewClient(client)
	if err != nil {
		log.Fatalf("unable to initialize: %v", err)
	}

	return &Handler{
		github: githubClient,
		sts:    sts.NewFromConfig(cfg),
	}
}

type requestBody struct {
	GitHubToken         string `json:"github_token"`
	IDToken             string `json:"id_token"`
	RoleToAssume        string `json:"role_to_assume"`
	RoleSessionName     string `json:"role_session_name"`
	DurationSeconds     int32  `json:"duration_seconds"`
	Repository          string `json:"repository"`
	UseNodeID           bool   `json:"use_node_id"`
	ObfuscateRepository string `json:"obfuscate_repository"`
	APIURL              string `json:"api_url"`
	SHA                 string `json:"sha"`
	RoleSessionTagging  bool   `json:"role_session_tagging"`
	RunID               string `json:"run_id"`
	Workflow            string `json:"workflow"`
	Actor               string `json:"actor"`
	Branch              string `json:"branch"`
}

type responseBody struct {
	AccessKeyId     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
	SessionToken    string `json:"session_token"`
	Message         string `json:"message,omitempty"`
	Warning         string `json:"warning,omitempty"`
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
	if err := h.github.ValidateAPIURL(req.APIURL); err != nil {
		return nil, err
	}

	var idToken *github.ActionsIDToken
	if req.IDToken != "" {
		var err error
		idToken, err = h.github.ParseIDToken(ctx, req.IDToken)
		if err != nil {
			return nil, &validationError{
				message: fmt.Sprintf("invalid oidc token: %v", err),
			}
		}
	}
	if err := h.validateGitHubToken(ctx, req); err != nil {
		return nil, err
	}

	// Use Next ID format
	resp0, err0 := h.assumeRole(ctx, true, idToken, req)
	if err0 == nil {
		return resp0, nil
	}
	if !req.UseNodeID {
		return nil, err0
	}

	// Use legacy or next ID format
	resp1, err1 := h.assumeRole(ctx, false, idToken, req)
	if err1 != nil {
		return nil, err0
	}
	resp1.Warning += "It looks that you use legacy node IDs. You need to migrate them. " +
		"See https://github.com/fuller-inc/actions-aws-assume-role#migrate-your-node-id-to-the-next-format for more detail.\n" +
		err0.Error()
	return resp1, nil
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
	// early check of the token prefix
	// ref. https://github.blog/changelog/2021-03-31-authentication-token-format-updates-are-generally-available/
	if len(req.GitHubToken) < 4 {
		return &validationError{
			message: "GITHUB_TOKEN has invalid format",
		}
	}
	switch req.GitHubToken[:4] {
	case "ghp_":
		// Personal Access Tokens
		return &validationError{
			message: "GITHUB_TOKEN looks like Personal Access Token. `github-token` must be `${{ github.token }}` or `${{ secrets.GITHUB_TOKEN }}`.",
		}
	case "gho_":
		// OAuth Access tokens
		return &validationError{
			message: "GITHUB_TOKEN looks like OAuth Access token. `github-token` must be `${{ github.token }}` or `${{ secrets.GITHUB_TOKEN }}`.",
		}
	case "ghu_":
		// GitHub App user-to-server tokens
		return &validationError{
			message: "GITHUB_TOKEN looks like GitHub App user-to-server token. `github-token` must be `${{ github.token }}` or `${{ secrets.GITHUB_TOKEN }}`.",
		}
	case "ghs_":
		// GitHub App server-to-server tokens
		// It's OK
	case "ghr_":
		// GitHub App refresh tokens
		return &validationError{
			message: "GITHUB_TOKEN looks like GitHub App refresh token. `github-token` must be `${{ github.token }}` or `${{ secrets.GITHUB_TOKEN }}`.",
		}
	default:
		// Old Format Personal Access Tokens
		return &validationError{
			message: "GITHUB_TOKEN looks like Personal Access Token. `github-token` must be `${{ github.token }}` or `${{ secrets.GITHUB_TOKEN }}`.",
		}
	}
	resp, err := h.updateCommitStatus(ctx, req, &github.CreateStatusRequest{
		State:       github.CommitStateSuccess,
		Description: "valid github token",
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
		return &validationError{
			message: "`github-token` isn't be generated by @github-actions[bot]. `github-token` must be `${{ github.token }}` or `${{ secrets.GITHUB_TOKEN }}`.",
		}
	}
	return nil
}

func splitOwnerRepo(fullname string) (owner, repo string, err error) {
	idx := strings.IndexByte(fullname, '/')
	if idx < 0 {
		err = &validationError{
			message: fmt.Sprintf("invalid repository name: %s", fullname),
		}
		return
	}
	owner = fullname[:idx]
	repo = fullname[idx+1:]
	return
}

func (h *Handler) updateCommitStatus(ctx context.Context, req *requestBody, status *github.CreateStatusRequest) (*github.CreateStatusResponse, error) {
	owner, repo, err := splitOwnerRepo(req.Repository)
	if err != nil {
		return nil, err
	}
	return h.github.CreateStatus(ctx, req.GitHubToken, owner, repo, req.SHA, status)
}

func (h *Handler) getRepo(ctx context.Context, nextIDFormat bool, idToken *github.ActionsIDToken, req *requestBody) (*github.GetRepoResponse, error) {
	var owner, repo string
	var err error
	if idToken != nil {
		// Get the information from the id token if it's avaliable.
		// They are more trustworthy because they are digitally signed.
		owner, repo, err = splitOwnerRepo(idToken.Repository)
	} else {
		owner, repo, err = splitOwnerRepo(req.Repository)
	}
	if err != nil {
		return nil, err
	}
	return h.github.GetRepo(ctx, false, req.GitHubToken, owner, repo)
}

func (h *Handler) getUser(ctx context.Context, nextIDFormat bool, idToken *github.ActionsIDToken, req *requestBody) (*github.GetUserResponse, error) {
	if idToken != nil {
		// Get the information from the id token if it's avaliable.
		// They are more trustworthy because they are digitally signed.
		return h.github.GetUser(ctx, false, req.GitHubToken, idToken.Actor)
	} else {
		return h.github.GetUser(ctx, false, req.GitHubToken, req.Actor)
	}
}

func (h *Handler) assumeRole(ctx context.Context, nextIDFormat bool, idToken *github.ActionsIDToken, req *requestBody) (*responseBody, error) {
	repo, err := h.getRepo(ctx, nextIDFormat, idToken, req)
	if err != nil {
		return nil, err
	}
	user, err := h.getUser(ctx, nextIDFormat, idToken, req)
	if err != nil {
		return nil, err
	}

	var repository, actor string
	if req.UseNodeID {
		repository = repo.NodeID
		actor = user.NodeID
	} else if idToken != nil {
		repository = idToken.Repository
		actor = idToken.Actor
	} else {
		repository = req.Repository
		actor = req.Actor
	}

	var tags []types.Tag
	if req.RoleSessionTagging {
		if idToken != nil {
			// Get the information from the id token if it's avaliable.
			// They are more trustworthy because they are digitally signed.
			subject := idToken.Subject
			if req.UseNodeID {
				fragment := strings.SplitN(subject, ":", 3)
				if len(fragment) != 3 {
					return nil, &validationError{
						message: fmt.Sprintf("invalid subject format: %q", subject),
					}
				}
				fragment[1] = repository
				subject = strings.Join(fragment, ":")
			}
			tags = []types.Tag{
				{
					Key:   aws.String("Audience"),
					Value: aws.String(sanitizeTagValue(idToken.Audience)),
				},
				{
					Key:   aws.String("Subject"),
					Value: aws.String(sanitizeTagValue(subject)),
				},
				{
					Key:   aws.String("GitHub"),
					Value: aws.String("Actions"),
				},
				{
					Key:   aws.String("Repository"),
					Value: aws.String(sanitizeTagValue(repository)),
				},
				{
					Key:   aws.String("Workflow"),
					Value: aws.String(sanitizeTagValue(idToken.Workflow)),
				},
				{
					Key:   aws.String("RunId"),
					Value: aws.String(sanitizeTagValue(idToken.RunID)),
				},
				{
					Key:   aws.String("Actor"),
					Value: aws.String(sanitizeTagValue(actor)),
				},
				{
					Key:   aws.String("Commit"),
					Value: aws.String(sanitizeTagValue(idToken.SHA)),
				},
			}
			if idToken.Ref != "" {
				tags = append(tags, types.Tag{
					Key:   aws.String("Branch"),
					Value: aws.String(sanitizeTagValue(idToken.Ref)),
				})
			}
			if idToken.Environment != "" {
				tags = append(tags, types.Tag{
					Key:   aws.String("Environment"),
					Value: aws.String(sanitizeTagValue(idToken.Environment)),
				})
			}
		} else {
			tags = []types.Tag{
				{
					Key:   aws.String("GitHub"),
					Value: aws.String("Actions"),
				},
				{
					Key:   aws.String("Repository"),
					Value: aws.String(sanitizeTagValue(repository)),
				},
				{
					Key:   aws.String("Workflow"),
					Value: aws.String(sanitizeTagValue(req.Workflow)),
				},
				{
					Key:   aws.String("RunId"),
					Value: aws.String(sanitizeTagValue(req.RunID)),
				},
				{
					Key:   aws.String("Actor"),
					Value: aws.String(sanitizeTagValue(actor)),
				},
				{
					Key:   aws.String("Commit"),
					Value: aws.String(sanitizeTagValue(req.SHA)),
				},
			}
			if req.Branch != "" {
				tags = append(tags, types.Tag{
					Key:   aws.String("Branch"),
					Value: aws.String(sanitizeTagValue(req.Branch)),
				})
			}
		}
	}

	// validate IAM Role
	// https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_create_for-user_externalid.html#external-id-use
	// > In addition when a customer gives you a role ARN, test whether you can assume the role both with and without the correct external ID.
	validationInput := &sts.AssumeRoleInput{
		RoleArn:         aws.String(req.RoleToAssume),
		RoleSessionName: aws.String(req.RoleSessionName),
		Tags:            tags,

		// set shortest duration seconds. because we don't use this credential actually.
		DurationSeconds: aws.Int32(900),

		// request without the correct external ID
	}
	_, err = h.sts.AssumeRole(ctx, validationInput)
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
	if req.UseNodeID {
		input.ExternalId = aws.String(repo.NodeID)
	} else {
		switch req.ObfuscateRepository {
		case "sha256":
			hash := sha256.Sum256([]byte(req.Repository))
			input.ExternalId = aws.String("sha256:" + hex.EncodeToString(hash[:]))
		case "":
			input.ExternalId = aws.String(req.Repository)
		default:
			return nil, &validationError{
				message: fmt.Sprintf("invalid obfuscate repository type: %s", req.ObfuscateRepository),
			}
		}
	}
	input.DurationSeconds = aws.Int32(req.DurationSeconds)
	resp, err := h.sts.AssumeRole(ctx, &input)
	if err != nil {
		var ae smithy.APIError
		if errors.As(err, &ae) && ae.ErrorCode() == "AccessDenied" {
			msg := fmt.Sprintf(
				"AWS denied your access: %s, please check your trust policy accepts %q as sts:ExternalId.",
				ae.ErrorMessage(),
				aws.ToString(input.ExternalId),
			)
			return nil, &validationError{
				message: msg,
			}
		}
		return nil, err
	}
	return &responseBody{
		AccessKeyId:     aws.ToString(resp.Credentials.AccessKeyId),
		SecretAccessKey: aws.ToString(resp.Credentials.SecretAccessKey),
		SessionToken:    aws.ToString(resp.Credentials.SessionToken),
	}, nil
}

// https://docs.aws.amazon.com/STS/latest/APIReference/API_Tag.html
const tagSanitizationCharactor = "_"
const tagMaxValueLength = 256

func sanitizeTagValue(s string) string {
	var builder strings.Builder
	builder.Grow(len(s))
	for i, r := range s {
		if i >= tagMaxValueLength {
			break
		}
		if validTagRune(r) {
			builder.WriteRune(r)
		} else {
			builder.WriteString(tagSanitizationCharactor)
		}
	}
	return builder.String()
}

// valid runes are match [\p{L}\p{Z}\p{N}_.:/=+\-@]
func validTagRune(r rune) bool {
	switch {
	case unicode.IsLetter(r):
		return true // \p{L}
	case unicode.IsSpace(r):
		return true // \p{Z}
	case unicode.IsNumber(r):
		return true // \p{N}
	}
	switch r {
	case '_', '.', ':', '/', '=', '+', '-', '@':
		return true
	}
	return false
}
