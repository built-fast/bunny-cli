package cmd

import (
	"context"

	"github.com/built-fast/bunny-cli/internal/client"
	"github.com/built-fast/bunny-cli/internal/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// appContextKey is an unexported type used as the key for storing App in a context.
type appContextKey struct{}

// App holds all API factory functions, allowing commands to obtain API clients
// without relying on package-level variables.
type App struct {
	NewPullZoneAPI    func(cmd *cobra.Command) (PullZoneAPI, error)
	NewStorageZoneAPI func(cmd *cobra.Command) (StorageZoneAPI, error)
	NewStorageAPI     func(cmd *cobra.Command, password, hostname string) (StorageAPI, error)
	NewDnsZoneAPI     func(cmd *cobra.Command) (DnsZoneAPI, error)
}

// NewAppContext returns a new context that carries the given App.
func NewAppContext(ctx context.Context, app *App) context.Context {
	return context.WithValue(ctx, appContextKey{}, app)
}

// AppFromContext returns the App stored in ctx. If no App is set,
// it returns a non-nil zero-value App to prevent nil panics.
func AppFromContext(ctx context.Context) *App {
	if app, ok := ctx.Value(appContextKey{}).(*App); ok && app != nil {
		return app
	}
	return &App{}
}

// newAPIFactory returns a factory function that creates an API client using
// the standard viper/output config pattern.
func newAPIFactory[T any](fn func(client.ClientConfig) (T, error)) func(cmd *cobra.Command) (T, error) {
	return func(cmd *cobra.Command) (T, error) {
		cfg := output.FromContext(cmd.Context())
		return fn(client.ClientConfig{
			APIKey: viper.GetString("api_key"),
			IsJSON: func() bool { return isJSONFormat(cfg.Format) },
		})
	}
}

// DefaultApp returns an App with all factory functions wired to the
// production implementations (viper config + real bunny.net client).
func DefaultApp() *App {
	return &App{
		NewPullZoneAPI: newAPIFactory(func(c client.ClientConfig) (PullZoneAPI, error) {
			return client.NewClient(c)
		}),
		NewStorageZoneAPI: newAPIFactory(func(c client.ClientConfig) (StorageZoneAPI, error) {
			return client.NewClient(c)
		}),
		NewStorageAPI: func(cmd *cobra.Command, password, hostname string) (StorageAPI, error) {
			cfg := output.FromContext(cmd.Context())
			return client.NewStorageClient(client.StorageClientConfig{
				Password: password,
				Hostname: hostname,
				IsJSON:   func() bool { return isJSONFormat(cfg.Format) },
			})
		},
		NewDnsZoneAPI: newAPIFactory(func(c client.ClientConfig) (DnsZoneAPI, error) {
			return client.NewClient(c)
		}),
	}
}
