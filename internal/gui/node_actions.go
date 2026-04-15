package gui

import (
	"errors"
	"fmt"

	"go-blockchain/internal/blockchain"
)

func (s *Service) nodeSessionByAddress(address string) (*nodeSession, error) {
	s.nodeMu.Lock()
	defer s.nodeMu.Unlock()
	s.ensureNodeMap()

	session, ok := s.nodes[address]
	if !ok {
		return nil, fmt.Errorf("节点不存在")
	}
	return session, nil
}

func (s *Service) InitializeNodeBlockchain(address, rewardAddress string) error {
	session, err := s.nodeSessionByAddress(address)
	if err != nil {
		return err
	}

	reward := rewardAddress
	if reward == "" {
		reward = session.node.MinerAddress
	}
	if !s.isValidAddress(reward) {
		return fmt.Errorf("无效创世奖励地址")
	}

	if err := session.node.EnsureBlockchain(reward); err != nil {
		return normalizeStorageError(err)
	}
	return nil
}

func (s *Service) SubmitNodeTransaction(nodeAddress, from, to string, amount int, fee int) (string, error) {
	if !s.isValidAddress(from) || !s.isValidAddress(to) {
		return "", fmt.Errorf("invalid wallet address")
	}

	session, err := s.nodeSessionByAddress(nodeAddress)
	if err != nil {
		return "", err
	}

	s.mu.Lock()
	wallets, err := s.loadWallets()
	s.mu.Unlock()
	if err != nil {
		return "", err
	}
	fromWallet, ok := wallets.GetWallet(from)
	if !ok {
		return "", fmt.Errorf("sender wallet not found")
	}

	tx, err := session.node.SubmitTransaction(fromWallet, to, amount, fee)
	if err != nil {
		return "", friendlyNodeError(err)
	}
	return tx.IDHex(), nil
}

func (s *Service) MineNodePending(address string) (string, error) {
	session, err := s.nodeSessionByAddress(address)
	if err != nil {
		return "", err
	}

	block, err := session.node.MinePending()
	if err != nil {
		return "", friendlyNodeError(err)
	}
	return block.HashHex(), nil
}

func friendlyNodeError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, blockchain.ErrBlockchainNotInitialized) {
		return fmt.Errorf("当前节点还没有链数据。请先为该节点初始化区块链，或先连接已有 seed 等待同步")
	}
	return normalizeStorageError(err)
}
