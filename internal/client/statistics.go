package client

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// Statistics represents the bunny.net global statistics response.
type Statistics struct {
	TotalBandwidthUsed                     int64              `json:"TotalBandwidthUsed"`
	TotalOriginTraffic                     int64              `json:"TotalOriginTraffic"`
	AverageOriginResponseTime              int32              `json:"AverageOriginResponseTime"`
	TotalRequestsServed                    int64              `json:"TotalRequestsServed"`
	CacheHitRate                           float64            `json:"CacheHitRate"`
	BandwidthUsedChart                     map[string]float64 `json:"BandwidthUsedChart"`
	BandwidthCachedChart                   map[string]float64 `json:"BandwidthCachedChart"`
	CacheHitRateChart                      map[string]float64 `json:"CacheHitRateChart"`
	RequestsServedChart                    map[string]float64 `json:"RequestsServedChart"`
	OriginResponseTimeChart                map[string]float64 `json:"OriginResponseTimeChart"`
	OriginTrafficChart                     map[string]float64 `json:"OriginTrafficChart"`
	OriginShieldBandwidthUsedChart         map[string]float64 `json:"OriginShieldBandwidthUsedChart"`
	OriginShieldInternalBandwidthUsedChart map[string]float64 `json:"OriginShieldInternalBandwidthUsedChart"`
	PullRequestsPulledChart                map[string]float64 `json:"PullRequestsPulledChart"`
	UserBalanceHistoryChart                map[string]float64 `json:"UserBalanceHistoryChart"`
	GeoTrafficDistribution                 map[string]int64   `json:"GeoTrafficDistribution"`
	Error3xxChart                          map[string]float64 `json:"Error3xxChart"`
	Error4xxChart                          map[string]float64 `json:"Error4xxChart"`
	Error5xxChart                          map[string]float64 `json:"Error5xxChart"`
}

// StatisticsOptions configures the statistics query.
type StatisticsOptions struct {
	DateFrom                          string
	DateTo                            string
	PullZone                          int64
	ServerZoneId                      int64
	LoadErrors                        bool
	Hourly                            bool
	LoadOriginResponseTimes           bool
	LoadOriginTraffic                 bool
	LoadRequestsServed                bool
	LoadBandwidthUsed                 bool
	LoadOriginShieldBandwidth         bool
	LoadGeographicTrafficDistribution bool
	LoadUserBalanceHistory            bool
}

// GetStatistics returns global CDN statistics.
func (c *Client) GetStatistics(ctx context.Context, opts StatisticsOptions) (*Statistics, error) {
	params := url.Values{}
	if opts.DateFrom != "" {
		params.Set("dateFrom", opts.DateFrom)
	}
	if opts.DateTo != "" {
		params.Set("dateTo", opts.DateTo)
	}
	if opts.PullZone > 0 {
		params.Set("pullZone", strconv.FormatInt(opts.PullZone, 10))
	}
	if opts.ServerZoneId > 0 {
		params.Set("serverZoneId", strconv.FormatInt(opts.ServerZoneId, 10))
	}
	if opts.LoadErrors {
		params.Set("loadErrors", "true")
	}
	if opts.Hourly {
		params.Set("hourly", "true")
	}
	if opts.LoadOriginResponseTimes {
		params.Set("loadOriginResponseTimes", "true")
	}
	if opts.LoadOriginTraffic {
		params.Set("loadOriginTraffic", "true")
	}
	if opts.LoadRequestsServed {
		params.Set("loadRequestsServed", "true")
	}
	if opts.LoadBandwidthUsed {
		params.Set("loadBandwidthUsed", "true")
	}
	if opts.LoadOriginShieldBandwidth {
		params.Set("loadOriginShieldBandwidth", "true")
	}
	if opts.LoadGeographicTrafficDistribution {
		params.Set("loadGeographicTrafficDistribution", "true")
	}
	if opts.LoadUserBalanceHistory {
		params.Set("loadUserBalanceHistory", "true")
	}

	path := "/statistics"
	if q := params.Encode(); q != "" {
		path += "?" + q
	}

	var stats Statistics
	if err := c.Get(ctx, path, &stats); err != nil {
		return nil, err
	}
	return &stats, nil
}

// FormatBytes formats bytes into a human-readable string.
func FormatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}
