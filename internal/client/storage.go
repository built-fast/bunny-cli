package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// StorageClientConfig holds the configuration needed to create an Edge Storage client.
type StorageClientConfig struct {
	// Password is the per-zone storage zone password used for authentication.
	Password string
	// Hostname is the storage API hostname (e.g., "storage.bunnycdn.com").
	Hostname string
	// IsJSON reports whether the current output format is JSON-based.
	IsJSON func() bool
}

// StorageClient is an HTTP client for the bunny.net Edge Storage API.
// Unlike the main Client, this handles binary file operations and uses
// per-zone password authentication.
type StorageClient struct {
	httpClient *http.Client
	baseURL    string
	password   string
}

// StorageObject represents a file or directory in Edge Storage.
type StorageObject struct {
	Guid            string `json:"Guid"`
	StorageZoneName string `json:"StorageZoneName"`
	Path            string `json:"Path"`
	ObjectName      string `json:"ObjectName"`
	Length          int64  `json:"Length"`
	LastChanged     string `json:"LastChanged"`
	IsDirectory     bool   `json:"IsDirectory"`
	ServerId        int    `json:"ServerId"`
	DateCreated     string `json:"DateCreated"`
	StorageZoneId   int64  `json:"StorageZoneId"`
}

// NewStorageClient creates a configured StorageClient for Edge Storage operations.
func NewStorageClient(cfg StorageClientConfig) (*StorageClient, error) {
	if cfg.Password == "" {
		return nil, fmt.Errorf("storage zone password is required")
	}
	if cfg.Hostname == "" {
		return nil, fmt.Errorf("storage hostname is required")
	}

	baseURL := cfg.Hostname
	if !strings.HasPrefix(baseURL, "http") {
		baseURL = "https://" + baseURL
	}
	baseURL = strings.TrimRight(baseURL, "/")

	// Allow overriding the storage base URL (e.g., for e2e tests with Prism).
	if envURL := os.Getenv("BUNNY_STORAGE_URL"); envURL != "" {
		baseURL = strings.TrimRight(envURL, "/")
	}

	base := http.DefaultTransport
	transport := newRetryTransport(base, os.Stderr, cfg.IsJSON)

	return &StorageClient{
		httpClient: &http.Client{Transport: transport},
		baseURL:    baseURL,
		password:   cfg.Password,
	}, nil
}

// ListFiles returns the contents of a directory in Edge Storage.
func (c *StorageClient) ListFiles(ctx context.Context, zoneName, path string) ([]StorageObject, error) {
	reqPath := "/" + zoneName + "/"
	if path != "" {
		reqPath += strings.TrimLeft(path, "/")
		if !strings.HasSuffix(reqPath, "/") {
			reqPath += "/"
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+reqPath, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("AccessKey", c.password)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		return nil, parseErrorResponse(resp)
	}

	var objects []StorageObject
	if err := json.NewDecoder(resp.Body).Decode(&objects); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return objects, nil
}

// DownloadFile downloads a file from Edge Storage.
// The caller is responsible for closing the returned ReadCloser.
func (c *StorageClient) DownloadFile(ctx context.Context, zoneName, path string) (io.ReadCloser, int64, error) {
	reqPath := "/" + zoneName + "/" + strings.TrimLeft(path, "/")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+reqPath, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("AccessKey", c.password)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}

	if resp.StatusCode >= 400 {
		defer func() { _ = resp.Body.Close() }()
		return nil, 0, parseErrorResponse(resp)
	}

	return resp.Body, resp.ContentLength, nil
}

// UploadFile uploads a file to Edge Storage.
// If checksum is non-empty, it is sent as an uppercase hex SHA256 Checksum header.
func (c *StorageClient) UploadFile(ctx context.Context, zoneName, path string, body io.Reader, size int64, checksum string) error {
	reqPath := "/" + zoneName + "/" + strings.TrimLeft(path, "/")

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.baseURL+reqPath, body)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("AccessKey", c.password)
	req.Header.Set("Content-Type", "application/octet-stream")
	req.ContentLength = size

	if checksum != "" {
		req.Header.Set("Checksum", strings.ToUpper(checksum))
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		return parseErrorResponse(resp)
	}

	return nil
}

// DeleteFile deletes a file or directory from Edge Storage.
// If the target is a directory, all contents are deleted recursively.
func (c *StorageClient) DeleteFile(ctx context.Context, zoneName, path string) error {
	reqPath := "/" + zoneName + "/" + strings.TrimLeft(path, "/")

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.baseURL+reqPath, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("AccessKey", c.password)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		return parseErrorResponse(resp)
	}

	return nil
}
