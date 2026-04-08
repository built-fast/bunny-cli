package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/built-fast/bunny-cli/internal/client"
	"github.com/built-fast/bunny-cli/internal/pagination"
	"github.com/spf13/cobra"
)

// --- mock EdgeScriptAPI ---

type mockEdgeScriptAPI struct {
	listEdgeScriptsFn               func(ctx context.Context, page, perPage int, search string, scriptTypes []int) (pagination.PageResponse[*client.EdgeScript], error)
	getEdgeScriptFn                 func(ctx context.Context, id int64) (*client.EdgeScript, error)
	createEdgeScriptFn              func(ctx context.Context, body *client.EdgeScriptCreate) (*client.EdgeScript, error)
	updateEdgeScriptFn              func(ctx context.Context, id int64, body *client.EdgeScriptUpdate) (*client.EdgeScript, error)
	deleteEdgeScriptFn              func(ctx context.Context, id int64, deleteLinkedPullZones bool) error
	getEdgeScriptStatisticsFn       func(ctx context.Context, id int64, dateFrom, dateTo string, loadLatest, hourly bool) (*client.EdgeScriptStatistics, error)
	rotateEdgeScriptDeploymentKeyFn func(ctx context.Context, id int64) error
	getEdgeScriptCodeFn             func(ctx context.Context, id int64) (*client.EdgeScriptCode, error)
	setEdgeScriptCodeFn             func(ctx context.Context, id int64, code string) error
	addEdgeScriptVariableFn         func(ctx context.Context, scriptId int64, body *client.EdgeScriptVariableCreate) (*client.EdgeScriptVariable, error)
	getEdgeScriptVariableFn         func(ctx context.Context, scriptId, variableId int64) (*client.EdgeScriptVariable, error)
	updateEdgeScriptVariableFn      func(ctx context.Context, scriptId, variableId int64, body *client.EdgeScriptVariableUpdate) error
	deleteEdgeScriptVariableFn      func(ctx context.Context, scriptId, variableId int64) error
	addEdgeScriptSecretFn           func(ctx context.Context, scriptId int64, body *client.EdgeScriptSecretCreate) (*client.EdgeScriptSecret, error)
	listEdgeScriptSecretsFn         func(ctx context.Context, scriptId int64) ([]*client.EdgeScriptSecret, error)
	updateEdgeScriptSecretFn        func(ctx context.Context, scriptId, secretId int64, body *client.EdgeScriptSecretUpdate) error
	deleteEdgeScriptSecretFn        func(ctx context.Context, scriptId, secretId int64) error
	listEdgeScriptReleasesFn        func(ctx context.Context, scriptId int64, page, perPage int) (pagination.PageResponse[*client.EdgeScriptRelease], error)
	getActiveEdgeScriptReleaseFn    func(ctx context.Context, scriptId int64) (*client.EdgeScriptRelease, error)
	publishEdgeScriptFn             func(ctx context.Context, scriptId int64, body *client.EdgeScriptPublish) error
	publishEdgeScriptReleaseFn      func(ctx context.Context, scriptId int64, uuid string) error
}

func (m *mockEdgeScriptAPI) ListEdgeScripts(ctx context.Context, page, perPage int, search string, scriptTypes []int) (pagination.PageResponse[*client.EdgeScript], error) {
	return m.listEdgeScriptsFn(ctx, page, perPage, search, scriptTypes)
}
func (m *mockEdgeScriptAPI) GetEdgeScript(ctx context.Context, id int64) (*client.EdgeScript, error) {
	return m.getEdgeScriptFn(ctx, id)
}
func (m *mockEdgeScriptAPI) CreateEdgeScript(ctx context.Context, body *client.EdgeScriptCreate) (*client.EdgeScript, error) {
	return m.createEdgeScriptFn(ctx, body)
}
func (m *mockEdgeScriptAPI) UpdateEdgeScript(ctx context.Context, id int64, body *client.EdgeScriptUpdate) (*client.EdgeScript, error) {
	return m.updateEdgeScriptFn(ctx, id, body)
}
func (m *mockEdgeScriptAPI) DeleteEdgeScript(ctx context.Context, id int64, deleteLinkedPullZones bool) error {
	return m.deleteEdgeScriptFn(ctx, id, deleteLinkedPullZones)
}
func (m *mockEdgeScriptAPI) GetEdgeScriptStatistics(ctx context.Context, id int64, dateFrom, dateTo string, loadLatest, hourly bool) (*client.EdgeScriptStatistics, error) {
	return m.getEdgeScriptStatisticsFn(ctx, id, dateFrom, dateTo, loadLatest, hourly)
}
func (m *mockEdgeScriptAPI) RotateEdgeScriptDeploymentKey(ctx context.Context, id int64) error {
	return m.rotateEdgeScriptDeploymentKeyFn(ctx, id)
}
func (m *mockEdgeScriptAPI) GetEdgeScriptCode(ctx context.Context, id int64) (*client.EdgeScriptCode, error) {
	return m.getEdgeScriptCodeFn(ctx, id)
}
func (m *mockEdgeScriptAPI) SetEdgeScriptCode(ctx context.Context, id int64, code string) error {
	return m.setEdgeScriptCodeFn(ctx, id, code)
}
func (m *mockEdgeScriptAPI) AddEdgeScriptVariable(ctx context.Context, scriptId int64, body *client.EdgeScriptVariableCreate) (*client.EdgeScriptVariable, error) {
	return m.addEdgeScriptVariableFn(ctx, scriptId, body)
}
func (m *mockEdgeScriptAPI) GetEdgeScriptVariable(ctx context.Context, scriptId, variableId int64) (*client.EdgeScriptVariable, error) {
	return m.getEdgeScriptVariableFn(ctx, scriptId, variableId)
}
func (m *mockEdgeScriptAPI) UpdateEdgeScriptVariable(ctx context.Context, scriptId, variableId int64, body *client.EdgeScriptVariableUpdate) error {
	return m.updateEdgeScriptVariableFn(ctx, scriptId, variableId, body)
}
func (m *mockEdgeScriptAPI) DeleteEdgeScriptVariable(ctx context.Context, scriptId, variableId int64) error {
	return m.deleteEdgeScriptVariableFn(ctx, scriptId, variableId)
}
func (m *mockEdgeScriptAPI) AddEdgeScriptSecret(ctx context.Context, scriptId int64, body *client.EdgeScriptSecretCreate) (*client.EdgeScriptSecret, error) {
	return m.addEdgeScriptSecretFn(ctx, scriptId, body)
}
func (m *mockEdgeScriptAPI) ListEdgeScriptSecrets(ctx context.Context, scriptId int64) ([]*client.EdgeScriptSecret, error) {
	return m.listEdgeScriptSecretsFn(ctx, scriptId)
}
func (m *mockEdgeScriptAPI) UpdateEdgeScriptSecret(ctx context.Context, scriptId, secretId int64, body *client.EdgeScriptSecretUpdate) error {
	return m.updateEdgeScriptSecretFn(ctx, scriptId, secretId, body)
}
func (m *mockEdgeScriptAPI) DeleteEdgeScriptSecret(ctx context.Context, scriptId, secretId int64) error {
	return m.deleteEdgeScriptSecretFn(ctx, scriptId, secretId)
}
func (m *mockEdgeScriptAPI) ListEdgeScriptReleases(ctx context.Context, scriptId int64, page, perPage int) (pagination.PageResponse[*client.EdgeScriptRelease], error) {
	return m.listEdgeScriptReleasesFn(ctx, scriptId, page, perPage)
}
func (m *mockEdgeScriptAPI) GetActiveEdgeScriptRelease(ctx context.Context, scriptId int64) (*client.EdgeScriptRelease, error) {
	return m.getActiveEdgeScriptReleaseFn(ctx, scriptId)
}
func (m *mockEdgeScriptAPI) PublishEdgeScript(ctx context.Context, scriptId int64, body *client.EdgeScriptPublish) error {
	return m.publishEdgeScriptFn(ctx, scriptId, body)
}
func (m *mockEdgeScriptAPI) PublishEdgeScriptRelease(ctx context.Context, scriptId int64, uuid string) error {
	return m.publishEdgeScriptReleaseFn(ctx, scriptId, uuid)
}

func newTestEdgeScriptApp(api EdgeScriptAPI) *App {
	return &App{NewEdgeScriptAPI: func(_ *cobra.Command) (EdgeScriptAPI, error) { return api, nil }}
}

func sampleEdgeScript() *client.EdgeScript {
	return &client.EdgeScript{
		Id:                  42,
		Name:                "my-script",
		ScriptType:          1,
		CurrentReleaseId:    10,
		DefaultHostname:     "my-script.b-cdn.net",
		SystemHostname:      "my-script.sys.b-cdn.net",
		DeploymentKey:       "deploy-key-123",
		MonthlyCost:         2.50,
		MonthlyRequestCount: 50000,
		MonthlyCpuTime:      1200,
		LastModified:        "2025-01-15T10:30:00Z",
		EdgeScriptVariables: []client.EdgeScriptVariable{
			{Id: 1, Name: "API_URL", Required: true, DefaultValue: "https://api.example.com"},
		},
		LinkedPullZones: []client.LinkedPullZone{
			{Id: 100, PullZoneName: "my-pz"},
		},
	}
}

// --- scripts help ---

func TestScripts_ShowsInHelp(t *testing.T) {
	t.Parallel()
	out, _, err := executeCommand(nil, "scripts", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, sub := range []string{"list", "get", "create", "update", "delete", "statistics", "rotate-key", "code", "publish", "releases", "variables", "secrets"} {
		if !strings.Contains(out, sub) {
			t.Errorf("expected scripts help to show %q subcommand", sub)
		}
	}
}

func TestScripts_Alias(t *testing.T) {
	t.Parallel()
	out, _, err := executeCommand(nil, "compute", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Manage edge scripts") {
		t.Error("expected compute alias to work")
	}
}

// --- scripts list ---

func TestScriptsList_ShowsFlags(t *testing.T) {
	t.Parallel()
	out, _, err := executeCommand(nil, "scripts", "list", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, flag := range []string{"--limit", "--all", "--search", "--type"} {
		if !strings.Contains(out, flag) {
			t.Errorf("expected help output to contain flag %q", flag)
		}
	}
}

func TestScriptsList_Table(t *testing.T) {
	t.Parallel()
	mock := &mockEdgeScriptAPI{
		listEdgeScriptsFn: func(_ context.Context, page, perPage int, search string, scriptTypes []int) (pagination.PageResponse[*client.EdgeScript], error) {
			return pagination.PageResponse[*client.EdgeScript]{
				Items:        []*client.EdgeScript{sampleEdgeScript()},
				HasMoreItems: false,
			}, nil
		},
	}
	app := newTestEdgeScriptApp(mock)

	out, _, err := executeCommand(app, "scripts", "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "my-script") {
		t.Error("expected output to contain script name")
	}
	if !strings.Contains(out, "CDN") {
		t.Error("expected output to contain script type")
	}
}

func TestScriptsList_JSON(t *testing.T) {
	t.Parallel()
	mock := &mockEdgeScriptAPI{
		listEdgeScriptsFn: func(_ context.Context, page, perPage int, search string, scriptTypes []int) (pagination.PageResponse[*client.EdgeScript], error) {
			return pagination.PageResponse[*client.EdgeScript]{
				Items:        []*client.EdgeScript{sampleEdgeScript()},
				HasMoreItems: false,
			}, nil
		},
	}
	app := newTestEdgeScriptApp(mock)

	out, _, err := executeCommand(app, "scripts", "list", "--output", "json")
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

func TestScriptsList_SearchParam(t *testing.T) {
	t.Parallel()
	var capturedSearch string
	mock := &mockEdgeScriptAPI{
		listEdgeScriptsFn: func(_ context.Context, page, perPage int, search string, scriptTypes []int) (pagination.PageResponse[*client.EdgeScript], error) {
			capturedSearch = search
			return pagination.PageResponse[*client.EdgeScript]{}, nil
		},
	}
	app := newTestEdgeScriptApp(mock)

	_, _, err := executeCommand(app, "scripts", "list", "--search", "test-script")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedSearch != "test-script" {
		t.Errorf("expected search='test-script', got %q", capturedSearch)
	}
}

func TestScriptsList_TypeFilter(t *testing.T) {
	t.Parallel()
	var capturedTypes []int
	mock := &mockEdgeScriptAPI{
		listEdgeScriptsFn: func(_ context.Context, page, perPage int, search string, scriptTypes []int) (pagination.PageResponse[*client.EdgeScript], error) {
			capturedTypes = scriptTypes
			return pagination.PageResponse[*client.EdgeScript]{}, nil
		},
	}
	app := newTestEdgeScriptApp(mock)

	_, _, err := executeCommand(app, "scripts", "list", "--type", "CDN")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(capturedTypes) != 1 || capturedTypes[0] != 1 {
		t.Errorf("expected scriptTypes=[1], got %v", capturedTypes)
	}
}

func TestScriptsList_ErrorPropagation(t *testing.T) {
	t.Parallel()
	mock := &mockEdgeScriptAPI{
		listEdgeScriptsFn: func(_ context.Context, page, perPage int, search string, scriptTypes []int) (pagination.PageResponse[*client.EdgeScript], error) {
			return pagination.PageResponse[*client.EdgeScript]{}, fmt.Errorf("API unavailable")
		},
	}
	app := newTestEdgeScriptApp(mock)

	_, stderr, err := executeCommand(app, "scripts", "list")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(stderr, "API unavailable") {
		t.Errorf("expected API error in stderr, got %q", stderr)
	}
}

// --- scripts get ---

func TestScriptsGet_Table(t *testing.T) {
	t.Parallel()
	mock := &mockEdgeScriptAPI{
		getEdgeScriptFn: func(_ context.Context, id int64) (*client.EdgeScript, error) {
			if id != 42 {
				t.Errorf("expected id=42, got %d", id)
			}
			return sampleEdgeScript(), nil
		},
	}
	app := newTestEdgeScriptApp(mock)

	out, _, err := executeCommand(app, "scripts", "get", "42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "my-script") {
		t.Error("expected output to contain script name")
	}
	if !strings.Contains(out, "deploy-key-123") {
		t.Error("expected output to contain deployment key")
	}
}

func TestScriptsGet_WithoutID_Fails(t *testing.T) {
	t.Parallel()
	_, _, err := executeCommand(nil, "scripts", "get")
	if err == nil {
		t.Fatal("expected error for missing ID argument")
	}
}

func TestScriptsGet_InvalidID_Fails(t *testing.T) {
	t.Parallel()
	mock := &mockEdgeScriptAPI{
		getEdgeScriptFn: func(_ context.Context, id int64) (*client.EdgeScript, error) {
			return sampleEdgeScript(), nil
		},
	}
	app := newTestEdgeScriptApp(mock)

	_, stderr, err := executeCommand(app, "scripts", "get", "abc")
	if err == nil {
		t.Fatal("expected error for invalid ID")
	}
	if !strings.Contains(stderr, "invalid script ID") {
		t.Errorf("expected 'invalid script ID' error, got %q", stderr)
	}
}

func TestScriptsGet_WatchFlag(t *testing.T) {
	t.Parallel()
	out, _, err := executeCommand(nil, "scripts", "get", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "--watch") {
		t.Error("expected get command to have --watch flag")
	}
}

// --- scripts create ---

func TestScriptsCreate_Success(t *testing.T) {
	t.Parallel()
	var capturedBody *client.EdgeScriptCreate
	mock := &mockEdgeScriptAPI{
		createEdgeScriptFn: func(_ context.Context, body *client.EdgeScriptCreate) (*client.EdgeScript, error) {
			capturedBody = body
			return &client.EdgeScript{Id: 99, Name: body.Name, ScriptType: body.ScriptType}, nil
		},
	}
	app := newTestEdgeScriptApp(mock)

	out, _, err := executeCommand(app, "scripts", "create", "--name", "new-script", "--type", "CDN")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedBody.Name != "new-script" {
		t.Errorf("expected name 'new-script', got %q", capturedBody.Name)
	}
	if capturedBody.ScriptType != 1 {
		t.Errorf("expected ScriptType=1 (CDN), got %d", capturedBody.ScriptType)
	}
	if !strings.Contains(out, "new-script") {
		t.Error("expected output to contain created script name")
	}
}

func TestScriptsCreate_FromFile(t *testing.T) {
	t.Parallel()
	out, _, err := executeCommand(nil, "scripts", "create", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "--from-file") {
		t.Error("expected create command to have --from-file flag")
	}
}

func TestScriptsCreate_FromFileStdin(t *testing.T) {
	t.Parallel()
	mock := &mockEdgeScriptAPI{
		createEdgeScriptFn: func(_ context.Context, body *client.EdgeScriptCreate) (*client.EdgeScript, error) {
			return &client.EdgeScript{Id: 99, Name: body.Name, ScriptType: body.ScriptType}, nil
		},
	}
	app := newTestEdgeScriptApp(mock)

	stdin := bytes.NewBufferString(`{"name":"stdin-script","type":"CDN"}`)
	out, _, err := executeCommandWithStdin(app, stdin, "scripts", "create", "--from-file", "-")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "stdin-script") {
		t.Error("expected output to contain script from stdin")
	}
}

// --- scripts update ---

func TestScriptsUpdate_Success(t *testing.T) {
	t.Parallel()
	var capturedId int64
	var capturedBody *client.EdgeScriptUpdate
	mock := &mockEdgeScriptAPI{
		updateEdgeScriptFn: func(_ context.Context, id int64, body *client.EdgeScriptUpdate) (*client.EdgeScript, error) {
			capturedId = id
			capturedBody = body
			return &client.EdgeScript{Id: id, Name: *body.Name}, nil
		},
	}
	app := newTestEdgeScriptApp(mock)

	out, _, err := executeCommand(app, "scripts", "update", "42", "--name", "updated-script")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedId != 42 {
		t.Errorf("expected id=42, got %d", capturedId)
	}
	if capturedBody.Name == nil || *capturedBody.Name != "updated-script" {
		t.Error("expected name to be set in body")
	}
	if !strings.Contains(out, "updated-script") {
		t.Error("expected output to show updated script")
	}
}

func TestScriptsUpdate_WithoutID_Fails(t *testing.T) {
	t.Parallel()
	_, _, err := executeCommand(nil, "scripts", "update")
	if err == nil {
		t.Fatal("expected error for missing ID argument")
	}
}

// --- scripts delete ---

func TestScriptsDelete_WithYes(t *testing.T) {
	t.Parallel()
	var deletedId int64
	mock := &mockEdgeScriptAPI{
		deleteEdgeScriptFn: func(_ context.Context, id int64, deleteLinkedPullZones bool) error {
			deletedId = id
			return nil
		},
	}
	app := newTestEdgeScriptApp(mock)

	out, _, err := executeCommand(app, "scripts", "delete", "42", "--yes")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deletedId != 42 {
		t.Errorf("expected deleted id=42, got %d", deletedId)
	}
	if !strings.Contains(out, "Edge script deleted") {
		t.Error("expected deletion confirmation message")
	}
}

func TestScriptsDelete_WithDeleteLinkedPullZones(t *testing.T) {
	t.Parallel()
	var capturedDeleteLinked bool
	mock := &mockEdgeScriptAPI{
		deleteEdgeScriptFn: func(_ context.Context, id int64, deleteLinkedPullZones bool) error {
			capturedDeleteLinked = deleteLinkedPullZones
			return nil
		},
	}
	app := newTestEdgeScriptApp(mock)

	_, _, err := executeCommand(app, "scripts", "delete", "42", "--yes", "--delete-linked-pullzones")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !capturedDeleteLinked {
		t.Error("expected deleteLinkedPullZones to be true")
	}
}

func TestScriptsDelete_WithoutYes_Canceled(t *testing.T) {
	t.Parallel()
	mock := &mockEdgeScriptAPI{
		deleteEdgeScriptFn: func(_ context.Context, id int64, deleteLinkedPullZones bool) error {
			t.Error("delete should not have been called")
			return nil
		},
	}
	app := newTestEdgeScriptApp(mock)

	stdin := bytes.NewBufferString("n\n")
	_, stderr, err := executeCommandWithStdin(app, stdin, "scripts", "delete", "42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "Deletion canceled") {
		t.Error("expected cancellation message")
	}
}

// --- scripts statistics ---

func TestScriptsStatistics_Success(t *testing.T) {
	t.Parallel()
	mock := &mockEdgeScriptAPI{
		getEdgeScriptStatisticsFn: func(_ context.Context, id int64, dateFrom, dateTo string, loadLatest, hourly bool) (*client.EdgeScriptStatistics, error) {
			return &client.EdgeScriptStatistics{
				TotalRequestsServed:        100000,
				TotalCpuUsed:               5.5,
				TotalMonthlyCost:           1.25,
				AverageCpuTimePerExecution: 0.05,
			}, nil
		},
	}
	app := newTestEdgeScriptApp(mock)

	out, _, err := executeCommand(app, "scripts", "statistics", "42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "100000") {
		t.Error("expected output to contain total requests")
	}
}

func TestScriptsStatistics_WithFlags(t *testing.T) {
	t.Parallel()
	var capturedFrom, capturedTo string
	var capturedHourly bool
	mock := &mockEdgeScriptAPI{
		getEdgeScriptStatisticsFn: func(_ context.Context, id int64, dateFrom, dateTo string, loadLatest, hourly bool) (*client.EdgeScriptStatistics, error) {
			capturedFrom = dateFrom
			capturedTo = dateTo
			capturedHourly = hourly
			return &client.EdgeScriptStatistics{}, nil
		},
	}
	app := newTestEdgeScriptApp(mock)

	_, _, err := executeCommand(app, "scripts", "statistics", "42", "--date-from", "2025-01-01", "--date-to", "2025-01-31", "--hourly")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedFrom != "2025-01-01" {
		t.Errorf("expected dateFrom='2025-01-01', got %q", capturedFrom)
	}
	if capturedTo != "2025-01-31" {
		t.Errorf("expected dateTo='2025-01-31', got %q", capturedTo)
	}
	if !capturedHourly {
		t.Error("expected hourly=true")
	}
}

// --- scripts rotate-key ---

func TestScriptsRotateKey_WithYes(t *testing.T) {
	t.Parallel()
	var rotatedId int64
	mock := &mockEdgeScriptAPI{
		rotateEdgeScriptDeploymentKeyFn: func(_ context.Context, id int64) error {
			rotatedId = id
			return nil
		},
	}
	app := newTestEdgeScriptApp(mock)

	out, _, err := executeCommand(app, "scripts", "rotate-key", "42", "--yes")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rotatedId != 42 {
		t.Errorf("expected rotated id=42, got %d", rotatedId)
	}
	if !strings.Contains(out, "Deployment key rotated") {
		t.Error("expected rotation confirmation message")
	}
}

func TestScriptsRotateKey_WithoutYes_Canceled(t *testing.T) {
	t.Parallel()
	mock := &mockEdgeScriptAPI{
		rotateEdgeScriptDeploymentKeyFn: func(_ context.Context, id int64) error {
			t.Error("rotate should not have been called")
			return nil
		},
	}
	app := newTestEdgeScriptApp(mock)

	stdin := bytes.NewBufferString("n\n")
	_, stderr, err := executeCommandWithStdin(app, stdin, "scripts", "rotate-key", "42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "Key rotation canceled") {
		t.Error("expected cancellation message")
	}
}

// --- scripts code get ---

func TestScriptsCodeGet_Success(t *testing.T) {
	t.Parallel()
	mock := &mockEdgeScriptAPI{
		getEdgeScriptCodeFn: func(_ context.Context, id int64) (*client.EdgeScriptCode, error) {
			return &client.EdgeScriptCode{
				Code:         "export default { fetch(req) { return new Response('Hello'); } }",
				LastModified: "2025-01-15T10:30:00Z",
			}, nil
		},
	}
	app := newTestEdgeScriptApp(mock)

	out, _, err := executeCommand(app, "scripts", "code", "get", "42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "export default") {
		t.Error("expected output to contain script code")
	}
}

// --- scripts code set ---

func TestScriptsCodeSet_FromStdin(t *testing.T) {
	t.Parallel()
	var capturedCode string
	mock := &mockEdgeScriptAPI{
		setEdgeScriptCodeFn: func(_ context.Context, id int64, code string) error {
			capturedCode = code
			return nil
		},
	}
	app := newTestEdgeScriptApp(mock)

	stdin := bytes.NewBufferString("export default { fetch(req) { return new Response('Hello'); } }")
	out, _, err := executeCommandWithStdin(app, stdin, "scripts", "code", "set", "42", "--file", "-")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(capturedCode, "export default") {
		t.Error("expected captured code to contain script code")
	}
	if !strings.Contains(out, "Code updated") {
		t.Error("expected confirmation message")
	}
}
