package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/built-fast/bunny-cli/internal/pagination"
)

// --- Shield response envelope types ---

// shieldPaginationInfo holds pagination metadata from Shield API responses.
type shieldPaginationInfo struct {
	TotalCount  int `json:"totalCount"`
	TotalPages  int `json:"totalPages"`
	CurrentPage int `json:"currentPage"`
	NextPage    int `json:"nextPage"`
	PageSize    int `json:"pageSize"`
}

// shieldPageResponse is the pagination envelope used by the Shield API.
type shieldPageResponse[T any] struct {
	Data  []T                  `json:"data"`
	Page  shieldPaginationInfo `json:"page"`
	Error any                  `json:"error,omitempty"`
}

// toPageResponse converts to the standard PageResponse used across the CLI.
func (s shieldPageResponse[T]) toPageResponse() pagination.PageResponse[T] {
	return pagination.PageResponse[T]{
		Items:        s.Data,
		CurrentPage:  s.Page.CurrentPage,
		TotalItems:   s.Page.TotalCount,
		HasMoreItems: s.Page.CurrentPage < s.Page.TotalPages,
	}
}

// shieldResponse wraps a single-item Shield API response.
type shieldResponse[T any] struct {
	Data  T   `json:"data"`
	Error any `json:"error,omitempty"`
}

// --- Shield Zone models ---

// ShieldZone represents a bunny.net Shield zone configuration.
type ShieldZone struct {
	ShieldZoneId                         int64                     `json:"shieldZoneId"`
	PullZoneId                           int64                     `json:"pullZoneId"`
	UserId                               string                    `json:"userId"`
	PlanType                             int                       `json:"planType"`
	DeletedDateTime                      string                    `json:"deletedDateTime"`
	LearningMode                         bool                      `json:"learningMode"`
	LearningModeUntil                    string                    `json:"learningModeUntil"`
	WafEnabled                           bool                      `json:"wafEnabled"`
	WafExecutionMode                     int                       `json:"wafExecutionMode"`
	WafDisabledRuleGroups                json.RawMessage           `json:"wafDisabledRuleGroups"`
	WafDisabledRules                     json.RawMessage           `json:"wafDisabledRules"`
	WafLogOnlyRules                      json.RawMessage           `json:"wafLogOnlyRules"`
	WafRequestHeaderLoggingEnabled       bool                      `json:"wafRequestHeaderLoggingEnabled"`
	WafRequestIgnoredHeaders             json.RawMessage           `json:"wafRequestIgnoredHeaders"`
	WafRealtimeThreatIntelligenceEnabled bool                      `json:"wafRealtimeThreatIntelligenceEnabled"`
	WafRequestBodyLimitAction            int                       `json:"wafRequestBodyLimitAction"`
	WafResponseBodyLimitAction           int                       `json:"wafResponseBodyLimitAction"`
	WafProfileId                         int                       `json:"wafProfileId"`
	DDoSEnabled                          bool                      `json:"dDoSEnabled"`
	DDoSShieldSensitivity                int                       `json:"dDoSShieldSensitivity"`
	DDoSExecutionMode                    int                       `json:"dDoSExecutionMode"`
	DDoSBlockingMode                     int                       `json:"dDoSBlockingMode"`
	DDoSChallengeWindow                  int                       `json:"dDoSChallengeWindow"`
	DDoSRequestVariationSensitivity      int                       `json:"dDoSRequestVariationSensitivity"`
	TotalWAFCustomRules                  int                       `json:"totalWAFCustomRules"`
	TotalRateLimitRules                  int                       `json:"totalRateLimitRules"`
	LastModified                         string                    `json:"lastModified"`
	CreatedDateTime                      string                    `json:"createdDateTime"`
	WhitelabelResponsePages              bool                      `json:"whitelabelResponsePages"`
	BotDetectionConfiguration            *BotDetectionConfig       `json:"botDetectionConfiguration"`
	UploadScanningConfiguration          *UploadScanningConfig     `json:"uploadScanningConfiguration"`
	AccessListConfigurations             []AccessListConfiguration `json:"accessListConfigurations"`
}

// ShieldZoneCreate holds the fields for creating a shield zone.
type ShieldZoneCreate struct {
	PullZoneId int64 `json:"pullZoneId"`
}

// ShieldZoneUpdate holds the fields for updating a shield zone.
type ShieldZoneUpdate struct {
	LearningMode                         *bool   `json:"learningMode,omitempty"`
	LearningModeUntil                    *string `json:"learningModeUntil,omitempty"`
	WafEnabled                           *bool   `json:"wafEnabled,omitempty"`
	WafExecutionMode                     *int    `json:"wafExecutionMode,omitempty"`
	WafDisabledRules                     *string `json:"wafDisabledRules,omitempty"`
	WafLogOnlyRules                      *string `json:"wafLogOnlyRules,omitempty"`
	WafRequestHeaderLoggingEnabled       *bool   `json:"wafRequestHeaderLoggingEnabled,omitempty"`
	WafRequestIgnoredHeaders             *string `json:"wafRequestIgnoredHeaders,omitempty"`
	WafRealtimeThreatIntelligenceEnabled *bool   `json:"wafRealtimeThreatIntelligenceEnabled,omitempty"`
	WafProfileId                         *int    `json:"wafProfileId,omitempty"`
	WafRequestBodyLimitAction            *int    `json:"wafRequestBodyLimitAction,omitempty"`
	WafResponseBodyLimitAction           *int    `json:"wafResponseBodyLimitAction,omitempty"`
	DDoSShieldSensitivity                *int    `json:"dDoSShieldSensitivity,omitempty"`
	DDoSExecutionMode                    *int    `json:"dDoSExecutionMode,omitempty"`
	DDoSChallengeWindow                  *int    `json:"dDoSChallengeWindow,omitempty"`
	WhitelabelResponsePages              *bool   `json:"whitelabelResponsePages,omitempty"`
}

// ShieldZoneResponse is the response returned by create/update operations (fewer fields than full ShieldZone).
type ShieldZoneResponse struct {
	ShieldZoneId                         int64  `json:"shieldZoneId"`
	PullZoneId                           int64  `json:"pullZoneId"`
	LearningMode                         bool   `json:"learningMode"`
	LearningModeUntil                    string `json:"learningModeUntil"`
	WafEnabled                           bool   `json:"wafEnabled"`
	WafExecutionMode                     int    `json:"wafExecutionMode"`
	WafDisabledRules                     string `json:"wafDisabledRules"`
	WafLogOnlyRules                      string `json:"wafLogOnlyRules"`
	WafRequestHeaderLoggingEnabled       bool   `json:"wafRequestHeaderLoggingEnabled"`
	WafRealtimeThreatIntelligenceEnabled bool   `json:"wafRealtimeThreatIntelligenceEnabled"`
	WafProfileId                         int    `json:"wafProfileId"`
	WafRequestBodyLimitAction            int    `json:"wafRequestBodyLimitAction"`
	WafResponseBodyLimitAction           int    `json:"wafResponseBodyLimitAction"`
	PlanType                             int    `json:"planType"`
	DDoSShieldSensitivity                int    `json:"dDoSShieldSensitivity"`
	DDoSExecutionMode                    int    `json:"dDoSExecutionMode"`
	DDoSChallengeWindow                  int    `json:"dDoSChallengeWindow"`
	WhitelabelResponsePages              bool   `json:"whitelabelResponsePages"`
	RateLimitRulesLimit                  int    `json:"rateLimitRulesLimit"`
	CustomWafRulesLimit                  int    `json:"customWafRulesLimit"`
}

// --- WAF models ---

// WafRuleMainGroup represents a top-level WAF rule group.
type WafRuleMainGroup struct {
	Name       string         `json:"name"`
	Ruleset    string         `json:"ruleset"`
	RuleGroups []WafRuleGroup `json:"ruleGroups"`
}

// WafRuleGroup represents a WAF rule group within a main group.
type WafRuleGroup struct {
	Id          int       `json:"id"`
	Name        string    `json:"name"`
	Code        string    `json:"code"`
	FileName    string    `json:"fileName"`
	MainGroup   string    `json:"mainGroup"`
	Ruleset     string    `json:"ruleset"`
	Description string    `json:"description"`
	Rules       []WafRule `json:"rules"`
}

// WafRule represents an individual managed WAF rule.
type WafRule struct {
	RuleId      int    `json:"ruleId"`
	Description string `json:"description"`
}

// CustomWafRule represents a user-created WAF rule.
type CustomWafRule struct {
	Id                int64          `json:"id"`
	ShieldZoneId      int64          `json:"shieldZoneId"`
	UserId            string         `json:"userId"`
	RuleName          string         `json:"ruleName"`
	RuleDescription   string         `json:"ruleDescription"`
	RuleJson          string         `json:"ruleJson"`
	RuleConfiguration *WafRuleConfig `json:"ruleConfiguration"`
}

// WafRuleConfig holds the configuration for a custom WAF rule.
type WafRuleConfig struct {
	ActionType            int                       `json:"actionType"`
	VariableTypes         map[string]string         `json:"variableTypes"`
	OperatorType          int                       `json:"operatorType"`
	SeverityType          int                       `json:"severityType"`
	TransformationTypes   []int                     `json:"transformationTypes"`
	Value                 string                    `json:"value"`
	ChainedRuleConditions []WafChainedRuleCondition `json:"chainedRuleConditions"`
}

// WafChainedRuleCondition represents a chained condition in a WAF rule.
type WafChainedRuleCondition struct {
	VariableTypes map[string]string `json:"variableTypes"`
	OperatorType  int               `json:"operatorType"`
	Value         string            `json:"value"`
}

// CustomWafRuleCreate holds the fields for creating a custom WAF rule.
type CustomWafRuleCreate struct {
	ShieldZoneId      int64          `json:"shieldZoneId"`
	RuleName          string         `json:"ruleName"`
	RuleDescription   string         `json:"ruleDescription,omitempty"`
	RuleConfiguration *WafRuleConfig `json:"ruleConfiguration"`
}

// CustomWafRuleUpdate holds the fields for updating a custom WAF rule.
type CustomWafRuleUpdate struct {
	RuleName          *string        `json:"ruleName,omitempty"`
	RuleDescription   *string        `json:"ruleDescription,omitempty"`
	RuleConfiguration *WafRuleConfig `json:"ruleConfiguration,omitempty"`
}

// WafProfile represents a WAF profile.
type WafProfile struct {
	Id              int    `json:"id"`
	Name            string `json:"name"`
	IsPremium       bool   `json:"isPremium"`
	ProfileCategory string `json:"profileCategory"`
	ImageUrl        string `json:"imageUrl"`
	Description     string `json:"description"`
	Features        string `json:"features"`
}

// WafEngineConfig holds WAF engine configuration data.
type WafEngineConfig struct {
	Variables []WafConfigVariable `json:"data"`
}

// WafConfigVariable represents a WAF engine configuration variable.
type WafConfigVariable struct {
	Name         string `json:"name"`
	ValueEncoded string `json:"valueEncoded"`
}

// TriggeredRule represents a WAF rule that has been triggered.
type TriggeredRule struct {
	RuleId                 string `json:"ruleId"`
	RuleDescription        string `json:"ruleDescription"`
	TotalTriggeredRequests int64  `json:"totalTriggeredRequests"`
}

// TriggeredRuleUpdate holds the fields for updating a triggered rule's review action.
type TriggeredRuleUpdate struct {
	RuleId string `json:"ruleId"`
	Action int    `json:"action"`
}

// --- Rate Limit models ---

// RateLimitRule represents a rate limiting rule.
type RateLimitRule struct {
	Id                int64            `json:"id"`
	ShieldZoneId      int64            `json:"shieldZoneId"`
	UserId            string           `json:"userId"`
	RuleName          string           `json:"ruleName"`
	RuleDescription   string           `json:"ruleDescription"`
	RuleJson          string           `json:"ruleJson"`
	RuleConfiguration *RateLimitConfig `json:"ruleConfiguration"`
}

// RateLimitConfig holds the configuration for a rate limit rule.
type RateLimitConfig struct {
	ActionType            int                       `json:"actionType"`
	VariableTypes         map[string]string         `json:"variableTypes"`
	OperatorType          int                       `json:"operatorType"`
	SeverityType          int                       `json:"severityType"`
	TransformationTypes   []int                     `json:"transformationTypes"`
	Value                 string                    `json:"value"`
	RequestCount          int                       `json:"requestCount"`
	CounterKeyType        int                       `json:"counterKeyType"`
	Timeframe             int                       `json:"timeframe"`
	BlockTime             int                       `json:"blockTime"`
	ChainedRuleConditions []WafChainedRuleCondition `json:"chainedRuleConditions"`
}

// RateLimitRuleCreate holds the fields for creating a rate limit rule.
type RateLimitRuleCreate struct {
	ShieldZoneId      int64            `json:"shieldZoneId"`
	RuleName          string           `json:"ruleName"`
	RuleDescription   string           `json:"ruleDescription,omitempty"`
	RuleConfiguration *RateLimitConfig `json:"ruleConfiguration"`
}

// RateLimitRuleUpdate holds the fields for updating a rate limit rule.
type RateLimitRuleUpdate struct {
	RuleName          *string          `json:"ruleName,omitempty"`
	RuleDescription   *string          `json:"ruleDescription,omitempty"`
	RuleConfiguration *RateLimitConfig `json:"ruleConfiguration,omitempty"`
}

// --- Access List models ---

// AccessListDetails holds details about a managed or custom access list.
type AccessListDetails struct {
	ListId          int64  `json:"listId"`
	ConfigurationId int64  `json:"configurationId"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	IsEnabled       bool   `json:"isEnabled"`
	Type            int    `json:"type"`
	Category        int    `json:"category"`
	Action          int    `json:"action"`
	RequiredPlan    int    `json:"requiredPlan"`
	EntryCount      int64  `json:"entryCount"`
	UpdateFrequency string `json:"updateFrequency"`
	LastUpdated     string `json:"lastUpdated"`
}

// AccessListsResponse holds the combined managed and custom access lists.
type AccessListsResponse struct {
	ManagedLists     []AccessListDetails `json:"managedLists"`
	CustomLists      []AccessListDetails `json:"customLists"`
	CustomEntryCount int                 `json:"customEntryCount"`
	CustomEntryLimit int                 `json:"customEntryLimit"`
	CustomListCount  int                 `json:"customListCount"`
	CustomListLimit  int                 `json:"customListLimit"`
}

// CustomAccessList represents a user-created access list.
type CustomAccessList struct {
	Id           int64  `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Type         int    `json:"type"`
	Content      string `json:"content"`
	Checksum     string `json:"checksum"`
	EntryCount   int64  `json:"entryCount"`
	LastModified string `json:"lastModified"`
}

// CustomAccessListCreate holds the fields for creating a custom access list.
type CustomAccessListCreate struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Type        int    `json:"type"`
	Content     string `json:"content"`
	Checksum    string `json:"checksum,omitempty"`
}

// CustomAccessListUpdate holds the fields for updating a custom access list.
type CustomAccessListUpdate struct {
	Name     *string `json:"name,omitempty"`
	Content  *string `json:"content,omitempty"`
	Checksum *string `json:"checksum,omitempty"`
}

// AccessListConfiguration holds the configuration/binding for an access list on a shield zone.
type AccessListConfiguration struct {
	Id               int64  `json:"id"`
	ShieldZoneId     int64  `json:"shieldZoneId"`
	UserId           string `json:"userId"`
	AccessListId     int64  `json:"accessListId"`
	UserAccessListId int64  `json:"userAccessListId"`
	IsEnabled        bool   `json:"isEnabled"`
	Action           int    `json:"action"`
}

// AccessListConfigUpdate holds the fields for updating an access list configuration.
type AccessListConfigUpdate struct {
	IsEnabled *bool `json:"isEnabled,omitempty"`
	Action    *int  `json:"action,omitempty"`
}

// --- Bot Detection models ---

// BotDetectionConfig represents bot detection configuration state for a shield zone.
// Maps to BotDetectionConfigurationState in the API spec.
type BotDetectionConfig struct {
	ShieldZoneId       int64                         `json:"shieldZoneId"`
	ExecutionMode      int                           `json:"executionMode"`
	RequestIntegrity   BotDetectionSensitivityConfig `json:"requestIntegrity"`
	IpAddress          BotDetectionSensitivityConfig `json:"ipAddress"`
	BrowserFingerprint BrowserFingerprintConfig      `json:"browserFingerprint"`
}

// BotDetectionUpdate holds the fields for updating bot detection configuration.
type BotDetectionUpdate struct {
	ShieldZoneId       int64                          `json:"shieldZoneId"`
	ExecutionMode      *int                           `json:"executionMode,omitempty"`
	RequestIntegrity   *BotDetectionSensitivityConfig `json:"requestIntegrity,omitempty"`
	IpAddress          *BotDetectionSensitivityConfig `json:"ipAddress,omitempty"`
	BrowserFingerprint *BrowserFingerprintConfig      `json:"browserFingerprint,omitempty"`
}

// BotDetectionSensitivityConfig holds a sensitivity level.
type BotDetectionSensitivityConfig struct {
	Sensitivity int `json:"sensitivity"`
}

// BrowserFingerprintConfig holds browser fingerprint configuration.
type BrowserFingerprintConfig struct {
	Sensitivity    int  `json:"sensitivity"`
	Aggression     int  `json:"aggression"`
	ComplexEnabled bool `json:"complexEnabled"`
}

// --- Upload Scanning models ---

// UploadScanningConfig represents upload scanning configuration for a shield zone.
type UploadScanningConfig struct {
	ShieldZoneId          int64 `json:"shieldZoneId"`
	IsEnabled             bool  `json:"isEnabled"`
	AntivirusScanningMode int   `json:"antivirusScanningMode"`
	CsamScanningMode      int   `json:"csamScanningMode"`
}

// UploadScanningUpdate holds the fields for updating upload scanning configuration.
type UploadScanningUpdate struct {
	ShieldZoneId          int64 `json:"shieldZoneId"`
	IsEnabled             *bool `json:"isEnabled,omitempty"`
	CsamScanningMode      *int  `json:"csamScanningMode,omitempty"`
	AntivirusScanningMode *int  `json:"antivirusScanningMode,omitempty"`
}

// --- Metrics models ---

// ShieldZoneMetrics holds overview metrics for a shield zone.
type ShieldZoneMetrics struct {
	Overview                *ShieldOverview      `json:"overview"`
	Waf                     *WafMetrics          `json:"waf"`
	DDoS                    *DDoSMetrics         `json:"dDoS"`
	Ratelimit               *RatelimitMetrics    `json:"ratelimit"`
	BotDetection            *BotDetectionMetrics `json:"botDetection"`
	AccessList              *AccessListMetrics   `json:"accessList"`
	TotalCleanRequestsLimit int64                `json:"totalCleanRequestsLimit"`
	TotalBillableRequests   int64                `json:"totalBillableRequests"`
}

// ShieldOverview holds high-level shield metric totals.
type ShieldOverview struct {
	DDoSMitigated          int64 `json:"dDoSMitigated"`
	WafTriggeredRules      int64 `json:"wafTriggeredRules"`
	RatelimitBreaches      int64 `json:"ratelimitBreaches"`
	BotDetectionChallenged int64 `json:"botDetectionChallenged"`
	AccessListActions      int64 `json:"accessListActions"`
	UploadScanningBlocks   int64 `json:"uploadScanningBlocks"`
}

// WafMetrics holds WAF-specific metrics.
type WafMetrics struct {
	TotalTriggeredRules int64 `json:"totalTriggeredRules"`
	BlockedRequests     int64 `json:"blockedRequests"`
	LoggedRequests      int64 `json:"loggedRequests"`
	ChallengedRequests  int64 `json:"challengedRequests"`
}

// DDoSMetrics holds DDoS-specific metrics.
type DDoSMetrics struct {
	LoggedRequests     int64 `json:"loggedRequests"`
	VerifiedRequests   int64 `json:"verifiedRequests"`
	BlockedRequests    int64 `json:"blockedRequests"`
	ChallengedRequests int64 `json:"challengedRequests"`
}

// RatelimitMetrics holds rate limit metrics.
type RatelimitMetrics struct {
	TotalBreaches      int64 `json:"totalBreaches"`
	LoggedBreaches     int64 `json:"loggedBreaches"`
	ChallengedBreaches int64 `json:"challengedBreaches"`
	BlockedBreaches    int64 `json:"blockedBreaches"`
}

// BotDetectionMetrics holds bot detection metrics.
type BotDetectionMetrics struct {
	LoggedRequests     int64 `json:"loggedRequests"`
	ChallengedRequests int64 `json:"challengedRequests"`
}

// AccessListMetrics holds access list metrics.
type AccessListMetrics struct {
	TotalActions       int64 `json:"totalActions"`
	BlockedRequests    int64 `json:"blockedRequests"`
	LoggedRequests     int64 `json:"loggedRequests"`
	ChallengedRequests int64 `json:"challengedRequests"`
}

// ShieldOverviewMetricsData holds detailed overview metrics with time-series data.
type ShieldOverviewMetricsData struct {
	Waf                            *OverviewMetric `json:"waf"`
	DDoS                           *OverviewMetric `json:"ddos"`
	RateLimit                      *OverviewMetric `json:"rateLimit"`
	AccessLists                    *OverviewMetric `json:"accessLists"`
	BotDetection                   *OverviewMetric `json:"botDetection"`
	UploadScanning                 *OverviewMetric `json:"uploadScanning"`
	TotalCleanRequestsLimit        int64           `json:"totalCleanRequestsLimit"`
	TotalBillableRequestsThisMonth int64           `json:"totalBillableRequestsThisMonth"`
	Resolution                     int             `json:"resolution"`
}

// OverviewMetric holds time-series metric data.
type OverviewMetric struct {
	Metrics map[string]map[string]int64 `json:"metrics"`
	Totals  map[string]int64            `json:"totals"`
}

// ShieldZoneRateLimitMetrics holds metrics for a single rate limit rule.
type ShieldZoneRateLimitMetrics struct {
	RatelimitId int64             `json:"ratelimitId"`
	Overview    *RatelimitMetrics `json:"overview"`
}

// WafRuleMetrics holds metrics for a single WAF rule.
type WafRuleMetrics struct {
	TotalTriggers      int64 `json:"totalTriggers"`
	BlockedRequests    int64 `json:"blockedRequests"`
	LoggedRequests     int64 `json:"loggedRequests"`
	ChallengedRequests int64 `json:"challengedRequests"`
}

// ShieldZoneBotDetectionMetrics holds bot detection metrics for a shield zone.
type ShieldZoneBotDetectionMetrics struct {
	TotalLoggedRequests     int64 `json:"totalLoggedRequests"`
	TotalChallengedRequests int64 `json:"totalChallengedRequests"`
}

// ShieldZoneUploadScanningMetrics holds upload scanning metrics for a shield zone.
type ShieldZoneUploadScanningMetrics struct {
	TotalLoggedRequests  int64 `json:"totalLoggedRequests"`
	TotalBlockedRequests int64 `json:"totalBlockedRequests"`
	TotalFilesScanned    int64 `json:"totalFilesScanned"`
}

// --- Event Log models ---

// EventLog represents a single Shield event log entry.
type EventLog struct {
	LogId     string            `json:"logId"`
	Timestamp int64             `json:"timestamp"`
	Log       string            `json:"log"`
	Labels    map[string]string `json:"labels"`
}

// EventLogResponse holds the response from the event logs endpoint.
type EventLogResponse struct {
	Logs              []EventLog `json:"logs"`
	HasMoreData       bool       `json:"hasMoreData"`
	ContinuationToken string     `json:"continuationToken"`
}

// --- Enum helpers ---

// ShieldPlanName returns a human-readable name for a Shield plan type.
func ShieldPlanName(planType int) string {
	switch planType {
	case 0:
		return "Free"
	case 1:
		return "Growth"
	case 2:
		return "Pro"
	case 3:
		return "Advanced"
	case 4:
		return "Enterprise"
	default:
		return fmt.Sprintf("Unknown(%d)", planType)
	}
}

// ShieldExecutionModeName returns a human-readable name for an execution mode.
func ShieldExecutionModeName(mode int) string {
	switch mode {
	case 0:
		return "Learn"
	case 1:
		return "Protect"
	default:
		return fmt.Sprintf("Unknown(%d)", mode)
	}
}

// WafActionTypeName returns a human-readable name for a WAF action type.
func WafActionTypeName(action int) string {
	switch action {
	case 0:
		return "Log"
	case 1:
		return "Block"
	case 2:
		return "Challenge"
	default:
		return fmt.Sprintf("Unknown(%d)", action)
	}
}

// RateLimitActionTypeName returns a human-readable name for a rate limit action type.
func RateLimitActionTypeName(action int) string {
	switch action {
	case 1:
		return "Log"
	case 2:
		return "Challenge"
	case 3:
		return "Block"
	default:
		return fmt.Sprintf("Unknown(%d)", action)
	}
}

// AccessListActionName returns a human-readable name for an access list action.
func AccessListActionName(action int) string {
	switch action {
	case 0:
		return "Allow"
	case 1:
		return "Block"
	case 2:
		return "Log"
	case 3:
		return "Challenge"
	default:
		return fmt.Sprintf("Unknown(%d)", action)
	}
}

// --- Shield Zone client methods ---

// ListShieldZones returns a paginated list of shield zones.
func (c *Client) ListShieldZones(ctx context.Context, page, perPage int) (pagination.PageResponse[*ShieldZone], error) {
	params := url.Values{}
	params.Set("page", fmt.Sprintf("%d", page))
	params.Set("perPage", fmt.Sprintf("%d", perPage))

	path := "/shield/shield-zones?" + params.Encode()
	var resp shieldPageResponse[*ShieldZone]
	if err := c.Get(ctx, path, &resp); err != nil {
		return pagination.PageResponse[*ShieldZone]{}, err
	}
	return resp.toPageResponse(), nil
}

// GetShieldZone returns a single shield zone by ID.
func (c *Client) GetShieldZone(ctx context.Context, id int64) (*ShieldZone, error) {
	var resp shieldResponse[*ShieldZone]
	err := c.Get(ctx, fmt.Sprintf("/shield/shield-zone/%d", id), &resp)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// GetShieldZoneByPullZone returns a shield zone by its associated pull zone ID.
func (c *Client) GetShieldZoneByPullZone(ctx context.Context, pullZoneId int64) (*ShieldZone, error) {
	var resp shieldResponse[*ShieldZone]
	err := c.Get(ctx, fmt.Sprintf("/shield/shield-zone/get-by-pullzone/%d", pullZoneId), &resp)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// CreateShieldZone creates a new shield zone.
func (c *Client) CreateShieldZone(ctx context.Context, body *ShieldZoneCreate) (*ShieldZoneResponse, error) {
	envelope := struct {
		PullZoneId int64 `json:"pullZoneId"`
	}{PullZoneId: body.PullZoneId}

	var resp shieldResponse[*ShieldZoneResponse]
	err := c.Post(ctx, "/shield/shield-zone", &envelope, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// UpdateShieldZone updates an existing shield zone.
func (c *Client) UpdateShieldZone(ctx context.Context, id int64, body *ShieldZoneUpdate) (*ShieldZoneResponse, error) {
	envelope := struct {
		ShieldZoneId int64             `json:"shieldZoneId"`
		ShieldZone   *ShieldZoneUpdate `json:"shieldZone"`
	}{ShieldZoneId: id, ShieldZone: body}

	var resp shieldResponse[*ShieldZoneResponse]
	err := c.Patch(ctx, "/shield/shield-zone", &envelope, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// --- WAF client methods ---

// ListWafRules returns the managed WAF rules for a shield zone.
func (c *Client) ListWafRules(ctx context.Context, shieldZoneId int64) ([]*WafRuleMainGroup, error) {
	var groups []*WafRuleMainGroup
	err := c.Get(ctx, fmt.Sprintf("/shield/waf/rules/%d", shieldZoneId), &groups)
	if err != nil {
		return nil, err
	}
	return groups, nil
}

// ListCustomWafRules returns a paginated list of custom WAF rules for a shield zone.
func (c *Client) ListCustomWafRules(ctx context.Context, shieldZoneId int64, page, perPage int) (pagination.PageResponse[*CustomWafRule], error) {
	params := url.Values{}
	params.Set("page", fmt.Sprintf("%d", page))
	params.Set("perPage", fmt.Sprintf("%d", perPage))

	path := fmt.Sprintf("/shield/waf/custom-rules/%d?%s", shieldZoneId, params.Encode())
	var resp shieldPageResponse[*CustomWafRule]
	if err := c.Get(ctx, path, &resp); err != nil {
		return pagination.PageResponse[*CustomWafRule]{}, err
	}
	return resp.toPageResponse(), nil
}

// GetCustomWafRule returns a single custom WAF rule by ID.
func (c *Client) GetCustomWafRule(ctx context.Context, id int64) (*CustomWafRule, error) {
	var rule CustomWafRule
	err := c.Get(ctx, fmt.Sprintf("/shield/waf/custom-rule/%d", id), &rule)
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

// CreateCustomWafRule creates a new custom WAF rule.
func (c *Client) CreateCustomWafRule(ctx context.Context, body *CustomWafRuleCreate) (*CustomWafRule, error) {
	var rule CustomWafRule
	err := c.Post(ctx, "/shield/waf/custom-rule", body, &rule)
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

// UpdateCustomWafRule updates a custom WAF rule.
func (c *Client) UpdateCustomWafRule(ctx context.Context, id int64, body *CustomWafRuleUpdate) (*CustomWafRule, error) {
	var rule CustomWafRule
	err := c.Patch(ctx, fmt.Sprintf("/shield/waf/custom-rule/%d", id), body, &rule)
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

// DeleteCustomWafRule deletes a custom WAF rule.
func (c *Client) DeleteCustomWafRule(ctx context.Context, id int64) error {
	return c.Delete(ctx, fmt.Sprintf("/shield/waf/custom-rule/%d", id))
}

// ListWafProfiles returns the available WAF profiles.
func (c *Client) ListWafProfiles(ctx context.Context) ([]*WafProfile, error) {
	// API returns data as [][]WafProfile (array of arrays grouped by category).
	var resp shieldResponse[[][]*WafProfile]
	err := c.Get(ctx, "/shield/waf/profiles", &resp)
	if err != nil {
		return nil, err
	}
	// Flatten into a single list.
	var profiles []*WafProfile
	for _, group := range resp.Data {
		profiles = append(profiles, group...)
	}
	return profiles, nil
}

// GetWafEngineConfig returns the WAF engine configuration variables.
func (c *Client) GetWafEngineConfig(ctx context.Context) ([]WafConfigVariable, error) {
	var resp shieldResponse[[]WafConfigVariable]
	err := c.Get(ctx, "/shield/waf/engine-config", &resp)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// ListTriggeredWafRules returns the triggered WAF rules for a shield zone.
func (c *Client) ListTriggeredWafRules(ctx context.Context, shieldZoneId int64) ([]*TriggeredRule, error) {
	var resp struct {
		TriggeredRules      []*TriggeredRule `json:"triggeredRules"`
		TotalTriggeredRules int              `json:"totalTriggeredRules"`
	}
	err := c.Get(ctx, fmt.Sprintf("/shield/waf/rules/review-triggered/%d", shieldZoneId), &resp)
	if err != nil {
		return nil, err
	}
	return resp.TriggeredRules, nil
}

// UpdateTriggeredWafRule updates the review action for a triggered WAF rule.
func (c *Client) UpdateTriggeredWafRule(ctx context.Context, shieldZoneId int64, body *TriggeredRuleUpdate) error {
	return c.Post(ctx, fmt.Sprintf("/shield/waf/rules/review-triggered/%d", shieldZoneId), body, nil)
}

// --- Rate Limit client methods ---

// ListRateLimits returns a paginated list of rate limit rules for a shield zone.
func (c *Client) ListRateLimits(ctx context.Context, shieldZoneId int64, page, perPage int) (pagination.PageResponse[*RateLimitRule], error) {
	params := url.Values{}
	params.Set("page", fmt.Sprintf("%d", page))
	params.Set("perPage", fmt.Sprintf("%d", perPage))

	path := fmt.Sprintf("/shield/rate-limits/%d?%s", shieldZoneId, params.Encode())
	var resp shieldPageResponse[*RateLimitRule]
	if err := c.Get(ctx, path, &resp); err != nil {
		return pagination.PageResponse[*RateLimitRule]{}, err
	}
	return resp.toPageResponse(), nil
}

// GetRateLimit returns a single rate limit rule by ID.
func (c *Client) GetRateLimit(ctx context.Context, id int64) (*RateLimitRule, error) {
	var rule RateLimitRule
	err := c.Get(ctx, fmt.Sprintf("/shield/rate-limit/%d", id), &rule)
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

// CreateRateLimit creates a new rate limit rule.
func (c *Client) CreateRateLimit(ctx context.Context, body *RateLimitRuleCreate) (*RateLimitRule, error) {
	var rule RateLimitRule
	err := c.Post(ctx, "/shield/rate-limit", body, &rule)
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

// UpdateRateLimit updates a rate limit rule.
func (c *Client) UpdateRateLimit(ctx context.Context, id int64, body *RateLimitRuleUpdate) (*RateLimitRule, error) {
	var rule RateLimitRule
	err := c.Patch(ctx, fmt.Sprintf("/shield/rate-limit/%d", id), body, &rule)
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

// DeleteRateLimit deletes a rate limit rule.
func (c *Client) DeleteRateLimit(ctx context.Context, id int64) error {
	return c.Delete(ctx, fmt.Sprintf("/shield/rate-limit/%d", id))
}

// --- Access List client methods ---

// ListAccessLists returns the access lists for a shield zone.
func (c *Client) ListAccessLists(ctx context.Context, shieldZoneId int64) (*AccessListsResponse, error) {
	var resp AccessListsResponse
	err := c.Get(ctx, fmt.Sprintf("/shield/shield-zone/%d/access-lists", shieldZoneId), &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetCustomAccessList returns a single custom access list.
func (c *Client) GetCustomAccessList(ctx context.Context, shieldZoneId, id int64) (*CustomAccessList, error) {
	var resp shieldResponse[*CustomAccessList]
	err := c.Get(ctx, fmt.Sprintf("/shield/shield-zone/%d/access-lists/%d", shieldZoneId, id), &resp)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// CreateCustomAccessList creates a new custom access list.
func (c *Client) CreateCustomAccessList(ctx context.Context, shieldZoneId int64, body *CustomAccessListCreate) (*CustomAccessList, error) {
	var resp shieldResponse[*CustomAccessList]
	err := c.Post(ctx, fmt.Sprintf("/shield/shield-zone/%d/access-lists", shieldZoneId), body, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// UpdateCustomAccessList updates a custom access list.
func (c *Client) UpdateCustomAccessList(ctx context.Context, shieldZoneId, id int64, body *CustomAccessListUpdate) (*CustomAccessList, error) {
	var resp shieldResponse[*CustomAccessList]
	err := c.Patch(ctx, fmt.Sprintf("/shield/shield-zone/%d/access-lists/%d", shieldZoneId, id), body, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// DeleteCustomAccessList deletes a custom access list.
func (c *Client) DeleteCustomAccessList(ctx context.Context, shieldZoneId, id int64) error {
	return c.Delete(ctx, fmt.Sprintf("/shield/shield-zone/%d/access-lists/%d", shieldZoneId, id))
}

// UpdateAccessListConfig updates an access list configuration.
func (c *Client) UpdateAccessListConfig(ctx context.Context, shieldZoneId, configId int64, body *AccessListConfigUpdate) error {
	var resp shieldResponse[*AccessListConfiguration]
	return c.Patch(ctx, fmt.Sprintf("/shield/shield-zone/%d/access-lists/configurations/%d", shieldZoneId, configId), body, &resp)
}

// --- Bot Detection client methods ---

// GetBotDetection returns the bot detection configuration for a shield zone.
func (c *Client) GetBotDetection(ctx context.Context, shieldZoneId int64) (*BotDetectionConfig, error) {
	var resp shieldResponse[*BotDetectionConfig]
	err := c.Get(ctx, fmt.Sprintf("/shield/shield-zone/%d/bot-detection", shieldZoneId), &resp)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// UpdateBotDetection updates the bot detection configuration for a shield zone.
func (c *Client) UpdateBotDetection(ctx context.Context, shieldZoneId int64, body *BotDetectionUpdate) error {
	body.ShieldZoneId = shieldZoneId
	var resp shieldResponse[*BotDetectionConfig]
	return c.Patch(ctx, fmt.Sprintf("/shield/shield-zone/%d/bot-detection", shieldZoneId), body, &resp)
}

// --- Upload Scanning client methods ---

// GetUploadScanning returns the upload scanning configuration for a shield zone.
func (c *Client) GetUploadScanning(ctx context.Context, shieldZoneId int64) (*UploadScanningConfig, error) {
	var resp shieldResponse[*UploadScanningConfig]
	err := c.Get(ctx, fmt.Sprintf("/shield/shield-zone/%d/upload-scanning", shieldZoneId), &resp)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// UpdateUploadScanning updates the upload scanning configuration for a shield zone.
func (c *Client) UpdateUploadScanning(ctx context.Context, shieldZoneId int64, body *UploadScanningUpdate) error {
	body.ShieldZoneId = shieldZoneId
	var resp shieldResponse[*UploadScanningConfig]
	return c.Patch(ctx, fmt.Sprintf("/shield/shield-zone/%d/upload-scanning", shieldZoneId), body, &resp)
}

// --- Metrics client methods ---

// GetShieldMetricsOverview returns overview metrics for a shield zone.
func (c *Client) GetShieldMetricsOverview(ctx context.Context, shieldZoneId int64) (*ShieldZoneMetrics, error) {
	var resp shieldResponse[*ShieldZoneMetrics]
	err := c.Get(ctx, fmt.Sprintf("/shield/metrics/overview/%d", shieldZoneId), &resp)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// GetShieldMetricsDetailed returns detailed overview metrics for a shield zone.
func (c *Client) GetShieldMetricsDetailed(ctx context.Context, shieldZoneId int64, startDate, endDate string, resolution int) (*ShieldOverviewMetricsData, error) {
	params := url.Values{}
	if startDate != "" {
		params.Set("StartDate", startDate)
	}
	if endDate != "" {
		params.Set("EndDate", endDate)
	}
	if resolution > 0 {
		params.Set("Resolution", fmt.Sprintf("%d", resolution))
	}

	path := fmt.Sprintf("/shield/metrics/overview/%d/detailed", shieldZoneId)
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	var resp shieldResponse[*ShieldOverviewMetricsData]
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// GetShieldRateLimitMetrics returns rate limit metrics for a shield zone.
func (c *Client) GetShieldRateLimitMetrics(ctx context.Context, shieldZoneId int64) ([]*ShieldZoneRateLimitMetrics, error) {
	var resp shieldResponse[[]*ShieldZoneRateLimitMetrics]
	err := c.Get(ctx, fmt.Sprintf("/shield/metrics/rate-limits/%d", shieldZoneId), &resp)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// GetShieldRateLimitMetric returns metrics for a single rate limit rule.
func (c *Client) GetShieldRateLimitMetric(ctx context.Context, id int64) (*RatelimitMetrics, error) {
	var resp shieldResponse[*RatelimitMetrics]
	err := c.Get(ctx, fmt.Sprintf("/shield/metrics/rate-limit/%d", id), &resp)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// GetShieldWafRuleMetrics returns metrics for a specific WAF rule on a shield zone.
func (c *Client) GetShieldWafRuleMetrics(ctx context.Context, shieldZoneId int64, ruleId int) (*WafRuleMetrics, error) {
	var resp shieldResponse[*WafRuleMetrics]
	err := c.Get(ctx, fmt.Sprintf("/shield/metrics/shield-zone/%d/waf-rule/%d", shieldZoneId, ruleId), &resp)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// GetShieldBotDetectionMetrics returns bot detection metrics for a shield zone.
func (c *Client) GetShieldBotDetectionMetrics(ctx context.Context, shieldZoneId int64) (*ShieldZoneBotDetectionMetrics, error) {
	var resp shieldResponse[*ShieldZoneBotDetectionMetrics]
	err := c.Get(ctx, fmt.Sprintf("/shield/metrics/shield-zone/%d/bot-detection", shieldZoneId), &resp)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// GetShieldUploadScanningMetrics returns upload scanning metrics for a shield zone.
func (c *Client) GetShieldUploadScanningMetrics(ctx context.Context, shieldZoneId int64) (*ShieldZoneUploadScanningMetrics, error) {
	var resp shieldResponse[*ShieldZoneUploadScanningMetrics]
	err := c.Get(ctx, fmt.Sprintf("/shield/metrics/shield-zone/%d/upload-scanning", shieldZoneId), &resp)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// --- Event Log client methods ---

// GetShieldEventLogs returns event logs for a shield zone on a specific date.
func (c *Client) GetShieldEventLogs(ctx context.Context, shieldZoneId int64, date, continuationToken string) (*EventLogResponse, error) {
	if continuationToken == "" {
		continuationToken = "start"
	}
	var resp EventLogResponse
	err := c.Get(ctx, fmt.Sprintf("/shield/event-logs/%d/%s/%s", shieldZoneId, date, continuationToken), &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}
