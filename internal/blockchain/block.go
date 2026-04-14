package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"strconv"
	"time"
)

// Block is the basic unit of the chain used in Plan 2.
type Block struct {
	Timestamp     int64
	Data          []byte
	PrevBlockHash []byte
	Hash          []byte
	Height        int
}

// NewBlock creates a new block from raw string data.
func NewBlock(data string, prevBlockHash []byte, height int) *Block {
	block := &Block{
		Timestamp:     time.Now().Unix(),
		Data:          []byte(data),
		PrevBlockHash: append([]byte(nil), prevBlockHash...),
		Height:        height,
	}
	block.Hash = block.CalculateHash()

	return block
}

// NewGenesisBlock creates the first block in the chain.
func NewGenesisBlock(data string) *Block {
	return NewBlock(data, nil, 0)
}

// CalculateHash derives the block hash from the current header fields.
func (b Block) CalculateHash() []byte {
	headers := bytes.Join(
		[][]byte{
			[]byte(strconv.FormatInt(b.Timestamp, 10)),
			b.Data,
			b.PrevBlockHash,
			[]byte(strconv.Itoa(b.Height)),
		},
		[]byte{},
	)
	sum := sha256.Sum256(headers)

	return sum[:]
}

// Serialize encodes the block for storage.
func (b Block) Serialize() ([]byte, error) {
	var result bytes.Buffer

	encoder := gob.NewEncoder(&result)
	if err := encoder.Encode(b); err != nil {
		return nil, err
	}

	return result.Bytes(), nil
}

// DeserializeBlock decodes a stored block.
func DeserializeBlock(data []byte) (*Block, error) {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&block); err != nil {
		return nil, err
	}

	return &block, nil
}

// HashHex returns the block hash in hex form for logging and CLI display.
func (b Block) HashHex() string {
	return hex.EncodeToString(b.Hash)
}

// PrevHashHex returns the previous hash in hex form for CLI display.
func (b Block) PrevHashHex() string {
	return hex.EncodeToString(b.PrevBlockHash)
}
