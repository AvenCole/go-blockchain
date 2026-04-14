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
}

// NewNode creates a network node with one listening address and local chain path.
func NewNode(address, dataDir, minerAddress string) *Node {
	node := &Node{
		Address:        address,
		DataDir:        dataDir,
		MinerAddress:   minerAddress,
		peers:          make(map[string]struct{}),
		orphanBlocks:   make(map[string]blockchain.Block),
		orphanChildren: make(map[string][]string),
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

	n.broadcast("block", blockMessage{
		From:  n.Address,
		Block: *block,
	}, n.Address)

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
	n.peers[peer] = struct{}{}
}

func (n *Node) importOrBufferBlock(from string, block blockchain.Block) error {
	blockCopy := block
	if err := blockchain.ImportBlockToDir(n.DataDir, &blockCopy); err != nil {
		switch {
		case errors.Is(err, blockchain.ErrOrphanBlock):
			n.bufferOrphan(block)
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

	n.processOrphanDescendants(from, block.Hash)
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
					_ = n.requestMissingParent(from, orphan.Height)
				}
				continue
			}
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
