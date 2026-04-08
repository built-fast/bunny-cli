package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// Base URL constants for bunny.net API servers.
const (
	BaseURLPlatform = "https://api.bunny.net"
	BaseURLStream   = "https://video.bunnycdn.com"
	BaseURLStorage  = "https://storage.bunnycdn.com"
)

// ClientConfig holds the configuration needed to create a bunny.net API client.
type ClientConfig struct {
	// APIKey is the bunny.net API key used for authentication.
	APIKey string
	// BaseURL overrides the default API base URL. Defaults to BaseURLPlatform.
	BaseURL string
	// IsJSON reports whether the current output format is JSON-based.
	// It is used to decide retry/logging behavior.
	IsJSON func() bool
}

// Client is a thin HTTP client for the bunny.net API.
type Client struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
}

// NewClient creates a configured Client.
// API key precedence: --api-key flag > BUNNY_API_KEY env var > config file api_key.
func NewClient(cfg ClientConfig) (*Client, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("API key not configured, run 'bunny configure' or set BUNNY_API_KEY")
	}

	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = BaseURLPlatform
	}

	// Allow overriding the API base URL (e.g., for local Prism mock server in e2e tests).
	if envURL := os.Getenv("BUNNY_API_URL"); envURL != "" {
		baseURL = envURL
	}

	base := http.DefaultTransport
	transport := newRetryTransport(base, os.Stderr, cfg.IsJSON)

	return &Client{
		httpClient: &http.Client{Transport: transport},
		baseURL:    baseURL,
		apiKey:     cfg.APIKey,
	}, nil
}

// Do executes an HTTP request, injecting the AccessKey header.
// If body is non-nil, it is marshaled to JSON. If result is non-nil,
// the response body is unmarshaled into it.
func (c *Client) Do(ctx context.Context, method, path string, body, result any) error {
	var bodyReader io.Reader
	var getBody func() (io.ReadCloser, error)

	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshaling request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
		getBody = func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(data)), nil
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("AccessKey", c.apiKey)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
		req.GetBody = getBody
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		return parseErrorResponse(resp)
	}

	if result != nil && resp.ContentLength != 0 {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("decoding response: %w", err)
		}
	}

	return nil
}

// Get is a convenience wrapper for Do with GET method.
func (c *Client) Get(ctx context.Context, path string, result any) error {
	return c.Do(ctx, http.MethodGet, path, nil, result)
}

// Post is a convenience wrapper for Do with POST method.
func (c *Client) Post(ctx context.Context, path string, body, result any) error {
	return c.Do(ctx, http.MethodPost, path, body, result)
}

// Put is a convenience wrapper for Do with PUT method.
func (c *Client) Put(ctx context.Context, path string, body, result any) error {
	return c.Do(ctx, http.MethodPut, path, body, result)
}

// Delete is a convenience wrapper for Do with DELETE method.
func (c *Client) Delete(ctx context.Context, path string) error {
	return c.Do(ctx, http.MethodDelete, path, nil, nil)
}
