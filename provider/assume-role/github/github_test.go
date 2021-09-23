package github

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
)

func TestCreateStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: want POST, got %s", r.Method)
		}
		path := "/repos/shogo82148/actions-aws-assume-role/statuses/496f02e29cc5760443becd7007049c1a2a502b6f"
		if r.URL.Path != path {
			t.Errorf("unexpected path: want %q, got %q", path, r.URL.Path)
		}

		data, err := os.ReadFile("testdata/status-created.json")
		if err != nil {
			panic(err)
		}
		rw.Header().Set("Content-Type", "application/json")
		rw.Header().Set("Content-Length", strconv.Itoa(len(data)))
		rw.WriteHeader(http.StatusCreated)
		rw.Write(data)
	}))
	defer ts.Close()
	c, err := NewClient(nil)
	if err != nil {
		t.Fatal(err)
	}
	c.baseURL = ts.URL

	resp, err := c.CreateStatus(context.Background(), "dummy-auth-token", "shogo82148", "actions-aws-assume-role", "496f02e29cc5760443becd7007049c1a2a502b6f", &CreateStatusRequest{
		State:   CommitStateSuccess,
		Context: "actions-aws-assume-role",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Creator.ID != 1157344 {
		t.Errorf("unexpected creator id: want %d, got %d", 1157344, resp.Creator.ID)
	}
	if resp.Creator.Login != "shogo82148" {
		t.Errorf("unexpected creator login: want %q, got %q", "shogo82148", resp.Creator.Login)
	}
	if resp.Creator.Type != "User" {
		t.Errorf("unexpected creator type: want %q, got %q", "User", resp.Creator.Type)
	}
}

func TestValidateGitHubToken(t *testing.T) {
	c, err := NewClient(nil)
	if err != nil {
		t.Fatal(err)
	}
	c.baseURL = defaultAPIBaseURL
	if err := c.ValidateAPIURL(""); err != nil {
		t.Error(err)
	}
	if err := c.ValidateAPIURL(defaultAPIBaseURL); err != nil {
		t.Error(err)
	}
	if err := c.ValidateAPIURL(defaultAPIBaseURL + "/"); err != nil {
		t.Error(err)
	}
	if err := c.ValidateAPIURL("https://example.com/api"); err == nil {
		t.Error("want error, but not")
	}
}

func TestCanonicalURL(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{
			input: "https://api.github.com",
			want:  "https://api.github.com",
		},
		{
			input: "https://API.GITHUB.COM",
			want:  "https://api.github.com",
		},
		{
			input: "https://api.github.com/",
			want:  "https://api.github.com",
		},
		{
			input: "http://example.com/API",
			want:  "http://example.com/API",
		},
		{
			input: "http://example.com/api/",
			want:  "http://example.com/api",
		},
		{
			input: "example.com/api",
			want:  "http://example.com/api",
		},
		{
			input: "http://example.com:80/api",
			want:  "http://example.com/api",
		},
		{
			input: "https://example.com:443/api",
			want:  "https://example.com/api",
		},
		{
			input: "http://example.com:443/api",
			want:  "http://example.com:443/api",
		},
		{
			input: "https://example.com:80/api",
			want:  "https://example.com:80/api",
		},
		{
			input: "https://[::1]:8080/api",
			want:  "https://[::1]:8080/api",
		},
	}
	for i, c := range cases {
		got, err := canonicalURL(c.input)
		if err != nil {
			t.Errorf("%d: canonicalURL(%q) returns error: %v", i, c.input, err)
			continue
		}
		if got != c.want {
			t.Errorf("%d: canonicalURL(%q) should be %q, but got %q", i, c.input, c.want, got)
		}
	}
}
