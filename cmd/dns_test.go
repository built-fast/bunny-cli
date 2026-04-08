package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/built-fast/bunny-cli/internal/client"
	"github.com/built-fast/bunny-cli/internal/pagination"
	"github.com/spf13/cobra"
)

// mockDnsZoneAPI implements DnsZoneAPI for testing.
type mockDnsZoneAPI struct {
	listDnsZonesFn    func(ctx context.Context, page, perPage int, search string) (pagination.PageResponse[*client.DnsZone], error)
	getDnsZoneFn      func(ctx context.Context, id int64) (*client.DnsZone, error)
	createDnsZoneFn   func(ctx context.Context, body *client.DnsZoneCreate) (*client.DnsZone, error)
	updateDnsZoneFn   func(ctx context.Context, id int64, body *client.DnsZoneUpdate) (*client.DnsZone, error)
	deleteDnsZoneFn   func(ctx context.Context, id int64) error
	addDnsRecordFn    func(ctx context.Context, zoneId int64, body *client.DnsRecordCreate) (*client.DnsRecord, error)
	updateDnsRecordFn func(ctx context.Context, zoneId, recordId int64, body *client.DnsRecordUpdate) error
	deleteDnsRecordFn func(ctx context.Context, zoneId, recordId int64) error
	importDnsZoneFn   func(ctx context.Context, zoneId int64, data io.Reader) (*client.DnsZoneImportResult, error)
	exportDnsZoneFn   func(ctx context.Context, zoneId int64) ([]byte, error)
	enableDnsSecFn    func(ctx context.Context, zoneId int64) (*client.DnsSecInfo, error)
	disableDnsSecFn   func(ctx context.Context, zoneId int64) (*client.DnsSecInfo, error)
}

func (m *mockDnsZoneAPI) ListDnsZones(ctx context.Context, page, perPage int, search string) (pagination.PageResponse[*client.DnsZone], error) {
	return m.listDnsZonesFn(ctx, page, perPage, search)
}

func (m *mockDnsZoneAPI) GetDnsZone(ctx context.Context, id int64) (*client.DnsZone, error) {
	return m.getDnsZoneFn(ctx, id)
}

func (m *mockDnsZoneAPI) CreateDnsZone(ctx context.Context, body *client.DnsZoneCreate) (*client.DnsZone, error) {
	return m.createDnsZoneFn(ctx, body)
}

func (m *mockDnsZoneAPI) UpdateDnsZone(ctx context.Context, id int64, body *client.DnsZoneUpdate) (*client.DnsZone, error) {
	return m.updateDnsZoneFn(ctx, id, body)
}

func (m *mockDnsZoneAPI) DeleteDnsZone(ctx context.Context, id int64) error {
	return m.deleteDnsZoneFn(ctx, id)
}

func (m *mockDnsZoneAPI) AddDnsRecord(ctx context.Context, zoneId int64, body *client.DnsRecordCreate) (*client.DnsRecord, error) {
	return m.addDnsRecordFn(ctx, zoneId, body)
}

func (m *mockDnsZoneAPI) UpdateDnsRecord(ctx context.Context, zoneId, recordId int64, body *client.DnsRecordUpdate) error {
	return m.updateDnsRecordFn(ctx, zoneId, recordId, body)
}

func (m *mockDnsZoneAPI) DeleteDnsRecord(ctx context.Context, zoneId, recordId int64) error {
	return m.deleteDnsRecordFn(ctx, zoneId, recordId)
}

func (m *mockDnsZoneAPI) ImportDnsZone(ctx context.Context, zoneId int64, data io.Reader) (*client.DnsZoneImportResult, error) {
	return m.importDnsZoneFn(ctx, zoneId, data)
}

func (m *mockDnsZoneAPI) ExportDnsZone(ctx context.Context, zoneId int64) ([]byte, error) {
	return m.exportDnsZoneFn(ctx, zoneId)
}

func (m *mockDnsZoneAPI) EnableDnsSec(ctx context.Context, zoneId int64) (*client.DnsSecInfo, error) {
	return m.enableDnsSecFn(ctx, zoneId)
}

func (m *mockDnsZoneAPI) DisableDnsSec(ctx context.Context, zoneId int64) (*client.DnsSecInfo, error) {
	return m.disableDnsSecFn(ctx, zoneId)
}

func newTestDnsZoneApp(api DnsZoneAPI) *App {
	return &App{NewDnsZoneAPI: func(_ *cobra.Command) (DnsZoneAPI, error) { return api, nil }}
}

func sampleDnsZone() *client.DnsZone {
	return &client.DnsZone{
		Id:                  100,
		Domain:              "example.com",
		DateCreated:         "2024-01-01T00:00:00Z",
		DateModified:        "2024-06-15T12:00:00Z",
		NameserversDetected: true,
		Nameserver1:         "kiki.bunny.net",
		Nameserver2:         "coco.bunny.net",
		SoaEmail:            "admin@example.com",
		DnsSecEnabled:       false,
		LoggingEnabled:      true,
		CertificateKeyType:  0,
		Records: []client.DnsRecord{
			{Id: 1, Type: 0, Name: "", Value: "93.184.216.34", Ttl: 300},
			{Id: 2, Type: 2, Name: "www", Value: "example.com", Ttl: 3600},
			{Id: 3, Type: 4, Name: "", Value: "mail.example.com", Ttl: 3600, Priority: 10},
		},
	}
}

func sampleDnsRecord() *client.DnsRecord {
	return &client.DnsRecord{
		Id:       1,
		Type:     0,
		Name:     "www",
		Value:    "93.184.216.34",
		Ttl:      300,
		Disabled: false,
	}
}

// --- dns help ---

func TestDns_ShowsInHelp(t *testing.T) {
	t.Parallel()
	out, _, err := executeCommand(nil, "dns", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, sub := range []string{"list", "get", "create", "update", "delete", "import", "export", "records", "dnssec"} {
		if !strings.Contains(out, sub) {
			t.Errorf("expected dns help to show %q subcommand", sub)
		}
	}
}

func TestDns_Alias(t *testing.T) {
	t.Parallel()
	out, _, err := executeCommand(nil, "dnszone", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Manage DNS zones") {
		t.Error("expected dnszone alias to work")
	}
}

// --- dns list ---

func TestDnsList_Table(t *testing.T) {
	t.Parallel()
	mock := &mockDnsZoneAPI{
		listDnsZonesFn: func(_ context.Context, page, perPage int, search string) (pagination.PageResponse[*client.DnsZone], error) {
			return pagination.PageResponse[*client.DnsZone]{
				Items:        []*client.DnsZone{sampleDnsZone()},
				HasMoreItems: false,
			}, nil
		},
	}
	app := newTestDnsZoneApp(mock)

	out, _, err := executeCommand(app, "dns", "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "example.com") {
		t.Error("expected output to contain domain name")
	}
}

func TestDnsList_JSON(t *testing.T) {
	t.Parallel()
	mock := &mockDnsZoneAPI{
		listDnsZonesFn: func(_ context.Context, page, perPage int, search string) (pagination.PageResponse[*client.DnsZone], error) {
			return pagination.PageResponse[*client.DnsZone]{
				Items:        []*client.DnsZone{sampleDnsZone()},
				HasMoreItems: false,
			}, nil
		},
	}
	app := newTestDnsZoneApp(mock)

	out, _, err := executeCommand(app, "dns", "list", "--output", "json")
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

func TestDnsList_SearchParam(t *testing.T) {
	t.Parallel()
	var capturedSearch string
	mock := &mockDnsZoneAPI{
		listDnsZonesFn: func(_ context.Context, page, perPage int, search string) (pagination.PageResponse[*client.DnsZone], error) {
			capturedSearch = search
			return pagination.PageResponse[*client.DnsZone]{}, nil
		},
	}
	app := newTestDnsZoneApp(mock)

	_, _, err := executeCommand(app, "dns", "list", "--search", "example")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedSearch != "example" {
		t.Errorf("expected search='example', got %q", capturedSearch)
	}
}

func TestDnsList_ErrorPropagation(t *testing.T) {
	t.Parallel()
	mock := &mockDnsZoneAPI{
		listDnsZonesFn: func(_ context.Context, page, perPage int, search string) (pagination.PageResponse[*client.DnsZone], error) {
			return pagination.PageResponse[*client.DnsZone]{}, fmt.Errorf("API unavailable")
		},
	}
	app := newTestDnsZoneApp(mock)

	_, stderr, err := executeCommand(app, "dns", "list")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(stderr, "API unavailable") {
		t.Errorf("expected API error in stderr, got %q", stderr)
	}
}

// --- dns get ---

func TestDnsGet_Table(t *testing.T) {
	t.Parallel()
	mock := &mockDnsZoneAPI{
		getDnsZoneFn: func(_ context.Context, id int64) (*client.DnsZone, error) {
			if id != 100 {
				t.Errorf("expected id=100, got %d", id)
			}
			return sampleDnsZone(), nil
		},
	}
	app := newTestDnsZoneApp(mock)

	out, _, err := executeCommand(app, "dns", "get", "100")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "example.com") {
		t.Error("expected output to contain domain name")
	}
	if !strings.Contains(out, "kiki.bunny.net") {
		t.Error("expected output to contain nameserver")
	}
}

func TestDnsGet_JSON(t *testing.T) {
	t.Parallel()
	mock := &mockDnsZoneAPI{
		getDnsZoneFn: func(_ context.Context, id int64) (*client.DnsZone, error) {
			return sampleDnsZone(), nil
		},
	}
	app := newTestDnsZoneApp(mock)

	out, _, err := executeCommand(app, "dns", "get", "100", "--output", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var result map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if result["Domain"] != "example.com" {
		t.Errorf("expected Domain=example.com, got %v", result["Domain"])
	}
}

func TestDnsGet_WithoutID_Fails(t *testing.T) {
	t.Parallel()
	_, _, err := executeCommand(nil, "dns", "get")
	if err == nil {
		t.Fatal("expected error for missing ID argument")
	}
}

func TestDnsGet_InvalidID_Fails(t *testing.T) {
	t.Parallel()
	mock := &mockDnsZoneAPI{
		getDnsZoneFn: func(_ context.Context, id int64) (*client.DnsZone, error) {
			return sampleDnsZone(), nil
		},
	}
	app := newTestDnsZoneApp(mock)

	_, stderr, err := executeCommand(app, "dns", "get", "abc")
	if err == nil {
		t.Fatal("expected error for invalid ID")
	}
	if !strings.Contains(stderr, "invalid DNS zone ID") {
		t.Errorf("expected 'invalid DNS zone ID' error, got %q", stderr)
	}
}

func TestDnsGet_WatchFlag(t *testing.T) {
	t.Parallel()
	out, _, err := executeCommand(nil, "dns", "get", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "--watch") {
		t.Error("expected get command to have --watch flag")
	}
}

// --- dns create ---

func TestDnsCreate_Success(t *testing.T) {
	t.Parallel()
	var capturedBody *client.DnsZoneCreate
	mock := &mockDnsZoneAPI{
		createDnsZoneFn: func(_ context.Context, body *client.DnsZoneCreate) (*client.DnsZone, error) {
			capturedBody = body
			return &client.DnsZone{Id: 200, Domain: body.Domain}, nil
		},
	}
	app := newTestDnsZoneApp(mock)

	out, _, err := executeCommand(app, "dns", "create", "--domain", "test.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedBody.Domain != "test.com" {
		t.Errorf("expected domain 'test.com', got %q", capturedBody.Domain)
	}
	if !strings.Contains(out, "test.com") {
		t.Error("expected output to contain created domain")
	}
}

func TestDnsCreate_FromFile(t *testing.T) {
	t.Parallel()
	out, _, err := executeCommand(nil, "dns", "create", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "--from-file") {
		t.Error("expected create command to have --from-file flag")
	}
}

func TestDnsCreate_FromFileStdin(t *testing.T) {
	t.Parallel()
	mock := &mockDnsZoneAPI{
		createDnsZoneFn: func(_ context.Context, body *client.DnsZoneCreate) (*client.DnsZone, error) {
			return &client.DnsZone{Id: 200, Domain: body.Domain}, nil
		},
	}
	app := newTestDnsZoneApp(mock)

	stdin := bytes.NewBufferString(`{"domain":"stdin.com"}`)
	out, _, err := executeCommandWithStdin(app, stdin, "dns", "create", "--from-file", "-")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "stdin.com") {
		t.Error("expected output to contain domain from stdin")
	}
}

// --- dns update ---

func TestDnsUpdate_Success(t *testing.T) {
	t.Parallel()
	var capturedId int64
	var capturedBody *client.DnsZoneUpdate
	mock := &mockDnsZoneAPI{
		updateDnsZoneFn: func(_ context.Context, id int64, body *client.DnsZoneUpdate) (*client.DnsZone, error) {
			capturedId = id
			capturedBody = body
			zone := sampleDnsZone()
			zone.SoaEmail = *body.SoaEmail
			return zone, nil
		},
	}
	app := newTestDnsZoneApp(mock)

	out, _, err := executeCommand(app, "dns", "update", "100", "--soa-email", "new@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedId != 100 {
		t.Errorf("expected id=100, got %d", capturedId)
	}
	if capturedBody.SoaEmail == nil || *capturedBody.SoaEmail != "new@example.com" {
		t.Error("expected soa-email to be set in body")
	}
	if !strings.Contains(out, "example.com") {
		t.Error("expected output to show updated zone")
	}
}

func TestDnsUpdate_WithoutID_Fails(t *testing.T) {
	t.Parallel()
	_, _, err := executeCommand(nil, "dns", "update")
	if err == nil {
		t.Fatal("expected error for missing ID argument")
	}
}

// --- dns delete ---

func TestDnsDelete_WithYes(t *testing.T) {
	t.Parallel()
	var deletedId int64
	mock := &mockDnsZoneAPI{
		deleteDnsZoneFn: func(_ context.Context, id int64) error {
			deletedId = id
			return nil
		},
	}
	app := newTestDnsZoneApp(mock)

	out, _, err := executeCommand(app, "dns", "delete", "100", "--yes")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deletedId != 100 {
		t.Errorf("expected deleted id=100, got %d", deletedId)
	}
	if !strings.Contains(out, "DNS zone deleted") {
		t.Error("expected deletion confirmation message")
	}
}

func TestDnsDelete_WithoutYes_Canceled(t *testing.T) {
	t.Parallel()
	mock := &mockDnsZoneAPI{
		deleteDnsZoneFn: func(_ context.Context, id int64) error {
			t.Error("delete should not have been called")
			return nil
		},
	}
	app := newTestDnsZoneApp(mock)

	stdin := bytes.NewBufferString("n\n")
	_, stderr, err := executeCommandWithStdin(app, stdin, "dns", "delete", "100")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "Deletion canceled") {
		t.Error("expected cancellation message")
	}
}

// --- dns import ---

func TestDnsImport_Success(t *testing.T) {
	t.Parallel()
	mock := &mockDnsZoneAPI{
		importDnsZoneFn: func(_ context.Context, zoneId int64, data io.Reader) (*client.DnsZoneImportResult, error) {
			if zoneId != 100 {
				t.Errorf("expected zoneId=100, got %d", zoneId)
			}
			return &client.DnsZoneImportResult{
				RecordsSuccessful: 5,
				RecordsFailed:     1,
				RecordsSkipped:    2,
			}, nil
		},
	}
	app := newTestDnsZoneApp(mock)

	// Use stdin for the zone file
	stdin := bytes.NewBufferString("@ IN A 93.184.216.34\n")
	out, _, err := executeCommandWithStdin(app, stdin, "dns", "import", "100", "--file", "-")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "5 successful") {
		t.Error("expected output to contain successful count")
	}
	if !strings.Contains(out, "1 failed") {
		t.Error("expected output to contain failed count")
	}
	if !strings.Contains(out, "2 skipped") {
		t.Error("expected output to contain skipped count")
	}
}

// --- dns export ---

func TestDnsExport_ToStdout(t *testing.T) {
	t.Parallel()
	zoneData := "$ORIGIN example.com.\n@ 300 IN A 93.184.216.34\n"
	mock := &mockDnsZoneAPI{
		exportDnsZoneFn: func(_ context.Context, zoneId int64) ([]byte, error) {
			if zoneId != 100 {
				t.Errorf("expected zoneId=100, got %d", zoneId)
			}
			return []byte(zoneData), nil
		},
	}
	app := newTestDnsZoneApp(mock)

	out, _, err := executeCommand(app, "dns", "export", "100")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != zoneData {
		t.Errorf("expected zone file content, got %q", out)
	}
}

func TestDnsExport_OutputFileFlag(t *testing.T) {
	t.Parallel()
	out, _, err := executeCommand(nil, "dns", "export", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "--output-file") {
		t.Error("expected export command to have --output-file flag")
	}
}

// --- dns dnssec ---

func TestDnsDnssecEnable(t *testing.T) {
	t.Parallel()
	mock := &mockDnsZoneAPI{
		enableDnsSecFn: func(_ context.Context, zoneId int64) (*client.DnsSecInfo, error) {
			if zoneId != 100 {
				t.Errorf("expected zoneId=100, got %d", zoneId)
			}
			return &client.DnsSecInfo{
				Enabled:  true,
				DsRecord: "example.com. 3600 IN DS 12345 13 2 ABCDEF",
				KeyTag:   12345,
			}, nil
		},
	}
	app := newTestDnsZoneApp(mock)

	out, _, err := executeCommand(app, "dns", "dnssec", "enable", "100")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "true") {
		t.Error("expected output to show enabled=true")
	}
	if !strings.Contains(out, "12345") {
		t.Error("expected output to show key tag")
	}
}

func TestDnsDnssecDisable_WithYes(t *testing.T) {
	t.Parallel()
	mock := &mockDnsZoneAPI{
		disableDnsSecFn: func(_ context.Context, zoneId int64) (*client.DnsSecInfo, error) {
			return &client.DnsSecInfo{Enabled: false}, nil
		},
	}
	app := newTestDnsZoneApp(mock)

	out, _, err := executeCommand(app, "dns", "dnssec", "disable", "100", "--yes")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "DNSSEC disabled") {
		t.Error("expected DNSSEC disabled message")
	}
}

func TestDnsDnssecDisable_Canceled(t *testing.T) {
	t.Parallel()
	mock := &mockDnsZoneAPI{
		disableDnsSecFn: func(_ context.Context, zoneId int64) (*client.DnsSecInfo, error) {
			t.Error("disable should not have been called")
			return nil, nil
		},
	}
	app := newTestDnsZoneApp(mock)

	stdin := bytes.NewBufferString("n\n")
	_, stderr, err := executeCommandWithStdin(app, stdin, "dns", "dnssec", "disable", "100")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "Operation canceled") {
		t.Error("expected cancellation message")
	}
}
