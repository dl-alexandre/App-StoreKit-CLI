package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/dl-alexandre/cli-tools/version"
)

func TestVersionCommand(t *testing.T) {
	version.Version = "0.0.10-test"
	version.BinaryName = "ask"
	version.GitCommit = "unknown"
	version.BuildTime = "unknown"

	cmd := newRootCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"version"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("version: %v\n%s", err, buf.String())
	}

	out := buf.String()
	if !strings.Contains(out, "ask version 0.0.10-test") {
		t.Fatalf("unexpected version output: %q", out)
	}
}
