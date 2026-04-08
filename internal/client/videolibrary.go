package client

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/built-fast/bunny-cli/internal/pagination"
)

// VideoLibrary represents a bunny.net video library.
type VideoLibrary struct {
	Id                 int64    `json:"Id"`
	Name               string   `json:"Name"`
	VideoCount         int64    `json:"VideoCount"`
	TrafficUsage       int64    `json:"TrafficUsage"`
	StorageUsage       int64    `json:"StorageUsage"`
	DateCreated        string   `json:"DateCreated"`
	DateModified       string   `json:"DateModified"`
	ApiKey             string   `json:"ApiKey"`
	ReadOnlyApiKey     string   `json:"ReadOnlyApiKey"`
	PullZoneId         int64    `json:"PullZoneId"`
	StorageZoneId      int64    `json:"StorageZoneId"`
	EnabledResolutions string   `json:"EnabledResolutions"`
	EnableMP4Fallback  bool     `json:"EnableMP4Fallback"`
	KeepOriginalFiles  bool     `json:"KeepOriginalFiles"`
	EnableDRM          bool     `json:"EnableDRM"`
	AllowDirectPlay    bool     `json:"AllowDirectPlay"`
	EnableTranscribing bool     `json:"EnableContentTagging"`
	ReplicationRegions []string `json:"ReplicationRegions"`
	WebhookUrl         string   `json:"WebhookUrl"`
	HasWatermark       bool     `json:"HasWatermark"`
}

// VideoLibraryCreate holds the fields for creating a video library.
type VideoLibraryCreate struct {
	Name               string   `json:"Name"`
	ReplicationRegions []string `json:"ReplicationRegions,omitempty"`
}

// VideoLibraryUpdate holds the fields for updating a video library.
// Pointer types allow distinguishing between "not set" and "set to zero value".
type VideoLibraryUpdate struct {
	Name               *string  `json:"Name,omitempty"`
	EnabledResolutions *string  `json:"EnabledResolutions,omitempty"`
	EnableMP4Fallback  *bool    `json:"EnableMP4Fallback,omitempty"`
	KeepOriginalFiles  *bool    `json:"KeepOriginalFiles,omitempty"`
	AllowDirectPlay    *bool    `json:"AllowDirectPlay,omitempty"`
	EnableDRM          *bool    `json:"EnableDRM,omitempty"`
	WebhookUrl         *string  `json:"WebhookUrl,omitempty"`
	ReplicationRegions []string `json:"ReplicationRegions,omitempty"`
}

// VideoLibraryLanguage represents a supported transcription language.
type VideoLibraryLanguage struct {
	ShortCode       string `json:"ShortCode"`
	Name            string `json:"Name"`
	SupportLevel    int    `json:"SupportLevel"`
	TranslateFromEn bool   `json:"TranslateFromEn"`
}

// ListVideoLibraries returns a paginated list of video libraries.
func (c *Client) ListVideoLibraries(ctx context.Context, page, perPage int, search string) (pagination.PageResponse[*VideoLibrary], error) {
	if perPage < 5 {
		perPage = 5
	}
	path := fmt.Sprintf("/videolibrary?page=%d&perPage=%d", page, perPage)
	if search != "" {
		path += "&search=" + search
	}

	var raw json.RawMessage
	if err := c.Get(ctx, path, &raw); err != nil {
		return pagination.PageResponse[*VideoLibrary]{}, err
	}

	// Try paginated object first
	var resp pagination.PageResponse[*VideoLibrary]
	if err := json.Unmarshal(raw, &resp); err == nil && len(raw) > 0 && raw[0] == '{' {
		return resp, nil
	}

	// Fall back to plain array
	var items []*VideoLibrary
	if err := json.Unmarshal(raw, &items); err != nil {
		return pagination.PageResponse[*VideoLibrary]{}, fmt.Errorf("decoding video library list: %w", err)
	}
	return pagination.PageResponse[*VideoLibrary]{
		Items:        items,
		CurrentPage:  page,
		TotalItems:   len(items),
		HasMoreItems: false,
	}, nil
}

// GetVideoLibrary returns a single video library by ID.
func (c *Client) GetVideoLibrary(ctx context.Context, id int64) (*VideoLibrary, error) {
	var lib VideoLibrary
	err := c.Get(ctx, fmt.Sprintf("/videolibrary/%d", id), &lib)
	if err != nil {
		return nil, err
	}
	return &lib, nil
}

// CreateVideoLibrary creates a new video library.
func (c *Client) CreateVideoLibrary(ctx context.Context, body *VideoLibraryCreate) (*VideoLibrary, error) {
	var lib VideoLibrary
	err := c.Post(ctx, "/videolibrary", body, &lib)
	if err != nil {
		return nil, err
	}
	return &lib, nil
}

// UpdateVideoLibrary updates an existing video library. Note: bunny.net uses POST for updates.
func (c *Client) UpdateVideoLibrary(ctx context.Context, id int64, body *VideoLibraryUpdate) (*VideoLibrary, error) {
	var lib VideoLibrary
	err := c.Post(ctx, fmt.Sprintf("/videolibrary/%d", id), body, &lib)
	if err != nil {
		return nil, err
	}
	return &lib, nil
}

// DeleteVideoLibrary deletes a video library by ID.
func (c *Client) DeleteVideoLibrary(ctx context.Context, id int64) error {
	return c.Delete(ctx, fmt.Sprintf("/videolibrary/%d", id))
}

// ResetVideoLibraryApiKey resets the API key for a video library.
func (c *Client) ResetVideoLibraryApiKey(ctx context.Context, id int64) error {
	return c.Post(ctx, fmt.Sprintf("/videolibrary/%d/resetApiKey", id), nil, nil)
}

// ListVideoLibraryLanguages returns the list of supported transcription languages.
func (c *Client) ListVideoLibraryLanguages(ctx context.Context) ([]VideoLibraryLanguage, error) {
	var langs []VideoLibraryLanguage
	err := c.Get(ctx, "/videolibrary/languages", &langs)
	if err != nil {
		return nil, err
	}
	return langs, nil
}
