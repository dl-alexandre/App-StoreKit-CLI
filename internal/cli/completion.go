package cli

import (
	"github.com/dl-alexandre/App-StoreKit-CLI/internal/validate"
)

func completeEnum(flag string) func(*Command, []string, string) ([]string, int) {
	return func(cmd *Command, args []string, toComplete string) ([]string, int) {
		return validate.Allowed(flag), 0
	}
}
