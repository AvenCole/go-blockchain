package gui

import (
	"fmt"
	"sort"

	"go-blockchain/internal/blockchain"
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
		return "", err
	}

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
		return "", err
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
		return "", err
	}

	block, _, err := bc.MineMempool(minerAddress)
	if err != nil {
		return "", err
	}

	return block.HashHex(), nil
}
