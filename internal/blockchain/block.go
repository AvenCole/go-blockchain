package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"strconv"
	"time"
)

// Block is the basic unit of the chain used in Plan 3.
type Block struct {
	Timestamp     int64
	Transactions  []Transaction
	MerkleRoot    []byte
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
	Difficulty    int
	Height        int
}

// NewBlock creates a new block from transactions.
func NewBlock(transactions []Transaction, prevBlockHash []byte, height int) *Block {
	block := &Block{
		Timestamp:     time.Now().Unix(),
		Transactions:  cloneTransactions(transactions),
		PrevBlockHash: append([]byte(nil), prevBlockHash...),
		Difficulty:    defaultDifficulty,
		Height:        height,
	}
	block.MerkleRoot = block.CalculateMerkleRoot()
	proof := NewProofOfWork(block)
	nonce, hash := proof.Run()
	block.Nonce = nonce
	block.Hash = hash

	return block
}

// NewGenesisBlock creates the first block in the chain.
func NewGenesisBlock(coinbase Transaction) *Block {
	return NewBlock([]Transaction{coinbase}, nil, 0)
}

// CalculateHash derives the block hash from the current header fields.
func (b Block) CalculateHash() []byte {
	return b.CalculateHashWithNonce(b.Nonce)
}

// CalculateHashWithNonce derives the block hash from the current header fields and one nonce.
func (b Block) CalculateHashWithNonce(nonce int) []byte {
	headers := bytes.Join(
		[][]byte{
			[]byte(strconv.FormatInt(b.Timestamp, 10)),
			b.MerkleRoot,
			b.PrevBlockHash,
			[]byte(strconv.Itoa(b.Height)),
			[]byte(strconv.Itoa(b.Difficulty)),
			[]byte(strconv.Itoa(nonce)),
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

// MerkleRootHex returns the Merkle root in hex form for CLI display.
func (b Block) MerkleRootHex() string {
	return hex.EncodeToString(b.MerkleRoot)
}

// CalculateMerkleRoot derives the Merkle root from all transaction IDs.
func (b Block) CalculateMerkleRoot() []byte {
	if len(b.Transactions) == 0 {
		return nil
	}

	var ids [][]byte
	for _, tx := range b.Transactions {
		ids = append(ids, tx.Hash())
	}

	return NewMerkleTree(ids).Root()
}

// VerifyMerkleRoot recomputes the Merkle root to validate transaction integrity.
func (b Block) VerifyMerkleRoot() bool {
	return bytes.Equal(b.MerkleRoot, b.CalculateMerkleRoot())
}

// VerifyProofOfWork validates that the stored hash/nonce satisfy the current difficulty.
func (b Block) VerifyProofOfWork() bool {
	proof := NewProofOfWork(&b)
	return proof.Validate()
}

func cloneTransactions(transactions []Transaction) []Transaction {
	cloned := make([]Transaction, len(transactions))
	for i, tx := range transactions {
		cloned[i] = tx.Clone()
	}

	return cloned
}
