package github

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
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
		idFormat := r.Header.Get("X-Github-Next-Global-ID")
		if idFormat != "0" {
			t.Errorf("unexpected X-Github-Next-Global-ID header: want %s, got %s", "0", idFormat)
		}

		data, err := os.ReadFile("testdata/get-user-current.json")
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
	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	c.baseURL = u

	resp, err := c.GetUser(context.Background(), false, "dummy-auth-token", "shogo82148")
	if err != nil {
		t.Fatal(err)
	}
	if resp.NodeID != "MDQ6VXNlcjExNTczNDQ=" {
		t.Errorf("unexpected creator type: want %q, got %q", "MDQ6VXNlcjExNTczNDQ=", resp.NodeID)
	}
}

func TestGetUser_WithNextNodeID(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("unexpected method: want GET, got %s", r.Method)
		}
		path := "/users/shogo82148"
		if r.URL.Path != path {
			t.Errorf("unexpected path: want %q, got %q", path, r.URL.Path)
		}
		idFormat := r.Header.Get("X-Github-Next-Global-ID")
		if idFormat != "1" {
			t.Errorf("unexpected X-Github-Next-Global-ID header: want %s, got %s", "1", idFormat)
		}

		data, err := os.ReadFile("testdata/get-user-next.json")
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
	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	c.baseURL = u

	resp, err := c.GetUser(context.Background(), true, "dummy-auth-token", "shogo82148")
	if err != nil {
		t.Fatal(err)
	}
	if resp.NodeID != "U_kgDOABGo4A" {
		t.Errorf("unexpected creator type: want %q, got %q", "U_kgDOABGo4A", resp.NodeID)
	}
}
