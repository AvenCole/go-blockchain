package wallet

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"math/big"
)

const version = byte(0x00)
const checksumLength = 4

// Wallet holds the private/public key pair used to derive an address.
type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

// New creates a fresh wallet using the P-256 curve.
func New() (*Wallet, error) {
	curve := elliptic.P256()
	privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generate key pair: %w", err)
	}

	publicKey := elliptic.Marshal(curve, privateKey.PublicKey.X, privateKey.PublicKey.Y)

	return &Wallet{
		PrivateKey: *privateKey,
		PublicKey:  publicKey,
	}, nil
}

// Address returns the wallet address encoded in Base58.
func (w Wallet) Address() string {
	pubKeyHash := HashPublicKey(w.PublicKey)
	versionedPayload := append([]byte{version}, pubKeyHash...)
	fullPayload := append(versionedPayload, checksum(versionedPayload)...)

	return Base58Encode(fullPayload)
}

// HashPublicKey derives a shortened public key hash for address generation.
func HashPublicKey(publicKey []byte) []byte {
	sum := sha256.Sum256(publicKey)
	trimmed := sum[:20]
	return append([]byte(nil), trimmed...)
}

// ValidateAddress checks the address version and checksum.
func ValidateAddress(address string) bool {
	decoded := Base58Decode(address)
	if len(decoded) < 1+checksumLength {
		return false
	}

	payload := decoded[:len(decoded)-checksumLength]
	actualChecksum := decoded[len(decoded)-checksumLength:]
	expectedChecksum := checksum(payload)

	if payload[0] != version {
		return false
	}

	return bytes.Equal(actualChecksum, expectedChecksum)
}

// PrivateKeyBytes exposes the scalar for persistence.
func (w Wallet) PrivateKeyBytes() []byte {
	return w.PrivateKey.D.Bytes()
}

// PublicKeyCoordinates exposes the X/Y coordinates for persistence.
func (w Wallet) PublicKeyCoordinates() ([]byte, []byte) {
	return w.PrivateKey.PublicKey.X.Bytes(), w.PrivateKey.PublicKey.Y.Bytes()
}

// FromRecord reconstructs a wallet from persisted key material.
func FromRecord(dBytes, xBytes, yBytes []byte) (*Wallet, error) {
	curve := elliptic.P256()
	privateKey := ecdsa.PrivateKey{}
	privateKey.PublicKey.Curve = curve
	privateKey.D = new(big.Int).SetBytes(dBytes)
	privateKey.PublicKey.X = new(big.Int).SetBytes(xBytes)
	privateKey.PublicKey.Y = new(big.Int).SetBytes(yBytes)

	if !curve.IsOnCurve(privateKey.PublicKey.X, privateKey.PublicKey.Y) {
		return nil, fmt.Errorf("wallet key material is not on curve")
	}

	publicKey := elliptic.Marshal(curve, privateKey.PublicKey.X, privateKey.PublicKey.Y)

	return &Wallet{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
	}, nil
}

func checksum(payload []byte) []byte {
	first := sha256.Sum256(payload)
	second := sha256.Sum256(first[:])

	return append([]byte(nil), second[:checksumLength]...)
}
