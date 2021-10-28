package github

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
)

func TestGetRepo(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("unexpected method: want GET, got %s", r.Method)
		}
		path := "/repos/fuller-inc/actions-aws-assume-role"
		if r.URL.Path != path {
			t.Errorf("unexpected path: want %q, got %q", path, r.URL.Path)
		}

		data, err := os.ReadFile("testdata/get-repo.json")
		if err != nil {
			panic(err)
		}
		rw.Header().Set("Content-Type", "application/json")
		rw.Header().Set("Content-Length", strconv.Itoa(len(data)))
		rw.WriteHeader(http.StatusOK)
		rw.Write(data)
	}))
	defer ts.Close()
	c, err := NewClient(nil)
	if err != nil {
		t.Fatal(err)
	}
	c.baseURL = ts.URL

	resp, err := c.GetRepo(context.Background(), "dummy-auth-token", "fuller-inc", "actions-aws-assume-role")
	if err != nil {
		t.Fatal(err)
	}
	if resp.NodeID != "MDEwOlJlcG9zaXRvcnkzNDg4NDkwMzk=" {
		t.Errorf("unexpected creator type: want %q, got %q", "MDEwOlJlcG9zaXRvcnkzNDg4NDkwMzk=", resp.NodeID)
	}
}
