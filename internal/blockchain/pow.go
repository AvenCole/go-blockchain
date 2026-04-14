package blockchain

import (
	"bytes"
	"math/big"
)

const defaultDifficulty = 12

// ProofOfWork mines and validates blocks against a target threshold.
type ProofOfWork struct {
	block  *Block
	target *big.Int
}

// NewProofOfWork prepares the target for one block.
func NewProofOfWork(block *Block) ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-block.Difficulty))

	return ProofOfWork{
		block:  block,
		target: target,
	}
}

// Run searches for a nonce that satisfies the configured target.
func (pow ProofOfWork) Run() (int, []byte) {
	var hashInt big.Int

	for nonce := 0; ; nonce++ {
		hash := pow.block.CalculateHashWithNonce(nonce)
		hashInt.SetBytes(hash)

		if hashInt.Cmp(pow.target) < 0 {
			return nonce, hash
		}
	}
}

// Validate checks whether the stored nonce/hash satisfy the target.
func (pow ProofOfWork) Validate() bool {
	var hashInt big.Int

	hash := pow.block.CalculateHashWithNonce(pow.block.Nonce)
	hashInt.SetBytes(hash)

	return hashInt.Cmp(pow.target) < 0 && bytes.Equal(hash, pow.block.Hash)
}
