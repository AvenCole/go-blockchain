package gui

import (
	"fmt"
	"path/filepath"
	"strconv"
	"time"

	"go-blockchain/internal/blockchain"
	"go-blockchain/internal/demo"
)

const demoWaitTimeout = 8 * time.Second

func (s *Service) RunNetworkQuickDemo() (NetworkDemoResult, error) {
	addresses, err := s.ensureWalletAddresses(2)
	if err != nil {
		return NetworkDemoResult{}, err
	}
	minerAddress := addresses[0]
	receiverAddress := addresses[1]

	sourceNode, err := s.StartNode("127.0.0.1:0", "", minerAddress)
	if err != nil {
		return NetworkDemoResult{}, err
	}
	peerNode, err := s.StartNode("127.0.0.1:0", "", "")
	if err != nil {
		return NetworkDemoResult{}, err
	}

	if err := s.InitializeNodeBlockchain(sourceNode, minerAddress); err != nil {
		return NetworkDemoResult{}, err
	}
	if err := s.ConnectNode(peerNode, sourceNode); err != nil {
		return NetworkDemoResult{}, err
	}

	if _, err := s.waitForNodeStatus(peerNode, demoWaitTimeout, func(node NodeStatus) bool {
		return node.Initialized && node.Height == 0
	}); err != nil {
		return NetworkDemoResult{}, err
	}

	txid, err := s.SubmitNodeTransaction(sourceNode, minerAddress, receiverAddress, 20, 1)
	if err != nil {
		return NetworkDemoResult{}, err
	}
	blockHash, err := s.MineNodePending(sourceNode)
	if err != nil {
		return NetworkDemoResult{}, err
	}

	peerStatus, err := s.waitForNodeStatus(peerNode, demoWaitTimeout, func(node NodeStatus) bool {
		return node.Initialized && node.Height >= 1
	})
	if err != nil {
		return NetworkDemoResult{}, err
	}
	sourceStatus, err := s.waitForNodeStatus(sourceNode, demoWaitTimeout, func(node NodeStatus) bool {
		for _, event := range node.RecentEvents {
			if event.Kind == "tip_announce" {
				return true
			}
		}
		return false
	})
	if err != nil {
		return NetworkDemoResult{}, err
	}

	result := NetworkDemoResult{
		SourceNode:      sourceNode,
		PeerNode:        peerNode,
		MinerAddress:    minerAddress,
		ReceiverAddress: receiverAddress,
		TxID:            txid,
		BlockHash:       blockHash,
		PeerHeight:      peerStatus.Height,
	}
	for _, event := range sourceStatus.RecentEvents {
		if event.Kind == "tip_announce" {
			result.TipAnnounced = true
			break
		}
	}
	return result, nil
}

func (s *Service) RunNetworkReorgDemo() (NetworkReorgDemoResult, error) {
	addresses, err := s.ensureWalletAddresses(2)
	if err != nil {
		return NetworkReorgDemoResult{}, err
	}
	minerAddress := addresses[0]
	receiverAddress := addresses[1]

	sourceNode, err := s.StartNode("127.0.0.1:0", "", minerAddress)
	if err != nil {
		return NetworkReorgDemoResult{}, err
	}
	peerNode, err := s.StartNode("127.0.0.1:0", "", "")
	if err != nil {
		return NetworkReorgDemoResult{}, err
	}

	if err := s.InitializeNodeBlockchain(sourceNode, minerAddress); err != nil {
		return NetworkReorgDemoResult{}, err
	}
	if err := s.ConnectNode(peerNode, sourceNode); err != nil {
		return NetworkReorgDemoResult{}, err
	}
	if _, err := s.waitForNodeStatus(peerNode, demoWaitTimeout, func(node NodeStatus) bool {
		return node.Initialized && node.Height == 0
	}); err != nil {
		return NetworkReorgDemoResult{}, err
	}

	txid, err := s.SubmitNodeTransaction(sourceNode, minerAddress, receiverAddress, 20, 1)
	if err != nil {
		return NetworkReorgDemoResult{}, err
	}
	blockHash, err := s.MineNodePending(sourceNode)
	if err != nil {
		return NetworkReorgDemoResult{}, err
	}
	minedStatus, err := s.waitForNodeStatus(sourceNode, demoWaitTimeout, func(node NodeStatus) bool {
		return node.Initialized && node.Height >= 1
	})
	if err != nil {
		return NetworkReorgDemoResult{}, err
	}
	if _, err := s.waitForNodeStatus(peerNode, demoWaitTimeout, func(node NodeStatus) bool {
		return node.Initialized && node.Height >= 1
	}); err != nil {
		return NetworkReorgDemoResult{}, err
	}

	sourceSession, err := s.nodeSessionByAddress(sourceNode)
	if err != nil {
		return NetworkReorgDemoResult{}, err
	}
	reorgResult, err := demo.ForceReorgToLongerGenesisFork(sourceSession.node.DataDir, minerAddress, txid, receiverAddress, 1)
	if err != nil {
		return NetworkReorgDemoResult{}, err
	}

	if err := s.ConnectNode(peerNode, sourceNode); err != nil {
		return NetworkReorgDemoResult{}, err
	}
	sourceStatus, err := s.waitForNodeStatus(sourceNode, demoWaitTimeout, func(node NodeStatus) bool {
		return node.LastReorg != nil && node.MempoolCount >= 1 && node.Height >= reorgResult.NewHeight
	})
	if err != nil {
		return NetworkReorgDemoResult{}, err
	}
	peerStatus, err := s.waitForNodeStatus(peerNode, demoWaitTimeout, func(node NodeStatus) bool {
		return node.LastReorg != nil && node.Height >= reorgResult.NewHeight
	})
	if err != nil {
		return NetworkReorgDemoResult{}, err
	}

	return NetworkReorgDemoResult{
		SourceNode:          sourceNode,
		PeerNode:            peerNode,
		MinerAddress:        minerAddress,
		ReceiverAddress:     receiverAddress,
		OriginalBlockHash:   blockHash,
		OriginalBlockHeight: minedStatus.Height,
		ReorgTxID:           reorgResult.ReorgTxID,
		Restored:            reorgResult.Restored && sourceStatus.MempoolCount >= 1,
		SourceOldHeight:     reorgResult.OldHeight,
		SourceNewHeight:     reorgResult.NewHeight,
		PeerHeight:          peerStatus.Height,
		PeerReorged:         peerStatus.LastReorg != nil,
	}, nil
}

func (s *Service) RunNetworkPartitionDemo() (NetworkPartitionDemoResult, error) {
	addresses, err := s.ensureWalletAddresses(2)
	if err != nil {
		return NetworkPartitionDemoResult{}, err
	}
	minerAddress := addresses[0]
	receiverAddress := addresses[1]

	sourceNode, err := s.StartNode("127.0.0.1:0", "", minerAddress)
	if err != nil {
		return NetworkPartitionDemoResult{}, err
	}
	peerNode, err := s.StartNode("127.0.0.1:0", "", "")
	if err != nil {
		return NetworkPartitionDemoResult{}, err
	}

	if err := s.InitializeNodeBlockchain(sourceNode, minerAddress); err != nil {
		return NetworkPartitionDemoResult{}, err
	}
	if err := s.ConnectNode(peerNode, sourceNode); err != nil {
		return NetworkPartitionDemoResult{}, err
	}
	if _, err := s.waitForNodeStatus(peerNode, demoWaitTimeout, func(node NodeStatus) bool {
		return node.Initialized && node.Height == 0
	}); err != nil {
		return NetworkPartitionDemoResult{}, err
	}

	forkNodeDir, err := s.cloneNodeDataDir(sourceNode)
	if err != nil {
		return NetworkPartitionDemoResult{}, err
	}
	forkNode, err := s.startManagedNode("127.0.0.1:0", "", minerAddress, forkNodeDir)
	if err != nil {
		return NetworkPartitionDemoResult{}, err
	}
	if _, err := s.waitForNodeStatus(forkNode, demoWaitTimeout, func(node NodeStatus) bool {
		return node.Initialized && node.Height == 0
	}); err != nil {
		return NetworkPartitionDemoResult{}, err
	}

	txid, err := s.SubmitNodeTransaction(sourceNode, minerAddress, receiverAddress, 20, 1)
	if err != nil {
		return NetworkPartitionDemoResult{}, err
	}
	if _, err := s.MineNodePending(sourceNode); err != nil {
		return NetworkPartitionDemoResult{}, err
	}
	sourceConfirmed, err := s.waitForNodeStatus(sourceNode, demoWaitTimeout, func(node NodeStatus) bool {
		return node.Initialized && node.Height >= 1
	})
	if err != nil {
		return NetworkPartitionDemoResult{}, err
	}
	if _, err := s.waitForNodeStatus(peerNode, demoWaitTimeout, func(node NodeStatus) bool {
		return node.Initialized && node.Height >= 1
	}); err != nil {
		return NetworkPartitionDemoResult{}, err
	}

	forkSession, err := s.nodeSessionByAddress(forkNode)
	if err != nil {
		return NetworkPartitionDemoResult{}, err
	}
	if _, err := demo.ForceReorgToLongerGenesisFork(forkSession.node.DataDir, minerAddress, "", receiverAddress, 2); err != nil {
		return NetworkPartitionDemoResult{}, err
	}
	forkStatus, err := s.waitForNodeStatus(forkNode, demoWaitTimeout, func(node NodeStatus) bool {
		return node.Initialized && node.Height >= 2 && node.TipHash != ""
	})
	if err != nil {
		return NetworkPartitionDemoResult{}, err
	}

	if err := s.ConnectNode(sourceNode, forkNode); err != nil {
		return NetworkPartitionDemoResult{}, err
	}
	sourceFinal, err := s.waitForNodeStatus(sourceNode, demoWaitTimeout, func(node NodeStatus) bool {
		return node.LastReorg != nil && node.Height >= forkStatus.Height && node.TipHash == forkStatus.TipHash
	})
	if err != nil {
		return NetworkPartitionDemoResult{}, err
	}
	peerFinal, err := s.waitForNodeStatus(peerNode, demoWaitTimeout, func(node NodeStatus) bool {
		return node.LastReorg != nil && node.Height >= forkStatus.Height && node.TipHash == forkStatus.TipHash
	})
	if err != nil {
		return NetworkPartitionDemoResult{}, err
	}

	sourceSession, err := s.nodeSessionByAddress(sourceNode)
	if err != nil {
		return NetworkPartitionDemoResult{}, err
	}
	restored, err := pendingTransactionExists(sourceSession.node.DataDir, txid)
	if err != nil {
		return NetworkPartitionDemoResult{}, err
	}

	return NetworkPartitionDemoResult{
		SourceNode:         sourceNode,
		PeerNode:           peerNode,
		ForkNode:           forkNode,
		MinerAddress:       minerAddress,
		ReceiverAddress:    receiverAddress,
		ConfirmedTxID:      txid,
		OldConfirmedHeight: sourceConfirmed.Height,
		ForkHeight:         forkStatus.Height,
		FinalTipHash:       forkStatus.TipHash,
		Restored:           restored,
		AllConverged:       restored && sourceFinal.TipHash == forkStatus.TipHash && peerFinal.TipHash == forkStatus.TipHash,
	}, nil
}

func (s *Service) ensureWalletAddresses(count int) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	wallets, err := s.loadWallets()
	if err != nil {
		return nil, err
	}

	addresses := wallets.Addresses()
	changed := false
	for len(addresses) < count {
		address, err := wallets.CreateWallet()
		if err != nil {
			return nil, err
		}
		addresses = append(addresses, address)
		changed = true
	}
	if changed {
		if err := wallets.SaveFile(s.cfg.DataDir); err != nil {
			return nil, err
		}
	}
	return addresses, nil
}

func (s *Service) waitForNodeStatus(address string, timeout time.Duration, predicate func(NodeStatus) bool) (NodeStatus, error) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		nodes, err := s.Nodes()
		if err != nil {
			return NodeStatus{}, err
		}
		for _, node := range nodes {
			if node.Address == address && predicate(node) {
				return node, nil
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
	return NodeStatus{}, fmt.Errorf("wait node status timeout: %s", address)
}

func (s *Service) cloneNodeDataDir(address string) (string, error) {
	session, err := s.nodeSessionByAddress(address)
	if err != nil {
		return "", err
	}

	cloneDir := filepath.Join(s.cfg.DataDir, "nodes", "demo-clone-"+strconv.FormatInt(time.Now().UnixNano(), 10))
	if err := demo.CopyDir(session.node.DataDir, cloneDir); err != nil {
		return "", err
	}
	return cloneDir, nil
}

func pendingTransactionExists(dataDir, txid string) (bool, error) {
	bc, err := blockchain.OpenBlockchain(dataDir)
	if err != nil {
		return false, err
	}
	defer bc.Close()

	pending, err := bc.PendingTransactions()
	if err != nil {
		return false, err
	}
	for _, tx := range pending {
		if tx.IDHex() == txid {
			return true, nil
		}
	}
	return false, nil
}
