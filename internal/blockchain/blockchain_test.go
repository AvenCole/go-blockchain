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
	t.Cleanup(func() { _ = created.Close() })

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
	if !first.VerifyMerkleRoot() || !first.VerifyProofOfWork() {
		t.Fatalf("genesis block verification failed")
	}

	tx, err := created.SendTransaction(miner, bob.Address(), 10, 0)
	if err != nil {
		t.Fatalf("SendTransaction() error = %v", err)
	}
	added, mined, err := created.MineMempool(miner.Address())
	if err != nil {
		t.Fatalf("MineMempool() error = %v", err)
	}
	if added.Height != 1 || mined != 1 {
		t.Fatalf("mined block height=%d txs=%d, want height=1 txs=1", added.Height, mined)
	}
	if !created.VerifyTransaction(tx) {
		t.Fatalf("VerifyTransaction(tx) = false, want true")
	}
	if !added.VerifyMerkleRoot() || !added.VerifyProofOfWork() {
		t.Fatalf("mined block verification failed")
	}

	blocks, err := created.Blocks()
	if err != nil {
		t.Fatalf("Blocks() error = %v", err)
	}
	if len(blocks) != 2 {
		t.Fatalf("len(Blocks()) = %d, want 2", len(blocks))
	}
	if len(blocks[0].Transactions) != 2 {
		t.Fatalf("len(blocks[0].Transactions) = %d, want 2", len(blocks[0].Transactions))
	}
	var minedTx Transaction
	for _, candidate := range blocks[0].Transactions {
		if candidate.IDHex() == tx.IDHex() {
			minedTx = candidate
		}
	}
	if len(minedTx.Inputs) != 1 || len(minedTx.Outputs) != 2 {
		t.Fatalf("mined transaction shape invalid")
	}
	if len(minedTx.Inputs[0].Signature) == 0 {
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
	t.Cleanup(func() { _ = opened.Close() })

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
	t.Cleanup(func() { _ = created.Close() })

	if _, err := created.SendTransaction(miner, alice.Address(), 20, 0); err != nil {
		t.Fatalf("SendTransaction() error = %v", err)
	}
	if _, _, err := created.MineMempool(miner.Address()); err != nil {
		t.Fatalf("MineMempool() error = %v", err)
	}

	minerBalance, err := created.BalanceOf(miner.Address())
	if err != nil {
		t.Fatalf("BalanceOf(miner) error = %v", err)
	}
	if minerBalance != 80 {
		t.Fatalf("BalanceOf(miner) = %d, want 80", minerBalance)
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
	t.Cleanup(func() { _ = created.Close() })

	if _, err := created.SendTransaction(miner, alice.Address(), 20, 0); err != nil {
		t.Fatalf("SendTransaction() error = %v", err)
	}
	if _, _, err := created.MineMempool(miner.Address()); err != nil {
		t.Fatalf("MineMempool() error = %v", err)
	}

	accumulated, spendable, err := created.FindSpendableOutputs(wallet.HashPublicKey(alice.PublicKey), 15)
	if err != nil {
		t.Fatalf("FindSpendableOutputs() error = %v", err)
	}
	if accumulated != 20 || len(spendable) != 1 {
		t.Fatalf("unexpected spendable outputs")
	}

	utxos, err := created.FindUTXO(wallet.HashPublicKey(alice.PublicKey))
	if err != nil {
		t.Fatalf("FindUTXO() error = %v", err)
	}
	if len(utxos) != 1 || utxos[0].Value != 20 {
		t.Fatalf("alice utxos invalid")
	}
}

func TestReindexUTXOPreservesBalances(t *testing.T) {
	dataDir := filepath.Join(t.TempDir(), "data")
	miner := mustNewWallet(t)
	alice := mustNewWallet(t)

	created, err := CreateBlockchain(dataDir, miner.Address())
	if err != nil {
		t.Fatalf("CreateBlockchain() error = %v", err)
	}
	t.Cleanup(func() { _ = created.Close() })

	if _, err := created.SendTransaction(miner, alice.Address(), 20, 0); err != nil {
		t.Fatalf("SendTransaction() error = %v", err)
	}
	if _, _, err := created.MineMempool(miner.Address()); err != nil {
		t.Fatalf("MineMempool() error = %v", err)
	}
	if err := created.ReindexUTXO(); err != nil {
		t.Fatalf("ReindexUTXO() error = %v", err)
	}

	minerBalance, _ := created.BalanceOf(miner.Address())
	aliceBalance, _ := created.BalanceOf(alice.Address())
	if minerBalance != 80 || aliceBalance != 20 {
		t.Fatalf("balances after reindex invalid: miner=%d alice=%d", minerBalance, aliceBalance)
	}
}

func TestSequentialSpendsWorkBeforeAndAfterReindex(t *testing.T) {
	dataDir := filepath.Join(t.TempDir(), "data")
	miner := mustNewWallet(t)
	alice := mustNewWallet(t)

	created, err := CreateBlockchain(dataDir, miner.Address())
	if err != nil {
		t.Fatalf("CreateBlockchain() error = %v", err)
	}
	t.Cleanup(func() { _ = created.Close() })

	if _, err := created.SendTransaction(miner, alice.Address(), 20, 0); err != nil {
		t.Fatalf("miner -> alice failed: %v", err)
	}
	if _, _, err := created.MineMempool(miner.Address()); err != nil {
		t.Fatalf("mine 1 failed: %v", err)
	}
	if _, err := created.SendTransaction(alice, miner.Address(), 10, 0); err != nil {
		t.Fatalf("alice -> miner failed: %v", err)
	}
	if _, _, err := created.MineMempool(alice.Address()); err != nil {
		t.Fatalf("mine 2 failed: %v", err)
	}
	if _, err := created.SendTransaction(miner, alice.Address(), 35, 0); err != nil {
		t.Fatalf("miner -> alice 35 before reindex failed: %v", err)
	}
	if _, _, err := created.MineMempool(miner.Address()); err != nil {
		t.Fatalf("mine 3 failed: %v", err)
	}

	if err := created.ReindexUTXO(); err != nil {
		t.Fatalf("ReindexUTXO() error = %v", err)
	}

	if _, err := created.SendTransaction(miner, alice.Address(), 5, 0); err != nil {
		t.Fatalf("miner -> alice 5 after reindex failed: %v", err)
	}
	if _, _, err := created.MineMempool(miner.Address()); err != nil {
		t.Fatalf("mine 4 failed: %v", err)
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
	t.Cleanup(func() { _ = created.Close() })

	tx, err := created.SendTransaction(miner, alice.Address(), 20, 0)
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
	t.Cleanup(func() { _ = created.Close() })

	tx, err := created.SendTransaction(miner, alice.Address(), 20, 0)
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
	t.Cleanup(func() { _ = created.Close() })

	tx, err := created.SendTransaction(miner, alice.Address(), 20, 0)
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

func TestValidateBlockRejectsDuplicateSpendInsideBlock(t *testing.T) {
	dataDir := filepath.Join(t.TempDir(), "data")
	miner := mustNewWallet(t)
	alice := mustNewWallet(t)

	created, err := CreateBlockchain(dataDir, miner.Address())
	if err != nil {
		t.Fatalf("CreateBlockchain() error = %v", err)
	}
	t.Cleanup(func() { _ = created.Close() })

	tx1, err := created.SendTransaction(miner, alice.Address(), 20, 0)
	if err != nil {
		t.Fatalf("SendTransaction(tx1) error = %v", err)
	}
	// Clone tx1 and force a second conflicting spend reference for block-level validation.
	conflict := tx1.Clone()
	conflict.Outputs[0].Value = 15
	conflict.Outputs[1].Value = 35
	conflict.ID = conflict.Hash()
	if err := created.SignTransaction(&conflict, miner.PrivateKey); err != nil {
		t.Fatalf("SignTransaction(conflict) error = %v", err)
	}

	block := NewBlock([]Transaction{NewCoinbaseTransaction(miner.Address(), "reward"), tx1, conflict}, created.tip, 1)
	if err := created.ValidateBlock(block); !errors.Is(err, ErrDoubleSpend) {
		t.Fatalf("ValidateBlock(double spend) error = %v, want ErrDoubleSpend", err)
	}
}

func TestMerkleRootRejectsTampering(t *testing.T) {
	dataDir := filepath.Join(t.TempDir(), "data")
	miner := mustNewWallet(t)
	alice := mustNewWallet(t)

	created, err := CreateBlockchain(dataDir, miner.Address())
	if err != nil {
		t.Fatalf("CreateBlockchain() error = %v", err)
	}
	t.Cleanup(func() { _ = created.Close() })

	if _, err := created.SendTransaction(miner, alice.Address(), 20, 0); err != nil {
		t.Fatalf("SendTransaction() error = %v", err)
	}
	block, _, err := created.MineMempool(miner.Address())
	if err != nil {
		t.Fatalf("MineMempool() error = %v", err)
	}
	if !block.VerifyMerkleRoot() {
		t.Fatalf("VerifyMerkleRoot() = false, want true")
	}

	block.Transactions[0].Outputs[0].Value = 999
	if block.VerifyMerkleRoot() {
		t.Fatalf("VerifyMerkleRoot(tampered) = true, want false")
	}
}

func TestImportInvalidGenesisBlockToEmptyDirFails(t *testing.T) {
	dataDir := filepath.Join(t.TempDir(), "data")
	miner := mustNewWallet(t)
	alice := mustNewWallet(t)

	coinbaseA := NewCoinbaseTransaction(miner.Address(), "reward-a")
	coinbaseB := NewCoinbaseTransaction(alice.Address(), "reward-b")
	badGenesis := NewBlock([]Transaction{coinbaseA, coinbaseB}, nil, 0)

	if err := ImportBlockToDir(dataDir, badGenesis); !errors.Is(err, ErrInvalidBlock) {
		t.Fatalf("ImportBlockToDir(invalid genesis) error = %v, want ErrInvalidBlock", err)
	}

	height, err := BestHeight(dataDir)
	if err != nil {
		t.Fatalf("BestHeight() error = %v", err)
	}
	if height != -1 {
		t.Fatalf("BestHeight() = %d, want -1 for no imported chain", height)
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
