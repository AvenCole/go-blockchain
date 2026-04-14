package wallet

import (
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type walletRecord struct {
	D []byte
	X []byte
	Y []byte
}

type walletFile struct {
	Records map[string]walletRecord
}

// Wallets manages a set of wallets stored on disk.
type Wallets struct {
	items map[string]*Wallet
}

// NewWallets loads or initializes a wallet collection for the given data dir.
func NewWallets(dataDir string) (*Wallets, error) {
	wallets := &Wallets{
		items: make(map[string]*Wallet),
	}

	if err := wallets.LoadFile(dataDir); err != nil {
		if os.IsNotExist(err) {
			return wallets, nil
		}

		return nil, err
	}

	return wallets, nil
}

// CreateWallet adds a new wallet and returns its address.
func (ws *Wallets) CreateWallet() (string, error) {
	wallet, err := New()
	if err != nil {
		return "", err
	}

	address := wallet.Address()
	ws.items[address] = wallet

	return address, nil
}

// Addresses returns the wallet addresses in sorted order.
func (ws *Wallets) Addresses() []string {
	addresses := make([]string, 0, len(ws.items))
	for address := range ws.items {
		addresses = append(addresses, address)
	}
	sort.Strings(addresses)

	return addresses
}

// GetWallet retrieves one wallet by address.
func (ws *Wallets) GetWallet(address string) (*Wallet, bool) {
	wallet, ok := ws.items[address]
	return wallet, ok
}

// SaveFile persists all wallets using gob.
func (ws *Wallets) SaveFile(dataDir string) error {
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return fmt.Errorf("create data dir: %w", err)
	}

	file, err := os.Create(walletFilePath(dataDir))
	if err != nil {
		return fmt.Errorf("create wallet file: %w", err)
	}
	defer file.Close()

	payload := walletFile{
		Records: make(map[string]walletRecord, len(ws.items)),
	}
	for address, wallet := range ws.items {
		x, y := wallet.PublicKeyCoordinates()
		payload.Records[address] = walletRecord{
			D: wallet.PrivateKeyBytes(),
			X: x,
			Y: y,
		}
	}

	if err := gob.NewEncoder(file).Encode(payload); err != nil {
		return fmt.Errorf("encode wallets: %w", err)
	}

	return nil
}

// LoadFile loads wallets from disk.
func (ws *Wallets) LoadFile(dataDir string) error {
	file, err := os.Open(walletFilePath(dataDir))
	if err != nil {
		return err
	}
	defer file.Close()

	var payload walletFile
	if err := gob.NewDecoder(file).Decode(&payload); err != nil {
		return fmt.Errorf("decode wallets: %w", err)
	}

	ws.items = make(map[string]*Wallet, len(payload.Records))
	for address, record := range payload.Records {
		wallet, err := FromRecord(record.D, record.X, record.Y)
		if err != nil {
			return fmt.Errorf("restore wallet %s: %w", address, err)
		}
		ws.items[address] = wallet
	}

	return nil
}

func walletFilePath(dataDir string) string {
	return filepath.Join(dataDir, "wallets.gob")
}
