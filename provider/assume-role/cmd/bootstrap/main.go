package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/shogo82148/ridgenative"
)

var svc *sts.Client

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

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx := r.Context()
	resp, err := svc.AssumeRole(ctx, &sts.AssumeRoleInput{
		RoleArn:         aws.String("arn:aws:iam::445285296882:role/assume-role-test"),
		RoleSessionName: aws.String("GitHubActions"),
	})
	if err != nil {
		log.Printf("failed to assume role: %v", err)
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}

	payload := &responseBody{
		AccessKeyId: aws.ToString(resp.Credentials.AccessKeyId),
	}
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Println(err)
	}
}

func main() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	svc = sts.NewFromConfig(cfg)

	http.HandleFunc("/", handler)
	ridgenative.ListenAndServe(":8080", nil)
}
