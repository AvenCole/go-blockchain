package gui

import (
	"os"
	"path/filepath"
	"testing"

	"go-blockchain/internal/config"
)

func TestNewServiceUsesDedicatedGUIDataDir(t *testing.T) {
	t.Setenv(guiDataDirEnv, "")

	service := NewService()
	base := config.Default().DataDir
	want := filepath.Join(base, "gui-desktop")

	if service.cfg.DataDir != want {
		t.Fatalf("GUI data dir = %q, want %q", service.cfg.DataDir, want)
	}
}

func TestNewServiceHonorsOverride(t *testing.T) {
	override := filepath.Join(os.TempDir(), "gui-override")
	t.Setenv(guiDataDirEnv, override)

	service := NewService()
	if service.cfg.DataDir != override {
		t.Fatalf("GUI data dir = %q, want override %q", service.cfg.DataDir, override)
	}
}
