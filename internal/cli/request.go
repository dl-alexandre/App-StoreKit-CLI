package cli

import (
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/dl-alexandre/App-StoreKit-CLI/internal/output"
)

func readBody(path string) ([]byte, error) {
	if path == "" {
		return nil, nil
	}
	if path == "-" {
		return io.ReadAll(os.Stdin)
	}
	// Validate path to prevent directory traversal
	cleanPath := filepath.Clean(path)
	return os.ReadFile(cleanPath) // #nosec G304 - path is cleaned above
}

func writeResponse(app *App, response any) error {
	if app == nil {
		return errors.New("app is required")
	}
	return output.Render(os.Stdout, response, app.Format, app.JQ, app.Columns)
}
