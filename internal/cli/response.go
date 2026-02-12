package cli

import (
	"github.com/dl-alexandre/App-Store-Server-CLI/internal/api"
	"github.com/dl-alexandre/App-Store-Server-CLI/internal/output"
)

func responseData(app *App, resp api.Response) any {
	if app.Format == output.FormatRaw {
		return resp.Body
	}
	if resp.JSON != nil {
		return resp.JSON
	}
	return string(resp.Body)
}
