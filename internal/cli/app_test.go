package cli

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"go-blockchain/internal/config"
)

func TestRunHelp(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	app := NewApp(config.Default(), &stdout, &stderr)
	code := app.Run([]string{"--help"})

	if code != 0 {
		t.Fatalf("Run(help) exit code = %d, want 0", code)
	}

	if !strings.Contains(stdout.String(), "Available commands:") {
		t.Fatalf("help output missing command list: %q", stdout.String())
	}
}

func TestRunVersion(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	app := NewApp(config.Default(), &stdout, &stderr)
	code := app.Run([]string{"version"})

	if code != 0 {
		t.Fatalf("Run(version) exit code = %d, want 0", code)
	}

	if !strings.Contains(stdout.String(), "go-blockchain version") {
		t.Fatalf("version output missing version banner: %q", stdout.String())
	}
}

func TestRunDoctor(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	app := NewApp(config.Default(), &stdout, &stderr)
	code := app.Run([]string{"doctor"})

	if code != 0 {
		t.Fatalf("Run(doctor) exit code = %d, want 0", code)
	}

	if !strings.Contains(stdout.String(), "blockchain skeleton is ready") {
		t.Fatalf("doctor output missing readiness text: %q", stdout.String())
	}
}

func TestCreateBlockchainAddBlockAndPrintChain(t *testing.T) {
	cfg := config.Default()
	cfg.DataDir = filepath.Join(t.TempDir(), "data")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	app := NewApp(cfg, &stdout, &stderr)

	if code := app.Run([]string{"createblockchain", "genesis"}); code != 0 {
		t.Fatalf("createblockchain exit code = %d, stderr=%q", code, stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	if code := app.Run([]string{"addblock", "block-1"}); code != 0 {
		t.Fatalf("addblock exit code = %d, stderr=%q", code, stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	if code := app.Run([]string{"printchain"}); code != 0 {
		t.Fatalf("printchain exit code = %d, stderr=%q", code, stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "Data: block-1") {
		t.Fatalf("printchain output missing newest block: %q", output)
	}

	if !strings.Contains(output, "Data: genesis") {
		t.Fatalf("printchain output missing genesis block: %q", output)
	}
}

func TestRunUnknownCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	app := NewApp(config.Default(), &stdout, &stderr)
	code := app.Run([]string{"unknown"})

	if code != 1 {
		t.Fatalf("Run(unknown) exit code = %d, want 1", code)
	}

	if !strings.Contains(stderr.String(), "unknown command") {
		t.Fatalf("stderr missing unknown command error: %q", stderr.String())
	}
}
