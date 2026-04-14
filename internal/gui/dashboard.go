package gui

import "go-blockchain/internal/blockchain"

func (s *Service) Dashboard() (DashboardData, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data := DashboardData{
		Height:      -1,
		DataDir:     s.cfg.DataDir,
		NetworkMode: s.cfg.NetworkMode,
	}

	wallets, err := s.loadWallets()
	if err == nil {
		data.WalletCount = len(wallets.Addresses())
	}

	bc, err := s.openBlockchain()
	if err == blockchain.ErrBlockchainNotInitialized {
		return data, nil
	}
	if err != nil {
		return DashboardData{}, err
	}
	defer bc.Close()
	current, err := bc.CurrentBlock()
	if err != nil {
		return DashboardData{}, err
	}

	mempoolSize, err := bc.MempoolSize()
	if err != nil {
		return DashboardData{}, err
	}

	data.Height = current.Height
	data.LatestHash = current.HashHex()
	data.MerkleRoot = current.MerkleRootHex()
	data.Difficulty = current.Difficulty
	data.Nonce = current.Nonce
	data.PendingTxCount = mempoolSize
	if status, err := bc.LastReorgStatus(); err == nil && status != nil {
		data.LastReorg = &ReorgStatusView{
			Timestamp:             status.Timestamp,
			OldHeight:             status.OldHeight,
			NewHeight:             status.NewHeight,
			OldTip:                status.OldTip,
			NewTip:                status.NewTip,
			RestoredTxCount:       status.RestoredTxCount,
			DroppedConfirmedCount: status.DroppedConfirmedCount,
		}
	}

	return data, nil
}
