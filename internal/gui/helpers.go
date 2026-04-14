package gui

import (
	"errors"
	"fmt"

	"go-blockchain/internal/blockchain"
	"go-blockchain/internal/wallet"
)

func (s *Service) openBlockchain() (*blockchain.Blockchain, error) {
	return s.ensureChain()
}

func (s *Service) isValidAddress(address string) bool {
	return wallet.ValidateAddress(address)
}

func friendlyGUIError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, blockchain.ErrBlockchainNotInitialized) {
		return fmt.Errorf("当前还没有初始化区块链。请先在命令行执行 createblockchain <钱包地址>，然后再回到 GUI 刷新")
	}
	return err
}
