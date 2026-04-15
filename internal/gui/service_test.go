package gui

import (
	"fmt"
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
	if len(nodes[0].RecentEvents) == 0 {
		t.Fatalf("len(nodes[0].RecentEvents) = 0, want events")
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
	if len(dashboard.RecentEvents) == 0 {
		t.Fatalf("len(dashboard.RecentEvents) = 0, want at least one event")
	}
	if dashboard.RecentEvents[0].Kind != "reorg" {
		t.Fatalf("dashboard.RecentEvents[0].Kind = %q, want reorg", dashboard.RecentEvents[0].Kind)
	}
}

func TestQueueP2PKAndMultiSigTransactions(t *testing.T) {
	t.Setenv(guiDataDirEnv, t.TempDir())

	service := NewService()
	minerAddr, err := service.CreateWallet()
	if err != nil {
		t.Fatalf("CreateWallet(miner) error = %v", err)
	}
	aliceAddr, err := service.CreateWallet()
	if err != nil {
		t.Fatalf("CreateWallet(alice) error = %v", err)
	}
	bobAddr, err := service.CreateWallet()
	if err != nil {
		t.Fatalf("CreateWallet(bob) error = %v", err)
	}

	chain, err := blockchain.CreateBlockchain(service.cfg.DataDir, minerAddr)
	if err != nil {
		t.Fatalf("CreateBlockchain() error = %v", err)
	}
	if err := chain.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	p2pkTxID, err := service.QueueP2PKTransaction(minerAddr, aliceAddr, 20, 1)
	if err != nil {
		t.Fatalf("QueueP2PKTransaction() error = %v", err)
	}
	if p2pkTxID == "" {
		t.Fatalf("QueueP2PKTransaction() returned empty txid")
	}
	if _, err := service.MinePending(minerAddr); err != nil {
		t.Fatalf("MinePending() error = %v", err)
	}

	multiTxID, err := service.QueueMultiSigTransaction(minerAddr, aliceAddr+","+bobAddr, 2, 10, 1)
	if err != nil {
		t.Fatalf("QueueMultiSigTransaction() error = %v", err)
	}
	if multiTxID == "" {
		t.Fatalf("QueueMultiSigTransaction() returned empty txid")
	}

	pending, err := service.PendingTransactions()
	if err != nil {
		t.Fatalf("PendingTransactions() error = %v", err)
	}
	if len(pending) != 1 {
		t.Fatalf("len(pending) = %d, want 1", len(pending))
	}
}

func TestNodeControlWorkflow(t *testing.T) {
	t.Setenv(guiDataDirEnv, t.TempDir())

	service := NewService()
	minerAddr, err := service.CreateWallet()
	if err != nil {
		t.Fatalf("CreateWallet(miner) error = %v", err)
	}
	aliceAddr, err := service.CreateWallet()
	if err != nil {
		t.Fatalf("CreateWallet(alice) error = %v", err)
	}

	addr, err := service.StartNode("127.0.0.1:0", "", minerAddr)
	if err != nil {
		t.Fatalf("StartNode() error = %v", err)
	}
	t.Cleanup(func() { _ = service.StopNode(addr) })

	nodes, err := service.Nodes()
	if err != nil {
		t.Fatalf("Nodes() before init error = %v", err)
	}
	if len(nodes) != 1 {
		t.Fatalf("len(nodes) before init = %d, want 1", len(nodes))
	}
	if nodes[0].Initialized {
		t.Fatalf("nodes[0].Initialized before init = true, want false")
	}

	if err := service.InitializeNodeBlockchain(addr, ""); err != nil {
		t.Fatalf("InitializeNodeBlockchain() error = %v", err)
	}

	nodes, err = service.Nodes()
	if err != nil {
		t.Fatalf("Nodes() after init error = %v", err)
	}
	if !nodes[0].Initialized {
		t.Fatalf("nodes[0].Initialized after init = false, want true")
	}
	if nodes[0].Height != 0 {
		t.Fatalf("nodes[0].Height after init = %d, want 0", nodes[0].Height)
	}

	txid, err := service.SubmitNodeTransaction(addr, minerAddr, aliceAddr, 20, 1)
	if err != nil {
		t.Fatalf("SubmitNodeTransaction() error = %v", err)
	}
	if txid == "" {
		t.Fatalf("SubmitNodeTransaction() returned empty txid")
	}

	nodes, err = service.Nodes()
	if err != nil {
		t.Fatalf("Nodes() after submit error = %v", err)
	}
	if nodes[0].MempoolCount != 1 {
		t.Fatalf("nodes[0].MempoolCount after submit = %d, want 1", nodes[0].MempoolCount)
	}

	hash, err := service.MineNodePending(addr)
	if err != nil {
		t.Fatalf("MineNodePending() error = %v", err)
	}
	if hash == "" {
		t.Fatalf("MineNodePending() returned empty hash")
	}

	nodes, err = service.Nodes()
	if err != nil {
		t.Fatalf("Nodes() after mine error = %v", err)
	}
	if nodes[0].Height != 1 {
		t.Fatalf("nodes[0].Height after mine = %d, want 1", nodes[0].Height)
	}
	if nodes[0].MempoolCount != 0 {
		t.Fatalf("nodes[0].MempoolCount after mine = %d, want 0", nodes[0].MempoolCount)
	}

	foundInit := false
	foundSubmit := false
	for _, event := range nodes[0].RecentEvents {
		if event.Kind == "chain_init" {
			foundInit = true
		}
		if event.Kind == "tx_submit" {
			foundSubmit = true
		}
	}
	if !foundInit {
		t.Fatalf("expected chain_init event in node recent events")
	}
	if !foundSubmit {
		t.Fatalf("expected tx_submit event in node recent events")
	}
}

func TestConsoleNodeCommands(t *testing.T) {
	t.Setenv(guiDataDirEnv, t.TempDir())

	service := NewService()
	minerAddr, err := service.CreateWallet()
	if err != nil {
		t.Fatalf("CreateWallet(miner) error = %v", err)
	}
	aliceAddr, err := service.CreateWallet()
	if err != nil {
		t.Fatalf("CreateWallet(alice) error = %v", err)
	}

	addr, err := service.StartNode("127.0.0.1:0", "", minerAddr)
	if err != nil {
		t.Fatalf("StartNode() error = %v", err)
	}
	t.Cleanup(func() { _ = service.StopNode(addr) })

	result, err := service.ExecuteCLI(fmt.Sprintf("nodeinit %s", addr))
	if err != nil {
		t.Fatalf("ExecuteCLI(nodeinit) error = %v", err)
	}
	if !strings.Contains(result.Stdout, "node chain ready") {
		t.Fatalf("nodeinit stdout = %q, want ready message", result.Stdout)
	}

	result, err = service.ExecuteCLI(fmt.Sprintf("nodesend %s %s %s 10 1", addr, minerAddr, aliceAddr))
	if err != nil {
		t.Fatalf("ExecuteCLI(nodesend) error = %v", err)
	}
	if !strings.Contains(result.Stdout, "node transaction queued") {
		t.Fatalf("nodesend stdout = %q, want queued message", result.Stdout)
	}

	result, err = service.ExecuteCLI(fmt.Sprintf("nodemine %s", addr))
	if err != nil {
		t.Fatalf("ExecuteCLI(nodemine) error = %v", err)
	}
	if !strings.Contains(result.Stdout, "node mined block") {
		t.Fatalf("nodemine stdout = %q, want mined message", result.Stdout)
	}
}

func TestRunNetworkQuickDemo(t *testing.T) {
	t.Setenv(guiDataDirEnv, t.TempDir())

	service := NewService()
	result, err := service.RunNetworkQuickDemo()
	if err != nil {
		t.Fatalf("RunNetworkQuickDemo() error = %v", err)
	}
	t.Cleanup(func() {
		_ = service.StopNode(result.SourceNode)
		_ = service.StopNode(result.PeerNode)
	})

	if result.SourceNode == "" || result.PeerNode == "" {
		t.Fatalf("RunNetworkQuickDemo() returned empty node addresses: %+v", result)
	}
	if result.TxID == "" || result.BlockHash == "" {
		t.Fatalf("RunNetworkQuickDemo() returned empty tx or block: %+v", result)
	}
	if result.PeerHeight < 1 {
		t.Fatalf("RunNetworkQuickDemo().PeerHeight = %d, want >= 1", result.PeerHeight)
	}
	if !result.TipAnnounced {
		t.Fatalf("RunNetworkQuickDemo().TipAnnounced = false, want true")
	}

	nodes, err := service.Nodes()
	if err != nil {
		t.Fatalf("Nodes() error = %v", err)
	}
	if len(nodes) != 2 {
		t.Fatalf("len(nodes) = %d, want 2", len(nodes))
	}
}

func TestConsoleRunNetworkDemoCommand(t *testing.T) {
	t.Setenv(guiDataDirEnv, t.TempDir())

	service := NewService()
	result, err := service.ExecuteCLI("runnetdemo")
	if err != nil {
		t.Fatalf("ExecuteCLI(runnetdemo) error = %v", err)
	}
	nodes, err := service.Nodes()
	if err != nil {
		t.Fatalf("Nodes() after runnetdemo error = %v", err)
	}
	for _, node := range nodes {
		_ = service.StopNode(node.Address)
	}
	if !strings.Contains(result.Stdout, "network demo ready") {
		t.Fatalf("runnetdemo stdout = %q, want ready message", result.Stdout)
	}
}

func TestRunNetworkReorgDemo(t *testing.T) {
	t.Setenv(guiDataDirEnv, t.TempDir())

	service := NewService()
	result, err := service.RunNetworkReorgDemo()
	if err != nil {
		t.Fatalf("RunNetworkReorgDemo() error = %v", err)
	}
	t.Cleanup(func() {
		nodes, _ := service.Nodes()
		for _, node := range nodes {
			_ = service.StopNode(node.Address)
		}
	})

	if result.SourceNode == "" || result.PeerNode == "" {
		t.Fatalf("RunNetworkReorgDemo() returned empty node addresses: %+v", result)
	}
	if result.OriginalBlockHash == "" || result.ReorgTxID == "" {
		t.Fatalf("RunNetworkReorgDemo() returned empty reorg data: %+v", result)
	}
	if !result.Restored {
		t.Fatalf("RunNetworkReorgDemo().Restored = false, want true")
	}
	if !result.PeerReorged {
		t.Fatalf("RunNetworkReorgDemo().PeerReorged = false, want true")
	}
	if result.SourceNewHeight <= result.SourceOldHeight {
		t.Fatalf("RunNetworkReorgDemo() heights = %d -> %d, want growth", result.SourceOldHeight, result.SourceNewHeight)
	}
	if result.PeerHeight < result.SourceNewHeight {
		t.Fatalf("RunNetworkReorgDemo().PeerHeight = %d, want >= %d", result.PeerHeight, result.SourceNewHeight)
	}

	nodes, err := service.Nodes()
	if err != nil {
		t.Fatalf("Nodes() after reorg demo error = %v", err)
	}
	if len(nodes) != 2 {
		t.Fatalf("len(nodes) after reorg demo = %d, want 2", len(nodes))
	}
	for _, node := range nodes {
		if node.LastReorg == nil {
			t.Fatalf("node %s LastReorg = nil, want value", node.Address)
		}
	}
}

func TestConsoleRunNetworkReorgDemoCommand(t *testing.T) {
	t.Setenv(guiDataDirEnv, t.TempDir())

	service := NewService()
	result, err := service.ExecuteCLI("runreorgdemo")
	if err != nil {
		t.Fatalf("ExecuteCLI(runreorgdemo) error = %v", err)
	}
	nodes, err := service.Nodes()
	if err != nil {
		t.Fatalf("Nodes() after runreorgdemo error = %v", err)
	}
	for _, node := range nodes {
		_ = service.StopNode(node.Address)
	}
	if !strings.Contains(result.Stdout, "network reorg demo ready") {
		t.Fatalf("runreorgdemo stdout = %q, want ready message", result.Stdout)
	}
}

func TestRunNetworkPartitionDemo(t *testing.T) {
	t.Setenv(guiDataDirEnv, t.TempDir())

	service := NewService()
	result, err := service.RunNetworkPartitionDemo()
	if err != nil {
		t.Fatalf("RunNetworkPartitionDemo() error = %v", err)
	}
	t.Cleanup(func() {
		nodes, _ := service.Nodes()
		for _, node := range nodes {
			_ = service.StopNode(node.Address)
		}
	})

	if result.SourceNode == "" || result.PeerNode == "" || result.ForkNode == "" {
		t.Fatalf("RunNetworkPartitionDemo() returned empty node addresses: %+v", result)
	}
	if result.ConfirmedTxID == "" || result.FinalTipHash == "" {
		t.Fatalf("RunNetworkPartitionDemo() returned empty tx or tip: %+v", result)
	}
	if result.ForkHeight <= result.OldConfirmedHeight {
		t.Fatalf("RunNetworkPartitionDemo() heights = old %d fork %d, want fork longer", result.OldConfirmedHeight, result.ForkHeight)
	}
	if !result.Restored {
		t.Fatalf("RunNetworkPartitionDemo().Restored = false, want true")
	}
	if !result.AllConverged {
		t.Fatalf("RunNetworkPartitionDemo().AllConverged = false, want true")
	}

	nodes, err := service.Nodes()
	if err != nil {
		t.Fatalf("Nodes() after partition demo error = %v", err)
	}
	if len(nodes) != 3 {
		t.Fatalf("len(nodes) after partition demo = %d, want 3", len(nodes))
	}
	for _, node := range nodes {
		if node.TipHash != result.FinalTipHash {
			t.Fatalf("node %s tip = %q, want %q", node.Address, node.TipHash, result.FinalTipHash)
		}
		if node.Height < result.ForkHeight {
			t.Fatalf("node %s height = %d, want >= %d", node.Address, node.Height, result.ForkHeight)
		}
	}
}

func TestConsoleRunNetworkPartitionDemoCommand(t *testing.T) {
	t.Setenv(guiDataDirEnv, t.TempDir())

	service := NewService()
	result, err := service.ExecuteCLI("runpartitiondemo")
	if err != nil {
		t.Fatalf("ExecuteCLI(runpartitiondemo) error = %v", err)
	}
	nodes, err := service.Nodes()
	if err != nil {
		t.Fatalf("Nodes() after runpartitiondemo error = %v", err)
	}
	for _, node := range nodes {
		_ = service.StopNode(node.Address)
	}
	if !strings.Contains(result.Stdout, "network partition demo ready") {
		t.Fatalf("runpartitiondemo stdout = %q, want ready message", result.Stdout)
	}
}
