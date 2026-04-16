package gui

import (
	"errors"
	"fmt"

	"go-blockchain/internal/blockchain"
	"go-blockchain/internal/wallet"
)

func (s *Service) openBlockchain() (*blockchain.Blockchain, error) {
	bc, err := blockchain.OpenBlockchain(s.cfg.DataDir)
	if err != nil {
		return nil, normalizeStorageError(err)
	}
	return bc, nil
}

func (s *Service) isValidAddress(address string) bool {
	return wallet.ValidateAddress(address)
}

func friendlyGUIError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, blockchain.ErrBlockchainNotInitialized) {
		return fmt.Errorf("当前还没有初始化区块链。请先在 GUI 中选择钱包并初始化主链")
	}
	return err
}
