package network

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"sort"
	"sync"
	"time"

	"go-blockchain/internal/blockchain"
	"go-blockchain/internal/wallet"
)

// Node is a lightweight local blockchain network simulator.
type Node struct {
	Address      string
	DataDir      string
	MinerAddress string

	mu             sync.RWMutex
	chainMu        sync.Mutex
	peers          map[string]struct{}
	orphanBlocks   map[string]blockchain.Block
	orphanChildren map[string][]string
	events         []NodeEvent
}

type NodeEvent struct {
	Timestamp string
	Kind      string
	Detail    string
}

type ChainSnapshot struct {
	Initialized  bool
	Height       int
	TipHash      string
	MempoolCount int
}

const maxNodeEvents = 12

// NewNode creates a network node with one listening address and local chain path.
func NewNode(address, dataDir, minerAddress string) *Node {
	node := &Node{
		Address:        address,
		DataDir:        dataDir,
		MinerAddress:   minerAddress,
		peers:          make(map[string]struct{}),
		orphanBlocks:   make(map[string]blockchain.Block),
		orphanChildren: make(map[string][]string),
		events:         make([]NodeEvent, 0, maxNodeEvents),
	}
	node.addPeer(address)
	return node
}

// Listen starts the TCP server until the context is cancelled.
func (n *Node) Listen(ctx context.Context) error {
	ln, err := net.Listen("tcp", n.Address)
	if err != nil {
		return err
	}
	defer ln.Close()

	n.mu.Lock()
	n.Address = ln.Addr().String()
	n.peers[n.Address] = struct{}{}
	n.mu.Unlock()
	n.recordEvent("listen", fmt.Sprintf("listening at %s", n.Address))

	go func() {
		<-ctx.Done()
		_ = ln.Close()
	}()

	for {
		conn, err := ln.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			var netErr net.Error
			if errors.As(err, &netErr) && netErr.Timeout() {
				continue
			}
			if stringsContains(err.Error(), "closed") {
				return nil
			}
			return err
		}

		go func() {
			defer conn.Close()
			_ = n.handleConn(conn)
		}()
	}
}

// Connect sends a version handshake to one peer.
func (n *Node) Connect(peer string) error {
	n.addPeer(peer)
	n.chainMu.Lock()
	height, err := blockchain.BestHeight(n.DataDir)
	n.chainMu.Unlock()
	if err != nil {
		return err
	}
	n.recordEvent("connect", fmt.Sprintf("send version to %s height=%d", peer, height))
	return n.send(peer, "version", versionMessage{
		From:       n.Address,
		BestHeight: height,
	})
}

// KnownPeers returns the known peer list in sorted order.
func (n *Node) KnownPeers() []string {
	n.mu.RLock()
	defer n.mu.RUnlock()

	peers := make([]string, 0, len(n.peers))
	for peer := range n.peers {
		peers = append(peers, peer)
	}
	sort.Strings(peers)
	return peers
}

// OrphanCount returns the number of buffered orphan blocks.
func (n *Node) OrphanCount() int {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return len(n.orphanBlocks)
}

// RecentEvents returns a snapshot of recent node events, newest first.
func (n *Node) RecentEvents() []NodeEvent {
	n.mu.RLock()
	defer n.mu.RUnlock()
	events := make([]NodeEvent, len(n.events))
	copy(events, n.events)
	return events
}

// EnsureBlockchain initializes one local chain when the node has none yet.
func (n *Node) EnsureBlockchain(genesisAddress string) error {
	n.chainMu.Lock()
	defer n.chainMu.Unlock()

	exists, err := blockchain.ChainExists(n.DataDir)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	bc, err := blockchain.CreateBlockchain(n.DataDir, genesisAddress)
	if err != nil {
		return err
	}
	defer bc.Close()

	n.recordEvent("chain_init", fmt.Sprintf("initialized local chain reward=%s", genesisAddress))
	return nil
}

// ChainSnapshot returns safe chain status information for one node.
func (n *Node) ChainSnapshot() (ChainSnapshot, error) {
	n.chainMu.Lock()
	defer n.chainMu.Unlock()

	snapshot := ChainSnapshot{Height: -1}
	exists, err := blockchain.ChainExists(n.DataDir)
	if err != nil {
		return snapshot, err
	}
	if !exists {
		return snapshot, nil
	}

	bc, err := blockchain.OpenBlockchain(n.DataDir)
	if err != nil {
		return snapshot, err
	}
	defer bc.Close()

	height, err := bc.Height()
	if err != nil {
		return snapshot, err
	}
	current, err := bc.CurrentBlock()
	if err != nil {
		return snapshot, err
	}
	mempoolCount, err := bc.MempoolSize()
	if err != nil {
		return snapshot, err
	}

	snapshot.Initialized = true
	snapshot.Height = height
	snapshot.TipHash = current.HashHex()
	snapshot.MempoolCount = mempoolCount
	return snapshot, nil
}

// LastReorgStatus returns the last recorded reorg status for one node chain.
func (n *Node) LastReorgStatus() (*blockchain.ReorgStatus, error) {
	n.chainMu.Lock()
	defer n.chainMu.Unlock()

	exists, err := blockchain.ChainExists(n.DataDir)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, nil
	}

	bc, err := blockchain.OpenBlockchain(n.DataDir)
	if err != nil {
		return nil, err
	}
	defer bc.Close()

	return bc.LastReorgStatus()
}

// SubmitTransaction creates one local transaction and broadcasts it.
func (n *Node) SubmitTransaction(from *wallet.Wallet, to string, amount int, fee int) (blockchain.Transaction, error) {
	n.chainMu.Lock()
	defer n.chainMu.Unlock()

	bc, err := blockchain.OpenBlockchain(n.DataDir)
	if err != nil {
		return blockchain.Transaction{}, err
	}
	defer bc.Close()

	tx, err := bc.SendTransaction(from, to, amount, fee)
	if err != nil {
		return blockchain.Transaction{}, err
	}
	n.recordEvent("tx_submit", fmt.Sprintf("submitted tx %s", tx.IDHex()))

	n.broadcast("tx", txMessage{
		From: n.Address,
		Tx:   tx,
	}, n.Address)

	return tx, nil
}

// MinePending mines all pending transactions and broadcasts the resulting block.
func (n *Node) MinePending() (*blockchain.Block, error) {
	if n.MinerAddress == "" {
		return nil, fmt.Errorf("miner address not configured")
	}

	n.chainMu.Lock()
	defer n.chainMu.Unlock()

	bc, err := blockchain.OpenBlockchain(n.DataDir)
	if err != nil {
		return nil, err
	}
	defer bc.Close()

	block, _, err := bc.MineMempool(n.MinerAddress)
	if err != nil {
		return nil, err
	}
	n.recordEvent("mine", fmt.Sprintf("mined block height=%d hash=%s", block.Height, block.HashHex()))

	n.broadcast("block", blockMessage{
		From:  n.Address,
		Block: *block,
	}, n.Address)
	_ = n.announceTipHeightExcept(block.Height, "")

	return block, nil
}

func (n *Node) handleConn(conn net.Conn) error {
	var env envelope
	if err := gob.NewDecoder(conn).Decode(&env); err != nil {
		return err
	}

	switch env.Type {
	case "version":
		var msg versionMessage
		if err := decodePayload(env.Data, &msg); err != nil {
			return err
		}
		return n.handleVersion(msg)
	case "addr":
		var msg addrMessage
		if err := decodePayload(env.Data, &msg); err != nil {
			return err
		}
		return n.handleAddr(msg)
	case "getblocks":
		var msg getBlocksMessage
		if err := decodePayload(env.Data, &msg); err != nil {
			return err
		}
		return n.handleGetBlocks(msg)
	case "blocks":
		var msg blocksMessage
		if err := decodePayload(env.Data, &msg); err != nil {
			return err
		}
		return n.handleBlocks(msg)
	case "tx":
		var msg txMessage
		if err := decodePayload(env.Data, &msg); err != nil {
			return err
		}
		return n.handleTx(msg)
	case "block":
		var msg blockMessage
		if err := decodePayload(env.Data, &msg); err != nil {
			return err
		}
		return n.handleBlock(msg)
	default:
		return fmt.Errorf("unknown message type %q", env.Type)
	}
}

func (n *Node) handleVersion(msg versionMessage) error {
	n.addPeer(msg.From)
	n.recordEvent("version", fmt.Sprintf("received version from %s height=%d", msg.From, msg.BestHeight))
	_ = n.send(msg.From, "addr", addrMessage{From: n.Address, Peers: n.KnownPeers()})

	n.chainMu.Lock()
	localHeight, err := blockchain.BestHeight(n.DataDir)
	n.chainMu.Unlock()
	if err != nil {
		return err
	}

	if localHeight > msg.BestHeight {
		return n.sendBlocks(msg.From, msg.BestHeight)
	}
	if localHeight < msg.BestHeight {
		return n.send(msg.From, "getblocks", getBlocksMessage{From: n.Address, FromHeight: localHeight})
	}

	return nil
}

func (n *Node) handleAddr(msg addrMessage) error {
	n.recordEvent("addr", fmt.Sprintf("received %d peers from %s", len(msg.Peers), msg.From))
	for _, peer := range msg.Peers {
		n.addPeer(peer)
	}
	return nil
}

func (n *Node) handleGetBlocks(msg getBlocksMessage) error {
	return n.sendBlocks(msg.From, msg.FromHeight)
}

func (n *Node) handleBlocks(msg blocksMessage) error {
	n.chainMu.Lock()
	defer n.chainMu.Unlock()

	for _, block := range msg.Blocks {
		if err := n.importOrBufferBlock(msg.From, block); err != nil {
			return err
		}
	}
	return nil
}

func (n *Node) handleTx(msg txMessage) error {
	n.addPeer(msg.From)

	n.chainMu.Lock()
	defer n.chainMu.Unlock()

	bc, err := blockchain.OpenBlockchain(n.DataDir)
	if err != nil {
		if errors.Is(err, blockchain.ErrBlockchainNotInitialized) {
			return nil
		}
		return err
	}
	defer bc.Close()

	if err := bc.AddToMempool(msg.Tx); err != nil {
		return nil
	}
	n.recordEvent("tx_receive", fmt.Sprintf("received tx %s from %s", msg.Tx.IDHex(), msg.From))

	n.broadcast("tx", msg, msg.From)
	return nil
}

func (n *Node) handleBlock(msg blockMessage) error {
	n.addPeer(msg.From)
	n.chainMu.Lock()
	defer n.chainMu.Unlock()

	if err := n.importOrBufferBlock(msg.From, msg.Block); err != nil {
		return err
	}
	n.recordEvent("block_receive", fmt.Sprintf("received block height=%d from %s", msg.Block.Height, msg.From))

	n.broadcast("block", msg, msg.From)
	return nil
}

func (n *Node) sendBlocks(peer string, fromHeight int) error {
	n.chainMu.Lock()
	defer n.chainMu.Unlock()

	bc, err := blockchain.OpenBlockchain(n.DataDir)
	if err != nil {
		if errors.Is(err, blockchain.ErrBlockchainNotInitialized) {
			return nil
		}
		return err
	}
	defer bc.Close()

	blocks, err := bc.BlocksFromHeight(fromHeight)
	if err != nil {
		return err
	}
	if len(blocks) == 0 {
		return nil
	}

	payload := make([]blockchain.Block, len(blocks))
	for i, block := range blocks {
		payload[i] = *block
	}

	return n.send(peer, "blocks", blocksMessage{
		From:   n.Address,
		Blocks: payload,
	})
}

func (n *Node) send(peer, msgType string, payload any) error {
	data, err := encodePayload(payload)
	if err != nil {
		return err
	}

	conn, err := net.DialTimeout("tcp", peer, 2*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()

	return gob.NewEncoder(conn).Encode(envelope{
		Type: msgType,
		Data: data,
	})
}

func (n *Node) broadcast(msgType string, payload any, except string) {
	for _, peer := range n.KnownPeers() {
		if peer == n.Address || peer == except {
			continue
		}
		_ = n.send(peer, msgType, payload)
	}
}

func (n *Node) addPeer(peer string) {
	if peer == "" {
		return
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	_, exists := n.peers[peer]
	n.peers[peer] = struct{}{}
	if !exists {
		n.events = append([]NodeEvent{{
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Kind:      "peer",
			Detail:    fmt.Sprintf("peer added %s", peer),
		}}, n.events...)
		if len(n.events) > maxNodeEvents {
			n.events = n.events[:maxNodeEvents]
		}
	}
}

func (n *Node) importOrBufferBlock(from string, block blockchain.Block) error {
	beforeHeight, _ := blockchain.BestHeight(n.DataDir)
	blockCopy := block
	if err := blockchain.ImportBlockToDir(n.DataDir, &blockCopy); err != nil {
		switch {
		case errors.Is(err, blockchain.ErrOrphanBlock):
			n.bufferOrphan(block)
			n.recordEvent("orphan", fmt.Sprintf("buffered orphan height=%d hash=%s", block.Height, block.HashHex()))
			_ = n.requestMissingParent(from, block.Height)
			return nil
		case errors.Is(err, blockchain.ErrInvalidBlock),
			errors.Is(err, blockchain.ErrBlockchainNotInitialized),
			errors.Is(err, blockchain.ErrDoubleSpend),
			errors.Is(err, blockchain.ErrInvalidTransaction):
			return nil
		default:
			return err
		}
	}

	n.recordEvent("block_import", fmt.Sprintf("imported block height=%d hash=%s", block.Height, block.HashHex()))
	n.processOrphanDescendants(from, block.Hash)
	afterHeight, _ := blockchain.BestHeight(n.DataDir)
	if afterHeight > beforeHeight {
		_ = n.announceTipHeightExcept(afterHeight, from)
	}
	return nil
}

func (n *Node) requestMissingParent(peer string, orphanHeight int) error {
	if peer == "" {
		return nil
	}
	fromHeight := orphanHeight - 2
	if fromHeight < -1 {
		fromHeight = -1
	}
	n.recordEvent("parent_request", fmt.Sprintf("request parent range from %s starting %d", peer, fromHeight))
	return n.send(peer, "getblocks", getBlocksMessage{
		From:       n.Address,
		FromHeight: fromHeight,
	})
}

func (n *Node) processOrphanDescendants(from string, parentHash []byte) {
	queue := [][]byte{append([]byte(nil), parentHash...)}
	for len(queue) > 0 {
		hash := queue[0]
		queue = queue[1:]

		children := n.takeOrphansByParent(hash)
		for _, orphan := range children {
			orphanCopy := orphan
			if err := blockchain.ImportBlockToDir(n.DataDir, &orphanCopy); err != nil {
				if errors.Is(err, blockchain.ErrOrphanBlock) {
					n.bufferOrphan(orphan)
					n.recordEvent("orphan", fmt.Sprintf("orphan still waiting height=%d hash=%s", orphan.Height, orphan.HashHex()))
					_ = n.requestMissingParent(from, orphan.Height)
				}
				continue
			}
			n.recordEvent("orphan_resolved", fmt.Sprintf("resolved orphan height=%d hash=%s", orphan.Height, orphan.HashHex()))
			queue = append(queue, append([]byte(nil), orphan.Hash...))
		}
	}
}

func (n *Node) bufferOrphan(block blockchain.Block) {
	hashHex := block.HashHex()
	prevHex := block.PrevHashHex()

	n.mu.Lock()
	defer n.mu.Unlock()
	if _, exists := n.orphanBlocks[hashHex]; exists {
		return
	}
	n.orphanBlocks[hashHex] = *cloneBlock(&block)
	n.orphanChildren[prevHex] = append(n.orphanChildren[prevHex], hashHex)
}

func (n *Node) takeOrphansByParent(parentHash []byte) []blockchain.Block {
	parentHex := hex.EncodeToString(parentHash)

	n.mu.Lock()
	defer n.mu.Unlock()

	hashes := append([]string(nil), n.orphanChildren[parentHex]...)
	delete(n.orphanChildren, parentHex)

	orphans := make([]blockchain.Block, 0, len(hashes))
	for _, hash := range hashes {
		block, exists := n.orphanBlocks[hash]
		if !exists {
			continue
		}
		delete(n.orphanBlocks, hash)
		orphans = append(orphans, block)
	}
	return orphans
}

func cloneBlock(block *blockchain.Block) *blockchain.Block {
	if block == nil {
		return nil
	}
	clone := *block
	clone.PrevBlockHash = append([]byte(nil), block.PrevBlockHash...)
	clone.Hash = append([]byte(nil), block.Hash...)
	clone.MerkleRoot = append([]byte(nil), block.MerkleRoot...)
	clone.Transactions = make([]blockchain.Transaction, len(block.Transactions))
	for i, tx := range block.Transactions {
		clone.Transactions[i] = tx.Clone()
	}
	return &clone
}

func (n *Node) recordEvent(kind, detail string) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.events = append([]NodeEvent{{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Kind:      kind,
		Detail:    detail,
	}}, n.events...)
	if len(n.events) > maxNodeEvents {
		n.events = n.events[:maxNodeEvents]
	}
}

func (n *Node) announceTipHeightExcept(height int, except string) error {
	announced := 0
	for _, peer := range n.KnownPeers() {
		if peer == n.Address || peer == except {
			continue
		}
		if err := n.send(peer, "version", versionMessage{
			From:       n.Address,
			BestHeight: height,
		}); err == nil {
			announced++
		}
	}
	if announced > 0 {
		n.recordEvent("tip_announce", fmt.Sprintf("announced height=%d to %d peer(s)", height, announced))
	}
	return nil
}

func encodePayload(v any) ([]byte, error) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func decodePayload(data []byte, out any) error {
	return gob.NewDecoder(bytes.NewReader(data)).Decode(out)
}

func stringsContains(s, substr string) bool {
	return bytes.Contains([]byte(s), []byte(substr))
}
