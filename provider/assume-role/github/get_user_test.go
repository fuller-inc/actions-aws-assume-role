package github

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
)

func TestGetUser(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("unexpected method: want GET, got %s", r.Method)
		}
		path := "/users/shogo82148"
		if r.URL.Path != path {
			t.Errorf("unexpected path: want %q, got %q", path, r.URL.Path)
		}

		data, err := os.ReadFile("testdata/get-user.json")
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

	resp, err := c.GetUser(context.Background(), "dummy-auth-token", "shogo82148")
	if err != nil {
		t.Fatal(err)
	}
	if resp.NodeID != "MDQ6VXNlcjExNTczNDQ=" {
		t.Errorf("unexpected creator type: want %q, got %q", "MDQ6VXNlcjExNTczNDQ=", resp.NodeID)
	}
}
