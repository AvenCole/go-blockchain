package wallet

import (
	"bytes"
	"crypto/sha256"
	"path/filepath"
	"testing"
)

func TestCreateWalletAndValidateAddress(t *testing.T) {
	wallet, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	address := wallet.Address()
	if address == "" {
		t.Fatalf("Address() returned empty string")
	}

	if !ValidateAddress(address) {
		t.Fatalf("ValidateAddress(%q) = false, want true", address)
	}
}

func TestValidateAddressRejectsWrongVersionAndChecksum(t *testing.T) {
	wallet, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	address := wallet.Address()
	decoded := Base58Decode(address)
	if len(decoded) < 1+checksumLength {
		t.Fatalf("decoded address too short")
	}

	payload := append([]byte(nil), decoded[:len(decoded)-checksumLength]...)
	payload[0] = 0x42
	wrongVersion := append(payload, testChecksum(payload)...)
	if ValidateAddress(Base58Encode(wrongVersion)) {
		t.Fatalf("ValidateAddress accepted wrong version")
	}

	badChecksum := append([]byte(nil), decoded...)
	badChecksum[len(badChecksum)-1] ^= 0x01
	if ValidateAddress(Base58Encode(badChecksum)) {
		t.Fatalf("ValidateAddress accepted bad checksum")
	}

	if ValidateAddress("0OIl-not-base58") {
		t.Fatalf("ValidateAddress accepted malformed Base58 input")
	}
}

func TestWalletsSaveAndLoad(t *testing.T) {
	dataDir := filepath.Join(t.TempDir(), "data")

	wallets, err := NewWallets(dataDir)
	if err != nil {
		t.Fatalf("NewWallets() error = %v", err)
	}

	address, err := wallets.CreateWallet()
	if err != nil {
		t.Fatalf("CreateWallet() error = %v", err)
	}

	if err := wallets.SaveFile(dataDir); err != nil {
		t.Fatalf("SaveFile() error = %v", err)
	}

	loaded, err := NewWallets(dataDir)
	if err != nil {
		t.Fatalf("NewWallets(load) error = %v", err)
	}

	addresses := loaded.Addresses()
	if len(addresses) != 1 {
		t.Fatalf("len(Addresses()) = %d, want 1", len(addresses))
	}

	if addresses[0] != address {
		t.Fatalf("loaded address = %q, want %q", addresses[0], address)
	}

	if _, ok := loaded.GetWallet(address); !ok {
		t.Fatalf("GetWallet(%q) = false, want true", address)
	}
}

func TestWalletAddressStableAfterLoad(t *testing.T) {
	dataDir := filepath.Join(t.TempDir(), "data")

	wallets, err := NewWallets(dataDir)
	if err != nil {
		t.Fatalf("NewWallets() error = %v", err)
	}

	address, err := wallets.CreateWallet()
	if err != nil {
		t.Fatalf("CreateWallet() error = %v", err)
	}

	before, ok := wallets.GetWallet(address)
	if !ok {
		t.Fatalf("GetWallet(before) = false, want true")
	}

	if err := wallets.SaveFile(dataDir); err != nil {
		t.Fatalf("SaveFile() error = %v", err)
	}

	loaded, err := NewWallets(dataDir)
	if err != nil {
		t.Fatalf("NewWallets(load) error = %v", err)
	}

	after, ok := loaded.GetWallet(address)
	if !ok {
		t.Fatalf("GetWallet(after) = false, want true")
	}

	if before.Address() != after.Address() {
		t.Fatalf("address before=%q after=%q, want stable round trip", before.Address(), after.Address())
	}

	if !bytes.Equal(before.PublicKey, after.PublicKey) {
		t.Fatalf("public key changed after load")
	}
}

func testChecksum(payload []byte) []byte {
	first := sha256.Sum256(payload)
	second := sha256.Sum256(first[:])
	return append([]byte(nil), second[:checksumLength]...)
}
