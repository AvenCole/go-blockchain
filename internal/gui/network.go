package gui

import (
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"go-blockchain/internal/blockchain"
	"go-blockchain/internal/network"
	"go-blockchain/internal/wallet"
)

type nodeSession struct {
	node   *network.Node
	cancel func()
}

var sanitizeNodeReplacer = strings.NewReplacer(":", "_", "/", "_", "\\", "_")

func (s *Service) ensureNodeMap() {
	if s.nodes == nil {
		s.nodes = make(map[string]*nodeSession)
	}
}

func (s *Service) StartNode(address, seed, miner string) (string, error) {
	if address == "" {
		return "", fmt.Errorf("节点地址不能为空")
	}
	if miner != "" && !wallet.ValidateAddress(miner) {
		return "", fmt.Errorf("无效矿工地址")
	}

	nodeKey := address
	if strings.HasSuffix(address, ":0") {
		nodeKey = address + "-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	}
	nodeDir := filepath.Join(s.cfg.DataDir, "nodes", sanitizeNodeReplacer.Replace(nodeKey))
	return s.startManagedNode(address, seed, miner, nodeDir)
}

func (s *Service) startManagedNode(address, seed, miner, nodeDir string) (string, error) {
	s.nodeMu.Lock()
	defer s.nodeMu.Unlock()
	s.ensureNodeMap()
	if _, exists := s.nodes[address]; exists {
		return address, fmt.Errorf("节点已存在")
	}

	node := network.NewNode(address, nodeDir, miner)
	ctx, cancel := contextFromBackground()
	errCh := make(chan error, 1)

	go func() {
		if err := node.Listen(ctx); err != nil {
			errCh <- err
		}
	}()
	shortWait()
	select {
	case err := <-errCh:
		cancel()
		return "", err
	default:
	}

	actualAddress := node.Address
	if actualAddress == "" {
		cancel()
		return "", fmt.Errorf("节点启动失败：监听地址为空")
	}
	if seed != "" {
		if err := node.Connect(seed); err != nil {
			cancel()
			return "", err
		}
	}

	s.nodes[actualAddress] = &nodeSession{
		node:   node,
		cancel: cancel,
	}
	return actualAddress, nil
}

func (s *Service) StopNode(address string) error {
	s.nodeMu.Lock()
	defer s.nodeMu.Unlock()
	session, ok := s.nodes[address]
	if !ok {
		return fmt.Errorf("节点不存在")
	}
	session.cancel()
	delete(s.nodes, address)
	return nil
}

func (s *Service) Nodes() ([]NodeStatus, error) {
	s.nodeMu.Lock()
	defer s.nodeMu.Unlock()
	s.ensureNodeMap()

	statuses := make([]NodeStatus, 0, len(s.nodes))
	for _, session := range s.nodes {
		snapshot, _ := session.node.ChainSnapshot()
		lastReorgStatus, _ := session.node.LastReorgStatus()
		events := session.node.RecentEvents()
		eventViews := make([]NodeEventView, 0, len(events))
		for _, event := range events {
			eventViews = append(eventViews, NodeEventView{
				Timestamp: event.Timestamp,
				Kind:      event.Kind,
				Detail:    event.Detail,
			})
		}
		statuses = append(statuses, NodeStatus{
			Address:      session.node.Address,
			MinerAddress: session.node.MinerAddress,
			Peers:        session.node.KnownPeers(),
			Initialized:  snapshot.Initialized,
			Height:       snapshot.Height,
			TipHash:      snapshot.TipHash,
			MempoolCount: snapshot.MempoolCount,
			Running:      true,
			OrphanCount:  session.node.OrphanCount(),
			LastReorg:    toReorgStatusView(lastReorgStatus),
			RecentEvents: eventViews,
		})
	}
	sort.Slice(statuses, func(i, j int) bool {
		return statuses[i].Address < statuses[j].Address
	})
	return statuses, nil
}

func toReorgStatusView(status *blockchain.ReorgStatus) *ReorgStatusView {
	if status == nil {
		return nil
	}
	return &ReorgStatusView{
		Timestamp:             status.Timestamp,
		OldHeight:             status.OldHeight,
		NewHeight:             status.NewHeight,
		OldTip:                status.OldTip,
		NewTip:                status.NewTip,
		RestoredTxCount:       status.RestoredTxCount,
		DroppedConfirmedCount: status.DroppedConfirmedCount,
	}
}

func (s *Service) ConnectNode(address, seed string) error {
	s.nodeMu.Lock()
	defer s.nodeMu.Unlock()
	s.ensureNodeMap()

	session, ok := s.nodes[address]
	if !ok {
		return fmt.Errorf("节点不存在")
	}
	if seed == "" {
		return fmt.Errorf("seed 不能为空")
	}
	return session.node.Connect(seed)
}
