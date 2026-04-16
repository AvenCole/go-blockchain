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
	tracker := newNetworkOperationTracker(s, "network.quick-demo", 7)
	tracker.start("prepare_wallets", "准备网络流程所需钱包")

	addresses, err := s.ensureWalletAddresses(2)
	if err != nil {
		tracker.fail("prepare_wallets", err)
		return NetworkDemoResult{}, err
	}
	minerAddress := addresses[0]
	receiverAddress := addresses[1]
	tracker.step("prepare_wallets", "钱包准备完成")

	sourceNode, err := s.StartNode("127.0.0.1:0", "", minerAddress)
	if err != nil {
		tracker.fail("start_source_node", err)
		return NetworkDemoResult{}, err
	}
	tracker.step("start_source_node", "主节点启动完成")
	peerNode, err := s.StartNode("127.0.0.1:0", "", "")
	if err != nil {
		tracker.fail("start_peer_node", err)
		return NetworkDemoResult{}, err
	}
	tracker.step("start_peer_node", "对端节点启动完成")

	if err := s.InitializeNodeBlockchain(sourceNode, minerAddress); err != nil {
		tracker.fail("initialize_source_chain", err)
		return NetworkDemoResult{}, err
	}
	if err := s.ConnectNode(peerNode, sourceNode); err != nil {
		tracker.fail("connect_peer", err)
		return NetworkDemoResult{}, err
	}
	tracker.step("connect_peer", "节点连接完成，等待同步创世链")

	if _, err := s.waitForNodeStatus(peerNode, demoWaitTimeout, func(node NodeStatus) bool {
		return node.Initialized && node.Height == 0
	}); err != nil {
		tracker.fail("wait_peer_sync", err)
		return NetworkDemoResult{}, err
	}
	tracker.step("wait_peer_sync", "对端节点已同步到初始链状态")

	txid, err := s.SubmitNodeTransaction(sourceNode, minerAddress, receiverAddress, 20, 1)
	if err != nil {
		tracker.fail("submit_transaction", err)
		return NetworkDemoResult{}, err
	}
	blockHash, err := s.MineNodePending(sourceNode)
	if err != nil {
		tracker.fail("mine_block", err)
		return NetworkDemoResult{}, err
	}
	tracker.step("mine_block", "交易提交并完成出块")

	peerStatus, err := s.waitForNodeStatus(peerNode, demoWaitTimeout, func(node NodeStatus) bool {
		return node.Initialized && node.Height >= 1
	})
	if err != nil {
		tracker.fail("wait_peer_block_sync", err)
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
		tracker.fail("wait_tip_announce", err)
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
	tracker.step("wait_tip_announce", "对端完成新区块同步并收到 tip 通告")
	tracker.complete(
		"completed",
		"快速同步流程完成",
		fmt.Sprintf("source=%s peer=%s tx=%s block=%s", sourceNode, peerNode, txid, blockHash),
	)
	return result, nil
}

func (s *Service) RunNetworkReorgDemo() (NetworkReorgDemoResult, error) {
	tracker := newNetworkOperationTracker(s, "network.reorg-demo", 7)
	tracker.start("prepare_wallets", "准备重组流程所需钱包")

	addresses, err := s.ensureWalletAddresses(2)
	if err != nil {
		tracker.fail("prepare_wallets", err)
		return NetworkReorgDemoResult{}, err
	}
	minerAddress := addresses[0]
	receiverAddress := addresses[1]
	tracker.step("prepare_wallets", "钱包准备完成")

	sourceNode, err := s.StartNode("127.0.0.1:0", "", minerAddress)
	if err != nil {
		tracker.fail("start_source_node", err)
		return NetworkReorgDemoResult{}, err
	}
	tracker.step("start_source_node", "主节点启动完成")
	peerNode, err := s.StartNode("127.0.0.1:0", "", "")
	if err != nil {
		tracker.fail("start_peer_node", err)
		return NetworkReorgDemoResult{}, err
	}
	tracker.step("start_peer_node", "对端节点启动完成")

	if err := s.InitializeNodeBlockchain(sourceNode, minerAddress); err != nil {
		tracker.fail("initialize_source_chain", err)
		return NetworkReorgDemoResult{}, err
	}
	if err := s.ConnectNode(peerNode, sourceNode); err != nil {
		tracker.fail("connect_peer", err)
		return NetworkReorgDemoResult{}, err
	}
	if _, err := s.waitForNodeStatus(peerNode, demoWaitTimeout, func(node NodeStatus) bool {
		return node.Initialized && node.Height == 0
	}); err != nil {
		tracker.fail("wait_peer_sync", err)
		return NetworkReorgDemoResult{}, err
	}
	tracker.step("wait_peer_sync", "对端节点已同步到初始链状态")

	txid, err := s.SubmitNodeTransaction(sourceNode, minerAddress, receiverAddress, 20, 1)
	if err != nil {
		tracker.fail("submit_transaction", err)
		return NetworkReorgDemoResult{}, err
	}
	blockHash, err := s.MineNodePending(sourceNode)
	if err != nil {
		tracker.fail("mine_original_block", err)
		return NetworkReorgDemoResult{}, err
	}
	minedStatus, err := s.waitForNodeStatus(sourceNode, demoWaitTimeout, func(node NodeStatus) bool {
		return node.Initialized && node.Height >= 1
	})
	if err != nil {
		tracker.fail("wait_source_confirm", err)
		return NetworkReorgDemoResult{}, err
	}
	if _, err := s.waitForNodeStatus(peerNode, demoWaitTimeout, func(node NodeStatus) bool {
		return node.Initialized && node.Height >= 1
	}); err != nil {
		tracker.fail("wait_peer_confirm", err)
		return NetworkReorgDemoResult{}, err
	}
	tracker.step("wait_peer_confirm", "原始确认区块已在双节点上完成同步")

	sourceSession, err := s.nodeSessionByAddress(sourceNode)
	if err != nil {
		tracker.fail("load_source_session", err)
		return NetworkReorgDemoResult{}, err
	}
	reorgResult, err := demo.ForceReorgToLongerGenesisFork(sourceSession.node.DataDir, minerAddress, txid, receiverAddress, 1)
	if err != nil {
		tracker.fail("force_reorg", err)
		return NetworkReorgDemoResult{}, err
	}
	tracker.step("force_reorg", "主节点已注入更长分叉链")

	if err := s.ConnectNode(peerNode, sourceNode); err != nil {
		tracker.fail("reconnect_peer", err)
		return NetworkReorgDemoResult{}, err
	}
	sourceStatus, err := s.waitForNodeStatus(sourceNode, demoWaitTimeout, func(node NodeStatus) bool {
		return node.LastReorg != nil && node.MempoolCount >= 1 && node.Height >= reorgResult.NewHeight
	})
	if err != nil {
		tracker.fail("wait_source_reorg", err)
		return NetworkReorgDemoResult{}, err
	}
	peerStatus, err := s.waitForNodeStatus(peerNode, demoWaitTimeout, func(node NodeStatus) bool {
		return node.LastReorg != nil && node.Height >= reorgResult.NewHeight
	})
	if err != nil {
		tracker.fail("wait_peer_reorg", err)
		return NetworkReorgDemoResult{}, err
	}
	tracker.step("wait_peer_reorg", "双节点完成重组并恢复交易")

	result := NetworkReorgDemoResult{
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
	}
	tracker.complete(
		"completed",
		"重组流程完成",
		fmt.Sprintf("source=%s peer=%s reorgTx=%s restored=%t", sourceNode, peerNode, reorgResult.ReorgTxID, result.Restored),
	)
	return result, nil
}

func (s *Service) RunNetworkPartitionDemo() (NetworkPartitionDemoResult, error) {
	tracker := newNetworkOperationTracker(s, "network.partition-demo", 8)
	tracker.start("prepare_wallets", "准备分区流程所需钱包")

	addresses, err := s.ensureWalletAddresses(2)
	if err != nil {
		tracker.fail("prepare_wallets", err)
		return NetworkPartitionDemoResult{}, err
	}
	minerAddress := addresses[0]
	receiverAddress := addresses[1]
	tracker.step("prepare_wallets", "钱包准备完成")

	sourceNode, err := s.StartNode("127.0.0.1:0", "", minerAddress)
	if err != nil {
		tracker.fail("start_source_node", err)
		return NetworkPartitionDemoResult{}, err
	}
	tracker.step("start_source_node", "主节点启动完成")
	peerNode, err := s.StartNode("127.0.0.1:0", "", "")
	if err != nil {
		tracker.fail("start_peer_node", err)
		return NetworkPartitionDemoResult{}, err
	}
	tracker.step("start_peer_node", "对端节点启动完成")

	if err := s.InitializeNodeBlockchain(sourceNode, minerAddress); err != nil {
		tracker.fail("initialize_source_chain", err)
		return NetworkPartitionDemoResult{}, err
	}
	if err := s.ConnectNode(peerNode, sourceNode); err != nil {
		tracker.fail("connect_peer", err)
		return NetworkPartitionDemoResult{}, err
	}
	if _, err := s.waitForNodeStatus(peerNode, demoWaitTimeout, func(node NodeStatus) bool {
		return node.Initialized && node.Height == 0
	}); err != nil {
		tracker.fail("wait_peer_sync", err)
		return NetworkPartitionDemoResult{}, err
	}
	tracker.step("wait_peer_sync", "source 与 peer 已同步初始链")

	forkNodeDir, err := s.cloneNodeDataDir(sourceNode)
	if err != nil {
		tracker.fail("clone_fork_data", err)
		return NetworkPartitionDemoResult{}, err
	}
	forkNode, err := s.startManagedNode("127.0.0.1:0", "", minerAddress, forkNodeDir)
	if err != nil {
		tracker.fail("start_fork_node", err)
		return NetworkPartitionDemoResult{}, err
	}
	if _, err := s.waitForNodeStatus(forkNode, demoWaitTimeout, func(node NodeStatus) bool {
		return node.Initialized && node.Height == 0
	}); err != nil {
		tracker.fail("wait_fork_ready", err)
		return NetworkPartitionDemoResult{}, err
	}
	tracker.step("wait_fork_ready", "隔离 fork 节点已从 source 初始状态克隆完成")

	txid, err := s.SubmitNodeTransaction(sourceNode, minerAddress, receiverAddress, 20, 1)
	if err != nil {
		tracker.fail("submit_transaction", err)
		return NetworkPartitionDemoResult{}, err
	}
	if _, err := s.MineNodePending(sourceNode); err != nil {
		tracker.fail("mine_confirmed_block", err)
		return NetworkPartitionDemoResult{}, err
	}
	sourceConfirmed, err := s.waitForNodeStatus(sourceNode, demoWaitTimeout, func(node NodeStatus) bool {
		return node.Initialized && node.Height >= 1
	})
	if err != nil {
		tracker.fail("wait_source_confirm", err)
		return NetworkPartitionDemoResult{}, err
	}
	if _, err := s.waitForNodeStatus(peerNode, demoWaitTimeout, func(node NodeStatus) bool {
		return node.Initialized && node.Height >= 1
	}); err != nil {
		tracker.fail("wait_peer_confirm", err)
		return NetworkPartitionDemoResult{}, err
	}
	tracker.step("wait_peer_confirm", "source 与 peer 已完成旧主链确认")

	forkSession, err := s.nodeSessionByAddress(forkNode)
	if err != nil {
		tracker.fail("load_fork_session", err)
		return NetworkPartitionDemoResult{}, err
	}
	if _, err := demo.ForceReorgToLongerGenesisFork(forkSession.node.DataDir, minerAddress, "", receiverAddress, 2); err != nil {
		tracker.fail("grow_fork_chain", err)
		return NetworkPartitionDemoResult{}, err
	}
	forkStatus, err := s.waitForNodeStatus(forkNode, demoWaitTimeout, func(node NodeStatus) bool {
		return node.Initialized && node.Height >= 2 && node.TipHash != ""
	})
	if err != nil {
		tracker.fail("wait_fork_growth", err)
		return NetworkPartitionDemoResult{}, err
	}
	tracker.step("wait_fork_growth", "隔离 fork 节点已长出更长分叉链")

	if err := s.ConnectNode(sourceNode, forkNode); err != nil {
		tracker.fail("merge_partition", err)
		return NetworkPartitionDemoResult{}, err
	}
	sourceFinal, err := s.waitForNodeStatus(sourceNode, demoWaitTimeout, func(node NodeStatus) bool {
		return node.LastReorg != nil && node.Height >= forkStatus.Height && node.TipHash == forkStatus.TipHash
	})
	if err != nil {
		tracker.fail("wait_source_merge", err)
		return NetworkPartitionDemoResult{}, err
	}
	peerFinal, err := s.waitForNodeStatus(peerNode, demoWaitTimeout, func(node NodeStatus) bool {
		return node.LastReorg != nil && node.Height >= forkStatus.Height && node.TipHash == forkStatus.TipHash
	})
	if err != nil {
		tracker.fail("wait_peer_merge", err)
		return NetworkPartitionDemoResult{}, err
	}
	tracker.step("wait_peer_merge", "三节点已完成合流并切换到新 tip")

	sourceSession, err := s.nodeSessionByAddress(sourceNode)
	if err != nil {
		tracker.fail("reload_source_session", err)
		return NetworkPartitionDemoResult{}, err
	}
	restored, err := pendingTransactionExists(sourceSession.node.DataDir, txid)
	if err != nil {
		tracker.fail("check_restored_transaction", err)
		return NetworkPartitionDemoResult{}, err
	}

	result := NetworkPartitionDemoResult{
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
	}
	tracker.complete(
		"completed",
		"分区 / 合流流程完成",
		fmt.Sprintf("source=%s peer=%s fork=%s finalTip=%s", sourceNode, peerNode, forkNode, forkStatus.TipHash),
	)
	return result, nil
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
