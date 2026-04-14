package gui

import (
	"fmt"
	"sort"
	"strings"

	"go-blockchain/internal/blockchain"
	"go-blockchain/internal/wallet"
)

func (s *Service) PendingTransactions() ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	bc, err := s.openBlockchain()
	if err == blockchain.ErrBlockchainNotInitialized {
		return []string{}, nil
	}
	if err != nil {
		return nil, err
	}
	defer bc.Close()
	txs, err := bc.PendingTransactions()
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(txs))
	for _, tx := range txs {
		ids = append(ids, tx.IDHex())
	}
	sort.Strings(ids)
	return ids, nil
}

func (s *Service) QueueTransaction(from, to string, amount int, fee int) (string, error) {
	if !s.isValidAddress(from) || !s.isValidAddress(to) {
		return "", fmt.Errorf("invalid wallet address")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	bc, err := s.openBlockchain()
	if err != nil {
		return "", friendlyGUIError(err)
	}
	defer bc.Close()

	wallets, err := s.loadWallets()
	if err != nil {
		return "", err
	}
	fromWallet, ok := wallets.GetWallet(from)
	if !ok {
		return "", fmt.Errorf("sender wallet not found")
	}

	tx, err := bc.SendTransaction(fromWallet, to, amount, fee)
	if err != nil {
		return "", friendlyGUIError(err)
	}

	return tx.IDHex(), nil
}

func (s *Service) QueueP2PKTransaction(from, to string, amount int, fee int) (string, error) {
	if !s.isValidAddress(from) || !s.isValidAddress(to) {
		return "", fmt.Errorf("invalid wallet address")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	bc, err := s.openBlockchain()
	if err != nil {
		return "", friendlyGUIError(err)
	}
	defer bc.Close()

	wallets, err := s.loadWallets()
	if err != nil {
		return "", err
	}
	fromWallet, ok := wallets.GetWallet(from)
	if !ok {
		return "", fmt.Errorf("sender wallet not found")
	}
	toWallet, ok := wallets.GetWallet(to)
	if !ok {
		return "", fmt.Errorf("recipient wallet not found locally")
	}

	tx, err := blockchain.NewP2PKTransaction(fromWallet, toWallet, amount, fee, bc)
	if err != nil {
		return "", friendlyGUIError(err)
	}
	if err := bc.AddToMempool(tx); err != nil {
		return "", friendlyGUIError(err)
	}
	return tx.IDHex(), nil
}

func (s *Service) QueueMultiSigTransaction(from string, recipientsCSV string, required int, amount int, fee int) (string, error) {
	if !s.isValidAddress(from) {
		return "", fmt.Errorf("invalid sender wallet address")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	bc, err := s.openBlockchain()
	if err != nil {
		return "", friendlyGUIError(err)
	}
	defer bc.Close()

	wallets, err := s.loadWallets()
	if err != nil {
		return "", err
	}
	fromWallet, ok := wallets.GetWallet(from)
	if !ok {
		return "", fmt.Errorf("sender wallet not found")
	}
	recipients, err := loadGUIWalletsFromCSV(wallets, recipientsCSV)
	if err != nil {
		return "", err
	}

	tx, err := blockchain.NewMultiSigTransaction(fromWallet, amount, fee, required, recipients, bc)
	if err != nil {
		return "", friendlyGUIError(err)
	}
	if err := bc.AddToMempool(tx); err != nil {
		return "", friendlyGUIError(err)
	}
	return tx.IDHex(), nil
}

func (s *Service) MinePending(minerAddress string) (string, error) {
	if !s.isValidAddress(minerAddress) {
		return "", fmt.Errorf("invalid miner address")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	bc, err := s.openBlockchain()
	if err != nil {
		return "", friendlyGUIError(err)
	}
	defer bc.Close()

	block, _, err := bc.MineMempool(minerAddress)
	if err != nil {
		return "", friendlyGUIError(err)
	}

	return block.HashHex(), nil
}

func loadGUIWalletsFromCSV(wallets walletProvider, csv string) ([]*wallet.Wallet, error) {
	items := strings.Split(csv, ",")
	result := make([]*wallet.Wallet, 0, len(items))
	for _, item := range items {
		address := strings.TrimSpace(item)
		if address == "" {
			continue
		}
		if !wallet.ValidateAddress(address) {
			return nil, fmt.Errorf("invalid recipient wallet address: %s", address)
		}
		w, ok := wallets.GetWallet(address)
		if !ok {
			return nil, fmt.Errorf("recipient wallet not found locally: %s", address)
		}
		result = append(result, w)
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("at least one recipient wallet is required")
	}
	return result, nil
}

type walletProvider interface {
	GetWallet(address string) (*wallet.Wallet, bool)
}
