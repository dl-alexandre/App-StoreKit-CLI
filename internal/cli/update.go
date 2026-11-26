package cli

import (
	"github.com/dl-alexandre/cli-tools/update"
	"github.com/dl-alexandre/cli-tools/version"
)

// AutoUpdateCheck performs a background update check (for use at startup)
// It returns immediately and doesn't block
func AutoUpdateCheck() {
	checker := update.New(update.Config{
		CurrentVersion: version.Version,
		BinaryName:     version.BinaryName,
		GitHubRepo:     "dl-alexandre/App-StoreKit-CLI",
		InstallCommand: "brew upgrade ask",
	})
	checker.AutoCheck()
}

// newCheckUpdateCommand creates the check-updates command
func newCheckUpdateCommand() *Command {
	var force bool
	var format string

	cmd := &Command{
		Use:   "check-updates",
		Short: "Check for available updates",
		RunE: func(cmd *Command, args []string) error {
			checker := update.New(update.Config{
				CurrentVersion: version.Version,
				BinaryName:     version.BinaryName,
				GitHubRepo:     "dl-alexandre/App-StoreKit-CLI",
				InstallCommand: "brew upgrade ask",
			})

			info, err := checker.Check(force)
			if err != nil {
				return err
			}

			return update.DisplayUpdate(info, version.BinaryName, format)
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Force check, bypassing cache")
	cmd.Flags().StringVar(&format, "format", "table", "Output format: table, json")

	return cmd
}

// UpdateInfo is re-exported from cli-tools for backward compatibility
type UpdateInfo = update.Info
