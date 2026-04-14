package blockchain

import "crypto/sha256"

// MerkleTree stores the current tree root used for block integrity.
type MerkleTree struct {
	root []byte
}

// NewMerkleTree builds a Merkle tree from transaction IDs.
func NewMerkleTree(data [][]byte) MerkleTree {
	if len(data) == 0 {
		return MerkleTree{}
	}

	level := make([][]byte, len(data))
	for i, item := range data {
		level[i] = hashLeaf(item)
	}

	for len(level) > 1 {
		if len(level)%2 != 0 {
			level = append(level, append([]byte(nil), level[len(level)-1]...))
		}

		next := make([][]byte, 0, len(level)/2)
		for i := 0; i < len(level); i += 2 {
			next = append(next, hashNode(level[i], level[i+1]))
		}
		level = next
	}

	return MerkleTree{
		root: level[0],
	}
}

// Root returns a copy of the root hash.
func (mt MerkleTree) Root() []byte {
	return append([]byte(nil), mt.root...)
}

func hashLeaf(data []byte) []byte {
	sum := sha256.Sum256(data)
	return sum[:]
}

func hashNode(left, right []byte) []byte {
	joined := append(append([]byte(nil), left...), right...)
	sum := sha256.Sum256(joined)
	return sum[:]
}
