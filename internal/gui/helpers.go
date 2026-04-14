package gui

import (
	"go-blockchain/internal/blockchain"
	"go-blockchain/internal/wallet"
)

func (s *Service) openBlockchain() (*blockchain.Blockchain, error) {
	return s.ensureChain()
}

func (s *Service) isValidAddress(address string) bool {
	return wallet.ValidateAddress(address)
}
