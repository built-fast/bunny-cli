package client

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/built-fast/bunny-cli/internal/pagination"
)

// StorageZone represents a bunny.net storage zone.
type StorageZone struct {
	Id                          int64      `json:"Id"`
	Name                        string     `json:"Name"`
	Password                    string     `json:"Password"`
	ReadOnlyPassword            string     `json:"ReadOnlyPassword"`
	DateModified                string     `json:"DateModified"`
	Deleted                     bool       `json:"Deleted"`
	StorageUsed                 int64      `json:"StorageUsed"`
	FilesStored                 int64      `json:"FilesStored"`
	Region                      string     `json:"Region"`
	ReplicationRegions          []string   `json:"ReplicationRegions"`
	StorageHostname             string     `json:"StorageHostname"`
	ZoneTier                    int        `json:"ZoneTier"`
	ReplicationChangeInProgress bool       `json:"ReplicationChangeInProgress"`
	Custom404FilePath           string     `json:"Custom404FilePath"`
	Rewrite404To200             bool       `json:"Rewrite404To200"`
	PullZones                   []PullZone `json:"PullZones"`
}

// StorageZoneCreate holds the fields for creating a storage zone.
type StorageZoneCreate struct {
	Name               string   `json:"Name"`
	Region             string   `json:"Region"`
	ReplicationRegions []string `json:"ReplicationRegions,omitempty"`
	ZoneTier           int      `json:"ZoneTier,omitempty"`
}

// StorageZoneUpdate holds the fields for updating a storage zone.
// Pointer types allow distinguishing between "not set" and "set to zero value".
type StorageZoneUpdate struct {
	ReplicationZones  []string `json:"ReplicationZones,omitempty"`
	OriginUrl         *string  `json:"OriginUrl,omitempty"`
	Custom404FilePath *string  `json:"Custom404FilePath,omitempty"`
	Rewrite404To200   *bool    `json:"Rewrite404To200,omitempty"`
}

// StorageZoneTierName returns a human-readable name for the storage zone tier.
func StorageZoneTierName(tier int) string {
	switch tier {
	case 0:
		return "Standard"
	case 1:
		return "Edge"
	default:
		return fmt.Sprintf("Unknown(%d)", tier)
	}
}

// ListStorageZones returns a paginated list of storage zones.
// The bunny.net API returns a paginated object when page > 0, but returns
// a plain array when page == 0. We handle both formats for compatibility
// with mock servers (e.g., Prism) that may return a plain array.
func (c *Client) ListStorageZones(ctx context.Context, page, perPage int, search string, includeDeleted bool) (pagination.PageResponse[*StorageZone], error) {
	if perPage < 5 {
		perPage = 5
	}
	path := fmt.Sprintf("/storagezone?page=%d&perPage=%d", page, perPage)
	if search != "" {
		path += "&search=" + search
	}
	if includeDeleted {
		path += "&includeDeleted=true"
	}

	var raw json.RawMessage
	if err := c.Get(ctx, path, &raw); err != nil {
		return pagination.PageResponse[*StorageZone]{}, err
	}

	// Try paginated object first (check if it has Items key)
	var resp pagination.PageResponse[*StorageZone]
	if err := json.Unmarshal(raw, &resp); err == nil && len(raw) > 0 && raw[0] == '{' {
		return resp, nil
	}

	// Fall back to plain array
	var items []*StorageZone
	if err := json.Unmarshal(raw, &items); err != nil {
		return pagination.PageResponse[*StorageZone]{}, fmt.Errorf("decoding storage zone list: %w", err)
	}
	return pagination.PageResponse[*StorageZone]{
		Items:        items,
		CurrentPage:  page,
		TotalItems:   len(items),
		HasMoreItems: false,
	}, nil
}

// GetStorageZone returns a single storage zone by ID.
func (c *Client) GetStorageZone(ctx context.Context, id int64) (*StorageZone, error) {
	var sz StorageZone
	err := c.Get(ctx, fmt.Sprintf("/storagezone/%d", id), &sz)
	if err != nil {
		return nil, err
	}
	return &sz, nil
}

// CreateStorageZone creates a new storage zone.
func (c *Client) CreateStorageZone(ctx context.Context, body *StorageZoneCreate) (*StorageZone, error) {
	var sz StorageZone
	err := c.Post(ctx, "/storagezone", body, &sz)
	if err != nil {
		return nil, err
	}
	return &sz, nil
}

// UpdateStorageZone updates an existing storage zone.
// Note: bunny.net uses POST for updates and returns 204 (no body).
func (c *Client) UpdateStorageZone(ctx context.Context, id int64, body *StorageZoneUpdate) error {
	return c.Post(ctx, fmt.Sprintf("/storagezone/%d", id), body, nil)
}

// DeleteStorageZone deletes a storage zone by ID.
func (c *Client) DeleteStorageZone(ctx context.Context, id int64, deleteLinkedPullZones bool) error {
	return c.Delete(ctx, fmt.Sprintf("/storagezone/%d?deleteLinkedPullZones=%t", id, deleteLinkedPullZones))
}

// ResetStorageZonePassword resets the password for a storage zone.
func (c *Client) ResetStorageZonePassword(ctx context.Context, id int64) error {
	return c.Post(ctx, fmt.Sprintf("/storagezone/%d/resetPassword", id), nil, nil)
}

// ResetStorageZoneReadOnlyPassword resets the read-only password for a storage zone.
func (c *Client) ResetStorageZoneReadOnlyPassword(ctx context.Context, id int64) error {
	return c.Post(ctx, fmt.Sprintf("/storagezone/resetReadOnlyPassword?id=%d", id), nil, nil)
}

// FindStorageZoneByName searches for a storage zone by exact name match.
func (c *Client) FindStorageZoneByName(ctx context.Context, name string) (*StorageZone, error) {
	resp, err := c.ListStorageZones(ctx, 1, 1000, name, false)
	if err != nil {
		return nil, err
	}
	for _, sz := range resp.Items {
		if sz.Name == name {
			return sz, nil
		}
	}
	return nil, fmt.Errorf("storage zone %q not found", name)
}
