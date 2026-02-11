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

	"github.com/barkhayot/request/pkg/throttler"
	"github.com/andybalholm/brotli"
)

const (
	defaultTimeout = 2 * time.Second
)

var (
	methods = map[string]struct{}{
		"GET":    {},
		"POST":   {},
		"PUT":    {},
		"PATCH":  {},
		"DELETE": {},
	}
)

type Config struct {
	Body           any
	BodyMarshalled []byte
	Headers        http.Header
	QueryParams    url.Values
	Endpoint       string
	Method         string
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

func WithMethod(m string) Options {
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

	if _, ok := methods[cfg.Method]; !ok {
		return nil, fmt.Errorf("invalid method: %s", cfg.Method)
	}

	req, err := http.NewRequestWithContext(ctx, cfg.Method, u.String(), body)
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
		Method:  "GET",
		Headers: make(http.Header),
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	return cfg
}
