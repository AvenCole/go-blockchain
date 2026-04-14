package network

import (
	"context"
	"net"
	"path/filepath"
	"testing"
	"time"

	"go-blockchain/internal/blockchain"
	"go-blockchain/internal/wallet"
)

func TestNodeDiscoveryAndChainSync(t *testing.T) {
	dirA := filepath.Join(t.TempDir(), "nodeA")
	dirB := filepath.Join(t.TempDir(), "nodeB")
	miner := mustWallet(t)

	chainA, err := blockchain.CreateBlockchain(dirA, miner.Address())
	if err != nil {
		t.Fatalf("CreateBlockchain(nodeA) error = %v", err)
	}
	_ = chainA.Close()

	addrA := freeAddress(t)
	addrB := freeAddress(t)

	nodeA := NewNode(addrA, dirA, miner.Address())
	nodeB := NewNode(addrB, dirB, "")

	ctxA, cancelA := context.WithCancel(context.Background())
	defer cancelA()
	ctxB, cancelB := context.WithCancel(context.Background())
	defer cancelB()

	go func() { _ = nodeA.Listen(ctxA) }()
	go func() { _ = nodeB.Listen(ctxB) }()
	waitUntilListening(t, nodeA.Address)
	waitUntilListening(t, nodeB.Address)

	if err := nodeB.Connect(nodeA.Address); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	waitUntil(t, func() bool { return len(nodeA.KnownPeers()) >= 2 && len(nodeB.KnownPeers()) >= 2 })
	if err := nodeA.sendBlocks(nodeB.Address, -1); err != nil {
		t.Fatalf("sendBlocks() error = %v", err)
	}

	waitUntil(t, func() bool {
		height, _ := blockchain.BestHeight(dirB)
		return height == 0
	})

	if len(nodeB.KnownPeers()) < 2 {
		t.Fatalf("nodeB peers = %v, want discovery of nodeA", nodeB.KnownPeers())
	}
}

func TestTransactionAndBlockBroadcast(t *testing.T) {
	dirA := filepath.Join(t.TempDir(), "nodeA")
	dirB := filepath.Join(t.TempDir(), "nodeB")
	miner := mustWallet(t)
	alice := mustWallet(t)

	chainA, err := blockchain.CreateBlockchain(dirA, miner.Address())
	if err != nil {
		t.Fatalf("CreateBlockchain(nodeA) error = %v", err)
	}
	_ = chainA.Close()

	addrA := freeAddress(t)
	addrB := freeAddress(t)

	nodeA := NewNode(addrA, dirA, miner.Address())
	nodeB := NewNode(addrB, dirB, "")

	ctxA, cancelA := context.WithCancel(context.Background())
	defer cancelA()
	ctxB, cancelB := context.WithCancel(context.Background())
	defer cancelB()

	go func() { _ = nodeA.Listen(ctxA) }()
	go func() { _ = nodeB.Listen(ctxB) }()
	waitUntilListening(t, nodeA.Address)
	waitUntilListening(t, nodeB.Address)

	if err := nodeB.Connect(nodeA.Address); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	waitUntil(t, func() bool { return len(nodeA.KnownPeers()) >= 2 && len(nodeB.KnownPeers()) >= 2 })
	if err := nodeA.sendBlocks(nodeB.Address, -1); err != nil {
		t.Fatalf("sendBlocks() error = %v", err)
	}
	waitUntil(t, func() bool {
		height, _ := blockchain.BestHeight(dirB)
		return height == 0
	})

	if _, err := nodeA.SubmitTransaction(miner, alice.Address(), 20, 2); err != nil {
		t.Fatalf("SubmitTransaction() error = %v", err)
	}
	waitUntil(t, func() bool {
		bc, err := blockchain.OpenBlockchain(dirB)
		if err != nil {
			return false
		}
		defer bc.Close()
		size, err := bc.MempoolSize()
		return err == nil && size == 1
	})
}

func TestNodeBuffersOrphanBlockUntilParentArrives(t *testing.T) {
	dirA := filepath.Join(t.TempDir(), "nodeA")
	dirB := filepath.Join(t.TempDir(), "nodeB")
	miner := mustWallet(t)

	chainA, err := blockchain.CreateBlockchain(dirA, miner.Address())
	if err != nil {
		t.Fatalf("CreateBlockchain(nodeA) error = %v", err)
	}
	defer chainA.Close()

	block1, err := chainA.AddBlock([]blockchain.Transaction{
		blockchain.NewCoinbaseTransaction(miner.Address(), "block-1"),
	})
	if err != nil {
		t.Fatalf("AddBlock(block1) error = %v", err)
	}
	block2, err := chainA.AddBlock([]blockchain.Transaction{
		blockchain.NewCoinbaseTransaction(miner.Address(), "block-2"),
	})
	if err != nil {
		t.Fatalf("AddBlock(block2) error = %v", err)
	}

	blocks, err := chainA.Blocks()
	if err != nil {
		t.Fatalf("Blocks() error = %v", err)
	}
	genesis := blocks[len(blocks)-1]
	if err := blockchain.ImportBlockToDir(dirB, genesis); err != nil {
		t.Fatalf("ImportBlockToDir(genesis) error = %v", err)
	}

	nodeB := NewNode("127.0.0.1:0", dirB, "")
	if err := nodeB.handleBlock(blockMessage{From: "peer", Block: *block2}); err != nil {
		t.Fatalf("handleBlock(orphan child) error = %v", err)
	}

	if nodeB.OrphanCount() != 1 {
		t.Fatalf("OrphanCount() = %d, want 1", nodeB.OrphanCount())
	}
	height, err := blockchain.BestHeight(dirB)
	if err != nil {
		t.Fatalf("BestHeight() error = %v", err)
	}
	if height != 0 {
		t.Fatalf("height before parent = %d, want 0", height)
	}

	if err := nodeB.handleBlock(blockMessage{From: "peer", Block: *block1}); err != nil {
		t.Fatalf("handleBlock(parent) error = %v", err)
	}
	if nodeB.OrphanCount() != 0 {
		t.Fatalf("OrphanCount() after parent = %d, want 0", nodeB.OrphanCount())
	}
	height, err = blockchain.BestHeight(dirB)
	if err != nil {
		t.Fatalf("BestHeight() after parent error = %v", err)
	}
	if height != 2 {
		t.Fatalf("height after parent = %d, want 2", height)
	}
}

func freeAddress(t *testing.T) string {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen() error = %v", err)
	}
	defer ln.Close()
	return ln.Addr().String()
}

func waitUntil(t *testing.T, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("condition not met before timeout")
}

func waitUntilListening(t *testing.T, addr string) {
	t.Helper()
	waitUntil(t, func() bool {
		conn, err := net.DialTimeout("tcp", addr, 100*time.Millisecond)
		if err != nil {
			return false
		}
		_ = conn.Close()
		return true
	})
}

func mustWallet(t *testing.T) *wallet.Wallet {
	t.Helper()
	w, err := wallet.New()
	if err != nil {
		t.Fatalf("wallet.New() error = %v", err)
	}
	return w
}
