package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/built-fast/bunny-cli/internal/pagination"
)

// streamPageResponse is the pagination envelope used by the Stream API.
// It uses lowercase JSON keys (unlike the platform API's PascalCase).
type streamPageResponse[T any] struct {
	TotalItems   int `json:"totalItems"`
	CurrentPage  int `json:"currentPage"`
	ItemsPerPage int `json:"itemsPerPage"`
	Items        []T `json:"items"`
}

// toPageResponse converts to the standard PageResponse used across the CLI.
func (s streamPageResponse[T]) toPageResponse() pagination.PageResponse[T] {
	hasMore := false
	if s.ItemsPerPage > 0 {
		hasMore = s.CurrentPage*s.ItemsPerPage < s.TotalItems
	}
	return pagination.PageResponse[T]{
		Items:        s.Items,
		CurrentPage:  s.CurrentPage,
		TotalItems:   s.TotalItems,
		HasMoreItems: hasMore,
	}
}

// VideoStatusName returns a human-readable name for a video status code.
func VideoStatusName(status int) string {
	switch status {
	case 0:
		return "Created"
	case 1:
		return "Uploaded"
	case 2:
		return "Processing"
	case 3:
		return "Transcoding"
	case 4:
		return "Finished"
	case 5:
		return "Error"
	case 6:
		return "UploadFailed"
	case 7:
		return "JitSegmenting"
	case 8:
		return "JitPlaylistsCreated"
	default:
		return fmt.Sprintf("Unknown(%d)", status)
	}
}

// --- Video models ---

// Video represents a video in a bunny.net stream library.
type Video struct {
	VideoLibraryId       int64     `json:"videoLibraryId"`
	Guid                 string    `json:"guid"`
	Title                string    `json:"title"`
	Description          string    `json:"description"`
	DateUploaded         string    `json:"dateUploaded"`
	Views                int64     `json:"views"`
	IsPublic             bool      `json:"isPublic"`
	Length               int       `json:"length"`
	Status               int       `json:"status"`
	Framerate            float64   `json:"framerate"`
	Width                int       `json:"width"`
	Height               int       `json:"height"`
	AvailableResolutions string    `json:"availableResolutions"`
	OutputCodecs         string    `json:"outputCodecs"`
	ThumbnailCount       int       `json:"thumbnailCount"`
	EncodeProgress       int       `json:"encodeProgress"`
	StorageSize          int64     `json:"storageSize"`
	HasMP4Fallback       bool      `json:"hasMP4Fallback"`
	CollectionId         string    `json:"collectionId"`
	ThumbnailFileName    string    `json:"thumbnailFileName"`
	AverageWatchTime     int64     `json:"averageWatchTime"`
	TotalWatchTime       int64     `json:"totalWatchTime"`
	Category             string    `json:"category"`
	Captions             []Caption `json:"captions"`
	Chapters             []Chapter `json:"chapters"`
	Moments              []Moment  `json:"moments"`
	MetaTags             []MetaTag `json:"metaTags"`
}

// VideoCreate holds the fields for creating a video record.
type VideoCreate struct {
	Title         string `json:"title"`
	CollectionId  string `json:"collectionId,omitempty"`
	ThumbnailTime int    `json:"thumbnailTime,omitempty"`
}

// VideoUpdate holds the fields for updating a video.
type VideoUpdate struct {
	Title        *string    `json:"title,omitempty"`
	CollectionId *string    `json:"collectionId,omitempty"`
	Chapters     *[]Chapter `json:"chapters,omitempty"`
	Moments      *[]Moment  `json:"moments,omitempty"`
	MetaTags     *[]MetaTag `json:"metaTags,omitempty"`
}

// VideoFetch holds the fields for fetching a video from a URL.
type VideoFetch struct {
	Url     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
	Title   string            `json:"title,omitempty"`
}

// StatusModel represents a generic status response from the Stream API.
type StatusModel struct {
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	StatusCode int    `json:"statusCode"`
}

// --- Collection models ---

// Collection represents a collection in a bunny.net stream library.
type Collection struct {
	VideoLibraryId   int64    `json:"videoLibraryId"`
	Guid             string   `json:"guid"`
	Name             string   `json:"name"`
	VideoCount       int64    `json:"videoCount"`
	TotalSize        int64    `json:"totalSize"`
	PreviewVideoIds  string   `json:"previewVideoIds"`
	PreviewImageUrls []string `json:"previewImageUrls"`
}

// CollectionCreate holds the fields for creating a collection.
type CollectionCreate struct {
	Name string `json:"name"`
}

// CollectionUpdate holds the fields for updating a collection.
type CollectionUpdate struct {
	Name string `json:"name"`
}

// --- Caption models ---

// Caption represents a caption track on a video.
type Caption struct {
	Srclang string `json:"srclang"`
	Label   string `json:"label"`
}

// CaptionAdd holds the fields for adding a caption to a video.
type CaptionAdd struct {
	Srclang      string `json:"srclang"`
	Label        string `json:"label"`
	CaptionsFile string `json:"captionsFile"`
}

// --- Supporting models ---

// Chapter represents a chapter marker in a video.
type Chapter struct {
	Title string `json:"title"`
	Start int    `json:"start"`
	End   int    `json:"end"`
}

// Moment represents a moment marker in a video.
type Moment struct {
	Label     string `json:"label"`
	Timestamp int    `json:"timestamp"`
}

// MetaTag represents a meta tag on a video.
type MetaTag struct {
	Property string `json:"property"`
	Value    string `json:"value"`
}

// TranscribeSettings holds settings for video transcription.
type TranscribeSettings struct {
	TargetLanguages     []string `json:"targetLanguages,omitempty"`
	SourceLanguage      string   `json:"sourceLanguage,omitempty"`
	GenerateTitle       bool     `json:"generateTitle,omitempty"`
	GenerateDescription bool     `json:"generateDescription,omitempty"`
	GenerateChapters    bool     `json:"generateChapters,omitempty"`
	GenerateMoments     bool     `json:"generateMoments,omitempty"`
}

// VideoStatistics holds statistics for a video library.
type VideoStatistics struct {
	ViewsChart        map[string]int64 `json:"viewsChart"`
	WatchTimeChart    map[string]int64 `json:"watchTimeChart"`
	CountryViewCounts map[string]int64 `json:"countryViewCounts"`
	CountryWatchTime  map[string]int64 `json:"countryWatchTime"`
	EngagementScore   int              `json:"engagementScore"`
}

// VideoHeatmap holds heatmap data for a video.
type VideoHeatmap struct {
	Heatmap map[string]int `json:"heatmap"`
}

// --- Stream API client methods ---

// ListVideos returns a paginated list of videos in a library.
func (c *Client) ListVideos(ctx context.Context, libraryId int64, page, itemsPerPage int, search, collection, orderBy string) (pagination.PageResponse[*Video], error) {
	params := url.Values{}
	params.Set("page", fmt.Sprintf("%d", page))
	params.Set("itemsPerPage", fmt.Sprintf("%d", itemsPerPage))
	if search != "" {
		params.Set("search", search)
	}
	if collection != "" {
		params.Set("collection", collection)
	}
	if orderBy != "" {
		params.Set("orderBy", orderBy)
	}

	path := fmt.Sprintf("/library/%d/videos?%s", libraryId, params.Encode())
	var resp streamPageResponse[*Video]
	if err := c.Get(ctx, path, &resp); err != nil {
		return pagination.PageResponse[*Video]{}, err
	}
	return resp.toPageResponse(), nil
}

// GetVideo returns a single video by library ID and video GUID.
func (c *Client) GetVideo(ctx context.Context, libraryId int64, videoId string) (*Video, error) {
	var v Video
	err := c.Get(ctx, fmt.Sprintf("/library/%d/videos/%s", libraryId, videoId), &v)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

// CreateVideo creates a new video record in a library.
func (c *Client) CreateVideo(ctx context.Context, libraryId int64, body *VideoCreate) (*Video, error) {
	var v Video
	err := c.Post(ctx, fmt.Sprintf("/library/%d/videos", libraryId), body, &v)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

// UpdateVideo updates an existing video. Note: bunny.net uses POST for updates.
func (c *Client) UpdateVideo(ctx context.Context, libraryId int64, videoId string, body *VideoUpdate) error {
	return c.Post(ctx, fmt.Sprintf("/library/%d/videos/%s", libraryId, videoId), body, nil)
}

// DeleteVideo deletes a video.
func (c *Client) DeleteVideo(ctx context.Context, libraryId int64, videoId string) error {
	return c.Delete(ctx, fmt.Sprintf("/library/%d/videos/%s", libraryId, videoId))
}

// UploadVideo uploads a binary video file to an existing video record.
func (c *Client) UploadVideo(ctx context.Context, libraryId int64, videoId string, body io.Reader, size int64) error {
	path := fmt.Sprintf("/library/%d/videos/%s", libraryId, videoId)

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.baseURL+path, body)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("AccessKey", c.apiKey)
	req.Header.Set("Content-Type", "application/octet-stream")
	if size > 0 {
		req.ContentLength = size
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

// FetchVideo creates and imports a video from a remote URL.
func (c *Client) FetchVideo(ctx context.Context, libraryId int64, body *VideoFetch, collectionId string, thumbnailTime int) (*StatusModel, error) {
	path := fmt.Sprintf("/library/%d/videos/fetch", libraryId)
	params := url.Values{}
	if collectionId != "" {
		params.Set("collectionId", collectionId)
	}
	if thumbnailTime > 0 {
		params.Set("thumbnailTime", fmt.Sprintf("%d", thumbnailTime))
	}
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	var status StatusModel
	err := c.Post(ctx, path, body, &status)
	if err != nil {
		return nil, err
	}
	return &status, nil
}

// ReencodeVideo re-encodes a video from the original file.
func (c *Client) ReencodeVideo(ctx context.Context, libraryId int64, videoId string) (*Video, error) {
	var v Video
	err := c.Post(ctx, fmt.Sprintf("/library/%d/videos/%s/reencode", libraryId, videoId), nil, &v)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

// TranscribeVideo triggers transcription for a video.
func (c *Client) TranscribeVideo(ctx context.Context, libraryId int64, videoId string, settings *TranscribeSettings) error {
	return c.Post(ctx, fmt.Sprintf("/library/%d/videos/%s/transcribe", libraryId, videoId), settings, nil)
}

// --- Collection client methods ---

// ListCollections returns a paginated list of collections in a library.
func (c *Client) ListCollections(ctx context.Context, libraryId int64, page, itemsPerPage int, search, orderBy string) (pagination.PageResponse[*Collection], error) {
	params := url.Values{}
	params.Set("page", fmt.Sprintf("%d", page))
	params.Set("itemsPerPage", fmt.Sprintf("%d", itemsPerPage))
	if search != "" {
		params.Set("search", search)
	}
	if orderBy != "" {
		params.Set("orderBy", orderBy)
	}

	path := fmt.Sprintf("/library/%d/collections?%s", libraryId, params.Encode())
	var resp streamPageResponse[*Collection]
	if err := c.Get(ctx, path, &resp); err != nil {
		return pagination.PageResponse[*Collection]{}, err
	}
	return resp.toPageResponse(), nil
}

// GetCollection returns a single collection by library ID and collection GUID.
func (c *Client) GetCollection(ctx context.Context, libraryId int64, collectionId string) (*Collection, error) {
	var col Collection
	err := c.Get(ctx, fmt.Sprintf("/library/%d/collections/%s", libraryId, collectionId), &col)
	if err != nil {
		return nil, err
	}
	return &col, nil
}

// CreateCollection creates a new collection in a library.
func (c *Client) CreateCollection(ctx context.Context, libraryId int64, body *CollectionCreate) (*Collection, error) {
	var col Collection
	err := c.Post(ctx, fmt.Sprintf("/library/%d/collections", libraryId), body, &col)
	if err != nil {
		return nil, err
	}
	return &col, nil
}

// UpdateCollection updates a collection. Note: bunny.net uses POST for updates.
func (c *Client) UpdateCollection(ctx context.Context, libraryId int64, collectionId string, body *CollectionUpdate) error {
	return c.Post(ctx, fmt.Sprintf("/library/%d/collections/%s", libraryId, collectionId), body, nil)
}

// DeleteCollection deletes a collection.
func (c *Client) DeleteCollection(ctx context.Context, libraryId int64, collectionId string) error {
	return c.Delete(ctx, fmt.Sprintf("/library/%d/collections/%s", libraryId, collectionId))
}

// --- Caption client methods ---

// AddCaption adds a caption track to a video.
func (c *Client) AddCaption(ctx context.Context, libraryId int64, videoId, srclang string, body *CaptionAdd) error {
	return c.Post(ctx, fmt.Sprintf("/library/%d/videos/%s/captions/%s", libraryId, videoId, srclang), body, nil)
}

// DeleteCaption removes a caption track from a video.
func (c *Client) DeleteCaption(ctx context.Context, libraryId int64, videoId, srclang string) error {
	return c.Delete(ctx, fmt.Sprintf("/library/%d/videos/%s/captions/%s", libraryId, videoId, srclang))
}

// --- Statistics client methods ---

// GetVideoStatistics returns statistics for a video library.
func (c *Client) GetVideoStatistics(ctx context.Context, libraryId int64, dateFrom, dateTo string, hourly bool, videoGuid string) (*VideoStatistics, error) {
	params := url.Values{}
	if dateFrom != "" {
		params.Set("dateFrom", dateFrom)
	}
	if dateTo != "" {
		params.Set("dateTo", dateTo)
	}
	if hourly {
		params.Set("hourly", "true")
	}
	if videoGuid != "" {
		params.Set("videoGuid", videoGuid)
	}

	path := fmt.Sprintf("/library/%d/statistics", libraryId)
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	var stats VideoStatistics
	if err := c.Get(ctx, path, &stats); err != nil {
		return nil, err
	}
	return &stats, nil
}

// GetVideoHeatmap returns the heatmap data for a video.
func (c *Client) GetVideoHeatmap(ctx context.Context, libraryId int64, videoId string) (*VideoHeatmap, error) {
	var hm VideoHeatmap
	err := c.Get(ctx, fmt.Sprintf("/library/%d/videos/%s/heatmap", libraryId, videoId), &hm)
	if err != nil {
		return nil, err
	}
	return &hm, nil
}
