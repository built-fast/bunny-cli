package cmd

import (
	"context"

	"github.com/built-fast/bunny-cli/internal/client"
	"github.com/built-fast/bunny-cli/internal/pagination"
)

// ShieldAPI abstracts the bunny.net Shield (security) API methods,
// allowing tests to inject mocks without making real API calls.
type ShieldAPI interface {
	// Shield Zones
	ListShieldZones(ctx context.Context, page, perPage int) (pagination.PageResponse[*client.ShieldZone], error)
	GetShieldZone(ctx context.Context, id int64) (*client.ShieldZone, error)
	GetShieldZoneByPullZone(ctx context.Context, pullZoneId int64) (*client.ShieldZone, error)
	CreateShieldZone(ctx context.Context, body *client.ShieldZoneCreate) (*client.ShieldZoneResponse, error)
	UpdateShieldZone(ctx context.Context, id int64, body *client.ShieldZoneUpdate) (*client.ShieldZoneResponse, error)

	// WAF
	ListWafRules(ctx context.Context, shieldZoneId int64) ([]*client.WafRuleMainGroup, error)
	ListCustomWafRules(ctx context.Context, shieldZoneId int64, page, perPage int) (pagination.PageResponse[*client.CustomWafRule], error)
	GetCustomWafRule(ctx context.Context, id int64) (*client.CustomWafRule, error)
	CreateCustomWafRule(ctx context.Context, body *client.CustomWafRuleCreate) (*client.CustomWafRule, error)
	UpdateCustomWafRule(ctx context.Context, id int64, body *client.CustomWafRuleUpdate) (*client.CustomWafRule, error)
	DeleteCustomWafRule(ctx context.Context, id int64) error
	ListWafProfiles(ctx context.Context) ([]*client.WafProfile, error)
	GetWafEngineConfig(ctx context.Context) ([]client.WafConfigVariable, error)
	ListTriggeredWafRules(ctx context.Context, shieldZoneId int64) ([]*client.TriggeredRule, error)
	UpdateTriggeredWafRule(ctx context.Context, shieldZoneId int64, body *client.TriggeredRuleUpdate) error

	// Rate Limits
	ListRateLimits(ctx context.Context, shieldZoneId int64, page, perPage int) (pagination.PageResponse[*client.RateLimitRule], error)
	GetRateLimit(ctx context.Context, id int64) (*client.RateLimitRule, error)
	CreateRateLimit(ctx context.Context, body *client.RateLimitRuleCreate) (*client.RateLimitRule, error)
	UpdateRateLimit(ctx context.Context, id int64, body *client.RateLimitRuleUpdate) (*client.RateLimitRule, error)
	DeleteRateLimit(ctx context.Context, id int64) error

	// Access Lists
	ListAccessLists(ctx context.Context, shieldZoneId int64) (*client.AccessListsResponse, error)
	GetCustomAccessList(ctx context.Context, shieldZoneId, id int64) (*client.CustomAccessList, error)
	CreateCustomAccessList(ctx context.Context, shieldZoneId int64, body *client.CustomAccessListCreate) (*client.CustomAccessList, error)
	UpdateCustomAccessList(ctx context.Context, shieldZoneId, id int64, body *client.CustomAccessListUpdate) (*client.CustomAccessList, error)
	DeleteCustomAccessList(ctx context.Context, shieldZoneId, id int64) error
	UpdateAccessListConfig(ctx context.Context, shieldZoneId, configId int64, body *client.AccessListConfigUpdate) error

	// Bot Detection
	GetBotDetection(ctx context.Context, shieldZoneId int64) (*client.BotDetectionConfig, error)
	UpdateBotDetection(ctx context.Context, shieldZoneId int64, body *client.BotDetectionUpdate) error

	// Upload Scanning
	GetUploadScanning(ctx context.Context, shieldZoneId int64) (*client.UploadScanningConfig, error)
	UpdateUploadScanning(ctx context.Context, shieldZoneId int64, body *client.UploadScanningUpdate) error

	// Metrics
	GetShieldMetricsOverview(ctx context.Context, shieldZoneId int64) (*client.ShieldZoneMetrics, error)
	GetShieldMetricsDetailed(ctx context.Context, shieldZoneId int64, startDate, endDate string, resolution int) (*client.ShieldOverviewMetricsData, error)
	GetShieldRateLimitMetrics(ctx context.Context, shieldZoneId int64) ([]*client.ShieldZoneRateLimitMetrics, error)
	GetShieldRateLimitMetric(ctx context.Context, id int64) (*client.RatelimitMetrics, error)
	GetShieldWafRuleMetrics(ctx context.Context, shieldZoneId int64, ruleId int) (*client.WafRuleMetrics, error)
	GetShieldBotDetectionMetrics(ctx context.Context, shieldZoneId int64) (*client.ShieldZoneBotDetectionMetrics, error)
	GetShieldUploadScanningMetrics(ctx context.Context, shieldZoneId int64) (*client.ShieldZoneUploadScanningMetrics, error)

	// Event Logs
	GetShieldEventLogs(ctx context.Context, shieldZoneId int64, date, continuationToken string) (*client.EventLogResponse, error)
}
