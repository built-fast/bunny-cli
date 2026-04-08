package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/built-fast/bunny-cli/internal/pagination"
)

// DnsZone represents a bunny.net DNS zone.
type DnsZone struct {
	Id                            int64       `json:"Id"`
	Domain                        string      `json:"Domain"`
	Records                       []DnsRecord `json:"Records"`
	DateModified                  string      `json:"DateModified"`
	DateCreated                   string      `json:"DateCreated"`
	NameserversDetected           bool        `json:"NameserversDetected"`
	CustomNameserversEnabled      bool        `json:"CustomNameserversEnabled"`
	Nameserver1                   string      `json:"Nameserver1"`
	Nameserver2                   string      `json:"Nameserver2"`
	SoaEmail                      string      `json:"SoaEmail"`
	NameserversNextCheck          string      `json:"NameserversNextCheck"`
	LoggingEnabled                bool        `json:"LoggingEnabled"`
	LoggingIPAnonymizationEnabled bool        `json:"LoggingIPAnonymizationEnabled"`
	LogAnonymizationType          int         `json:"LogAnonymizationType"`
	DnsSecEnabled                 bool        `json:"DnsSecEnabled"`
	CertificateKeyType            int         `json:"CertificateKeyType"`
}

// DnsRecord represents a DNS record within a zone.
type DnsRecord struct {
	Id                    int64   `json:"Id"`
	Type                  int     `json:"Type"`
	Ttl                   int     `json:"Ttl"`
	Value                 string  `json:"Value"`
	Name                  string  `json:"Name"`
	Weight                int     `json:"Weight"`
	Priority              int     `json:"Priority"`
	Port                  int     `json:"Port"`
	Flags                 int     `json:"Flags"`
	Tag                   string  `json:"Tag"`
	Accelerated           bool    `json:"Accelerated"`
	AcceleratedPullZoneId int64   `json:"AcceleratedPullZoneId"`
	LinkName              string  `json:"LinkName"`
	MonitorStatus         int     `json:"MonitorStatus"`
	MonitorType           int     `json:"MonitorType"`
	GeolocationLatitude   float64 `json:"GeolocationLatitude"`
	GeolocationLongitude  float64 `json:"GeolocationLongitude"`
	LatencyZone           string  `json:"LatencyZone"`
	SmartRoutingType      int     `json:"SmartRoutingType"`
	Disabled              bool    `json:"Disabled"`
	Comment               string  `json:"Comment"`
}

// DnsZoneCreate holds the fields for creating a DNS zone.
type DnsZoneCreate struct {
	Domain string `json:"Domain"`
}

// DnsZoneUpdate holds the fields for updating a DNS zone.
// Pointer types allow distinguishing between "not set" and "set to zero value".
type DnsZoneUpdate struct {
	CustomNameserversEnabled      *bool   `json:"CustomNameserversEnabled,omitempty"`
	Nameserver1                   *string `json:"Nameserver1,omitempty"`
	Nameserver2                   *string `json:"Nameserver2,omitempty"`
	SoaEmail                      *string `json:"SoaEmail,omitempty"`
	LoggingEnabled                *bool   `json:"LoggingEnabled,omitempty"`
	LogAnonymizationType          *int    `json:"LogAnonymizationType,omitempty"`
	LoggingIPAnonymizationEnabled *bool   `json:"LoggingIPAnonymizationEnabled,omitempty"`
	CertificateKeyType            *int    `json:"CertificateKeyType,omitempty"`
}

// DnsRecordCreate holds the fields for adding a DNS record.
type DnsRecordCreate struct {
	Type                 int     `json:"Type"`
	Ttl                  int     `json:"Ttl"`
	Value                string  `json:"Value"`
	Name                 string  `json:"Name,omitempty"`
	Weight               int     `json:"Weight,omitempty"`
	Priority             int     `json:"Priority,omitempty"`
	Port                 int     `json:"Port,omitempty"`
	Flags                int     `json:"Flags,omitempty"`
	Tag                  string  `json:"Tag,omitempty"`
	Accelerated          bool    `json:"Accelerated,omitempty"`
	PullZoneId           int64   `json:"PullZoneId,omitempty"`
	ScriptId             int64   `json:"ScriptId,omitempty"`
	MonitorType          int     `json:"MonitorType,omitempty"`
	GeolocationLatitude  float64 `json:"GeolocationLatitude,omitempty"`
	GeolocationLongitude float64 `json:"GeolocationLongitude,omitempty"`
	LatencyZone          string  `json:"LatencyZone,omitempty"`
	SmartRoutingType     int     `json:"SmartRoutingType,omitempty"`
	Disabled             bool    `json:"Disabled,omitempty"`
	Comment              string  `json:"Comment,omitempty"`
}

// DnsRecordUpdate holds the fields for updating a DNS record.
// Pointer types allow distinguishing between "not set" and "set to zero value".
type DnsRecordUpdate struct {
	Type                 *int     `json:"Type,omitempty"`
	Ttl                  *int     `json:"Ttl,omitempty"`
	Value                *string  `json:"Value,omitempty"`
	Name                 *string  `json:"Name,omitempty"`
	Weight               *int     `json:"Weight,omitempty"`
	Priority             *int     `json:"Priority,omitempty"`
	Port                 *int     `json:"Port,omitempty"`
	Flags                *int     `json:"Flags,omitempty"`
	Tag                  *string  `json:"Tag,omitempty"`
	Accelerated          *bool    `json:"Accelerated,omitempty"`
	PullZoneId           *int64   `json:"PullZoneId,omitempty"`
	ScriptId             *int64   `json:"ScriptId,omitempty"`
	MonitorType          *int     `json:"MonitorType,omitempty"`
	GeolocationLatitude  *float64 `json:"GeolocationLatitude,omitempty"`
	GeolocationLongitude *float64 `json:"GeolocationLongitude,omitempty"`
	LatencyZone          *string  `json:"LatencyZone,omitempty"`
	SmartRoutingType     *int     `json:"SmartRoutingType,omitempty"`
	Disabled             *bool    `json:"Disabled,omitempty"`
	Comment              *string  `json:"Comment,omitempty"`
}

// DnsSecInfo holds the DNSSEC DS record information returned when enabling/disabling DNSSEC.
type DnsSecInfo struct {
	Enabled      bool   `json:"Enabled"`
	DsRecord     string `json:"DsRecord"`
	Digest       string `json:"Digest"`
	DigestType   string `json:"DigestType"`
	Algorithm    int    `json:"Algorithm"`
	PublicKey    string `json:"PublicKey"`
	KeyTag       int    `json:"KeyTag"`
	Flags        int    `json:"Flags"`
	DsConfigured bool   `json:"DsConfigured"`
}

// DnsZoneImportResult holds the result of a DNS zone file import.
type DnsZoneImportResult struct {
	RecordsSuccessful int `json:"RecordsSuccessful"`
	RecordsFailed     int `json:"RecordsFailed"`
	RecordsSkipped    int `json:"RecordsSkipped"`
}

// dnsRecordTypeNames maps record type integers to their string names.
var dnsRecordTypeNames = map[int]string{
	0:  "A",
	1:  "AAAA",
	2:  "CNAME",
	3:  "TXT",
	4:  "MX",
	5:  "Redirect",
	6:  "Flatten",
	7:  "PullZone",
	8:  "SRV",
	9:  "CAA",
	10: "PTR",
	11: "Script",
	12: "NS",
}

// DnsRecordTypeName returns a human-readable name for a DNS record type.
func DnsRecordTypeName(t int) string {
	if name, ok := dnsRecordTypeNames[t]; ok {
		return name
	}
	return fmt.Sprintf("Unknown(%d)", t)
}

// DnsRecordTypeFromName converts a record type name (case-insensitive) to its integer value.
func DnsRecordTypeFromName(name string) (int, error) {
	upper := strings.ToUpper(strings.TrimSpace(name))
	for k, v := range dnsRecordTypeNames {
		if strings.ToUpper(v) == upper {
			return k, nil
		}
	}
	return 0, fmt.Errorf("unknown DNS record type: %q", name)
}

// LogAnonymizationTypeName returns a human-readable name for the log anonymization type.
func LogAnonymizationTypeName(t int) string {
	switch t {
	case 0:
		return "OneDigit"
	case 1:
		return "Drop"
	default:
		return fmt.Sprintf("Unknown(%d)", t)
	}
}

// CertificateKeyTypeName returns a human-readable name for the certificate key type.
func CertificateKeyTypeName(t int) string {
	switch t {
	case 0:
		return "ECDSA"
	case 1:
		return "RSA"
	default:
		return fmt.Sprintf("Unknown(%d)", t)
	}
}

// ListDnsZones returns a paginated list of DNS zones.
func (c *Client) ListDnsZones(ctx context.Context, page, perPage int, search string) (pagination.PageResponse[*DnsZone], error) {
	if perPage < 5 {
		perPage = 5
	}
	path := fmt.Sprintf("/dnszone?page=%d&perPage=%d", page, perPage)
	if search != "" {
		path += "&search=" + search
	}

	var raw json.RawMessage
	if err := c.Get(ctx, path, &raw); err != nil {
		return pagination.PageResponse[*DnsZone]{}, err
	}

	// Try paginated object first
	var resp pagination.PageResponse[*DnsZone]
	if err := json.Unmarshal(raw, &resp); err == nil && len(raw) > 0 && raw[0] == '{' {
		return resp, nil
	}

	// Fall back to plain array
	var items []*DnsZone
	if err := json.Unmarshal(raw, &items); err != nil {
		return pagination.PageResponse[*DnsZone]{}, fmt.Errorf("decoding DNS zone list: %w", err)
	}
	return pagination.PageResponse[*DnsZone]{
		Items:        items,
		CurrentPage:  page,
		TotalItems:   len(items),
		HasMoreItems: false,
	}, nil
}

// GetDnsZone returns a single DNS zone by ID.
func (c *Client) GetDnsZone(ctx context.Context, id int64) (*DnsZone, error) {
	var zone DnsZone
	err := c.Get(ctx, fmt.Sprintf("/dnszone/%d", id), &zone)
	if err != nil {
		return nil, err
	}
	return &zone, nil
}

// CreateDnsZone creates a new DNS zone.
func (c *Client) CreateDnsZone(ctx context.Context, body *DnsZoneCreate) (*DnsZone, error) {
	var zone DnsZone
	err := c.Post(ctx, "/dnszone", body, &zone)
	if err != nil {
		return nil, err
	}
	return &zone, nil
}

// UpdateDnsZone updates an existing DNS zone. Note: bunny.net uses POST for updates.
func (c *Client) UpdateDnsZone(ctx context.Context, id int64, body *DnsZoneUpdate) (*DnsZone, error) {
	var zone DnsZone
	err := c.Post(ctx, fmt.Sprintf("/dnszone/%d", id), body, &zone)
	if err != nil {
		return nil, err
	}
	return &zone, nil
}

// DeleteDnsZone deletes a DNS zone by ID.
func (c *Client) DeleteDnsZone(ctx context.Context, id int64) error {
	return c.Delete(ctx, fmt.Sprintf("/dnszone/%d", id))
}

// AddDnsRecord adds a DNS record to a zone. Note: bunny.net uses PUT for adding records.
func (c *Client) AddDnsRecord(ctx context.Context, zoneId int64, body *DnsRecordCreate) (*DnsRecord, error) {
	var record DnsRecord
	err := c.Put(ctx, fmt.Sprintf("/dnszone/%d/records", zoneId), body, &record)
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// UpdateDnsRecord updates a DNS record. Returns no body (204).
func (c *Client) UpdateDnsRecord(ctx context.Context, zoneId, recordId int64, body *DnsRecordUpdate) error {
	return c.Post(ctx, fmt.Sprintf("/dnszone/%d/records/%d", zoneId, recordId), body, nil)
}

// DeleteDnsRecord deletes a DNS record from a zone.
func (c *Client) DeleteDnsRecord(ctx context.Context, zoneId, recordId int64) error {
	return c.Delete(ctx, fmt.Sprintf("/dnszone/%d/records/%d", zoneId, recordId))
}

// ImportDnsZone imports DNS records from a zone file.
func (c *Client) ImportDnsZone(ctx context.Context, zoneId int64, data io.Reader) (*DnsZoneImportResult, error) {
	raw, err := c.DoRaw(ctx, http.MethodPost, fmt.Sprintf("/dnszone/%d/import", zoneId), "application/octet-stream", data)
	if err != nil {
		return nil, err
	}
	var result DnsZoneImportResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("decoding import result: %w", err)
	}
	return &result, nil
}

// ExportDnsZone exports a DNS zone as a zone file.
func (c *Client) ExportDnsZone(ctx context.Context, zoneId int64) ([]byte, error) {
	return c.DoRaw(ctx, http.MethodGet, fmt.Sprintf("/dnszone/%d/export", zoneId), "", nil)
}

// EnableDnsSec enables DNSSEC for a DNS zone and returns the DS record info.
func (c *Client) EnableDnsSec(ctx context.Context, zoneId int64) (*DnsSecInfo, error) {
	var info DnsSecInfo
	err := c.Post(ctx, fmt.Sprintf("/dnszone/%d/dnssec", zoneId), nil, &info)
	if err != nil {
		return nil, err
	}
	return &info, nil
}

// DisableDnsSec disables DNSSEC for a DNS zone.
func (c *Client) DisableDnsSec(ctx context.Context, zoneId int64) (*DnsSecInfo, error) {
	var info DnsSecInfo
	err := c.Do(ctx, http.MethodDelete, fmt.Sprintf("/dnszone/%d/dnssec", zoneId), nil, &info)
	if err != nil {
		return nil, err
	}
	return &info, nil
}
