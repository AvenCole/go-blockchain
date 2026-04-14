package blockchain

import (
	"errors"
	"path/filepath"
	"testing"

	"go-blockchain/internal/wallet"
)

func TestCreateBlockchainAndIterate(t *testing.T) {
	dataDir := filepath.Join(t.TempDir(), "data")
	miner := mustNewWallet(t)
	bob := mustNewWallet(t)

	created, err := CreateBlockchain(dataDir, miner.Address())
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
	if len(first.Transactions) != 1 || !first.Transactions[0].IsCoinbase() {
		t.Fatalf("genesis transaction should be one coinbase")
	}

	added, tx, err := created.SendTransaction(miner, bob.Address(), 10)
	if err != nil {
		t.Fatalf("SendTransaction() error = %v", err)
	}
	if added.Height != 1 {
		t.Fatalf("new block height = %d, want 1", added.Height)
	}
	if !created.VerifyTransaction(tx) {
		t.Fatalf("VerifyTransaction(tx) = false, want true")
	}

	blocks, err := created.Blocks()
	if err != nil {
		t.Fatalf("Blocks() error = %v", err)
	}
	if len(blocks) != 2 {
		t.Fatalf("len(Blocks()) = %d, want 2", len(blocks))
	}
	if len(blocks[0].Transactions) != 1 {
		t.Fatalf("len(blocks[0].Transactions) = %d, want 1", len(blocks[0].Transactions))
	}
	if blocks[0].Transactions[0].IDHex() != tx.IDHex() {
		t.Fatalf("latest tx id = %s, want %s", blocks[0].Transactions[0].IDHex(), tx.IDHex())
	}
	if len(blocks[0].Transactions[0].Inputs) != 1 {
		t.Fatalf("len(tx.Inputs) = %d, want 1", len(blocks[0].Transactions[0].Inputs))
	}
	if len(blocks[0].Transactions[0].Outputs) != 2 {
		t.Fatalf("len(tx.Outputs) = %d, want 2 (recipient + change)", len(blocks[0].Transactions[0].Outputs))
	}
	if len(blocks[0].Transactions[0].Inputs[0].Signature) == 0 {
		t.Fatalf("tx signature missing")
	}
}

func TestOpenBlockchain(t *testing.T) {
	dataDir := filepath.Join(t.TempDir(), "data")
	miner := mustNewWallet(t)

	created, err := CreateBlockchain(dataDir, miner.Address())
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
	miner := mustNewWallet(t)

	exists, err := ChainExists(dataDir)
	if err != nil {
		t.Fatalf("ChainExists() error = %v", err)
	}
	if exists {
		t.Fatalf("ChainExists() = true, want false before initialization")
	}

	created, err := CreateBlockchain(dataDir, miner.Address())
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

func TestBalanceOf(t *testing.T) {
	dataDir := filepath.Join(t.TempDir(), "data")
	miner := mustNewWallet(t)
	alice := mustNewWallet(t)

	created, err := CreateBlockchain(dataDir, miner.Address())
	if err != nil {
		t.Fatalf("CreateBlockchain() error = %v", err)
	}
	t.Cleanup(func() {
		_ = created.Close()
	})

	if _, _, err := created.SendTransaction(miner, alice.Address(), 20); err != nil {
		t.Fatalf("SendTransaction() error = %v", err)
	}

	minerBalance, err := created.BalanceOf(miner.Address())
	if err != nil {
		t.Fatalf("BalanceOf(miner) error = %v", err)
	}
	if minerBalance != 30 {
		t.Fatalf("BalanceOf(miner) = %d, want 30", minerBalance)
	}

	aliceBalance, err := created.BalanceOf(alice.Address())
	if err != nil {
		t.Fatalf("BalanceOf(alice) error = %v", err)
	}
	if aliceBalance != 20 {
		t.Fatalf("BalanceOf(alice) = %d, want 20", aliceBalance)
	}
}

func TestFindSpendableOutputsAndUTXO(t *testing.T) {
	dataDir := filepath.Join(t.TempDir(), "data")
	miner := mustNewWallet(t)
	alice := mustNewWallet(t)

	created, err := CreateBlockchain(dataDir, miner.Address())
	if err != nil {
		t.Fatalf("CreateBlockchain() error = %v", err)
	}
	t.Cleanup(func() {
		_ = created.Close()
	})

	if _, _, err := created.SendTransaction(miner, alice.Address(), 20); err != nil {
		t.Fatalf("SendTransaction() error = %v", err)
	}

	accumulated, spendable, err := created.FindSpendableOutputs(wallet.HashPublicKey(alice.PublicKey), 15)
	if err != nil {
		t.Fatalf("FindSpendableOutputs() error = %v", err)
	}
	if accumulated != 20 {
		t.Fatalf("accumulated = %d, want 20", accumulated)
	}
	if len(spendable) != 1 {
		t.Fatalf("len(spendable) = %d, want 1", len(spendable))
	}

	utxos, err := created.FindUTXO(wallet.HashPublicKey(alice.PublicKey))
	if err != nil {
		t.Fatalf("FindUTXO() error = %v", err)
	}
	if len(utxos) != 1 || utxos[0].Value != 20 {
		t.Fatalf("alice utxos = %+v, want one output of 20", utxos)
	}
}

func TestVerifyTransactionRejectsTampering(t *testing.T) {
	dataDir := filepath.Join(t.TempDir(), "data")
	miner := mustNewWallet(t)
	alice := mustNewWallet(t)

	created, err := CreateBlockchain(dataDir, miner.Address())
	if err != nil {
		t.Fatalf("CreateBlockchain() error = %v", err)
	}
	t.Cleanup(func() {
		_ = created.Close()
	})

	_, tx, err := created.SendTransaction(miner, alice.Address(), 20)
	if err != nil {
		t.Fatalf("SendTransaction() error = %v", err)
	}
	if !created.VerifyTransaction(tx) {
		t.Fatalf("VerifyTransaction(tx) = false, want true")
	}

	tx.Inputs[0].Signature[0] ^= 0x01
	if created.VerifyTransaction(tx) {
		t.Fatalf("VerifyTransaction(tampered) = true, want false")
	}
}

func TestVerifyTransactionRejectsTamperedPubKey(t *testing.T) {
	dataDir := filepath.Join(t.TempDir(), "data")
	miner := mustNewWallet(t)
	alice := mustNewWallet(t)
	mallory := mustNewWallet(t)

	created, err := CreateBlockchain(dataDir, miner.Address())
	if err != nil {
		t.Fatalf("CreateBlockchain() error = %v", err)
	}
	t.Cleanup(func() {
		_ = created.Close()
	})

	_, tx, err := created.SendTransaction(miner, alice.Address(), 20)
	if err != nil {
		t.Fatalf("SendTransaction() error = %v", err)
	}

	tx.Inputs[0].PubKey = append([]byte(nil), mallory.PublicKey...)
	if created.VerifyTransaction(tx) {
		t.Fatalf("VerifyTransaction(tampered pubkey) = true, want false")
	}
}

func TestVerifyTransactionRejectsInvalidOutputIndexWithoutPanic(t *testing.T) {
	dataDir := filepath.Join(t.TempDir(), "data")
	miner := mustNewWallet(t)
	alice := mustNewWallet(t)

	created, err := CreateBlockchain(dataDir, miner.Address())
	if err != nil {
		t.Fatalf("CreateBlockchain() error = %v", err)
	}
	t.Cleanup(func() {
		_ = created.Close()
	})

	_, tx, err := created.SendTransaction(miner, alice.Address(), 20)
	if err != nil {
		t.Fatalf("SendTransaction() error = %v", err)
	}

	tx.Inputs[0].Out = 999
	if created.VerifyTransaction(tx) {
		t.Fatalf("VerifyTransaction(invalid out index) = true, want false")
	}

	if _, err := created.AddBlock([]Transaction{tx}); !errors.Is(err, ErrInvalidTransaction) {
		t.Fatalf("AddBlock(invalid tx) error = %v, want ErrInvalidTransaction", err)
	}
}

func mustNewWallet(t *testing.T) *wallet.Wallet {
	t.Helper()

	w, err := wallet.New()
	if err != nil {
		t.Fatalf("wallet.New() error = %v", err)
	}

	return w
}
