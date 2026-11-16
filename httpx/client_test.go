package httpx

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestNewRequest(t *testing.T) {
	req, err := NewRequest(context.Background(), http.MethodPost, "https://example.test/api", "/items", url.Values{"id": {"7"}}, strings.NewReader("x"))
	if err != nil {
		t.Fatalf("NewRequest() = %v, want nil", err)
	}
	if got := req.URL.String(); got != "https://example.test/api/items?id=7" {
		t.Errorf("URL = %q, want %q", got, "https://example.test/api/items?id=7")
	}
	if req.Method != http.MethodPost {
		t.Errorf("Method = %q, want %q", req.Method, http.MethodPost)
	}
}

func TestNewRequestNoQuery(t *testing.T) {
	req, err := NewRequest(context.Background(), http.MethodGet, "http://127.0.0.1:80", "/health", nil, nil)
	if err != nil {
		t.Fatalf("NewRequest() = %v, want nil", err)
	}
	if got := req.URL.String(); got != "http://127.0.0.1:80/health" {
		t.Errorf("URL = %q, want %q", got, "http://127.0.0.1:80/health")
	}
}

func TestNewRequestInvalidBase(t *testing.T) {
	if _, err := NewRequest(context.Background(), http.MethodGet, "://bad", "/x", nil, nil); !errors.Is(err, ErrInvalidURL) {
		t.Errorf("NewRequest() = %v, want %v", err, ErrInvalidURL)
	}
}

func TestErrorBody(t *testing.T) {
	resp := &http.Response{Body: io.NopCloser(strings.NewReader("  not found \n"))}
	if got := ErrorBody(resp); got != "not found" {
		t.Errorf("ErrorBody() = %q, want %q", got, "not found")
	}
}

func TestErrorBodyLimit(t *testing.T) {
	resp := &http.Response{Body: io.NopCloser(strings.NewReader(strings.Repeat("x", 2*maxErrorBody)))}
	if got := ErrorBody(resp); len(got) != maxErrorBody {
		t.Errorf("len(ErrorBody()) = %d, want %d", len(got), maxErrorBody)
	}
}
