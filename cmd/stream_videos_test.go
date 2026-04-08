package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/built-fast/bunny-cli/internal/client"
	"github.com/built-fast/bunny-cli/internal/pagination"
	"github.com/spf13/cobra"
)

// mockStreamAPI implements StreamAPI for testing.
type mockStreamAPI struct {
	listVideosFn       func(ctx context.Context, libraryId int64, page, itemsPerPage int, search, collection, orderBy string) (pagination.PageResponse[*client.Video], error)
	getVideoFn         func(ctx context.Context, libraryId int64, videoId string) (*client.Video, error)
	createVideoFn      func(ctx context.Context, libraryId int64, body *client.VideoCreate) (*client.Video, error)
	updateVideoFn      func(ctx context.Context, libraryId int64, videoId string, body *client.VideoUpdate) error
	deleteVideoFn      func(ctx context.Context, libraryId int64, videoId string) error
	uploadVideoFn      func(ctx context.Context, libraryId int64, videoId string, body io.Reader, size int64) error
	fetchVideoFn       func(ctx context.Context, libraryId int64, body *client.VideoFetch, collectionId string, thumbnailTime int) (*client.StatusModel, error)
	reencodeVideoFn    func(ctx context.Context, libraryId int64, videoId string) (*client.Video, error)
	transcribeVideoFn  func(ctx context.Context, libraryId int64, videoId string, settings *client.TranscribeSettings) error
	listCollectionsFn  func(ctx context.Context, libraryId int64, page, itemsPerPage int, search, orderBy string) (pagination.PageResponse[*client.Collection], error)
	getCollectionFn    func(ctx context.Context, libraryId int64, collectionId string) (*client.Collection, error)
	createCollectionFn func(ctx context.Context, libraryId int64, body *client.CollectionCreate) (*client.Collection, error)
	updateCollectionFn func(ctx context.Context, libraryId int64, collectionId string, body *client.CollectionUpdate) error
	deleteCollectionFn func(ctx context.Context, libraryId int64, collectionId string) error
	addCaptionFn       func(ctx context.Context, libraryId int64, videoId, srclang string, body *client.CaptionAdd) error
	deleteCaptionFn    func(ctx context.Context, libraryId int64, videoId, srclang string) error
	getVideoStatsFn    func(ctx context.Context, libraryId int64, dateFrom, dateTo string, hourly bool, videoGuid string) (*client.VideoStatistics, error)
	getVideoHeatmapFn  func(ctx context.Context, libraryId int64, videoId string) (*client.VideoHeatmap, error)
}

func (m *mockStreamAPI) ListVideos(ctx context.Context, libraryId int64, page, itemsPerPage int, search, collection, orderBy string) (pagination.PageResponse[*client.Video], error) {
	return m.listVideosFn(ctx, libraryId, page, itemsPerPage, search, collection, orderBy)
}

func (m *mockStreamAPI) GetVideo(ctx context.Context, libraryId int64, videoId string) (*client.Video, error) {
	return m.getVideoFn(ctx, libraryId, videoId)
}

func (m *mockStreamAPI) CreateVideo(ctx context.Context, libraryId int64, body *client.VideoCreate) (*client.Video, error) {
	return m.createVideoFn(ctx, libraryId, body)
}

func (m *mockStreamAPI) UpdateVideo(ctx context.Context, libraryId int64, videoId string, body *client.VideoUpdate) error {
	return m.updateVideoFn(ctx, libraryId, videoId, body)
}

func (m *mockStreamAPI) DeleteVideo(ctx context.Context, libraryId int64, videoId string) error {
	return m.deleteVideoFn(ctx, libraryId, videoId)
}

func (m *mockStreamAPI) UploadVideo(ctx context.Context, libraryId int64, videoId string, body io.Reader, size int64) error {
	return m.uploadVideoFn(ctx, libraryId, videoId, body, size)
}

func (m *mockStreamAPI) FetchVideo(ctx context.Context, libraryId int64, body *client.VideoFetch, collectionId string, thumbnailTime int) (*client.StatusModel, error) {
	return m.fetchVideoFn(ctx, libraryId, body, collectionId, thumbnailTime)
}

func (m *mockStreamAPI) ReencodeVideo(ctx context.Context, libraryId int64, videoId string) (*client.Video, error) {
	return m.reencodeVideoFn(ctx, libraryId, videoId)
}

func (m *mockStreamAPI) TranscribeVideo(ctx context.Context, libraryId int64, videoId string, settings *client.TranscribeSettings) error {
	return m.transcribeVideoFn(ctx, libraryId, videoId, settings)
}

func (m *mockStreamAPI) ListCollections(ctx context.Context, libraryId int64, page, itemsPerPage int, search, orderBy string) (pagination.PageResponse[*client.Collection], error) {
	return m.listCollectionsFn(ctx, libraryId, page, itemsPerPage, search, orderBy)
}

func (m *mockStreamAPI) GetCollection(ctx context.Context, libraryId int64, collectionId string) (*client.Collection, error) {
	return m.getCollectionFn(ctx, libraryId, collectionId)
}

func (m *mockStreamAPI) CreateCollection(ctx context.Context, libraryId int64, body *client.CollectionCreate) (*client.Collection, error) {
	return m.createCollectionFn(ctx, libraryId, body)
}

func (m *mockStreamAPI) UpdateCollection(ctx context.Context, libraryId int64, collectionId string, body *client.CollectionUpdate) error {
	return m.updateCollectionFn(ctx, libraryId, collectionId, body)
}

func (m *mockStreamAPI) DeleteCollection(ctx context.Context, libraryId int64, collectionId string) error {
	return m.deleteCollectionFn(ctx, libraryId, collectionId)
}

func (m *mockStreamAPI) AddCaption(ctx context.Context, libraryId int64, videoId, srclang string, body *client.CaptionAdd) error {
	return m.addCaptionFn(ctx, libraryId, videoId, srclang, body)
}

func (m *mockStreamAPI) DeleteCaption(ctx context.Context, libraryId int64, videoId, srclang string) error {
	return m.deleteCaptionFn(ctx, libraryId, videoId, srclang)
}

func (m *mockStreamAPI) GetVideoStatistics(ctx context.Context, libraryId int64, dateFrom, dateTo string, hourly bool, videoGuid string) (*client.VideoStatistics, error) {
	return m.getVideoStatsFn(ctx, libraryId, dateFrom, dateTo, hourly, videoGuid)
}

func (m *mockStreamAPI) GetVideoHeatmap(ctx context.Context, libraryId int64, videoId string) (*client.VideoHeatmap, error) {
	return m.getVideoHeatmapFn(ctx, libraryId, videoId)
}

func newTestStreamApp(api StreamAPI) *App {
	return &App{NewStreamAPI: func(_ *cobra.Command) (StreamAPI, error) { return api, nil }}
}

func sampleVideo() *client.Video {
	return &client.Video{
		VideoLibraryId:       100,
		Guid:                 "abc-123-def",
		Title:                "My Test Video",
		Description:          "A test video",
		DateUploaded:         "2025-03-01T08:00:00Z",
		Views:                1500,
		IsPublic:             true,
		Length:               120,
		Status:               4, // Finished
		Framerate:            30.0,
		Width:                1920,
		Height:               1080,
		AvailableResolutions: "240p,360p,480p,720p,1080p",
		EncodeProgress:       100,
		StorageSize:          50000000,
		HasMP4Fallback:       true,
		CollectionId:         "col-abc-123",
		AverageWatchTime:     90,
		TotalWatchTime:       135000,
		Category:             "technology",
	}
}

// --- stream videos help ---

func TestStreamVideos_ShowsInHelp(t *testing.T) {
	t.Parallel()
	out, _, err := executeCommand(nil, "stream", "videos", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, sub := range []string{"list", "get", "create", "update", "delete", "upload", "fetch", "reencode", "transcribe"} {
		if !strings.Contains(out, sub) {
			t.Errorf("expected videos help to show %q subcommand", sub)
		}
	}
}

// --- stream videos list ---

func TestStreamVideosList_Table(t *testing.T) {
	t.Parallel()
	mock := &mockStreamAPI{
		listVideosFn: func(_ context.Context, libraryId int64, page, itemsPerPage int, search, collection, orderBy string) (pagination.PageResponse[*client.Video], error) {
			if libraryId != 100 {
				t.Errorf("expected libraryId=100, got %d", libraryId)
			}
			return pagination.PageResponse[*client.Video]{
				Items:        []*client.Video{sampleVideo()},
				HasMoreItems: false,
			}, nil
		},
	}
	app := newTestStreamApp(mock)

	out, _, err := executeCommand(app, "stream", "videos", "list", "100")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "My Test Video") {
		t.Error("expected output to contain video title")
	}
	if !strings.Contains(out, "Finished") {
		t.Error("expected output to contain status name")
	}
}

func TestStreamVideosList_JSON(t *testing.T) {
	t.Parallel()
	mock := &mockStreamAPI{
		listVideosFn: func(_ context.Context, libraryId int64, page, itemsPerPage int, search, collection, orderBy string) (pagination.PageResponse[*client.Video], error) {
			return pagination.PageResponse[*client.Video]{
				Items:        []*client.Video{sampleVideo()},
				HasMoreItems: false,
			}, nil
		},
	}
	app := newTestStreamApp(mock)

	out, _, err := executeCommand(app, "stream", "videos", "list", "100", "--output", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var result map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if result["object"] != "list" {
		t.Errorf("expected object=list, got %v", result["object"])
	}
}

func TestStreamVideosList_MissingLibraryID(t *testing.T) {
	t.Parallel()
	_, _, err := executeCommand(nil, "stream", "videos", "list")
	if err == nil {
		t.Fatal("expected error for missing library ID")
	}
}

func TestStreamVideosList_SearchAndCollection(t *testing.T) {
	t.Parallel()
	var capturedSearch, capturedCollection string
	mock := &mockStreamAPI{
		listVideosFn: func(_ context.Context, libraryId int64, page, itemsPerPage int, search, collection, orderBy string) (pagination.PageResponse[*client.Video], error) {
			capturedSearch = search
			capturedCollection = collection
			return pagination.PageResponse[*client.Video]{}, nil
		},
	}
	app := newTestStreamApp(mock)

	_, _, err := executeCommand(app, "stream", "videos", "list", "100", "--search", "test", "--collection", "col-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedSearch != "test" {
		t.Errorf("expected search='test', got %q", capturedSearch)
	}
	if capturedCollection != "col-123" {
		t.Errorf("expected collection='col-123', got %q", capturedCollection)
	}
}

// --- stream videos get ---

func TestStreamVideosGet_Table(t *testing.T) {
	t.Parallel()
	mock := &mockStreamAPI{
		getVideoFn: func(_ context.Context, libraryId int64, videoId string) (*client.Video, error) {
			if libraryId != 100 {
				t.Errorf("expected libraryId=100, got %d", libraryId)
			}
			if videoId != "abc-123-def" {
				t.Errorf("expected videoId=abc-123-def, got %s", videoId)
			}
			return sampleVideo(), nil
		},
	}
	app := newTestStreamApp(mock)

	out, _, err := executeCommand(app, "stream", "videos", "get", "100", "abc-123-def")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "My Test Video") {
		t.Error("expected output to contain video title")
	}
	if !strings.Contains(out, "1920") {
		t.Error("expected output to contain width")
	}
}

func TestStreamVideosGet_MissingVideoID(t *testing.T) {
	t.Parallel()
	_, _, err := executeCommand(nil, "stream", "videos", "get", "100")
	if err == nil {
		t.Fatal("expected error for missing video ID")
	}
}

// --- stream videos create ---

func TestStreamVideosCreate_Success(t *testing.T) {
	t.Parallel()
	var capturedBody *client.VideoCreate
	mock := &mockStreamAPI{
		createVideoFn: func(_ context.Context, libraryId int64, body *client.VideoCreate) (*client.Video, error) {
			capturedBody = body
			return &client.Video{Guid: "new-guid", Title: body.Title, VideoLibraryId: libraryId}, nil
		},
	}
	app := newTestStreamApp(mock)

	out, _, err := executeCommand(app, "stream", "videos", "create", "100", "--title", "New Video")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedBody.Title != "New Video" {
		t.Errorf("expected title 'New Video', got %q", capturedBody.Title)
	}
	if !strings.Contains(out, "new-guid") {
		t.Error("expected output to contain video GUID")
	}
}

func TestStreamVideosCreate_RequiresTitle(t *testing.T) {
	t.Parallel()
	_, _, err := executeCommand(nil, "stream", "videos", "create", "100")
	if err == nil {
		t.Fatal("expected error for missing required --title flag")
	}
}

// --- stream videos update ---

func TestStreamVideosUpdate_Success(t *testing.T) {
	t.Parallel()
	var capturedTitle string
	mock := &mockStreamAPI{
		updateVideoFn: func(_ context.Context, libraryId int64, videoId string, body *client.VideoUpdate) error {
			if body.Title != nil {
				capturedTitle = *body.Title
			}
			return nil
		},
		getVideoFn: func(_ context.Context, libraryId int64, videoId string) (*client.Video, error) {
			v := sampleVideo()
			v.Title = "Updated Title"
			return v, nil
		},
	}
	app := newTestStreamApp(mock)

	out, _, err := executeCommand(app, "stream", "videos", "update", "100", "abc-123-def", "--title", "Updated Title")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedTitle != "Updated Title" {
		t.Errorf("expected title='Updated Title', got %q", capturedTitle)
	}
	if !strings.Contains(out, "Updated Title") {
		t.Error("expected output to show updated title")
	}
}

// --- stream videos delete ---

func TestStreamVideosDelete_WithYes(t *testing.T) {
	t.Parallel()
	var deletedVideoId string
	mock := &mockStreamAPI{
		deleteVideoFn: func(_ context.Context, libraryId int64, videoId string) error {
			deletedVideoId = videoId
			return nil
		},
	}
	app := newTestStreamApp(mock)

	out, _, err := executeCommand(app, "stream", "videos", "delete", "100", "abc-123-def", "--yes")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deletedVideoId != "abc-123-def" {
		t.Errorf("expected deleted videoId=abc-123-def, got %s", deletedVideoId)
	}
	if !strings.Contains(out, "Video deleted") {
		t.Error("expected deletion confirmation message")
	}
}

func TestStreamVideosDelete_Canceled(t *testing.T) {
	t.Parallel()
	mock := &mockStreamAPI{
		deleteVideoFn: func(_ context.Context, libraryId int64, videoId string) error {
			t.Error("delete should not have been called")
			return nil
		},
	}
	app := newTestStreamApp(mock)

	stdin := bytes.NewBufferString("n\n")
	_, stderr, err := executeCommandWithStdin(app, stdin, "stream", "videos", "delete", "100", "abc-123-def")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "Deletion canceled") {
		t.Error("expected cancellation message")
	}
}

// --- stream videos fetch ---

func TestStreamVideosFetch_Success(t *testing.T) {
	t.Parallel()
	var capturedURL string
	mock := &mockStreamAPI{
		fetchVideoFn: func(_ context.Context, libraryId int64, body *client.VideoFetch, collectionId string, thumbnailTime int) (*client.StatusModel, error) {
			capturedURL = body.Url
			return &client.StatusModel{Success: true}, nil
		},
	}
	app := newTestStreamApp(mock)

	out, _, err := executeCommand(app, "stream", "videos", "fetch", "100", "--url", "https://example.com/video.mp4")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedURL != "https://example.com/video.mp4" {
		t.Errorf("expected URL='https://example.com/video.mp4', got %q", capturedURL)
	}
	if !strings.Contains(out, "Video fetch initiated") {
		t.Error("expected fetch confirmation message")
	}
}

func TestStreamVideosFetch_RequiresURL(t *testing.T) {
	t.Parallel()
	_, _, err := executeCommand(nil, "stream", "videos", "fetch", "100")
	if err == nil {
		t.Fatal("expected error for missing required --url flag")
	}
}

// --- stream videos reencode ---

func TestStreamVideosReencode_Success(t *testing.T) {
	t.Parallel()
	var capturedVideoId string
	mock := &mockStreamAPI{
		reencodeVideoFn: func(_ context.Context, libraryId int64, videoId string) (*client.Video, error) {
			capturedVideoId = videoId
			return sampleVideo(), nil
		},
	}
	app := newTestStreamApp(mock)

	out, _, err := executeCommand(app, "stream", "videos", "reencode", "100", "abc-123-def")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedVideoId != "abc-123-def" {
		t.Errorf("expected videoId=abc-123-def, got %s", capturedVideoId)
	}
	if !strings.Contains(out, "My Test Video") {
		t.Error("expected output to contain video title")
	}
}

// --- stream videos transcribe ---

func TestStreamVideosTranscribe_Success(t *testing.T) {
	t.Parallel()
	var capturedSettings *client.TranscribeSettings
	mock := &mockStreamAPI{
		transcribeVideoFn: func(_ context.Context, libraryId int64, videoId string, settings *client.TranscribeSettings) error {
			capturedSettings = settings
			return nil
		},
	}
	app := newTestStreamApp(mock)

	out, _, err := executeCommand(app, "stream", "videos", "transcribe", "100", "abc-123-def", "--source-language", "en", "--generate-title")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedSettings.SourceLanguage != "en" {
		t.Errorf("expected source-language='en', got %q", capturedSettings.SourceLanguage)
	}
	if !capturedSettings.GenerateTitle {
		t.Error("expected generate-title=true")
	}
	if !strings.Contains(out, "Transcription started") {
		t.Error("expected transcription confirmation message")
	}
}

// --- stream videos error propagation ---

func TestStreamVideosList_ErrorPropagation(t *testing.T) {
	t.Parallel()
	mock := &mockStreamAPI{
		listVideosFn: func(_ context.Context, libraryId int64, page, itemsPerPage int, search, collection, orderBy string) (pagination.PageResponse[*client.Video], error) {
			return pagination.PageResponse[*client.Video]{}, fmt.Errorf("stream API error")
		},
	}
	app := newTestStreamApp(mock)

	_, stderr, err := executeCommand(app, "stream", "videos", "list", "100")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(stderr, "stream API error") {
		t.Errorf("expected API error in stderr, got %q", stderr)
	}
}
