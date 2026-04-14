package blockchain

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBalanceOfByScanMatchesCached(t *testing.T) {
	dataDir := filepath.Join(t.TempDir(), "data")
	miner := mustNewWallet(t)
	alice := mustNewWallet(t)

	bc, err := CreateBlockchain(dataDir, miner.Address())
	if err != nil {
		t.Fatalf("CreateBlockchain() error = %v", err)
	}
	t.Cleanup(func() { _ = bc.Close() })

	if _, err := bc.SendTransaction(miner, alice.Address(), 20, 2); err != nil {
		t.Fatalf("SendTransaction() error = %v", err)
	}
	if _, _, err := bc.MineMempool(miner.Address()); err != nil {
		t.Fatalf("MineMempool() error = %v", err)
	}

	cached, err := bc.BalanceOf(miner.Address())
	if err != nil {
		t.Fatalf("BalanceOf() error = %v", err)
	}
	scanned, err := bc.BalanceOfByScan(miner.Address())
	if err != nil {
		t.Fatalf("BalanceOfByScan() error = %v", err)
	}

	if cached != scanned {
		t.Fatalf("cached=%d scanned=%d, want equal", cached, scanned)
	}
}

func TestRunBalanceBenchmarkAndWrite(t *testing.T) {
	dataDir := filepath.Join(t.TempDir(), "data")
	outputDir := filepath.Join(t.TempDir(), "perf")
	miner := mustNewWallet(t)
	alice := mustNewWallet(t)

	bc, err := CreateBlockchain(dataDir, miner.Address())
	if err != nil {
		t.Fatalf("CreateBlockchain() error = %v", err)
	}
	t.Cleanup(func() { _ = bc.Close() })

	if _, err := bc.SendTransaction(miner, alice.Address(), 20, 2); err != nil {
		t.Fatalf("SendTransaction() error = %v", err)
	}
	if _, _, err := bc.MineMempool(miner.Address()); err != nil {
		t.Fatalf("MineMempool() error = %v", err)
	}

	result, err := bc.RunBalanceBenchmark([]string{miner.Address(), alice.Address()}, 5)
	if err != nil {
		t.Fatalf("RunBalanceBenchmark() error = %v", err)
	}
	if result.AddressesTested != 2 || result.TotalLookups != 10 {
		t.Fatalf("unexpected benchmark result: %+v", result)
	}

	if err := WriteBalanceBenchmark(result, outputDir); err != nil {
		t.Fatalf("WriteBalanceBenchmark() error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(outputDir, "latest.json")); err != nil {
		t.Fatalf("latest.json missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outputDir, "latest.md")); err != nil {
		t.Fatalf("latest.md missing: %v", err)
	}
}
