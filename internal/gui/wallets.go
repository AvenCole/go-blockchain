package gui

import (
	"go-blockchain/internal/blockchain"
	"go-blockchain/internal/wallet"
)

func (s *Service) Wallets() ([]WalletView, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	wallets, err := s.loadWallets()
	if err != nil {
		return nil, err
	}

	addresses := wallets.Addresses()
	views := make([]WalletView, 0, len(addresses))

	bc, err := s.openBlockchain()
	if err == blockchain.ErrBlockchainNotInitialized {
		bc = nil
	} else if err != nil {
		return nil, err
	}
	for _, address := range addresses {
		balance := 0
		if bc != nil {
			balance, err = bc.BalanceOf(address)
			if err != nil {
				return nil, err
			}
		}

		views = append(views, WalletView{
			Address: address,
			Balance: balance,
		})
	}

	return views, nil
}

func (s *Service) CreateWallet() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	wallets, err := s.loadWallets()
	if err != nil {
		return "", err
	}

	address, err := wallets.CreateWallet()
	if err != nil {
		return "", err
	}
	if err := wallets.SaveFile(s.cfg.DataDir); err != nil {
		return "", err
	}

	return address, nil
}

func (s *Service) loadWallets() (*wallet.Wallets, error) {
	return wallet.NewWallets(s.cfg.DataDir)
}
