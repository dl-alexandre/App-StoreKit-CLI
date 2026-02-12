package cli

import (
	"github.com/dl-alexandre/App-StoreKit-CLI/internal/validate"
	"github.com/spf13/cobra"
)

func completeEnum(flag string) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return validate.Allowed(flag), cobra.ShellCompDirectiveNoFileComp
	}
}
