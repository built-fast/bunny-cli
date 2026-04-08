package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/built-fast/bunny-cli/internal/pagination"
)

// PullZone represents a bunny.net pull zone.
type PullZone struct {
	Id                    int64      `json:"Id"`
	Name                  string     `json:"Name"`
	OriginUrl             string     `json:"OriginUrl"`
	Enabled               bool       `json:"Enabled"`
	Suspended             bool       `json:"Suspended"`
	CnameDomain           string     `json:"CnameDomain"`
	StorageZoneId         int64      `json:"StorageZoneId"`
	MonthlyBandwidthUsed  int64      `json:"MonthlyBandwidthUsed"`
	MonthlyCharges        float64    `json:"MonthlyCharges"`
	Type                  int        `json:"Type"` // 0=Premium, 1=Volume
	Hostnames             []Hostname `json:"Hostnames"`
	EdgeRules             []EdgeRule `json:"EdgeRules"`
	IgnoreQueryStrings    bool       `json:"IgnoreQueryStrings"`
	EnableGeoZoneUS       bool       `json:"EnableGeoZoneUS"`
	EnableGeoZoneEU       bool       `json:"EnableGeoZoneEU"`
	EnableGeoZoneASIA     bool       `json:"EnableGeoZoneASIA"`
	EnableGeoZoneSA       bool       `json:"EnableGeoZoneSA"`
	EnableGeoZoneAF       bool       `json:"EnableGeoZoneAF"`
	ZoneSecurityEnabled   bool       `json:"ZoneSecurityEnabled"`
	OriginHostHeader      string     `json:"OriginHostHeader"`
	AddHostHeader         bool       `json:"AddHostHeader"`
	VerifyOriginSSL       bool       `json:"VerifyOriginSSL"`
	EnableLogging         bool       `json:"EnableLogging"`
	MonthlyBandwidthLimit int64      `json:"MonthlyBandwidthLimit"`
	EnableOriginShield    bool       `json:"EnableOriginShield"`
	FollowRedirects       bool       `json:"FollowRedirects"`
	DisableCookies        bool       `json:"DisableCookies"`
}

// PullZoneTypeName returns a human-readable name for the pull zone type.
func PullZoneTypeName(t int) string {
	switch t {
	case 0:
		return "Premium"
	case 1:
		return "Volume"
	default:
		return fmt.Sprintf("Unknown(%d)", t)
	}
}

// Hostname represents a hostname attached to a pull zone.
type Hostname struct {
	Id               int64  `json:"Id"`
	Value            string `json:"Value"`
	ForceSSL         bool   `json:"ForceSSL"`
	IsSystemHostname bool   `json:"IsSystemHostname"`
	HasCertificate   bool   `json:"HasCertificate"`
}

// EdgeRule represents an edge rule on a pull zone.
type EdgeRule struct {
	Guid                string            `json:"Guid"`
	ActionType          int               `json:"ActionType"`
	ActionParameter1    string            `json:"ActionParameter1"`
	ActionParameter2    string            `json:"ActionParameter2"`
	Triggers            []EdgeRuleTrigger `json:"Triggers"`
	TriggerMatchingType int               `json:"TriggerMatchingType"`
	Description         string            `json:"Description"`
	Enabled             bool              `json:"Enabled"`
}

// EdgeRuleTrigger represents a trigger condition for an edge rule.
type EdgeRuleTrigger struct {
	Type                int      `json:"Type"`
	PatternMatches      []string `json:"PatternMatches"`
	PatternMatchingType int      `json:"PatternMatchingType"`
	Parameter1          string   `json:"Parameter1"`
}

// PullZoneCreate holds the fields for creating a pull zone.
type PullZoneCreate struct {
	Name      string `json:"Name"`
	OriginUrl string `json:"OriginUrl,omitempty"`
	Type      int    `json:"Type,omitempty"`
}

// PullZoneUpdate holds the fields for updating a pull zone.
// Pointer types allow distinguishing between "not set" and "set to zero value".
type PullZoneUpdate struct {
	OriginUrl             *string `json:"OriginUrl,omitempty"`
	OriginHostHeader      *string `json:"OriginHostHeader,omitempty"`
	AddHostHeader         *bool   `json:"AddHostHeader,omitempty"`
	VerifyOriginSSL       *bool   `json:"VerifyOriginSSL,omitempty"`
	EnableGeoZoneUS       *bool   `json:"EnableGeoZoneUS,omitempty"`
	EnableGeoZoneEU       *bool   `json:"EnableGeoZoneEU,omitempty"`
	EnableGeoZoneASIA     *bool   `json:"EnableGeoZoneASIA,omitempty"`
	EnableGeoZoneSA       *bool   `json:"EnableGeoZoneSA,omitempty"`
	EnableGeoZoneAF       *bool   `json:"EnableGeoZoneAF,omitempty"`
	IgnoreQueryStrings    *bool   `json:"IgnoreQueryStrings,omitempty"`
	ZoneSecurityEnabled   *bool   `json:"ZoneSecurityEnabled,omitempty"`
	EnableLogging         *bool   `json:"EnableLogging,omitempty"`
	MonthlyBandwidthLimit *int64  `json:"MonthlyBandwidthLimit,omitempty"`
	EnableOriginShield    *bool   `json:"EnableOriginShield,omitempty"`
	FollowRedirects       *bool   `json:"FollowRedirects,omitempty"`
	DisableCookies        *bool   `json:"DisableCookies,omitempty"`
}

// ListPullZones returns a paginated list of pull zones.
// The bunny.net API returns a paginated object when page > 0, but returns
// a plain array when page == 0. We handle both formats for compatibility
// with mock servers (e.g., Prism) that may return a plain array.
func (c *Client) ListPullZones(ctx context.Context, page, perPage int, search string) (pagination.PageResponse[*PullZone], error) {
	// bunny.net API requires perPage >= 5
	if perPage < 5 {
		perPage = 5
	}
	path := fmt.Sprintf("/pullzone?page=%d&perPage=%d", page, perPage)
	if search != "" {
		path += "&search=" + search
	}
	var raw json.RawMessage
	if err := c.Get(ctx, path, &raw); err != nil {
		return pagination.PageResponse[*PullZone]{}, err
	}

	// Try paginated object first (check if it has Items key)
	var resp pagination.PageResponse[*PullZone]
	if err := json.Unmarshal(raw, &resp); err == nil && len(raw) > 0 && raw[0] == '{' {
		return resp, nil
	}

	// Fall back to plain array
	var items []*PullZone
	if err := json.Unmarshal(raw, &items); err != nil {
		return pagination.PageResponse[*PullZone]{}, fmt.Errorf("decoding pull zone list: %w", err)
	}
	return pagination.PageResponse[*PullZone]{
		Items:        items,
		CurrentPage:  page,
		TotalItems:   len(items),
		HasMoreItems: false,
	}, nil
}

// GetPullZone returns a single pull zone by ID.
func (c *Client) GetPullZone(ctx context.Context, id int64) (*PullZone, error) {
	var pz PullZone
	err := c.Get(ctx, fmt.Sprintf("/pullzone/%d", id), &pz)
	if err != nil {
		return nil, err
	}
	return &pz, nil
}

// CreatePullZone creates a new pull zone.
func (c *Client) CreatePullZone(ctx context.Context, body *PullZoneCreate) (*PullZone, error) {
	var pz PullZone
	err := c.Post(ctx, "/pullzone", body, &pz)
	if err != nil {
		return nil, err
	}
	return &pz, nil
}

// UpdatePullZone updates an existing pull zone. Note: bunny.net uses POST for updates.
func (c *Client) UpdatePullZone(ctx context.Context, id int64, body *PullZoneUpdate) (*PullZone, error) {
	var pz PullZone
	err := c.Post(ctx, fmt.Sprintf("/pullzone/%d", id), body, &pz)
	if err != nil {
		return nil, err
	}
	return &pz, nil
}

// DeletePullZone deletes a pull zone by ID.
func (c *Client) DeletePullZone(ctx context.Context, id int64) error {
	return c.Delete(ctx, fmt.Sprintf("/pullzone/%d", id))
}

// AddPullZoneHostname adds a custom hostname to a pull zone.
func (c *Client) AddPullZoneHostname(ctx context.Context, id int64, hostname string) error {
	body := struct {
		Hostname string `json:"Hostname"`
	}{Hostname: hostname}
	return c.Post(ctx, fmt.Sprintf("/pullzone/%d/addHostname", id), &body, nil)
}

// RemovePullZoneHostname removes a custom hostname from a pull zone.
func (c *Client) RemovePullZoneHostname(ctx context.Context, id int64, hostname string) error {
	body := struct {
		Hostname string `json:"Hostname"`
	}{Hostname: hostname}
	return c.Do(ctx, http.MethodDelete, fmt.Sprintf("/pullzone/%d/removeHostname", id), &body, nil)
}

// PurgePullZoneCache purges the cache for a pull zone. If cacheTag is empty, purges all.
func (c *Client) PurgePullZoneCache(ctx context.Context, id int64, cacheTag string) error {
	var body any
	if cacheTag != "" {
		body = &struct {
			CacheTag string `json:"CacheTag"`
		}{CacheTag: cacheTag}
	}
	return c.Post(ctx, fmt.Sprintf("/pullzone/%d/purgeCache", id), body, nil)
}

// AddOrUpdateEdgeRule creates or updates an edge rule on a pull zone.
func (c *Client) AddOrUpdateEdgeRule(ctx context.Context, pullZoneId int64, rule *EdgeRule) error {
	return c.Post(ctx, fmt.Sprintf("/pullzone/%d/edgerules/addOrUpdate", pullZoneId), rule, nil)
}

// DeleteEdgeRule deletes an edge rule from a pull zone.
func (c *Client) DeleteEdgeRule(ctx context.Context, pullZoneId int64, edgeRuleId string) error {
	return c.Delete(ctx, fmt.Sprintf("/pullzone/%d/edgerules/%s", pullZoneId, edgeRuleId))
}

// SetEdgeRuleEnabled enables or disables an edge rule.
func (c *Client) SetEdgeRuleEnabled(ctx context.Context, pullZoneId int64, edgeRuleId string, enabled bool) error {
	body := struct {
		Id    int64 `json:"Id"`
		Value bool  `json:"Value"`
	}{Id: pullZoneId, Value: enabled}
	return c.Post(ctx, fmt.Sprintf("/pullzone/%d/edgerules/%s/setEdgeRuleEnabled", pullZoneId, edgeRuleId), &body, nil)
}
