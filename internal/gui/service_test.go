package gui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go-blockchain/internal/blockchain"
	"go-blockchain/internal/config"
	"go-blockchain/internal/wallet"
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

func TestDashboardIncludesLastReorgStatus(t *testing.T) {
	t.Setenv(guiDataDirEnv, t.TempDir())

	service := NewService()
	miner, err := wallet.New()
	if err != nil {
		t.Fatalf("wallet.New() error = %v", err)
	}
	alice, err := wallet.New()
	if err != nil {
		t.Fatalf("wallet.New() error = %v", err)
	}

	chain, err := blockchain.CreateBlockchain(service.cfg.DataDir, miner.Address())
	if err != nil {
		t.Fatalf("CreateBlockchain() error = %v", err)
	}

	tx, err := chain.SendTransaction(miner, alice.Address(), 20, 0)
	if err != nil {
		t.Fatalf("SendTransaction() error = %v", err)
	}
	if _, _, err := chain.MineMempool(miner.Address()); err != nil {
		t.Fatalf("MineMempool() error = %v", err)
	}

	blocks, err := chain.Blocks()
	if err != nil {
		t.Fatalf("Blocks() error = %v", err)
	}
	genesis := blocks[len(blocks)-1]

	fork1 := blockchain.NewBlock([]blockchain.Transaction{blockchain.NewCoinbaseTransaction(miner.Address(), "fork-1")}, genesis.Hash, 1)
	if err := chain.ImportBlock(fork1); err != nil {
		t.Fatalf("ImportBlock(fork1) error = %v", err)
	}
	fork2 := blockchain.NewBlock([]blockchain.Transaction{blockchain.NewCoinbaseTransaction(miner.Address(), "fork-2")}, fork1.Hash, 2)
	if err := chain.ImportBlock(fork2); err != nil {
		t.Fatalf("ImportBlock(fork2) error = %v", err)
	}

	pending, err := chain.PendingTransactions()
	if err != nil {
		t.Fatalf("PendingTransactions() error = %v", err)
	}
	if len(pending) != 1 || pending[0].IDHex() != tx.IDHex() {
		t.Fatalf("pending tx not restored after reorg")
	}
	if err := chain.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	dashboard, err := service.Dashboard()
	if err != nil {
		t.Fatalf("Dashboard() error = %v", err)
	}
	if dashboard.LastReorg == nil {
		t.Fatalf("dashboard.LastReorg = nil, want value")
	}
	if dashboard.LastReorg.RestoredTxCount != 1 {
		t.Fatalf("dashboard.LastReorg.RestoredTxCount = %d, want 1", dashboard.LastReorg.RestoredTxCount)
	}
}
