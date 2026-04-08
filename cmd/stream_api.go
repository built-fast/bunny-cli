package cmd

import (
	"context"
	"io"

	"github.com/built-fast/bunny-cli/internal/client"
	"github.com/built-fast/bunny-cli/internal/pagination"
)

// StreamAPI abstracts the bunny.net Stream (video) API methods,
// allowing tests to inject mocks without making real API calls.
type StreamAPI interface {
	// Videos
	ListVideos(ctx context.Context, libraryId int64, page, itemsPerPage int, search, collection, orderBy string) (pagination.PageResponse[*client.Video], error)
	GetVideo(ctx context.Context, libraryId int64, videoId string) (*client.Video, error)
	CreateVideo(ctx context.Context, libraryId int64, body *client.VideoCreate) (*client.Video, error)
	UpdateVideo(ctx context.Context, libraryId int64, videoId string, body *client.VideoUpdate) error
	DeleteVideo(ctx context.Context, libraryId int64, videoId string) error
	UploadVideo(ctx context.Context, libraryId int64, videoId string, body io.Reader, size int64) error
	FetchVideo(ctx context.Context, libraryId int64, body *client.VideoFetch, collectionId string, thumbnailTime int) (*client.StatusModel, error)
	ReencodeVideo(ctx context.Context, libraryId int64, videoId string) (*client.Video, error)
	TranscribeVideo(ctx context.Context, libraryId int64, videoId string, settings *client.TranscribeSettings) error

	// Collections
	ListCollections(ctx context.Context, libraryId int64, page, itemsPerPage int, search, orderBy string) (pagination.PageResponse[*client.Collection], error)
	GetCollection(ctx context.Context, libraryId int64, collectionId string) (*client.Collection, error)
	CreateCollection(ctx context.Context, libraryId int64, body *client.CollectionCreate) (*client.Collection, error)
	UpdateCollection(ctx context.Context, libraryId int64, collectionId string, body *client.CollectionUpdate) error
	DeleteCollection(ctx context.Context, libraryId int64, collectionId string) error

	// Captions
	AddCaption(ctx context.Context, libraryId int64, videoId, srclang string, body *client.CaptionAdd) error
	DeleteCaption(ctx context.Context, libraryId int64, videoId, srclang string) error

	// Statistics
	GetVideoStatistics(ctx context.Context, libraryId int64, dateFrom, dateTo string, hourly bool, videoGuid string) (*client.VideoStatistics, error)
	GetVideoHeatmap(ctx context.Context, libraryId int64, videoId string) (*client.VideoHeatmap, error)
}
