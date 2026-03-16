package main

import (
	"os"

	"github.com/dl-alexandre/App-StoreKit-CLI/internal/cli"
)

// Build-time variables (set by GoReleaser or build flags)
var (
	version   = "dev"
	gitCommit = "unknown"
	buildTime = "unknown"
)

func main() {
	// Set build-time variables in the cli package
	cli.Version = version
	cli.GitCommit = gitCommit
	cli.BuildTime = buildTime

	os.Exit(cli.Execute())
}
