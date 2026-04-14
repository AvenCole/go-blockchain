package gui

import (
	"os"
	"path/filepath"
	"strings"
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

func TestWalletsExposeLockingScript(t *testing.T) {
	t.Setenv(guiDataDirEnv, t.TempDir())

	service := NewService()
	if _, err := service.CreateWallet(); err != nil {
		t.Fatalf("CreateWallet() error = %v", err)
	}

	wallets, err := service.Wallets()
	if err != nil {
		t.Fatalf("Wallets() error = %v", err)
	}
	if len(wallets) != 1 {
		t.Fatalf("len(wallets) = %d, want 1", len(wallets))
	}
	if !strings.Contains(wallets[0].LockingScript, "OP_DUP OP_HASH160") {
		t.Fatalf("locking script = %q, want standard P2PKH", wallets[0].LockingScript)
	}
}
