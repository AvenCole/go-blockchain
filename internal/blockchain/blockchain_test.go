package blockchain

import (
	"bytes"
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
	if minedTx.Inputs[0].EffectiveScriptSig().IsEmpty() {
		t.Fatalf("tx scriptSig missing")
	}
	if minedTx.Outputs[0].EffectiveScriptPubKey().IsEmpty() {
		t.Fatalf("tx scriptPubKey missing")
	}

	events, err := created.RecentChainEvents(5)
	if err != nil {
		t.Fatalf("RecentChainEvents() error = %v", err)
	}
	if len(events) < 2 {
		t.Fatalf("len(events) = %d, want at least 2", len(events))
	}
	if events[0].Kind != "main_block" {
		t.Fatalf("events[0].Kind = %q, want main_block", events[0].Kind)
	}
	if events[1].Kind != "genesis" {
		t.Fatalf("events[1].Kind = %q, want genesis", events[1].Kind)
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

func TestImportBlockSwitchesToLongerFork(t *testing.T) {
	dataDir := filepath.Join(t.TempDir(), "data")
	miner := mustNewWallet(t)

	created, err := CreateBlockchain(dataDir, miner.Address())
	if err != nil {
		t.Fatalf("CreateBlockchain() error = %v", err)
	}
	t.Cleanup(func() { _ = created.Close() })

	if _, err := created.AddBlock([]Transaction{NewCoinbaseTransaction(miner.Address(), "main-1")}); err != nil {
		t.Fatalf("AddBlock(main-1) error = %v", err)
	}
	if _, err := created.AddBlock([]Transaction{NewCoinbaseTransaction(miner.Address(), "main-2")}); err != nil {
		t.Fatalf("AddBlock(main-2) error = %v", err)
	}

	mainTip, err := created.CurrentBlock()
	if err != nil {
		t.Fatalf("CurrentBlock() error = %v", err)
	}
	if mainTip.Height != 2 {
		t.Fatalf("main tip height = %d, want 2", mainTip.Height)
	}

	blocks, err := created.Blocks()
	if err != nil {
		t.Fatalf("Blocks() error = %v", err)
	}
	genesis := blocks[len(blocks)-1]

	fork1 := NewBlock([]Transaction{NewCoinbaseTransaction(miner.Address(), "fork-1")}, genesis.Hash, 1)
	if err := created.ImportBlock(fork1); err != nil {
		t.Fatalf("ImportBlock(fork1) error = %v", err)
	}
	current, _ := created.CurrentBlock()
	if !bytes.Equal(current.Hash, mainTip.Hash) {
		t.Fatalf("current tip switched on shorter fork")
	}

	fork2 := NewBlock([]Transaction{NewCoinbaseTransaction(miner.Address(), "fork-2")}, fork1.Hash, 2)
	if err := created.ImportBlock(fork2); err != nil {
		t.Fatalf("ImportBlock(fork2) error = %v", err)
	}
	current, _ = created.CurrentBlock()
	if !bytes.Equal(current.Hash, mainTip.Hash) {
		t.Fatalf("current tip switched on equal-height fork")
	}

	fork3 := NewBlock([]Transaction{NewCoinbaseTransaction(miner.Address(), "fork-3")}, fork2.Hash, 3)
	if err := created.ImportBlock(fork3); err != nil {
		t.Fatalf("ImportBlock(fork3) error = %v", err)
	}

	current, err = created.CurrentBlock()
	if err != nil {
		t.Fatalf("CurrentBlock() after fork error = %v", err)
	}
	if !bytes.Equal(current.Hash, fork3.Hash) {
		t.Fatalf("current tip hash = %x, want fork3 %x", current.Hash, fork3.Hash)
	}

	allBlocks, err := created.Blocks()
	if err != nil {
		t.Fatalf("Blocks() after fork error = %v", err)
	}
	if len(allBlocks) != 4 {
		t.Fatalf("len(Blocks()) = %d, want 4 for fork branch", len(allBlocks))
	}
	if allBlocks[0].Height != 3 || allBlocks[1].Height != 2 || allBlocks[2].Height != 1 || allBlocks[3].Height != 0 {
		t.Fatalf("unexpected branch heights after switch")
	}

	events, err := created.RecentChainEvents(10)
	if err != nil {
		t.Fatalf("RecentChainEvents() error = %v", err)
	}
	foundForkStore := false
	foundReorg := false
	for _, event := range events {
		if event.Kind == "fork_block" {
			foundForkStore = true
		}
		if event.Kind == "reorg" {
			foundReorg = true
		}
	}
	if !foundForkStore {
		t.Fatalf("expected fork_block event in recent history")
	}
	if !foundReorg {
		t.Fatalf("expected reorg event in recent history")
	}
}

func TestImportBlockRejectsUnknownParent(t *testing.T) {
	dataDir := filepath.Join(t.TempDir(), "data")
	miner := mustNewWallet(t)

	created, err := CreateBlockchain(dataDir, miner.Address())
	if err != nil {
		t.Fatalf("CreateBlockchain() error = %v", err)
	}
	t.Cleanup(func() { _ = created.Close() })

	orphan := NewBlock([]Transaction{NewCoinbaseTransaction(miner.Address(), "orphan")}, bytes.Repeat([]byte{0x42}, 32), 1)
	if err := created.ImportBlock(orphan); !errors.Is(err, ErrOrphanBlock) {
		t.Fatalf("ImportBlock(orphan) error = %v, want ErrOrphanBlock", err)
	}
}

func TestReorgRestoresDisconnectedTransactionsToMempool(t *testing.T) {
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
	if _, _, err := created.MineMempool(miner.Address()); err != nil {
		t.Fatalf("MineMempool() error = %v", err)
	}
	size, err := created.MempoolSize()
	if err != nil {
		t.Fatalf("MempoolSize() error = %v", err)
	}
	if size != 0 {
		t.Fatalf("mempool size after mine = %d, want 0", size)
	}

	blocks, err := created.Blocks()
	if err != nil {
		t.Fatalf("Blocks() error = %v", err)
	}
	genesis := blocks[len(blocks)-1]

	fork1 := NewBlock([]Transaction{NewCoinbaseTransaction(miner.Address(), "fork-1")}, genesis.Hash, 1)
	if err := created.ImportBlock(fork1); err != nil {
		t.Fatalf("ImportBlock(fork1) error = %v", err)
	}
	fork2 := NewBlock([]Transaction{NewCoinbaseTransaction(miner.Address(), "fork-2")}, fork1.Hash, 2)
	if err := created.ImportBlock(fork2); err != nil {
		t.Fatalf("ImportBlock(fork2) error = %v", err)
	}

	pending, err := created.PendingTransactions()
	if err != nil {
		t.Fatalf("PendingTransactions() error = %v", err)
	}
	if len(pending) != 1 {
		t.Fatalf("len(pending) = %d, want 1 restored tx", len(pending))
	}
	if pending[0].IDHex() != tx.IDHex() {
		t.Fatalf("restored txid = %s, want %s", pending[0].IDHex(), tx.IDHex())
	}

	aliceBalance, err := created.BalanceOf(alice.Address())
	if err != nil {
		t.Fatalf("BalanceOf(alice) error = %v", err)
	}
	if aliceBalance != 0 {
		t.Fatalf("alice balance after reorg = %d, want 0", aliceBalance)
	}

	status, err := created.LastReorgStatus()
	if err != nil {
		t.Fatalf("LastReorgStatus() error = %v", err)
	}
	if status == nil {
		t.Fatalf("LastReorgStatus() = nil, want status")
	}
	if status.RestoredTxCount != 1 {
		t.Fatalf("status.RestoredTxCount = %d, want 1", status.RestoredTxCount)
	}
	if status.NewHeight != 2 || status.OldHeight != 1 {
		t.Fatalf("unexpected reorg heights old=%d new=%d", status.OldHeight, status.NewHeight)
	}

	events, err := created.RecentChainEvents(5)
	if err != nil {
		t.Fatalf("RecentChainEvents() error = %v", err)
	}
	if len(events) == 0 {
		t.Fatalf("len(events) = 0, want at least one reorg event")
	}
	if events[0].Kind != "reorg" {
		t.Fatalf("events[0].Kind = %q, want reorg", events[0].Kind)
	}
	if events[0].RestoredTxCount != 1 {
		t.Fatalf("events[0].RestoredTxCount = %d, want 1", events[0].RestoredTxCount)
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
