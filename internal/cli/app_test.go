package cli

import (
	"bytes"
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

	if !strings.Contains(stdout.String(), "initialization skeleton is ready") {
		t.Fatalf("doctor output missing readiness text: %q", stdout.String())
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
