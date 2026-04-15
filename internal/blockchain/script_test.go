package blockchain

import (
	"bytes"
	"path/filepath"
	"testing"
)

func TestNewTransactionUsesScriptVM(t *testing.T) {
	dataDir := filepath.Join(t.TempDir(), "data")
	miner := mustNewWallet(t)
	alice := mustNewWallet(t)

	created, err := CreateBlockchain(dataDir, miner.Address())
	if err != nil {
		t.Fatalf("CreateBlockchain() error = %v", err)
	}
	t.Cleanup(func() { _ = created.Close() })

	tx, err := created.SendTransaction(miner, alice.Address(), 20, 1)
	if err != nil {
		t.Fatalf("SendTransaction() error = %v", err)
	}

	if !tx.UsesScriptVM() {
		t.Fatalf("tx.UsesScriptVM() = false, want true")
	}
	if tx.Inputs[0].EffectiveScriptSig().IsEmpty() {
		t.Fatalf("input scriptSig is empty")
	}
	if tx.Outputs[0].EffectiveScriptPubKey().IsEmpty() {
		t.Fatalf("output scriptPubKey is empty")
	}
	if !created.VerifyTransaction(tx) {
		t.Fatalf("VerifyTransaction(tx) = false, want true")
	}
}

func TestVerifyTransactionRejectsTamperedScriptSig(t *testing.T) {
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

	tampered := tx.Clone()
	tampered.Inputs[0].ScriptSig.Commands[0].Data[0] ^= 0x01
	if created.VerifyTransaction(tampered) {
		t.Fatalf("VerifyTransaction(tampered scriptSig) = true, want false")
	}
}

func TestLegacyTransactionVerificationStillWorks(t *testing.T) {
	dataDir := filepath.Join(t.TempDir(), "data")
	miner := mustNewWallet(t)
	alice := mustNewWallet(t)

	created, err := CreateBlockchain(dataDir, miner.Address())
	if err != nil {
		t.Fatalf("CreateBlockchain() error = %v", err)
	}
	t.Cleanup(func() { _ = created.Close() })

	mainOutput, err := NewTXOutput(20, alice.Address())
	if err != nil {
		t.Fatalf("NewTXOutput() error = %v", err)
	}
	legacyTx, err := buildSpendTransaction(miner, mainOutput, 20, 0, created, 0)
	if err != nil {
		t.Fatalf("buildSpendTransaction(legacy) error = %v", err)
	}
	if legacyTx.UsesScriptVM() {
		t.Fatalf("legacyTx.UsesScriptVM() = true, want false")
	}

	if !created.VerifyTransaction(legacyTx) {
		t.Fatalf("VerifyTransaction(legacyTx) = false, want true")
	}

	prevTx, err := created.FindTransaction(legacyTx.Inputs[0].TxID)
	if err != nil {
		t.Fatalf("FindTransaction() error = %v", err)
	}
	prevTx.Outputs[legacyTx.Inputs[0].Out].ScriptPubKey = Script{}
	legacyTx.Inputs[0].ScriptSig = Script{}

	if !legacyTx.Verify(map[string]Transaction{
		legacyTx.Inputs[0].TxIDHex(): prevTx,
	}) {
		t.Fatalf("legacy verification fallback = false, want true")
	}
}

func TestExtractP2PKHPubKeyHash(t *testing.T) {
	pubKeyHash := bytes.Repeat([]byte{0x42}, 20)
	script := NewP2PKHLockingScript(pubKeyHash)

	got, ok := ExtractP2PKHPubKeyHash(script)
	if !ok {
		t.Fatalf("ExtractP2PKHPubKeyHash() ok = false, want true")
	}
	if !bytes.Equal(got, pubKeyHash) {
		t.Fatalf("ExtractP2PKHPubKeyHash() = %x, want %x", got, pubKeyHash)
	}
}

func TestP2PKTransactionUsesScriptVM(t *testing.T) {
	dataDir := filepath.Join(t.TempDir(), "data")
	miner := mustNewWallet(t)
	alice := mustNewWallet(t)

	created, err := CreateBlockchain(dataDir, miner.Address())
	if err != nil {
		t.Fatalf("CreateBlockchain() error = %v", err)
	}
	t.Cleanup(func() { _ = created.Close() })

	tx, err := NewP2PKTransaction(miner, alice, 20, 0, created)
	if err != nil {
		t.Fatalf("NewP2PKTransaction() error = %v", err)
	}
	if !tx.UsesScriptVM() {
		t.Fatalf("tx.UsesScriptVM() = false, want true")
	}
	if _, ok := ExtractP2PKPubKey(tx.Outputs[0].EffectiveScriptPubKey()); !ok {
		t.Fatalf("main output is not P2PK")
	}
	if !created.VerifyTransaction(tx) {
		t.Fatalf("VerifyTransaction(tx) = false, want true")
	}
}

func TestVerifyP2PKTransactionRejectsTamperedSignature(t *testing.T) {
	dataDir := filepath.Join(t.TempDir(), "data")
	miner := mustNewWallet(t)
	alice := mustNewWallet(t)

	created, err := CreateBlockchain(dataDir, miner.Address())
	if err != nil {
		t.Fatalf("CreateBlockchain() error = %v", err)
	}
	t.Cleanup(func() { _ = created.Close() })

	tx, err := NewP2PKTransaction(miner, alice, 20, 0, created)
	if err != nil {
		t.Fatalf("NewP2PKTransaction() error = %v", err)
	}

	tampered := tx.Clone()
	tampered.Inputs[0].ScriptSig.Commands[0].Data[0] ^= 0x01
	if created.VerifyTransaction(tampered) {
		t.Fatalf("VerifyTransaction(tampered p2pk) = true, want false")
	}
}

func TestSpendingP2PKOutputUsesP2PKUnlockingScript(t *testing.T) {
	dataDir := filepath.Join(t.TempDir(), "data")
	miner := mustNewWallet(t)
	alice := mustNewWallet(t)

	created, err := CreateBlockchain(dataDir, miner.Address())
	if err != nil {
		t.Fatalf("CreateBlockchain() error = %v", err)
	}
	t.Cleanup(func() { _ = created.Close() })

	firstTx, err := NewP2PKTransaction(miner, alice, 20, 0, created)
	if err != nil {
		t.Fatalf("NewP2PKTransaction() error = %v", err)
	}
	if err := created.AddToMempool(firstTx); err != nil {
		t.Fatalf("AddToMempool(firstTx) error = %v", err)
	}
	if _, _, err := created.MineMempool(miner.Address()); err != nil {
		t.Fatalf("MineMempool() error = %v", err)
	}

	secondTx, err := NewUTXOTransaction(alice, miner.Address(), 10, 0, created)
	if err != nil {
		t.Fatalf("NewUTXOTransaction() error = %v", err)
	}
	if _, ok := ExtractP2PKSignature(secondTx.Inputs[0].EffectiveScriptSig()); !ok {
		t.Fatalf("alice unlocking script is not P2PK")
	}
	if !created.VerifyTransaction(secondTx) {
		t.Fatalf("VerifyTransaction(secondTx) = false, want true")
	}
}
