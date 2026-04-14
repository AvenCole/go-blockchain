package blockchain

import (
	"bytes"
	"testing"
)

func TestMerkleTreeHandlesOddLeafCount(t *testing.T) {
	leafA := []byte("a")
	leafB := []byte("b")
	leafC := []byte("c")

	tree := NewMerkleTree([][]byte{leafA, leafB, leafC})
	if len(tree.Root()) == 0 {
		t.Fatalf("Merkle root should not be empty")
	}

	left := hashNode(hashLeaf(leafA), hashLeaf(leafB))
	right := hashNode(hashLeaf(leafC), hashLeaf(leafC))
	expected := hashNode(left, right)

	if !bytes.Equal(tree.Root(), expected) {
		t.Fatalf("Merkle root mismatch for odd leaf count")
	}
}
