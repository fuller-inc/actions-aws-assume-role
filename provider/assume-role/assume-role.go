package assumerole

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type stsClient interface {
	AssumeRole(ctx context.Context, params *sts.AssumeRoleInput, optFns ...func(*sts.Options)) (*sts.AssumeRoleOutput, error)
}

type Handler struct {
	sts stsClient
}

func NewHandler() *Handler {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	return &Handler{
		sts: sts.NewFromConfig(cfg),
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
	AccessKeyId     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
	SessionToken    string `json:"session_token"`
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	data, err := io.ReadAll(r.Body)
	if err != nil {
		// TODO: better handling error
		log.Printf("failed to read the request body: %v", err)
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}
	var payload *requestBody
	if err := json.Unmarshal(data, &payload); err != nil {
		// TODO: better handling error
		log.Printf("failed to unmarshal the request body: %v", err)
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}

	resp, err := h.handle(ctx, payload)
	if err != nil {
		// TODO: better handling error
		log.Printf("failed to handle the request: %v", err)
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("failed to write the response: %v", err)
	}
}

func (h *Handler) handle(ctx context.Context, req *requestBody) (*responseBody, error) {
	resp, err := h.sts.AssumeRole(ctx, &sts.AssumeRoleInput{
		RoleArn:         aws.String("arn:aws:iam::445285296882:role/assume-role-test"),
		RoleSessionName: aws.String("GitHubActions"),
	})
	if err != nil {
		return nil, err
	}
	return &responseBody{
		AccessKeyId: aws.ToString(resp.Credentials.AccessKeyId),
	}, nil
}
