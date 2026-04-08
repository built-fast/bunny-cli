package cmd

import (
	"context"
	"io"

	"github.com/built-fast/bunny-cli/internal/client"
)

// StorageAPI abstracts the bunny.net Edge Storage API methods,
// allowing tests to inject mocks without making real API calls.
type StorageAPI interface {
	ListFiles(ctx context.Context, zoneName, path string) ([]client.StorageObject, error)
	DownloadFile(ctx context.Context, zoneName, path string) (io.ReadCloser, int64, error)
	UploadFile(ctx context.Context, zoneName, path string, body io.Reader, size int64, checksum string) error
	DeleteFile(ctx context.Context, zoneName, path string) error
}
