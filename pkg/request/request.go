package request

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/andybalholm/brotli"
	"github.com/barkhayot/request/pkg/throttler"
)

const (
	defaultTimeout = 2 * time.Second
)

type Method string

const (
	MethodGet    Method = "GET"
	MethodPost   Method = "POST"
	MethodPut    Method = "PUT"
	MethodPatch  Method = "PATCH"
	MethodDelete Method = "DELETE"
	MethodQuery  Method = "QUERY"
)

// methodInfo describes the semantics we care about when building a request.
// bodyIsData marks methods whose body is the semantic payload rather than an
// oddity, which is what separates QUERY from GET.
type methodInfo struct {
	safe       bool
	idempotent bool
	bodyIsData bool
}

// safeWithBody reports whether a redirect may not drop the request body
// without changing what the request means.
func (m methodInfo) safeWithBody() bool {
	return m.safe && m.bodyIsData
}

var (
	methods = map[Method]methodInfo{
		MethodGet:    {safe: true, idempotent: true},
		MethodPost:   {bodyIsData: true},
		MethodPut:    {idempotent: true, bodyIsData: true},
		MethodPatch:  {bodyIsData: true},
		MethodDelete: {idempotent: true},
		MethodQuery:  {safe: true, idempotent: true, bodyIsData: true},
	}
)

type Config struct {
	Body           any
	BodyMarshalled []byte
	Headers        http.Header
	QueryParams    url.Values
	Endpoint       string
	Method         Method
	Timeout        time.Duration

	Throttler throttler.Throttler
	Proxy     string
}

type Options func(*Config)

func WithTimeout(t time.Duration) Options {
	return func(c *Config) {
		c.Timeout = t
	}
}

func WithEndpoint(e string) Options {
	return func(c *Config) {
		c.Endpoint = e
	}
}

func WithMethod(m Method) Options {
	return func(c *Config) {
		c.Method = m
	}
}

func WithBody(b any) Options {
	return func(c *Config) {
		c.Body = b
	}
}

func WithBodyMarshalled(b []byte) Options {
	return func(c *Config) {
		c.BodyMarshalled = b
	}
}

func WithHeaders(h http.Header) Options {
	return func(c *Config) {
		c.Headers = h
	}
}

func WithQueryParams(q url.Values) Options {
	return func(c *Config) {
		c.QueryParams = q
	}
}

func WithThrottler(t throttler.Throttler) Options {
	return func(c *Config) {
		c.Throttler = t
	}
}

func WithProxy(proxy string) Options {
	return func(c *Config) {
		c.Proxy = proxy
	}
}

func Request[T any](ctx context.Context, opts ...Options) (T, error) {
	var out T
	resp, err := requestRaw(ctx, newConfig(opts))
	if err != nil {
		return out, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		b, _ := io.ReadAll(resp.Body)
		return out, fmt.Errorf("http %d: %s", resp.StatusCode, string(b))
	}

	// Handle Brotli decompression
	// TODO: extend it later
	var reader io.Reader = resp.Body
	if resp.Header.Get("Content-Encoding") == "br" {
		reader = brotli.NewReader(resp.Body)
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		return out, err
	}

	if err := json.Unmarshal(body, &out); err != nil {
		return out, err
	}

	return out, nil
}

func RequestRaw(ctx context.Context, opts ...Options) (*http.Response, error) {
	return requestRaw(ctx, newConfig(opts))
}

func requestRaw(ctx context.Context, cfg Config) (*http.Response, error) {
	var body io.Reader

	if cfg.Body != nil && cfg.BodyMarshalled != nil {
		return nil, errors.New("cannot set both Body and BodyMarshalled")
	}

	if cfg.Body != nil {
		b, err := json.Marshal(cfg.Body)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(b)
	}

	if cfg.BodyMarshalled != nil {
		body = bytes.NewReader(cfg.BodyMarshalled)
	}

	if cfg.Throttler != nil {
		if err := cfg.Throttler.Wait(ctx); err != nil {
			return nil, err
		}
	}

	u, err := url.Parse(cfg.Endpoint)
	if err != nil {
		return nil, err
	}

	if len(cfg.QueryParams) > 0 {
		u.RawQuery = cfg.QueryParams.Encode()
	}

	info, ok := methods[cfg.Method]
	if !ok {
		return nil, fmt.Errorf("invalid method: %s", cfg.Method)
	}

	req, err := http.NewRequestWithContext(ctx, string(cfg.Method), u.String(), body)
	if err != nil {
		return nil, err
	}

	if body != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	for k, v := range cfg.Headers {
		for _, vv := range v {
			req.Header.Add(k, vv)
		}
	}

	client := &http.Client{
		Timeout: cfg.Timeout,
	}

	// net/http rewrites any non-GET/HEAD method to GET and drops the body on
	// 301, 302 and 303. For a safe method whose body *is* the request, that
	// silently turns a query into a bare GET, so hand those responses back to
	// the caller instead. 303 genuinely means "GET this other resource", and
	// 307/308 preserve the method and replay the body, so both still follow.
	if info.safeWithBody() {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			if resp := req.Response; resp != nil {
				switch resp.StatusCode {
				case http.StatusMovedPermanently, http.StatusFound:
					return http.ErrUseLastResponse
				}
			}
			return nil
		}
	}

	if cfg.Proxy != "" {
		proxyURL, err := validateProxy(cfg.Proxy)
		if err != nil {
			return nil, err
		}
		client.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
	}

	return client.Do(req)
}

func newConfig(opts []Options) Config {
	cfg := Config{
		Timeout: defaultTimeout,
		Method:  MethodGet,
		Headers: make(http.Header),
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	return cfg
}
