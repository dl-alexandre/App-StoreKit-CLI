package cli

import "context"

type contextKey struct{}

func withApp(ctx context.Context, app *App) context.Context {
	return context.WithValue(ctx, contextKey{}, app)
}

func appFromContext(ctx context.Context) *App {
	if ctx == nil {
		return nil
	}
	if app, ok := ctx.Value(contextKey{}).(*App); ok {
		return app
	}
	return nil
}
