package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/built-fast/bunny-cli/internal/pagination"
)

// --- Edge Script models ---

// EdgeScript represents a bunny.net edge script (compute).
type EdgeScript struct {
	Id                  int64                `json:"Id"`
	Name                string               `json:"Name"`
	LastModified        string               `json:"LastModified"`
	ScriptType          int                  `json:"ScriptType"` // 0=DNS, 1=CDN, 2=Middleware
	CurrentReleaseId    int64                `json:"CurrentReleaseId"`
	EdgeScriptVariables []EdgeScriptVariable `json:"EdgeScriptVariables"`
	Deleted             bool                 `json:"Deleted"`
	LinkedPullZones     []LinkedPullZone     `json:"LinkedPullZones"`
	DefaultHostname     string               `json:"DefaultHostname"`
	SystemHostname      string               `json:"SystemHostname"`
	DeploymentKey       string               `json:"DeploymentKey"`
	MonthlyCost         float64              `json:"MonthlyCost"`
	MonthlyRequestCount int64                `json:"MonthlyRequestCount"`
	MonthlyCpuTime      int64                `json:"MonthlyCpuTime"`
}

// LinkedPullZone is an abbreviated pull zone reference embedded in an edge script.
type LinkedPullZone struct {
	Id              int64  `json:"Id"`
	PullZoneName    string `json:"PullZoneName"`
	DefaultHostname string `json:"DefaultHostname"`
}

// EdgeScriptCreate holds the fields for creating an edge script.
type EdgeScriptCreate struct {
	Name                 string `json:"Name"`
	ScriptType           int    `json:"ScriptType"`
	Code                 string `json:"Code,omitempty"`
	CreateLinkedPullZone bool   `json:"CreateLinkedPullZone,omitempty"`
	LinkedPullZoneName   string `json:"LinkedPullZoneName,omitempty"`
}

// EdgeScriptUpdate holds the fields for updating an edge script.
type EdgeScriptUpdate struct {
	Name       *string `json:"Name,omitempty"`
	ScriptType *int    `json:"ScriptType,omitempty"`
}

// --- Variable models ---

// EdgeScriptVariable represents a variable attached to an edge script.
type EdgeScriptVariable struct {
	Id           int64  `json:"Id"`
	Name         string `json:"Name"`
	Required     bool   `json:"Required"`
	DefaultValue string `json:"DefaultValue"`
}

// EdgeScriptVariableCreate holds the fields for adding a variable.
type EdgeScriptVariableCreate struct {
	Name         string `json:"Name"`
	Required     bool   `json:"Required,omitempty"`
	DefaultValue string `json:"DefaultValue,omitempty"`
}

// EdgeScriptVariableUpdate holds the fields for updating a variable.
type EdgeScriptVariableUpdate struct {
	DefaultValue *string `json:"DefaultValue,omitempty"`
	Required     *bool   `json:"Required,omitempty"`
}

// --- Secret models ---

// EdgeScriptSecret represents a secret attached to an edge script.
type EdgeScriptSecret struct {
	Id           int64  `json:"Id"`
	Name         string `json:"Name"`
	LastModified string `json:"LastModified"`
}

// EdgeScriptSecretCreate holds the fields for adding a secret.
type EdgeScriptSecretCreate struct {
	Name   string `json:"Name"`
	Secret string `json:"Secret"`
}

// EdgeScriptSecretUpdate holds the fields for updating a secret.
type EdgeScriptSecretUpdate struct {
	Secret string `json:"Secret"`
}

// edgeScriptSecretsResponse wraps the list secrets response.
type edgeScriptSecretsResponse struct {
	Secrets []*EdgeScriptSecret `json:"Secrets"`
}

// --- Code models ---

// EdgeScriptCode represents the code content of an edge script.
type EdgeScriptCode struct {
	Code         string `json:"Code"`
	LastModified string `json:"LastModified"`
}

// edgeScriptCodeSet holds the body for setting script code.
type edgeScriptCodeSet struct {
	Code string `json:"Code"`
}

// --- Release models ---

// EdgeScriptRelease represents a release of an edge script.
type EdgeScriptRelease struct {
	Id            int64  `json:"Id"`
	Deleted       bool   `json:"Deleted"`
	Code          string `json:"Code"`
	Uuid          string `json:"Uuid"`
	Note          string `json:"Note"`
	Author        string `json:"Author"`
	AuthorEmail   string `json:"AuthorEmail"`
	CommitSha     string `json:"CommitSha"`
	Status        int    `json:"Status"` // 0=Archived, 1=Live
	DateReleased  string `json:"DateReleased"`
	DatePublished string `json:"DatePublished"`
}

// EdgeScriptPublish holds the body for publishing an edge script release.
type EdgeScriptPublish struct {
	Note string `json:"Note,omitempty"`
}

// --- Statistics models ---

// EdgeScriptStatistics holds statistics for an edge script.
type EdgeScriptStatistics struct {
	TotalRequestsServed        int64              `json:"TotalRequestsServed"`
	TotalCpuUsed               float64            `json:"TotalCpuUsed"`
	TotalMonthlyCost           float64            `json:"TotalMonthlyCost"`
	AverageCpuTimePerExecution float64            `json:"AverageCpuTimePerExecution"`
	RequestsServedChart        map[string]int64   `json:"RequestsServedChart"`
	AverageCpuTimeChart        map[string]float64 `json:"AverageCpuTimeChart"`
	TotalCpuTimeChart          map[string]int64   `json:"TotalCpuTimeChart"`
}

// --- Enum helpers ---

// scriptTypeNames maps script type integers to their string names.
var scriptTypeNames = map[int]string{
	0: "DNS",
	1: "CDN",
	2: "Middleware",
}

// ScriptTypeName returns a human-readable name for a script type.
func ScriptTypeName(t int) string {
	if name, ok := scriptTypeNames[t]; ok {
		return name
	}
	return fmt.Sprintf("Unknown(%d)", t)
}

// ScriptTypeFromName converts a script type name (case-insensitive) to its integer value.
func ScriptTypeFromName(name string) (int, error) {
	upper := strings.ToUpper(strings.TrimSpace(name))
	for k, v := range scriptTypeNames {
		if strings.ToUpper(v) == upper {
			return k, nil
		}
	}
	return 0, fmt.Errorf("unknown script type: %q (valid: DNS, CDN, Middleware)", name)
}

// ReleaseStatusName returns a human-readable name for a release status.
func ReleaseStatusName(s int) string {
	switch s {
	case 0:
		return "Archived"
	case 1:
		return "Live"
	default:
		return fmt.Sprintf("Unknown(%d)", s)
	}
}

// --- Edge Script client methods ---

// ListEdgeScripts returns a paginated list of edge scripts.
func (c *Client) ListEdgeScripts(ctx context.Context, page, perPage int, search string, scriptTypes []int) (pagination.PageResponse[*EdgeScript], error) {
	if perPage < 5 {
		perPage = 5
	}
	params := url.Values{}
	params.Set("page", fmt.Sprintf("%d", page))
	params.Set("perPage", fmt.Sprintf("%d", perPage))
	if search != "" {
		params.Set("search", search)
	}
	for _, t := range scriptTypes {
		params.Add("type", fmt.Sprintf("%d", t))
	}

	path := "/compute/script?" + params.Encode()

	var raw json.RawMessage
	if err := c.Get(ctx, path, &raw); err != nil {
		return pagination.PageResponse[*EdgeScript]{}, err
	}

	var resp pagination.PageResponse[*EdgeScript]
	if err := json.Unmarshal(raw, &resp); err == nil && len(raw) > 0 && raw[0] == '{' {
		return resp, nil
	}

	var items []*EdgeScript
	if err := json.Unmarshal(raw, &items); err != nil {
		return pagination.PageResponse[*EdgeScript]{}, fmt.Errorf("decoding edge script list: %w", err)
	}
	return pagination.PageResponse[*EdgeScript]{
		Items:        items,
		CurrentPage:  page,
		TotalItems:   len(items),
		HasMoreItems: false,
	}, nil
}

// GetEdgeScript returns a single edge script by ID.
func (c *Client) GetEdgeScript(ctx context.Context, id int64) (*EdgeScript, error) {
	var s EdgeScript
	err := c.Get(ctx, fmt.Sprintf("/compute/script/%d", id), &s)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// CreateEdgeScript creates a new edge script.
func (c *Client) CreateEdgeScript(ctx context.Context, body *EdgeScriptCreate) (*EdgeScript, error) {
	var s EdgeScript
	err := c.Post(ctx, "/compute/script", body, &s)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// UpdateEdgeScript updates an existing edge script. Note: bunny.net uses POST for updates.
func (c *Client) UpdateEdgeScript(ctx context.Context, id int64, body *EdgeScriptUpdate) (*EdgeScript, error) {
	var s EdgeScript
	err := c.Post(ctx, fmt.Sprintf("/compute/script/%d", id), body, &s)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// DeleteEdgeScript deletes an edge script by ID.
func (c *Client) DeleteEdgeScript(ctx context.Context, id int64, deleteLinkedPullZones bool) error {
	path := fmt.Sprintf("/compute/script/%d", id)
	if deleteLinkedPullZones {
		path += "?deleteLinkedPullZones=true"
	}
	return c.Do(ctx, http.MethodDelete, path, nil, nil)
}

// GetEdgeScriptStatistics returns statistics for an edge script.
func (c *Client) GetEdgeScriptStatistics(ctx context.Context, id int64, dateFrom, dateTo string, loadLatest, hourly bool) (*EdgeScriptStatistics, error) {
	params := url.Values{}
	if dateFrom != "" {
		params.Set("dateFrom", dateFrom)
	}
	if dateTo != "" {
		params.Set("dateTo", dateTo)
	}
	if loadLatest {
		params.Set("loadLatest", "true")
	}
	if hourly {
		params.Set("hourly", "true")
	}

	path := fmt.Sprintf("/compute/script/%d/statistics", id)
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	var stats EdgeScriptStatistics
	if err := c.Get(ctx, path, &stats); err != nil {
		return nil, err
	}
	return &stats, nil
}

// RotateEdgeScriptDeploymentKey rotates the deployment key for an edge script.
func (c *Client) RotateEdgeScriptDeploymentKey(ctx context.Context, id int64) error {
	return c.Post(ctx, fmt.Sprintf("/compute/script/%d/deploymentKey/rotate", id), nil, nil)
}

// --- Code client methods ---

// GetEdgeScriptCode returns the code content of an edge script.
func (c *Client) GetEdgeScriptCode(ctx context.Context, id int64) (*EdgeScriptCode, error) {
	var code EdgeScriptCode
	err := c.Get(ctx, fmt.Sprintf("/compute/script/%d/code", id), &code)
	if err != nil {
		return nil, err
	}
	return &code, nil
}

// SetEdgeScriptCode sets the code content of an edge script.
func (c *Client) SetEdgeScriptCode(ctx context.Context, id int64, code string) error {
	return c.Post(ctx, fmt.Sprintf("/compute/script/%d/code", id), &edgeScriptCodeSet{Code: code}, nil)
}

// --- Variable client methods ---

// AddEdgeScriptVariable adds a variable to an edge script.
func (c *Client) AddEdgeScriptVariable(ctx context.Context, scriptId int64, body *EdgeScriptVariableCreate) (*EdgeScriptVariable, error) {
	var v EdgeScriptVariable
	err := c.Post(ctx, fmt.Sprintf("/compute/script/%d/variables/add", scriptId), body, &v)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

// GetEdgeScriptVariable returns a single variable by script and variable ID.
func (c *Client) GetEdgeScriptVariable(ctx context.Context, scriptId, variableId int64) (*EdgeScriptVariable, error) {
	var v EdgeScriptVariable
	err := c.Get(ctx, fmt.Sprintf("/compute/script/%d/variables/%d", scriptId, variableId), &v)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

// UpdateEdgeScriptVariable updates a variable on an edge script.
func (c *Client) UpdateEdgeScriptVariable(ctx context.Context, scriptId, variableId int64, body *EdgeScriptVariableUpdate) error {
	return c.Post(ctx, fmt.Sprintf("/compute/script/%d/variables/%d", scriptId, variableId), body, nil)
}

// DeleteEdgeScriptVariable deletes a variable from an edge script.
func (c *Client) DeleteEdgeScriptVariable(ctx context.Context, scriptId, variableId int64) error {
	return c.Delete(ctx, fmt.Sprintf("/compute/script/%d/variables/%d", scriptId, variableId))
}

// --- Secret client methods ---

// AddEdgeScriptSecret adds a secret to an edge script.
func (c *Client) AddEdgeScriptSecret(ctx context.Context, scriptId int64, body *EdgeScriptSecretCreate) (*EdgeScriptSecret, error) {
	var s EdgeScriptSecret
	err := c.Post(ctx, fmt.Sprintf("/compute/script/%d/secrets", scriptId), body, &s)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// ListEdgeScriptSecrets returns all secrets for an edge script.
func (c *Client) ListEdgeScriptSecrets(ctx context.Context, scriptId int64) ([]*EdgeScriptSecret, error) {
	var resp edgeScriptSecretsResponse
	err := c.Get(ctx, fmt.Sprintf("/compute/script/%d/secrets", scriptId), &resp)
	if err != nil {
		return nil, err
	}
	return resp.Secrets, nil
}

// UpdateEdgeScriptSecret updates a secret on an edge script.
func (c *Client) UpdateEdgeScriptSecret(ctx context.Context, scriptId, secretId int64, body *EdgeScriptSecretUpdate) error {
	return c.Post(ctx, fmt.Sprintf("/compute/script/%d/secrets/%d", scriptId, secretId), body, nil)
}

// DeleteEdgeScriptSecret deletes a secret from an edge script.
func (c *Client) DeleteEdgeScriptSecret(ctx context.Context, scriptId, secretId int64) error {
	return c.Delete(ctx, fmt.Sprintf("/compute/script/%d/secrets/%d", scriptId, secretId))
}

// --- Release client methods ---

// ListEdgeScriptReleases returns a paginated list of releases for an edge script.
func (c *Client) ListEdgeScriptReleases(ctx context.Context, scriptId int64, page, perPage int) (pagination.PageResponse[*EdgeScriptRelease], error) {
	if perPage < 5 {
		perPage = 5
	}
	params := url.Values{}
	params.Set("page", fmt.Sprintf("%d", page))
	params.Set("perPage", fmt.Sprintf("%d", perPage))

	path := fmt.Sprintf("/compute/script/%d/releases?%s", scriptId, params.Encode())

	var raw json.RawMessage
	if err := c.Get(ctx, path, &raw); err != nil {
		return pagination.PageResponse[*EdgeScriptRelease]{}, err
	}

	var resp pagination.PageResponse[*EdgeScriptRelease]
	if err := json.Unmarshal(raw, &resp); err == nil && len(raw) > 0 && raw[0] == '{' {
		return resp, nil
	}

	var items []*EdgeScriptRelease
	if err := json.Unmarshal(raw, &items); err != nil {
		return pagination.PageResponse[*EdgeScriptRelease]{}, fmt.Errorf("decoding edge script release list: %w", err)
	}
	return pagination.PageResponse[*EdgeScriptRelease]{
		Items:        items,
		CurrentPage:  page,
		TotalItems:   len(items),
		HasMoreItems: false,
	}, nil
}

// GetActiveEdgeScriptRelease returns the currently active release for an edge script.
func (c *Client) GetActiveEdgeScriptRelease(ctx context.Context, scriptId int64) (*EdgeScriptRelease, error) {
	var r EdgeScriptRelease
	err := c.Get(ctx, fmt.Sprintf("/compute/script/%d/releases/active", scriptId), &r)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// --- Publish client methods ---

// PublishEdgeScript publishes the current code of an edge script as a new release.
func (c *Client) PublishEdgeScript(ctx context.Context, scriptId int64, body *EdgeScriptPublish) error {
	return c.Post(ctx, fmt.Sprintf("/compute/script/%d/publish", scriptId), body, nil)
}

// PublishEdgeScriptRelease publishes a specific release by UUID.
func (c *Client) PublishEdgeScriptRelease(ctx context.Context, scriptId int64, uuid string) error {
	return c.Post(ctx, fmt.Sprintf("/compute/script/%d/publish/%s", scriptId, uuid), nil, nil)
}
