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

func TestSplitCommandLinePreservesQuotedArguments(t *testing.T) {
	args, err := splitCommandLine(`createblockchain "quoted address"`)
	if err != nil {
		t.Fatalf("splitCommandLine() error = %v", err)
	}
	if len(args) != 2 {
		t.Fatalf("len(args) = %d, want 2", len(args))
	}
	if args[1] != "quoted address" {
		t.Fatalf("args[1] = %q, want quoted address", args[1])
	}
}

func TestStartAndStopNodeLifecycle(t *testing.T) {
	t.Setenv(guiDataDirEnv, t.TempDir())

	service := NewService()
	addr, err := service.StartNode("127.0.0.1:0", "", "")
	if err != nil {
		t.Fatalf("StartNode() error = %v", err)
	}
	if !strings.Contains(addr, ":") || strings.HasSuffix(addr, ":0") {
		t.Fatalf("StartNode() addr = %q, want bound address", addr)
	}

	nodes, err := service.Nodes()
	if err != nil {
		t.Fatalf("Nodes() error = %v", err)
	}
	if len(nodes) != 1 {
		t.Fatalf("len(nodes) = %d, want 1", len(nodes))
	}
	if nodes[0].Address != addr {
		t.Fatalf("nodes[0].Address = %q, want %q", nodes[0].Address, addr)
	}

	if err := service.StopNode(addr); err != nil {
		t.Fatalf("StopNode() error = %v", err)
	}

	nodes, err = service.Nodes()
	if err != nil {
		t.Fatalf("Nodes() after stop error = %v", err)
	}
	if len(nodes) != 0 {
		t.Fatalf("len(nodes) after stop = %d, want 0", len(nodes))
	}
}
