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

	if !strings.Contains(stdout.String(), "blockchain transaction demo is ready") {
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
	if !strings.Contains(output, "Transactions: 1") {
		t.Fatalf("printchain output missing transaction count: %q", output)
	}

	if !strings.Contains(output, "Output: to=system amount=50") {
		t.Fatalf("printchain output missing debug coinbase output: %q", output)
	}
}

func TestRunSendAndGetBalance(t *testing.T) {
	cfg := config.Default()
	cfg.DataDir = filepath.Join(t.TempDir(), "data")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	app := NewApp(cfg, &stdout, &stderr)

	if code := app.Run([]string{"createblockchain", "miner"}); code != 0 {
		t.Fatalf("createblockchain exit code = %d, stderr=%q", code, stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	if code := app.Run([]string{"send", "miner", "alice", "20"}); code != 0 {
		t.Fatalf("send exit code = %d, stderr=%q", code, stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	if code := app.Run([]string{"getbalance", "alice"}); code != 0 {
		t.Fatalf("getbalance exit code = %d, stderr=%q", code, stderr.String())
	}

	if !strings.Contains(stdout.String(), "balance[alice]=20") {
		t.Fatalf("getbalance output = %q, want alice balance", stdout.String())
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
