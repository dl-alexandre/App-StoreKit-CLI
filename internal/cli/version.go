package cli

import (
	"fmt"

	"github.com/dl-alexandre/cli-tools/version"
)

// Build-time variables (re-exported from cli-tools/version for backward compatibility)
var (
	// Version is the current version of the CLI
	Version = version.Version

	// BinaryName is the name of the binary
	BinaryName = version.BinaryName

	// GitHubRepo is the GitHub repository name
	GitHubRepo = "App-StoreKit-CLI"

	// GitCommit is the git commit hash
	GitCommit = version.GitCommit

	// BuildTime is the build timestamp
	BuildTime = version.BuildTime
)

func init() {
	// Set CLI-specific metadata
	version.BinaryName = "ask"
}

func newVersionCommand() *Command {
	return &Command{
		Use:   "version",
		Short: "Print the ask version",
		Args:  NoArgs,
		Run: func(cmd *Command, args []string) {
			fmt.Fprintln(cmd.OutOrStdout(), version.DetailedString())
		},
	}
}
