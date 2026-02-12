package cli

import (
	"errors"
	"io"
	"os"

	"github.com/dl-alexandre/App-Store-Server-CLI/internal/output"
)

func readBody(path string) ([]byte, error) {
	if path == "" {
		return nil, nil
	}
	if path == "-" {
		return io.ReadAll(os.Stdin)
	}
	return os.ReadFile(path)
}

func writeResponse(app *App, response any) error {
	if app == nil {
		return errors.New("app is required")
	}
	return output.Render(os.Stdout, response, app.Format, app.JQ, app.Columns)
}
