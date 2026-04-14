package blockchain

import "testing"

func TestProofOfWorkValidate(t *testing.T) {
	block := NewBlock(nil, nil, 1)

	if block.Difficulty != defaultDifficulty {
		t.Fatalf("Difficulty = %d, want %d", block.Difficulty, defaultDifficulty)
	}

	if !block.VerifyProofOfWork() {
		t.Fatalf("VerifyProofOfWork() = false, want true")
	}
}

func TestProofOfWorkRejectsTampering(t *testing.T) {
	block := NewBlock(nil, nil, 1)
	if !block.VerifyProofOfWork() {
		t.Fatalf("VerifyProofOfWork() = false, want true")
	}

	block.Nonce++
	if block.VerifyProofOfWork() {
		t.Fatalf("VerifyProofOfWork(tampered nonce) = true, want false")
	}
}
