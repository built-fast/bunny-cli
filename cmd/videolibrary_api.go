package cmd

import (
	"context"

	"github.com/built-fast/bunny-cli/internal/client"
	"github.com/built-fast/bunny-cli/internal/pagination"
)

// VideoLibraryAPI abstracts the bunny.net video library API methods,
// allowing tests to inject mocks without making real API calls.
type VideoLibraryAPI interface {
	ListVideoLibraries(ctx context.Context, page, perPage int, search string) (pagination.PageResponse[*client.VideoLibrary], error)
	GetVideoLibrary(ctx context.Context, id int64) (*client.VideoLibrary, error)
	CreateVideoLibrary(ctx context.Context, body *client.VideoLibraryCreate) (*client.VideoLibrary, error)
	UpdateVideoLibrary(ctx context.Context, id int64, body *client.VideoLibraryUpdate) (*client.VideoLibrary, error)
	DeleteVideoLibrary(ctx context.Context, id int64) error
	ResetVideoLibraryApiKey(ctx context.Context, id int64) error
	ListVideoLibraryLanguages(ctx context.Context) ([]client.VideoLibraryLanguage, error)
}
