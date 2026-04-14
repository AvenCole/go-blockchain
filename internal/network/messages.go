package network

import "go-blockchain/internal/blockchain"

type envelope struct {
	Type string
	Data []byte
}

type versionMessage struct {
	From       string
	BestHeight int
}

type addrMessage struct {
	From  string
	Peers []string
}

type getBlocksMessage struct {
	From       string
	FromHeight int
}

type blocksMessage struct {
	From   string
	Blocks []blockchain.Block
}

type txMessage struct {
	From string
	Tx   blockchain.Transaction
}

type blockMessage struct {
	From  string
	Block blockchain.Block
}
