package config

import "testing"

func TestDefault(t *testing.T) {
	cfg := Default()

	if cfg.ProjectName != "go-blockchain" {
		t.Fatalf("ProjectName = %q, want %q", cfg.ProjectName, "go-blockchain")
	}

	if cfg.DataDir != "./data" {
		t.Fatalf("DataDir = %q, want %q", cfg.DataDir, "./data")
	}

	if cfg.DefaultPort != 3000 {
		t.Fatalf("DefaultPort = %d, want %d", cfg.DefaultPort, 3000)
	}

	if cfg.LogLevel != "info" {
		t.Fatalf("LogLevel = %q, want %q", cfg.LogLevel, "info")
	}

	if cfg.NetworkMode != "local" {
		t.Fatalf("NetworkMode = %q, want %q", cfg.NetworkMode, "local")
	}
}
