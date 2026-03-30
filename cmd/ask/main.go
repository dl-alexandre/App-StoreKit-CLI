package main

import (
	"os"

	"github.com/dl-alexandre/App-StoreKit-CLI/internal/cli"
	cliver "github.com/dl-alexandre/cli-tools/version"
)

// Build-time variables (set by GoReleaser or build flags)
var (
	version   = "dev"
	gitCommit = "unknown"
	buildTime = "unknown"
)

func main() {
	// Set version info in cli-tools
	cliver.Version = version
	cliver.GitCommit = gitCommit
	cliver.BuildTime = buildTime
	cliver.BinaryName = "ask"

	os.Exit(cli.Execute())
}
