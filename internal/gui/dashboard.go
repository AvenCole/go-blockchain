package gui

import "go-blockchain/internal/blockchain"

func (s *Service) Dashboard() (DashboardData, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data := DashboardData{
		Height:       -1,
		DataDir:      s.cfg.DataDir,
		NetworkMode:  s.cfg.NetworkMode,
		RecentEvents: []ChainEventView{},
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
	if events, err := bc.RecentChainEvents(5); err == nil && len(events) > 0 {
		data.RecentEvents = make([]ChainEventView, 0, len(events))
		for _, event := range events {
			data.RecentEvents = append(data.RecentEvents, ChainEventView{
				Timestamp:             event.Timestamp,
				Kind:                  event.Kind,
				Summary:               event.Summary,
				OldHeight:             event.OldHeight,
				NewHeight:             event.NewHeight,
				OldTip:                event.OldTip,
				NewTip:                event.NewTip,
				RestoredTxCount:       event.RestoredTxCount,
				DroppedConfirmedCount: event.DroppedConfirmedCount,
			})
		}
	}

	return data, nil
}
