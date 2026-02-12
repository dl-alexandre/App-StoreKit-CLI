package cli

import "github.com/dl-alexandre/App-Store-Server-CLI/internal/api"

const (
	exitSuccess     = 0
	exitUnknown     = 1
	exitAuth        = 2
	exitValidation  = 3
	exitRateLimited = 4
	exitNotFound    = 5
)

func exitCode(err error) int {
	if err == nil {
		return exitSuccess
	}
	if apiErr, ok := err.(api.APIError); ok {
		switch apiErr.Status {
		case 401, 403:
			return exitAuth
		case 400, 422:
			return exitValidation
		case 404:
			return exitNotFound
		case 429:
			return exitRateLimited
		default:
			return exitUnknown
		}
	}
	return exitUnknown
}

func formatError(err error) string {
	if err == nil {
		return ""
	}
	if apiErr, ok := err.(api.APIError); ok {
		if apiErr.Code != "" || apiErr.Message != "" {
			return apiErr.Error()
		}
		if apiErr.Status != 0 {
			return apiErr.Error()
		}
	}
	return err.Error()
}
