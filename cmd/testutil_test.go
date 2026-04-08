package cmd

import (
	"context"

	"github.com/built-fast/bunny-cli/internal/client"
	"github.com/built-fast/bunny-cli/internal/pagination"
	"github.com/spf13/cobra"
)

// mockPullZoneAPI implements PullZoneAPI for testing.
type mockPullZoneAPI struct {
	listPullZonesFn          func(ctx context.Context, page, perPage int, search string) (pagination.PageResponse[*client.PullZone], error)
	getPullZoneFn            func(ctx context.Context, id int64) (*client.PullZone, error)
	createPullZoneFn         func(ctx context.Context, body *client.PullZoneCreate) (*client.PullZone, error)
	updatePullZoneFn         func(ctx context.Context, id int64, body *client.PullZoneUpdate) (*client.PullZone, error)
	deletePullZoneFn         func(ctx context.Context, id int64) error
	addPullZoneHostnameFn    func(ctx context.Context, id int64, hostname string) error
	removePullZoneHostnameFn func(ctx context.Context, id int64, hostname string) error
	purgePullZoneCacheFn     func(ctx context.Context, id int64, cacheTag string) error
	addOrUpdateEdgeRuleFn    func(ctx context.Context, pullZoneId int64, rule *client.EdgeRule) error
	deleteEdgeRuleFn         func(ctx context.Context, pullZoneId int64, edgeRuleId string) error
	setEdgeRuleEnabledFn     func(ctx context.Context, pullZoneId int64, edgeRuleId string, enabled bool) error
}

func (m *mockPullZoneAPI) ListPullZones(ctx context.Context, page, perPage int, search string) (pagination.PageResponse[*client.PullZone], error) {
	return m.listPullZonesFn(ctx, page, perPage, search)
}

func (m *mockPullZoneAPI) GetPullZone(ctx context.Context, id int64) (*client.PullZone, error) {
	return m.getPullZoneFn(ctx, id)
}

func (m *mockPullZoneAPI) CreatePullZone(ctx context.Context, body *client.PullZoneCreate) (*client.PullZone, error) {
	return m.createPullZoneFn(ctx, body)
}

func (m *mockPullZoneAPI) UpdatePullZone(ctx context.Context, id int64, body *client.PullZoneUpdate) (*client.PullZone, error) {
	return m.updatePullZoneFn(ctx, id, body)
}

func (m *mockPullZoneAPI) DeletePullZone(ctx context.Context, id int64) error {
	return m.deletePullZoneFn(ctx, id)
}

func (m *mockPullZoneAPI) AddPullZoneHostname(ctx context.Context, id int64, hostname string) error {
	return m.addPullZoneHostnameFn(ctx, id, hostname)
}

func (m *mockPullZoneAPI) RemovePullZoneHostname(ctx context.Context, id int64, hostname string) error {
	return m.removePullZoneHostnameFn(ctx, id, hostname)
}

func (m *mockPullZoneAPI) PurgePullZoneCache(ctx context.Context, id int64, cacheTag string) error {
	return m.purgePullZoneCacheFn(ctx, id, cacheTag)
}

func (m *mockPullZoneAPI) AddOrUpdateEdgeRule(ctx context.Context, pullZoneId int64, rule *client.EdgeRule) error {
	return m.addOrUpdateEdgeRuleFn(ctx, pullZoneId, rule)
}

func (m *mockPullZoneAPI) DeleteEdgeRule(ctx context.Context, pullZoneId int64, edgeRuleId string) error {
	return m.deleteEdgeRuleFn(ctx, pullZoneId, edgeRuleId)
}

func (m *mockPullZoneAPI) SetEdgeRuleEnabled(ctx context.Context, pullZoneId int64, edgeRuleId string, enabled bool) error {
	return m.setEdgeRuleEnabledFn(ctx, pullZoneId, edgeRuleId, enabled)
}

func newTestPullZoneApp(api PullZoneAPI) *App {
	return &App{NewPullZoneAPI: func(_ *cobra.Command) (PullZoneAPI, error) { return api, nil }}
}

// samplePullZone returns a PullZone for use in tests.
func samplePullZone() *client.PullZone {
	return &client.PullZone{
		Id:                   42,
		Name:                 "my-zone",
		OriginUrl:            "https://origin.example.com",
		Enabled:              true,
		CnameDomain:          "my-zone.b-cdn.net",
		MonthlyBandwidthUsed: 1024000,
		MonthlyCharges:       4.99,
		Type:                 0,
		Hostnames: []client.Hostname{
			{Id: 1, Value: "cdn.example.com", ForceSSL: true, HasCertificate: true},
			{Id: 2, Value: "my-zone.b-cdn.net", IsSystemHostname: true},
		},
		EdgeRules: []client.EdgeRule{
			{Guid: "rule-1", Description: "Force HTTPS", ActionType: 0, Enabled: true},
		},
	}
}
