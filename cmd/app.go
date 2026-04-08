package cmd

import (
	"context"
)

// appContextKey is an unexported type used as the key for storing App in a context.
type appContextKey struct{}

// App holds all API factory functions, allowing commands to obtain API clients
// without relying on package-level variables.
type App struct {
	// API factory functions will be added as resources are implemented.
	// e.g. NewPullZoneAPI func(cmd *cobra.Command) (PullZoneAPI, error)
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

// DefaultApp returns an App with all factory functions wired to the
// production implementations (viper config + real bunny.net client).
func DefaultApp() *App {
	return &App{}
}
