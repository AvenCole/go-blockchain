package blockchain

import (
	"path/filepath"
	"testing"
)

func TestCreateBlockchainAndIterate(t *testing.T) {
	dataDir := filepath.Join(t.TempDir(), "data")

	created, err := CreateBlockchain(dataDir, "genesis")
	if err != nil {
		t.Fatalf("CreateBlockchain() error = %v", err)
	}
	t.Cleanup(func() {
		_ = created.Close()
	})

	first, err := created.CurrentBlock()
	if err != nil {
		t.Fatalf("CurrentBlock() error = %v", err)
	}

	if first.Height != 0 {
		t.Fatalf("genesis height = %d, want 0", first.Height)
	}

	if string(first.Data) != "genesis" {
		t.Fatalf("genesis data = %q, want %q", string(first.Data), "genesis")
	}

	added, err := created.AddBlock("block-1")
	if err != nil {
		t.Fatalf("AddBlock() error = %v", err)
	}

	if added.Height != 1 {
		t.Fatalf("new block height = %d, want 1", added.Height)
	}

	blocks, err := created.Blocks()
	if err != nil {
		t.Fatalf("Blocks() error = %v", err)
	}

	if len(blocks) != 2 {
		t.Fatalf("len(Blocks()) = %d, want 2", len(blocks))
	}

	if string(blocks[0].Data) != "block-1" {
		t.Fatalf("latest block data = %q, want %q", string(blocks[0].Data), "block-1")
	}

	if string(blocks[1].Data) != "genesis" {
		t.Fatalf("genesis block data = %q, want %q", string(blocks[1].Data), "genesis")
	}
}

func TestOpenBlockchain(t *testing.T) {
	dataDir := filepath.Join(t.TempDir(), "data")

	created, err := CreateBlockchain(dataDir, "genesis")
	if err != nil {
		t.Fatalf("CreateBlockchain() error = %v", err)
	}

	if err := created.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	opened, err := OpenBlockchain(dataDir)
	if err != nil {
		t.Fatalf("OpenBlockchain() error = %v", err)
	}
	t.Cleanup(func() {
		_ = opened.Close()
	})

	current, err := opened.CurrentBlock()
	if err != nil {
		t.Fatalf("CurrentBlock() error = %v", err)
	}

	if current.Height != 0 {
		t.Fatalf("opened chain height = %d, want 0", current.Height)
	}
}

func TestChainExists(t *testing.T) {
	dataDir := filepath.Join(t.TempDir(), "data")

	exists, err := ChainExists(dataDir)
	if err != nil {
		t.Fatalf("ChainExists() error = %v", err)
	}

	if exists {
		t.Fatalf("ChainExists() = true, want false before initialization")
	}

	created, err := CreateBlockchain(dataDir, "genesis")
	if err != nil {
		t.Fatalf("CreateBlockchain() error = %v", err)
	}
	if err := created.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	exists, err = ChainExists(dataDir)
	if err != nil {
		t.Fatalf("ChainExists() after init error = %v", err)
	}

	if !exists {
		t.Fatalf("ChainExists() = false, want true after initialization")
	}
}
