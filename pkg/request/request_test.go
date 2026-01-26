package request

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestNewConfig_Defaults(t *testing.T) {
	cfg := newConfig(nil)

	if cfg.Timeout != defaultTimeout {
		t.Fatalf("expected default timeout %v, got %v", defaultTimeout, cfg.Timeout)
	}

	if cfg.Method != "GET" {
		t.Fatalf("expected GET, got %s", cfg.Method)
	}

	if cfg.Headers == nil {
		t.Fatal("expected headers to be initialized")
	}
}

func TestNewConfig_Overrides(t *testing.T) {
	cfg := newConfig([]Options{
		WithMethod("POST"),
		WithTimeout(5 * time.Second),
	})

	if cfg.Method != "POST" {
		t.Fatalf("expected POST, got %s", cfg.Method)
	}

	if cfg.Timeout != 5*time.Second {
		t.Fatalf("unexpected timeout")
	}
}

func TestWithHeaders(t *testing.T) {
	h := http.Header{}
	h.Add("X-Test", "123")

	cfg := newConfig([]Options{
		WithHeaders(h),
	})

	if cfg.Headers.Get("X-Test") != "123" {
		t.Fatal("header not set")
	}
}

func TestWithQueryParams(t *testing.T) {
	q := url.Values{}
	q.Set("page", "1")

	cfg := newConfig([]Options{
		WithQueryParams(q),
	})

	if cfg.QueryParams.Get("page") != "1" {
		t.Fatal("query param not set")
	}
}

func TestRequestRaw_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	_, err := RequestRaw(context.Background(),
		WithEndpoint(srv.URL),
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRequestRaw_InvalidMethod(t *testing.T) {
	_, err := RequestRaw(context.Background(),
		WithEndpoint("http://example.com"),
		WithMethod("INVALID"),
	)

	if err == nil {
		t.Fatal("expected error for invalid method")
	}
}

func TestRequestRaw_BodyConflict(t *testing.T) {
	_, err := RequestRaw(context.Background(),
		WithEndpoint("http://example.com"),
		WithBody(map[string]string{"a": "b"}),
		WithBodyMarshalled([]byte(`{}`)),
	)

	if err == nil {
		t.Fatal("expected body conflict error")
	}
}

func TestRequest_JSONDecode(t *testing.T) {
	type Resp struct {
		Message string `json:"message"`
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message":"ok"}`))
	}))
	defer srv.Close()

	resp, err := Request[Resp](context.Background(),
		WithEndpoint(srv.URL),
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Message != "ok" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestRequest_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad request", http.StatusBadRequest)
	}))
	defer srv.Close()

	_, err := Request[any](context.Background(),
		WithEndpoint(srv.URL),
	)

	if err == nil {
		t.Fatal("expected http error")
	}
}

func TestRequest_ContextCancelled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := Request[any](
		ctx,
		WithEndpoint(server.URL),
		WithMethod("GET"),
	)

	if err == nil {
		t.Fatalf("expected error due to context cancellation, got nil")
	}

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestRequestRaw_ProxyUsed(t *testing.T) {
	target := "http://example.com/ok"

	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI != target {
			t.Fatalf("expected proxy request URI %s, got %s", target, r.RequestURI)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer proxy.Close()

	_, err := RequestRaw(context.Background(),
		WithEndpoint(target),
		WithProxy(proxy.URL),
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
