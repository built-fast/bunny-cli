package cmd

import (
	"context"

	"github.com/built-fast/bunny-cli/internal/client"
	"github.com/built-fast/bunny-cli/internal/pagination"
)

// EdgeScriptAPI abstracts the bunny.net Edge Scripting (compute) API methods,
// allowing tests to inject mocks without making real API calls.
type EdgeScriptAPI interface {
	// Scripts
	ListEdgeScripts(ctx context.Context, page, perPage int, search string, scriptTypes []int) (pagination.PageResponse[*client.EdgeScript], error)
	GetEdgeScript(ctx context.Context, id int64) (*client.EdgeScript, error)
	CreateEdgeScript(ctx context.Context, body *client.EdgeScriptCreate) (*client.EdgeScript, error)
	UpdateEdgeScript(ctx context.Context, id int64, body *client.EdgeScriptUpdate) (*client.EdgeScript, error)
	DeleteEdgeScript(ctx context.Context, id int64, deleteLinkedPullZones bool) error
	GetEdgeScriptStatistics(ctx context.Context, id int64, dateFrom, dateTo string, loadLatest, hourly bool) (*client.EdgeScriptStatistics, error)
	RotateEdgeScriptDeploymentKey(ctx context.Context, id int64) error

	// Code
	GetEdgeScriptCode(ctx context.Context, id int64) (*client.EdgeScriptCode, error)
	SetEdgeScriptCode(ctx context.Context, id int64, code string) error

	// Variables
	AddEdgeScriptVariable(ctx context.Context, scriptId int64, body *client.EdgeScriptVariableCreate) (*client.EdgeScriptVariable, error)
	GetEdgeScriptVariable(ctx context.Context, scriptId, variableId int64) (*client.EdgeScriptVariable, error)
	UpdateEdgeScriptVariable(ctx context.Context, scriptId, variableId int64, body *client.EdgeScriptVariableUpdate) error
	DeleteEdgeScriptVariable(ctx context.Context, scriptId, variableId int64) error

	// Secrets
	AddEdgeScriptSecret(ctx context.Context, scriptId int64, body *client.EdgeScriptSecretCreate) (*client.EdgeScriptSecret, error)
	ListEdgeScriptSecrets(ctx context.Context, scriptId int64) ([]*client.EdgeScriptSecret, error)
	UpdateEdgeScriptSecret(ctx context.Context, scriptId, secretId int64, body *client.EdgeScriptSecretUpdate) error
	DeleteEdgeScriptSecret(ctx context.Context, scriptId, secretId int64) error

	// Releases
	ListEdgeScriptReleases(ctx context.Context, scriptId int64, page, perPage int) (pagination.PageResponse[*client.EdgeScriptRelease], error)
	GetActiveEdgeScriptRelease(ctx context.Context, scriptId int64) (*client.EdgeScriptRelease, error)

	// Publish
	PublishEdgeScript(ctx context.Context, scriptId int64, body *client.EdgeScriptPublish) error
	PublishEdgeScriptRelease(ctx context.Context, scriptId int64, uuid string) error
}
