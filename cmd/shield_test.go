package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/built-fast/bunny-cli/internal/client"
	"github.com/built-fast/bunny-cli/internal/pagination"
	"github.com/spf13/cobra"
)

// mockShieldAPI implements ShieldAPI for testing.
type mockShieldAPI struct {
	listShieldZonesFn                func(ctx context.Context, page, perPage int) (pagination.PageResponse[*client.ShieldZone], error)
	getShieldZoneFn                  func(ctx context.Context, id int64) (*client.ShieldZone, error)
	getShieldZoneByPullZoneFn        func(ctx context.Context, pullZoneId int64) (*client.ShieldZone, error)
	createShieldZoneFn               func(ctx context.Context, body *client.ShieldZoneCreate) (*client.ShieldZoneResponse, error)
	updateShieldZoneFn               func(ctx context.Context, id int64, body *client.ShieldZoneUpdate) (*client.ShieldZoneResponse, error)
	listWafRulesFn                   func(ctx context.Context, shieldZoneId int64) ([]*client.WafRuleMainGroup, error)
	listCustomWafRulesFn             func(ctx context.Context, shieldZoneId int64, page, perPage int) (pagination.PageResponse[*client.CustomWafRule], error)
	getCustomWafRuleFn               func(ctx context.Context, id int64) (*client.CustomWafRule, error)
	createCustomWafRuleFn            func(ctx context.Context, body *client.CustomWafRuleCreate) (*client.CustomWafRule, error)
	updateCustomWafRuleFn            func(ctx context.Context, id int64, body *client.CustomWafRuleUpdate) (*client.CustomWafRule, error)
	deleteCustomWafRuleFn            func(ctx context.Context, id int64) error
	listWafProfilesFn                func(ctx context.Context) ([]*client.WafProfile, error)
	getWafEngineConfigFn             func(ctx context.Context) ([]client.WafConfigVariable, error)
	listTriggeredWafRulesFn          func(ctx context.Context, shieldZoneId int64) ([]*client.TriggeredRule, error)
	updateTriggeredWafRuleFn         func(ctx context.Context, shieldZoneId int64, body *client.TriggeredRuleUpdate) error
	listRateLimitsFn                 func(ctx context.Context, shieldZoneId int64, page, perPage int) (pagination.PageResponse[*client.RateLimitRule], error)
	getRateLimitFn                   func(ctx context.Context, id int64) (*client.RateLimitRule, error)
	createRateLimitFn                func(ctx context.Context, body *client.RateLimitRuleCreate) (*client.RateLimitRule, error)
	updateRateLimitFn                func(ctx context.Context, id int64, body *client.RateLimitRuleUpdate) (*client.RateLimitRule, error)
	deleteRateLimitFn                func(ctx context.Context, id int64) error
	listAccessListsFn                func(ctx context.Context, shieldZoneId int64) (*client.AccessListsResponse, error)
	getCustomAccessListFn            func(ctx context.Context, shieldZoneId, id int64) (*client.CustomAccessList, error)
	createCustomAccessListFn         func(ctx context.Context, shieldZoneId int64, body *client.CustomAccessListCreate) (*client.CustomAccessList, error)
	updateCustomAccessListFn         func(ctx context.Context, shieldZoneId, id int64, body *client.CustomAccessListUpdate) (*client.CustomAccessList, error)
	deleteCustomAccessListFn         func(ctx context.Context, shieldZoneId, id int64) error
	updateAccessListConfigFn         func(ctx context.Context, shieldZoneId, configId int64, body *client.AccessListConfigUpdate) error
	getBotDetectionFn                func(ctx context.Context, shieldZoneId int64) (*client.BotDetectionConfig, error)
	updateBotDetectionFn             func(ctx context.Context, shieldZoneId int64, body *client.BotDetectionUpdate) error
	getUploadScanningFn              func(ctx context.Context, shieldZoneId int64) (*client.UploadScanningConfig, error)
	updateUploadScanningFn           func(ctx context.Context, shieldZoneId int64, body *client.UploadScanningUpdate) error
	getShieldMetricsOverviewFn       func(ctx context.Context, shieldZoneId int64) (*client.ShieldZoneMetrics, error)
	getShieldMetricsDetailedFn       func(ctx context.Context, shieldZoneId int64, startDate, endDate string, resolution int) (*client.ShieldOverviewMetricsData, error)
	getShieldRateLimitMetricsFn      func(ctx context.Context, shieldZoneId int64) ([]*client.ShieldZoneRateLimitMetrics, error)
	getShieldRateLimitMetricFn       func(ctx context.Context, id int64) (*client.RatelimitMetrics, error)
	getShieldWafRuleMetricsFn        func(ctx context.Context, shieldZoneId int64, ruleId int) (*client.WafRuleMetrics, error)
	getShieldBotDetectionMetricsFn   func(ctx context.Context, shieldZoneId int64) (*client.ShieldZoneBotDetectionMetrics, error)
	getShieldUploadScanningMetricsFn func(ctx context.Context, shieldZoneId int64) (*client.ShieldZoneUploadScanningMetrics, error)
	getShieldEventLogsFn             func(ctx context.Context, shieldZoneId int64, date, continuationToken string) (*client.EventLogResponse, error)
}

func (m *mockShieldAPI) ListShieldZones(ctx context.Context, page, perPage int) (pagination.PageResponse[*client.ShieldZone], error) {
	return m.listShieldZonesFn(ctx, page, perPage)
}

func (m *mockShieldAPI) GetShieldZone(ctx context.Context, id int64) (*client.ShieldZone, error) {
	return m.getShieldZoneFn(ctx, id)
}

func (m *mockShieldAPI) GetShieldZoneByPullZone(ctx context.Context, pullZoneId int64) (*client.ShieldZone, error) {
	return m.getShieldZoneByPullZoneFn(ctx, pullZoneId)
}

func (m *mockShieldAPI) CreateShieldZone(ctx context.Context, body *client.ShieldZoneCreate) (*client.ShieldZoneResponse, error) {
	return m.createShieldZoneFn(ctx, body)
}

func (m *mockShieldAPI) UpdateShieldZone(ctx context.Context, id int64, body *client.ShieldZoneUpdate) (*client.ShieldZoneResponse, error) {
	return m.updateShieldZoneFn(ctx, id, body)
}

func (m *mockShieldAPI) ListWafRules(ctx context.Context, shieldZoneId int64) ([]*client.WafRuleMainGroup, error) {
	return m.listWafRulesFn(ctx, shieldZoneId)
}

func (m *mockShieldAPI) ListCustomWafRules(ctx context.Context, shieldZoneId int64, page, perPage int) (pagination.PageResponse[*client.CustomWafRule], error) {
	return m.listCustomWafRulesFn(ctx, shieldZoneId, page, perPage)
}

func (m *mockShieldAPI) GetCustomWafRule(ctx context.Context, id int64) (*client.CustomWafRule, error) {
	return m.getCustomWafRuleFn(ctx, id)
}

func (m *mockShieldAPI) CreateCustomWafRule(ctx context.Context, body *client.CustomWafRuleCreate) (*client.CustomWafRule, error) {
	return m.createCustomWafRuleFn(ctx, body)
}

func (m *mockShieldAPI) UpdateCustomWafRule(ctx context.Context, id int64, body *client.CustomWafRuleUpdate) (*client.CustomWafRule, error) {
	return m.updateCustomWafRuleFn(ctx, id, body)
}

func (m *mockShieldAPI) DeleteCustomWafRule(ctx context.Context, id int64) error {
	return m.deleteCustomWafRuleFn(ctx, id)
}

func (m *mockShieldAPI) ListWafProfiles(ctx context.Context) ([]*client.WafProfile, error) {
	return m.listWafProfilesFn(ctx)
}

func (m *mockShieldAPI) GetWafEngineConfig(ctx context.Context) ([]client.WafConfigVariable, error) {
	return m.getWafEngineConfigFn(ctx)
}

func (m *mockShieldAPI) ListTriggeredWafRules(ctx context.Context, shieldZoneId int64) ([]*client.TriggeredRule, error) {
	return m.listTriggeredWafRulesFn(ctx, shieldZoneId)
}

func (m *mockShieldAPI) UpdateTriggeredWafRule(ctx context.Context, shieldZoneId int64, body *client.TriggeredRuleUpdate) error {
	return m.updateTriggeredWafRuleFn(ctx, shieldZoneId, body)
}

func (m *mockShieldAPI) ListRateLimits(ctx context.Context, shieldZoneId int64, page, perPage int) (pagination.PageResponse[*client.RateLimitRule], error) {
	return m.listRateLimitsFn(ctx, shieldZoneId, page, perPage)
}

func (m *mockShieldAPI) GetRateLimit(ctx context.Context, id int64) (*client.RateLimitRule, error) {
	return m.getRateLimitFn(ctx, id)
}

func (m *mockShieldAPI) CreateRateLimit(ctx context.Context, body *client.RateLimitRuleCreate) (*client.RateLimitRule, error) {
	return m.createRateLimitFn(ctx, body)
}

func (m *mockShieldAPI) UpdateRateLimit(ctx context.Context, id int64, body *client.RateLimitRuleUpdate) (*client.RateLimitRule, error) {
	return m.updateRateLimitFn(ctx, id, body)
}

func (m *mockShieldAPI) DeleteRateLimit(ctx context.Context, id int64) error {
	return m.deleteRateLimitFn(ctx, id)
}

func (m *mockShieldAPI) ListAccessLists(ctx context.Context, shieldZoneId int64) (*client.AccessListsResponse, error) {
	return m.listAccessListsFn(ctx, shieldZoneId)
}

func (m *mockShieldAPI) GetCustomAccessList(ctx context.Context, shieldZoneId, id int64) (*client.CustomAccessList, error) {
	return m.getCustomAccessListFn(ctx, shieldZoneId, id)
}

func (m *mockShieldAPI) CreateCustomAccessList(ctx context.Context, shieldZoneId int64, body *client.CustomAccessListCreate) (*client.CustomAccessList, error) {
	return m.createCustomAccessListFn(ctx, shieldZoneId, body)
}

func (m *mockShieldAPI) UpdateCustomAccessList(ctx context.Context, shieldZoneId, id int64, body *client.CustomAccessListUpdate) (*client.CustomAccessList, error) {
	return m.updateCustomAccessListFn(ctx, shieldZoneId, id, body)
}

func (m *mockShieldAPI) DeleteCustomAccessList(ctx context.Context, shieldZoneId, id int64) error {
	return m.deleteCustomAccessListFn(ctx, shieldZoneId, id)
}

func (m *mockShieldAPI) UpdateAccessListConfig(ctx context.Context, shieldZoneId, configId int64, body *client.AccessListConfigUpdate) error {
	return m.updateAccessListConfigFn(ctx, shieldZoneId, configId, body)
}

func (m *mockShieldAPI) GetBotDetection(ctx context.Context, shieldZoneId int64) (*client.BotDetectionConfig, error) {
	return m.getBotDetectionFn(ctx, shieldZoneId)
}

func (m *mockShieldAPI) UpdateBotDetection(ctx context.Context, shieldZoneId int64, body *client.BotDetectionUpdate) error {
	return m.updateBotDetectionFn(ctx, shieldZoneId, body)
}

func (m *mockShieldAPI) GetUploadScanning(ctx context.Context, shieldZoneId int64) (*client.UploadScanningConfig, error) {
	return m.getUploadScanningFn(ctx, shieldZoneId)
}

func (m *mockShieldAPI) UpdateUploadScanning(ctx context.Context, shieldZoneId int64, body *client.UploadScanningUpdate) error {
	return m.updateUploadScanningFn(ctx, shieldZoneId, body)
}

func (m *mockShieldAPI) GetShieldMetricsOverview(ctx context.Context, shieldZoneId int64) (*client.ShieldZoneMetrics, error) {
	return m.getShieldMetricsOverviewFn(ctx, shieldZoneId)
}

func (m *mockShieldAPI) GetShieldMetricsDetailed(ctx context.Context, shieldZoneId int64, startDate, endDate string, resolution int) (*client.ShieldOverviewMetricsData, error) {
	return m.getShieldMetricsDetailedFn(ctx, shieldZoneId, startDate, endDate, resolution)
}

func (m *mockShieldAPI) GetShieldRateLimitMetrics(ctx context.Context, shieldZoneId int64) ([]*client.ShieldZoneRateLimitMetrics, error) {
	return m.getShieldRateLimitMetricsFn(ctx, shieldZoneId)
}

func (m *mockShieldAPI) GetShieldRateLimitMetric(ctx context.Context, id int64) (*client.RatelimitMetrics, error) {
	return m.getShieldRateLimitMetricFn(ctx, id)
}

func (m *mockShieldAPI) GetShieldWafRuleMetrics(ctx context.Context, shieldZoneId int64, ruleId int) (*client.WafRuleMetrics, error) {
	return m.getShieldWafRuleMetricsFn(ctx, shieldZoneId, ruleId)
}

func (m *mockShieldAPI) GetShieldBotDetectionMetrics(ctx context.Context, shieldZoneId int64) (*client.ShieldZoneBotDetectionMetrics, error) {
	return m.getShieldBotDetectionMetricsFn(ctx, shieldZoneId)
}

func (m *mockShieldAPI) GetShieldUploadScanningMetrics(ctx context.Context, shieldZoneId int64) (*client.ShieldZoneUploadScanningMetrics, error) {
	return m.getShieldUploadScanningMetricsFn(ctx, shieldZoneId)
}

func (m *mockShieldAPI) GetShieldEventLogs(ctx context.Context, shieldZoneId int64, date, continuationToken string) (*client.EventLogResponse, error) {
	return m.getShieldEventLogsFn(ctx, shieldZoneId, date, continuationToken)
}

func newTestShieldApp(api ShieldAPI) *App {
	return &App{NewShieldAPI: func(_ *cobra.Command) (ShieldAPI, error) { return api, nil }}
}

// --- Sample data helpers ---

func sampleShieldZone() *client.ShieldZone {
	return &client.ShieldZone{
		ShieldZoneId:          10,
		PullZoneId:            200,
		PlanType:              1,
		LearningMode:          true,
		WafEnabled:            true,
		WafExecutionMode:      1,
		WafProfileId:          5,
		DDoSEnabled:           true,
		DDoSShieldSensitivity: 2,
		DDoSExecutionMode:     1,
		DDoSChallengeWindow:   300,
		TotalWAFCustomRules:   3,
		TotalRateLimitRules:   2,
		CreatedDateTime:       "2024-01-01T00:00:00Z",
		LastModified:          "2024-06-15T12:00:00Z",
	}
}

func sampleShieldZoneResponse() *client.ShieldZoneResponse {
	return &client.ShieldZoneResponse{
		ShieldZoneId:          10,
		PullZoneId:            200,
		PlanType:              1,
		LearningMode:          true,
		WafEnabled:            true,
		WafExecutionMode:      1,
		WafProfileId:          5,
		DDoSShieldSensitivity: 2,
		DDoSExecutionMode:     1,
		RateLimitRulesLimit:   10,
		CustomWafRulesLimit:   20,
	}
}

func sampleCustomWafRule() *client.CustomWafRule {
	return &client.CustomWafRule{
		Id:              50,
		ShieldZoneId:    10,
		RuleName:        "Block SQL Injection",
		RuleDescription: "Blocks common SQL injection patterns",
		RuleJson:        `{"action":"block"}`,
	}
}

func sampleRateLimitRule() *client.RateLimitRule {
	return &client.RateLimitRule{
		Id:              60,
		ShieldZoneId:    10,
		RuleName:        "API Rate Limit",
		RuleDescription: "Limit API requests per minute",
		RuleJson:        `{"limit":100}`,
	}
}

func sampleAccessListDetails() client.AccessListDetails {
	return client.AccessListDetails{
		ListId:          70,
		ConfigurationId: 71,
		Name:            "IP Blocklist",
		IsEnabled:       true,
		Action:          1,
		Category:        1,
		EntryCount:      15,
	}
}

func sampleCustomAccessList() *client.CustomAccessList {
	return &client.CustomAccessList{
		Id:           80,
		Name:         "My IP List",
		Description:  "Custom IP allow list",
		Type:         0,
		Content:      "10.0.0.1\n10.0.0.2",
		Checksum:     "abc123",
		EntryCount:   2,
		LastModified: "2024-06-15T12:00:00Z",
	}
}

func sampleBotDetectionConfig() *client.BotDetectionConfig {
	return &client.BotDetectionConfig{
		ShieldZoneId:       10,
		ExecutionMode:      1,
		RequestIntegrity:   client.BotDetectionSensitivityConfig{Sensitivity: 2},
		IpAddress:          client.BotDetectionSensitivityConfig{Sensitivity: 1},
		BrowserFingerprint: client.BrowserFingerprintConfig{Sensitivity: 3, Aggression: 2},
	}
}

func sampleUploadScanningConfig() *client.UploadScanningConfig {
	return &client.UploadScanningConfig{
		ShieldZoneId:          10,
		IsEnabled:             true,
		AntivirusScanningMode: 1,
		CsamScanningMode:      1,
	}
}

func sampleShieldZoneMetrics() *client.ShieldZoneMetrics {
	return &client.ShieldZoneMetrics{
		Overview: &client.ShieldOverview{
			DDoSMitigated:          100,
			WafTriggeredRules:      50,
			RatelimitBreaches:      25,
			BotDetectionChallenged: 10,
			AccessListActions:      5,
			UploadScanningBlocks:   3,
		},
		TotalCleanRequestsLimit: 1000000,
		TotalBillableRequests:   500000,
	}
}

func sampleWafRuleMetrics() *client.WafRuleMetrics {
	return &client.WafRuleMetrics{
		TotalTriggers:      200,
		BlockedRequests:    150,
		LoggedRequests:     30,
		ChallengedRequests: 20,
	}
}

func sampleEventLogResponse() *client.EventLogResponse {
	return &client.EventLogResponse{
		Logs: []client.EventLog{
			{
				LogId:     "log-001",
				Timestamp: 1718452800,
				Log:       "WAF rule triggered",
				Labels:    map[string]string{"ruleId": "123"},
			},
		},
		HasMoreData:       false,
		ContinuationToken: "",
	}
}

// --- Shield help ---

func TestShield_ShowsInHelp(t *testing.T) {
	t.Parallel()
	out, _, err := executeCommand(nil, "shield", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, sub := range []string{"zones", "waf", "rate-limits", "access-lists", "bot-detection", "upload-scanning", "metrics", "event-logs"} {
		if !strings.Contains(out, sub) {
			t.Errorf("expected shield help to show %q subcommand", sub)
		}
	}
}

// --- Shield Zones ---

func TestShieldZones_ShowsInHelp(t *testing.T) {
	t.Parallel()
	out, _, err := executeCommand(nil, "shield", "zones", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, sub := range []string{"list", "get", "get-by-pullzone", "create", "update"} {
		if !strings.Contains(out, sub) {
			t.Errorf("expected shield zones help to show %q subcommand", sub)
		}
	}
}

func TestShieldZones_Alias(t *testing.T) {
	t.Parallel()
	out, _, err := executeCommand(nil, "shield", "zone", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Manage Shield zones") {
		t.Error("expected zone alias to work")
	}
}

func TestShieldZonesList_Table(t *testing.T) {
	t.Parallel()
	mock := &mockShieldAPI{
		listShieldZonesFn: func(_ context.Context, page, perPage int) (pagination.PageResponse[*client.ShieldZone], error) {
			return pagination.PageResponse[*client.ShieldZone]{
				Items:        []*client.ShieldZone{sampleShieldZone()},
				HasMoreItems: false,
			}, nil
		},
	}
	app := newTestShieldApp(mock)

	out, _, err := executeCommand(app, "shield", "zones", "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "10") {
		t.Error("expected output to contain shield zone ID")
	}
	if !strings.Contains(out, "200") {
		t.Error("expected output to contain pull zone ID")
	}
	if !strings.Contains(out, "Growth") {
		t.Error("expected output to contain plan name")
	}
}

func TestShieldZonesList_JSON(t *testing.T) {
	t.Parallel()
	mock := &mockShieldAPI{
		listShieldZonesFn: func(_ context.Context, page, perPage int) (pagination.PageResponse[*client.ShieldZone], error) {
			return pagination.PageResponse[*client.ShieldZone]{
				Items:        []*client.ShieldZone{sampleShieldZone()},
				HasMoreItems: false,
			}, nil
		},
	}
	app := newTestShieldApp(mock)

	out, _, err := executeCommand(app, "shield", "zones", "list", "--output", "json")
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

func TestShieldZonesGet_Table(t *testing.T) {
	t.Parallel()
	mock := &mockShieldAPI{
		getShieldZoneFn: func(_ context.Context, id int64) (*client.ShieldZone, error) {
			if id != 10 {
				t.Errorf("expected id=10, got %d", id)
			}
			return sampleShieldZone(), nil
		},
	}
	app := newTestShieldApp(mock)

	out, _, err := executeCommand(app, "shield", "zones", "get", "10")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "10") {
		t.Error("expected output to contain shield zone ID")
	}
	if !strings.Contains(out, "200") {
		t.Error("expected output to contain pull zone ID")
	}
	if !strings.Contains(out, "Growth") {
		t.Error("expected output to contain plan name")
	}
}

func TestShieldZonesGetByPullZone_Table(t *testing.T) {
	t.Parallel()
	var capturedPullZoneId int64
	mock := &mockShieldAPI{
		getShieldZoneByPullZoneFn: func(_ context.Context, pullZoneId int64) (*client.ShieldZone, error) {
			capturedPullZoneId = pullZoneId
			return sampleShieldZone(), nil
		},
	}
	app := newTestShieldApp(mock)

	out, _, err := executeCommand(app, "shield", "zones", "get-by-pullzone", "200")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedPullZoneId != 200 {
		t.Errorf("expected pullZoneId=200, got %d", capturedPullZoneId)
	}
	if !strings.Contains(out, "10") {
		t.Error("expected output to contain shield zone ID")
	}
}

func TestShieldZonesCreate_Success(t *testing.T) {
	t.Parallel()
	var capturedBody *client.ShieldZoneCreate
	mock := &mockShieldAPI{
		createShieldZoneFn: func(_ context.Context, body *client.ShieldZoneCreate) (*client.ShieldZoneResponse, error) {
			capturedBody = body
			return sampleShieldZoneResponse(), nil
		},
	}
	app := newTestShieldApp(mock)

	out, _, err := executeCommand(app, "shield", "zones", "create", "--pull-zone-id", "200")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedBody.PullZoneId != 200 {
		t.Errorf("expected pull zone ID=200, got %d", capturedBody.PullZoneId)
	}
	if !strings.Contains(out, "10") {
		t.Error("expected output to contain shield zone ID")
	}
}

func TestShieldZonesUpdate_Success(t *testing.T) {
	t.Parallel()
	var capturedId int64
	var capturedBody *client.ShieldZoneUpdate
	mock := &mockShieldAPI{
		updateShieldZoneFn: func(_ context.Context, id int64, body *client.ShieldZoneUpdate) (*client.ShieldZoneResponse, error) {
			capturedId = id
			capturedBody = body
			return sampleShieldZoneResponse(), nil
		},
	}
	app := newTestShieldApp(mock)

	out, _, err := executeCommand(app, "shield", "zones", "update", "10", "--waf-enabled", "--learning-mode")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedId != 10 {
		t.Errorf("expected id=10, got %d", capturedId)
	}
	if capturedBody.WafEnabled == nil || !*capturedBody.WafEnabled {
		t.Error("expected waf-enabled to be set in body")
	}
	if capturedBody.LearningMode == nil || !*capturedBody.LearningMode {
		t.Error("expected learning-mode to be set in body")
	}
	if !strings.Contains(out, "10") {
		t.Error("expected output to contain shield zone ID")
	}
}

// --- Shield WAF ---

func TestShieldWaf_ShowsInHelp(t *testing.T) {
	t.Parallel()
	out, _, err := executeCommand(nil, "shield", "waf", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, sub := range []string{"rules", "custom-rules", "profiles", "engine", "triggered"} {
		if !strings.Contains(out, sub) {
			t.Errorf("expected shield waf help to show %q subcommand", sub)
		}
	}
}

func TestShieldWafCustomRulesList_Table(t *testing.T) {
	t.Parallel()
	mock := &mockShieldAPI{
		listCustomWafRulesFn: func(_ context.Context, shieldZoneId int64, page, perPage int) (pagination.PageResponse[*client.CustomWafRule], error) {
			if shieldZoneId != 10 {
				t.Errorf("expected shieldZoneId=10, got %d", shieldZoneId)
			}
			return pagination.PageResponse[*client.CustomWafRule]{
				Items:        []*client.CustomWafRule{sampleCustomWafRule()},
				HasMoreItems: false,
			}, nil
		},
	}
	app := newTestShieldApp(mock)

	out, _, err := executeCommand(app, "shield", "waf", "custom-rules", "list", "10")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Block SQL Injection") {
		t.Error("expected output to contain rule name")
	}
	if !strings.Contains(out, "50") {
		t.Error("expected output to contain rule ID")
	}
}

func TestShieldWafCustomRulesGet_Table(t *testing.T) {
	t.Parallel()
	mock := &mockShieldAPI{
		getCustomWafRuleFn: func(_ context.Context, id int64) (*client.CustomWafRule, error) {
			if id != 50 {
				t.Errorf("expected id=50, got %d", id)
			}
			return sampleCustomWafRule(), nil
		},
	}
	app := newTestShieldApp(mock)

	out, _, err := executeCommand(app, "shield", "waf", "custom-rules", "get", "50")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Block SQL Injection") {
		t.Error("expected output to contain rule name")
	}
	if !strings.Contains(out, "Blocks common SQL injection patterns") {
		t.Error("expected output to contain rule description")
	}
}

func TestShieldWafCustomRulesCreate_Success(t *testing.T) {
	t.Parallel()
	var capturedBody *client.CustomWafRuleCreate
	mock := &mockShieldAPI{
		createCustomWafRuleFn: func(_ context.Context, body *client.CustomWafRuleCreate) (*client.CustomWafRule, error) {
			capturedBody = body
			return sampleCustomWafRule(), nil
		},
	}
	app := newTestShieldApp(mock)

	out, _, err := executeCommand(app, "shield", "waf", "custom-rules", "create",
		"--shield-zone-id", "10",
		"--rule-name", "Block SQL Injection",
		"--rule-description", "Blocks common SQL injection patterns",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedBody.ShieldZoneId != 10 {
		t.Errorf("expected shield zone ID=10, got %d", capturedBody.ShieldZoneId)
	}
	if capturedBody.RuleName != "Block SQL Injection" {
		t.Errorf("expected rule name 'Block SQL Injection', got %q", capturedBody.RuleName)
	}
	if !strings.Contains(out, "Block SQL Injection") {
		t.Error("expected output to contain rule name")
	}
}

func TestShieldWafCustomRulesDelete_WithYes(t *testing.T) {
	t.Parallel()
	var deletedId int64
	mock := &mockShieldAPI{
		deleteCustomWafRuleFn: func(_ context.Context, id int64) error {
			deletedId = id
			return nil
		},
	}
	app := newTestShieldApp(mock)

	out, _, err := executeCommand(app, "shield", "waf", "custom-rules", "delete", "50", "--yes")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deletedId != 50 {
		t.Errorf("expected deleted id=50, got %d", deletedId)
	}
	if !strings.Contains(out, "Custom WAF rule deleted") {
		t.Error("expected deletion confirmation message")
	}
}

func TestShieldWafCustomRulesDelete_WithoutYes_Canceled(t *testing.T) {
	t.Parallel()
	mock := &mockShieldAPI{
		deleteCustomWafRuleFn: func(_ context.Context, id int64) error {
			t.Error("delete should not have been called")
			return nil
		},
	}
	app := newTestShieldApp(mock)

	stdin := bytes.NewBufferString("n\n")
	_, stderr, err := executeCommandWithStdin(app, stdin, "shield", "waf", "custom-rules", "delete", "50")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "Deletion canceled") {
		t.Error("expected cancellation message")
	}
}

func TestShieldWafProfilesList_Table(t *testing.T) {
	t.Parallel()
	mock := &mockShieldAPI{
		listWafProfilesFn: func(_ context.Context) ([]*client.WafProfile, error) {
			return []*client.WafProfile{
				{Id: 1, Name: "OWASP Core", IsPremium: false, ProfileCategory: "General", Description: "OWASP core ruleset"},
			}, nil
		},
	}
	app := newTestShieldApp(mock)

	out, _, err := executeCommand(app, "shield", "waf", "profiles")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "OWASP Core") {
		t.Error("expected output to contain profile name")
	}
	if !strings.Contains(out, "General") {
		t.Error("expected output to contain profile category")
	}
}

func TestShieldWafEngine_Table(t *testing.T) {
	t.Parallel()
	mock := &mockShieldAPI{
		getWafEngineConfigFn: func(_ context.Context) ([]client.WafConfigVariable, error) {
			return []client.WafConfigVariable{
				{Name: "max_body_size", ValueEncoded: "1048576"},
				{Name: "paranoia_level", ValueEncoded: "2"},
			}, nil
		},
	}
	app := newTestShieldApp(mock)

	out, _, err := executeCommand(app, "shield", "waf", "engine")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "max_body_size") {
		t.Error("expected output to contain config variable name")
	}
	if !strings.Contains(out, "1048576") {
		t.Error("expected output to contain config variable value")
	}
}

func TestShieldWafTriggeredList_Table(t *testing.T) {
	t.Parallel()
	mock := &mockShieldAPI{
		listTriggeredWafRulesFn: func(_ context.Context, shieldZoneId int64) ([]*client.TriggeredRule, error) {
			if shieldZoneId != 10 {
				t.Errorf("expected shieldZoneId=10, got %d", shieldZoneId)
			}
			return []*client.TriggeredRule{
				{RuleId: "999", RuleDescription: "XSS Attack Detected", TotalTriggeredRequests: 42},
			}, nil
		},
	}
	app := newTestShieldApp(mock)

	out, _, err := executeCommand(app, "shield", "waf", "triggered", "list", "10")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "999") {
		t.Error("expected output to contain rule ID")
	}
	if !strings.Contains(out, "XSS Attack Detected") {
		t.Error("expected output to contain rule description")
	}
	if !strings.Contains(out, "42") {
		t.Error("expected output to contain triggered count")
	}
}

// --- Shield Rate Limits ---

func TestShieldRateLimitsList_Table(t *testing.T) {
	t.Parallel()
	mock := &mockShieldAPI{
		listRateLimitsFn: func(_ context.Context, shieldZoneId int64, page, perPage int) (pagination.PageResponse[*client.RateLimitRule], error) {
			if shieldZoneId != 10 {
				t.Errorf("expected shieldZoneId=10, got %d", shieldZoneId)
			}
			return pagination.PageResponse[*client.RateLimitRule]{
				Items:        []*client.RateLimitRule{sampleRateLimitRule()},
				HasMoreItems: false,
			}, nil
		},
	}
	app := newTestShieldApp(mock)

	out, _, err := executeCommand(app, "shield", "rate-limits", "list", "10")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "API Rate Limit") {
		t.Error("expected output to contain rule name")
	}
	if !strings.Contains(out, "60") {
		t.Error("expected output to contain rule ID")
	}
}

func TestShieldRateLimitsGet_Table(t *testing.T) {
	t.Parallel()
	mock := &mockShieldAPI{
		getRateLimitFn: func(_ context.Context, id int64) (*client.RateLimitRule, error) {
			if id != 60 {
				t.Errorf("expected id=60, got %d", id)
			}
			return sampleRateLimitRule(), nil
		},
	}
	app := newTestShieldApp(mock)

	out, _, err := executeCommand(app, "shield", "rate-limits", "get", "60")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "API Rate Limit") {
		t.Error("expected output to contain rule name")
	}
	if !strings.Contains(out, "Limit API requests per minute") {
		t.Error("expected output to contain rule description")
	}
}

func TestShieldRateLimitsCreate_Success(t *testing.T) {
	t.Parallel()
	var capturedBody *client.RateLimitRuleCreate
	mock := &mockShieldAPI{
		createRateLimitFn: func(_ context.Context, body *client.RateLimitRuleCreate) (*client.RateLimitRule, error) {
			capturedBody = body
			return sampleRateLimitRule(), nil
		},
	}
	app := newTestShieldApp(mock)

	out, _, err := executeCommand(app, "shield", "rate-limits", "create",
		"--shield-zone-id", "10",
		"--rule-name", "API Rate Limit",
		"--rule-description", "Limit API requests per minute",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedBody.ShieldZoneId != 10 {
		t.Errorf("expected shield zone ID=10, got %d", capturedBody.ShieldZoneId)
	}
	if capturedBody.RuleName != "API Rate Limit" {
		t.Errorf("expected rule name 'API Rate Limit', got %q", capturedBody.RuleName)
	}
	if !strings.Contains(out, "API Rate Limit") {
		t.Error("expected output to contain rule name")
	}
}

func TestShieldRateLimitsDelete_WithYes(t *testing.T) {
	t.Parallel()
	var deletedId int64
	mock := &mockShieldAPI{
		deleteRateLimitFn: func(_ context.Context, id int64) error {
			deletedId = id
			return nil
		},
	}
	app := newTestShieldApp(mock)

	out, _, err := executeCommand(app, "shield", "rate-limits", "delete", "60", "--yes")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deletedId != 60 {
		t.Errorf("expected deleted id=60, got %d", deletedId)
	}
	if !strings.Contains(out, "Rate limit rule deleted") {
		t.Error("expected deletion confirmation message")
	}
}

// --- Shield Access Lists ---

func TestShieldAccessListsList_Table(t *testing.T) {
	t.Parallel()
	mock := &mockShieldAPI{
		listAccessListsFn: func(_ context.Context, shieldZoneId int64) (*client.AccessListsResponse, error) {
			if shieldZoneId != 10 {
				t.Errorf("expected shieldZoneId=10, got %d", shieldZoneId)
			}
			return &client.AccessListsResponse{
				ManagedLists: []client.AccessListDetails{},
				CustomLists:  []client.AccessListDetails{sampleAccessListDetails()},
			}, nil
		},
	}
	app := newTestShieldApp(mock)

	out, _, err := executeCommand(app, "shield", "access-lists", "list", "10")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "IP Blocklist") {
		t.Error("expected output to contain access list name")
	}
	if !strings.Contains(out, "70") {
		t.Error("expected output to contain list ID")
	}
}

func TestShieldAccessListsGet_Table(t *testing.T) {
	t.Parallel()
	var capturedShieldZoneId, capturedId int64
	mock := &mockShieldAPI{
		getCustomAccessListFn: func(_ context.Context, shieldZoneId, id int64) (*client.CustomAccessList, error) {
			capturedShieldZoneId = shieldZoneId
			capturedId = id
			return sampleCustomAccessList(), nil
		},
	}
	app := newTestShieldApp(mock)

	out, _, err := executeCommand(app, "shield", "access-lists", "get", "10", "80")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedShieldZoneId != 10 {
		t.Errorf("expected shieldZoneId=10, got %d", capturedShieldZoneId)
	}
	if capturedId != 80 {
		t.Errorf("expected id=80, got %d", capturedId)
	}
	if !strings.Contains(out, "My IP List") {
		t.Error("expected output to contain access list name")
	}
	if !strings.Contains(out, "Custom IP allow list") {
		t.Error("expected output to contain description")
	}
}

func TestShieldAccessListsCreate_Success(t *testing.T) {
	t.Parallel()
	var capturedShieldZoneId int64
	var capturedBody *client.CustomAccessListCreate
	mock := &mockShieldAPI{
		createCustomAccessListFn: func(_ context.Context, shieldZoneId int64, body *client.CustomAccessListCreate) (*client.CustomAccessList, error) {
			capturedShieldZoneId = shieldZoneId
			capturedBody = body
			return sampleCustomAccessList(), nil
		},
	}
	app := newTestShieldApp(mock)

	out, _, err := executeCommand(app, "shield", "access-lists", "create", "10",
		"--name", "My IP List",
		"--content", "10.0.0.1\n10.0.0.2",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedShieldZoneId != 10 {
		t.Errorf("expected shieldZoneId=10, got %d", capturedShieldZoneId)
	}
	if capturedBody.Name != "My IP List" {
		t.Errorf("expected name 'My IP List', got %q", capturedBody.Name)
	}
	if !strings.Contains(out, "My IP List") {
		t.Error("expected output to contain access list name")
	}
}

func TestShieldAccessListsDelete_WithYes(t *testing.T) {
	t.Parallel()
	var deletedShieldZoneId, deletedId int64
	mock := &mockShieldAPI{
		deleteCustomAccessListFn: func(_ context.Context, shieldZoneId, id int64) error {
			deletedShieldZoneId = shieldZoneId
			deletedId = id
			return nil
		},
	}
	app := newTestShieldApp(mock)

	out, _, err := executeCommand(app, "shield", "access-lists", "delete", "10", "80", "--yes")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deletedShieldZoneId != 10 {
		t.Errorf("expected shieldZoneId=10, got %d", deletedShieldZoneId)
	}
	if deletedId != 80 {
		t.Errorf("expected deleted id=80, got %d", deletedId)
	}
	if !strings.Contains(out, "Access list deleted") {
		t.Error("expected deletion confirmation message")
	}
}

func TestShieldAccessListsConfigUpdate_Success(t *testing.T) {
	t.Parallel()
	var capturedShieldZoneId, capturedConfigId int64
	var capturedBody *client.AccessListConfigUpdate
	mock := &mockShieldAPI{
		updateAccessListConfigFn: func(_ context.Context, shieldZoneId, configId int64, body *client.AccessListConfigUpdate) error {
			capturedShieldZoneId = shieldZoneId
			capturedConfigId = configId
			capturedBody = body
			return nil
		},
	}
	app := newTestShieldApp(mock)

	out, _, err := executeCommand(app, "shield", "access-lists", "config", "update", "10", "71", "--enabled", "--action", "1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedShieldZoneId != 10 {
		t.Errorf("expected shieldZoneId=10, got %d", capturedShieldZoneId)
	}
	if capturedConfigId != 71 {
		t.Errorf("expected configId=71, got %d", capturedConfigId)
	}
	if capturedBody.IsEnabled == nil || !*capturedBody.IsEnabled {
		t.Error("expected enabled to be set in body")
	}
	if capturedBody.Action == nil || *capturedBody.Action != 1 {
		t.Error("expected action=1 in body")
	}
	if !strings.Contains(out, "Access list configuration updated") {
		t.Error("expected update confirmation message")
	}
}

// --- Shield Bot Detection ---

func TestShieldBotDetectionGet_Table(t *testing.T) {
	t.Parallel()
	mock := &mockShieldAPI{
		getBotDetectionFn: func(_ context.Context, shieldZoneId int64) (*client.BotDetectionConfig, error) {
			if shieldZoneId != 10 {
				t.Errorf("expected shieldZoneId=10, got %d", shieldZoneId)
			}
			return sampleBotDetectionConfig(), nil
		},
	}
	app := newTestShieldApp(mock)

	out, _, err := executeCommand(app, "shield", "bot-detection", "get", "10")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "10") {
		t.Error("expected output to contain shield zone ID")
	}
	if !strings.Contains(out, "Protect") {
		t.Error("expected output to contain execution mode name")
	}
}

func TestShieldBotDetectionUpdate_Success(t *testing.T) {
	t.Parallel()
	var capturedShieldZoneId int64
	var capturedBody *client.BotDetectionUpdate
	mock := &mockShieldAPI{
		updateBotDetectionFn: func(_ context.Context, shieldZoneId int64, body *client.BotDetectionUpdate) error {
			capturedShieldZoneId = shieldZoneId
			capturedBody = body
			return nil
		},
	}
	app := newTestShieldApp(mock)

	out, _, err := executeCommand(app, "shield", "bot-detection", "update", "10", "--execution-mode", "1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedShieldZoneId != 10 {
		t.Errorf("expected shieldZoneId=10, got %d", capturedShieldZoneId)
	}
	if capturedBody.ExecutionMode == nil || *capturedBody.ExecutionMode != 1 {
		t.Error("expected execution-mode=1 in body")
	}
	if !strings.Contains(out, "Bot detection configuration updated") {
		t.Error("expected update confirmation message")
	}
}

// --- Shield Upload Scanning ---

func TestShieldUploadScanningGet_Table(t *testing.T) {
	t.Parallel()
	mock := &mockShieldAPI{
		getUploadScanningFn: func(_ context.Context, shieldZoneId int64) (*client.UploadScanningConfig, error) {
			if shieldZoneId != 10 {
				t.Errorf("expected shieldZoneId=10, got %d", shieldZoneId)
			}
			return sampleUploadScanningConfig(), nil
		},
	}
	app := newTestShieldApp(mock)

	out, _, err := executeCommand(app, "shield", "upload-scanning", "get", "10")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "10") {
		t.Error("expected output to contain shield zone ID")
	}
	if !strings.Contains(out, "true") {
		t.Error("expected output to contain enabled status")
	}
}

func TestShieldUploadScanningUpdate_Success(t *testing.T) {
	t.Parallel()
	var capturedShieldZoneId int64
	var capturedBody *client.UploadScanningUpdate
	mock := &mockShieldAPI{
		updateUploadScanningFn: func(_ context.Context, shieldZoneId int64, body *client.UploadScanningUpdate) error {
			capturedShieldZoneId = shieldZoneId
			capturedBody = body
			return nil
		},
	}
	app := newTestShieldApp(mock)

	out, _, err := executeCommand(app, "shield", "upload-scanning", "update", "10", "--enabled", "--antivirus-scanning-mode", "1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedShieldZoneId != 10 {
		t.Errorf("expected shieldZoneId=10, got %d", capturedShieldZoneId)
	}
	if capturedBody.IsEnabled == nil || !*capturedBody.IsEnabled {
		t.Error("expected enabled to be set in body")
	}
	if capturedBody.AntivirusScanningMode == nil || *capturedBody.AntivirusScanningMode != 1 {
		t.Error("expected antivirus-scanning-mode=1 in body")
	}
	if !strings.Contains(out, "Upload scanning configuration updated") {
		t.Error("expected update confirmation message")
	}
}

// --- Shield Metrics ---

func TestShieldMetricsOverview_Table(t *testing.T) {
	t.Parallel()
	mock := &mockShieldAPI{
		getShieldMetricsOverviewFn: func(_ context.Context, shieldZoneId int64) (*client.ShieldZoneMetrics, error) {
			if shieldZoneId != 10 {
				t.Errorf("expected shieldZoneId=10, got %d", shieldZoneId)
			}
			return sampleShieldZoneMetrics(), nil
		},
	}
	app := newTestShieldApp(mock)

	out, _, err := executeCommand(app, "shield", "metrics", "overview", "10")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "100") {
		t.Error("expected output to contain DDoS mitigated count")
	}
	if !strings.Contains(out, "50") {
		t.Error("expected output to contain WAF triggered count")
	}
	if !strings.Contains(out, "1000000") {
		t.Error("expected output to contain clean requests limit")
	}
}

func TestShieldMetricsDetailed_Table(t *testing.T) {
	t.Parallel()
	var capturedStartDate, capturedEndDate string
	var capturedResolution int
	mock := &mockShieldAPI{
		getShieldMetricsDetailedFn: func(_ context.Context, shieldZoneId int64, startDate, endDate string, resolution int) (*client.ShieldOverviewMetricsData, error) {
			capturedStartDate = startDate
			capturedEndDate = endDate
			capturedResolution = resolution
			return &client.ShieldOverviewMetricsData{
				TotalCleanRequestsLimit:        2000000,
				TotalBillableRequestsThisMonth: 750000,
				Resolution:                     1,
			}, nil
		},
	}
	app := newTestShieldApp(mock)

	out, _, err := executeCommand(app, "shield", "metrics", "detailed", "10",
		"--start-date", "2024-01-01",
		"--end-date", "2024-01-31",
		"--resolution", "1",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedStartDate != "2024-01-01" {
		t.Errorf("expected start-date='2024-01-01', got %q", capturedStartDate)
	}
	if capturedEndDate != "2024-01-31" {
		t.Errorf("expected end-date='2024-01-31', got %q", capturedEndDate)
	}
	if capturedResolution != 1 {
		t.Errorf("expected resolution=1, got %d", capturedResolution)
	}
	if !strings.Contains(out, "2000000") {
		t.Error("expected output to contain clean requests limit")
	}
	if !strings.Contains(out, "750000") {
		t.Error("expected output to contain billable requests")
	}
}

func TestShieldMetricsRateLimits_Table(t *testing.T) {
	t.Parallel()
	mock := &mockShieldAPI{
		getShieldRateLimitMetricsFn: func(_ context.Context, shieldZoneId int64) ([]*client.ShieldZoneRateLimitMetrics, error) {
			if shieldZoneId != 10 {
				t.Errorf("expected shieldZoneId=10, got %d", shieldZoneId)
			}
			return []*client.ShieldZoneRateLimitMetrics{
				{
					RatelimitId: 60,
					Overview: &client.RatelimitMetrics{
						TotalBreaches:      100,
						LoggedBreaches:     40,
						ChallengedBreaches: 30,
						BlockedBreaches:    30,
					},
				},
			}, nil
		},
	}
	app := newTestShieldApp(mock)

	out, _, err := executeCommand(app, "shield", "metrics", "rate-limits", "10")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "60") {
		t.Error("expected output to contain rate limit ID")
	}
	if !strings.Contains(out, "100") {
		t.Error("expected output to contain total breaches")
	}
}

func TestShieldMetricsWafRule_Table(t *testing.T) {
	t.Parallel()
	var capturedShieldZoneId int64
	var capturedRuleId int
	mock := &mockShieldAPI{
		getShieldWafRuleMetricsFn: func(_ context.Context, shieldZoneId int64, ruleId int) (*client.WafRuleMetrics, error) {
			capturedShieldZoneId = shieldZoneId
			capturedRuleId = ruleId
			return sampleWafRuleMetrics(), nil
		},
	}
	app := newTestShieldApp(mock)

	out, _, err := executeCommand(app, "shield", "metrics", "waf-rule", "10", "999")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedShieldZoneId != 10 {
		t.Errorf("expected shieldZoneId=10, got %d", capturedShieldZoneId)
	}
	if capturedRuleId != 999 {
		t.Errorf("expected ruleId=999, got %d", capturedRuleId)
	}
	if !strings.Contains(out, "200") {
		t.Error("expected output to contain total triggers")
	}
	if !strings.Contains(out, "150") {
		t.Error("expected output to contain blocked requests")
	}
}

func TestShieldMetricsBotDetection_Table(t *testing.T) {
	t.Parallel()
	mock := &mockShieldAPI{
		getShieldBotDetectionMetricsFn: func(_ context.Context, shieldZoneId int64) (*client.ShieldZoneBotDetectionMetrics, error) {
			if shieldZoneId != 10 {
				t.Errorf("expected shieldZoneId=10, got %d", shieldZoneId)
			}
			return &client.ShieldZoneBotDetectionMetrics{
				TotalLoggedRequests:     500,
				TotalChallengedRequests: 150,
			}, nil
		},
	}
	app := newTestShieldApp(mock)

	out, _, err := executeCommand(app, "shield", "metrics", "bot-detection", "10")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "500") {
		t.Error("expected output to contain logged requests")
	}
	if !strings.Contains(out, "150") {
		t.Error("expected output to contain challenged requests")
	}
}

func TestShieldMetricsUploadScanning_Table(t *testing.T) {
	t.Parallel()
	mock := &mockShieldAPI{
		getShieldUploadScanningMetricsFn: func(_ context.Context, shieldZoneId int64) (*client.ShieldZoneUploadScanningMetrics, error) {
			if shieldZoneId != 10 {
				t.Errorf("expected shieldZoneId=10, got %d", shieldZoneId)
			}
			return &client.ShieldZoneUploadScanningMetrics{
				TotalLoggedRequests:  300,
				TotalBlockedRequests: 50,
				TotalFilesScanned:    1000,
			}, nil
		},
	}
	app := newTestShieldApp(mock)

	out, _, err := executeCommand(app, "shield", "metrics", "upload-scanning", "10")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "300") {
		t.Error("expected output to contain logged requests")
	}
	if !strings.Contains(out, "50") {
		t.Error("expected output to contain blocked requests")
	}
	if !strings.Contains(out, "1000") {
		t.Error("expected output to contain files scanned")
	}
}

// --- Shield Event Logs ---

func TestShieldEventLogs_Table(t *testing.T) {
	t.Parallel()
	var capturedShieldZoneId int64
	var capturedDate string
	mock := &mockShieldAPI{
		getShieldEventLogsFn: func(_ context.Context, shieldZoneId int64, date, continuationToken string) (*client.EventLogResponse, error) {
			capturedShieldZoneId = shieldZoneId
			capturedDate = date
			return sampleEventLogResponse(), nil
		},
	}
	app := newTestShieldApp(mock)

	out, _, err := executeCommand(app, "shield", "event-logs", "10", "2024-06-15")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedShieldZoneId != 10 {
		t.Errorf("expected shieldZoneId=10, got %d", capturedShieldZoneId)
	}
	if capturedDate != "2024-06-15" {
		t.Errorf("expected date='2024-06-15', got %q", capturedDate)
	}
	if !strings.Contains(out, "log-001") {
		t.Error("expected output to contain log ID")
	}
	if !strings.Contains(out, "WAF rule triggered") {
		t.Error("expected output to contain log message")
	}
}
