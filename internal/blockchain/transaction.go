package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"math/big"
	"sort"
	"strconv"

	"go-blockchain/internal/wallet"
)

const coinbaseInputSource = "COINBASE"
const subsidy = 50

// Transaction is the signed UTXO-style transaction model used in Plan 6.
type Transaction struct {
	ID      []byte
	Inputs  []TXInput
	Outputs []TXOutput
}

// TXInput references one previous transaction output and carries signature proof.
type TXInput struct {
	TxID      []byte
	Out       int
	Signature []byte
	PubKey    []byte
}

// TXOutput locks one value to the recipient public key hash.
type TXOutput struct {
	Value      int
	PubKeyHash []byte
}

// NewCoinbaseTransaction creates the prototype mining reward transaction.
func NewCoinbaseTransaction(to, data string) Transaction {
	if data == "" {
		data = fmt.Sprintf("reward to %s", to)
	}

	output, err := NewTXOutput(subsidy, to)
	if err != nil {
		panic(err)
	}

	tx := Transaction{
		Inputs: []TXInput{
			{
				TxID:      nil,
				Out:       -1,
				Signature: nil,
				PubKey:    []byte(data),
			},
		},
		Outputs: []TXOutput{output},
	}
	tx.ID = tx.Hash()
	return tx
}

// NewUTXOTransaction creates a signed UTXO transaction from spendable outputs.
func NewUTXOTransaction(fromWallet *wallet.Wallet, to string, amount int, bc *Blockchain) (Transaction, error) {
	if fromWallet == nil {
		return Transaction{}, fmt.Errorf("from wallet must not be nil")
	}
	if to == "" {
		return Transaction{}, fmt.Errorf("to must not be empty")
	}
	if amount <= 0 {
		return Transaction{}, fmt.Errorf("amount must be positive")
	}

	fromAddress := fromWallet.Address()
	fromPubKeyHash := wallet.HashPublicKey(fromWallet.PublicKey)
	accumulated, spendable, err := bc.FindSpendableOutputs(fromPubKeyHash, amount)
	if err != nil {
		return Transaction{}, err
	}
	if accumulated < amount {
		return Transaction{}, ErrInsufficientFunds
	}

	var inputs []TXInput
	txIDs := make([]string, 0, len(spendable))
	for txIDHex := range spendable {
		txIDs = append(txIDs, txIDHex)
	}
	sort.Strings(txIDs)

	for _, txIDHex := range txIDs {
		txID, err := hex.DecodeString(txIDHex)
		if err != nil {
			return Transaction{}, fmt.Errorf("decode txid %s: %w", txIDHex, err)
		}

		indexes := append([]int(nil), spendable[txIDHex]...)
		sort.Ints(indexes)

		for _, index := range indexes {
			inputs = append(inputs, TXInput{
				TxID:      append([]byte(nil), txID...),
				Out:       index,
				Signature: nil,
				PubKey:    append([]byte(nil), fromWallet.PublicKey...),
			})
		}
	}

	mainOutput, err := NewTXOutput(amount, to)
	if err != nil {
		return Transaction{}, err
	}
	outputs := []TXOutput{mainOutput}
	if accumulated > amount {
		changeOutput, err := NewTXOutput(accumulated-amount, fromAddress)
		if err != nil {
			return Transaction{}, err
		}
		outputs = append(outputs, changeOutput)
	}

	tx := Transaction{
		Inputs:  inputs,
		Outputs: outputs,
	}
	tx.ID = tx.Hash()

	if err := bc.SignTransaction(&tx, fromWallet.PrivateKey); err != nil {
		return Transaction{}, err
	}

	return tx, nil
}

// NewTXOutput creates one output locked to an address.
func NewTXOutput(value int, address string) (TXOutput, error) {
	pubKeyHash, err := wallet.PublicKeyHashFromAddress(address)
	if err != nil {
		return TXOutput{}, err
	}

	return TXOutput{
		Value:      value,
		PubKeyHash: pubKeyHash,
	}, nil
}

// Hash calculates a deterministic transaction hash.
func (tx Transaction) Hash() []byte {
	copyTx := tx.Clone()
	copyTx.ID = nil

	var encoded bytes.Buffer
	_ = gob.NewEncoder(&encoded).Encode(copyTx)
	sum := sha256.Sum256(encoded.Bytes())
	return sum[:]
}

// Serialize converts the transaction into bytes for optional reuse.
func (tx Transaction) Serialize() ([]byte, error) {
	var result bytes.Buffer

	encoder := gob.NewEncoder(&result)
	if err := encoder.Encode(tx); err != nil {
		return nil, err
	}

	return result.Bytes(), nil
}

// DeserializeTransaction decodes a transaction.
func DeserializeTransaction(data []byte) (*Transaction, error) {
	var tx Transaction

	decoder := gob.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&tx); err != nil {
		return nil, err
	}

	return &tx, nil
}

// Sign signs each input against the previous referenced outputs.
func (tx *Transaction) Sign(privKey ecdsa.PrivateKey, prevTXs map[string]Transaction) error {
	if tx.IsCoinbase() {
		return nil
	}

	for _, input := range tx.Inputs {
		prevTx := prevTXs[input.TxIDHex()]
		if len(prevTx.ID) == 0 {
			return fmt.Errorf("previous transaction not found for input %s", input.TxIDHex())
		}
		if input.Out < 0 || input.Out >= len(prevTx.Outputs) {
			return fmt.Errorf("referenced output index out of range for input %s", input.TxIDHex())
		}
	}

	txCopy := tx.TrimmedCopy()
	for inID, input := range txCopy.Inputs {
		prevTx := prevTXs[input.TxIDHex()]
		txCopy.Inputs[inID].PubKey = append([]byte(nil), prevTx.Outputs[input.Out].PubKeyHash...)
		txCopy.ID = txCopy.Hash()
		txCopy.Inputs[inID].PubKey = nil

		r, s, err := ecdsa.Sign(rand.Reader, &privKey, txCopy.ID)
		if err != nil {
			return fmt.Errorf("sign input %d: %w", inID, err)
		}

		signature := append(r.Bytes(), s.Bytes()...)
		tx.Inputs[inID].Signature = signature
	}

	return nil
}

// Verify checks the transaction signature set against previous outputs.
func (tx Transaction) Verify(prevTXs map[string]Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}

	for _, input := range tx.Inputs {
		prevTx := prevTXs[input.TxIDHex()]
		if len(prevTx.ID) == 0 {
			return false
		}
		if input.Out < 0 || input.Out >= len(prevTx.Outputs) {
			return false
		}
	}

	txCopy := tx.TrimmedCopy()
	curve := elliptic.P256()

	for inID, input := range tx.Inputs {
		prevTx := prevTXs[input.TxIDHex()]
		txCopy.Inputs[inID].PubKey = append([]byte(nil), prevTx.Outputs[input.Out].PubKeyHash...)
		txCopy.ID = txCopy.Hash()
		txCopy.Inputs[inID].PubKey = nil

		x, y := elliptic.Unmarshal(curve, input.PubKey)
		if x == nil || y == nil {
			return false
		}

		if !bytes.Equal(wallet.HashPublicKey(input.PubKey), prevTx.Outputs[input.Out].PubKeyHash) {
			return false
		}

		r := big.Int{}
		s := big.Int{}
		sigLen := len(input.Signature)
		if sigLen == 0 || sigLen%2 != 0 {
			return false
		}
		r.SetBytes(input.Signature[:sigLen/2])
		s.SetBytes(input.Signature[sigLen/2:])

		publicKey := ecdsa.PublicKey{
			Curve: curve,
			X:     x,
			Y:     y,
		}
		if !ecdsa.Verify(&publicKey, txCopy.ID, &r, &s) {
			return false
		}
	}

	return true
}

// IsCoinbase reports whether a transaction is the prototype mining reward.
func (tx Transaction) IsCoinbase() bool {
	return len(tx.Inputs) == 1 && len(tx.Inputs[0].TxID) == 0 && tx.Inputs[0].Out == -1
}

// IDHex returns the transaction ID in hex for CLI display.
func (tx Transaction) IDHex() string {
	return hex.EncodeToString(tx.ID)
}

// Clone returns a deep copy of the transaction.
func (tx Transaction) Clone() Transaction {
	cloned := Transaction{
		ID:      append([]byte(nil), tx.ID...),
		Inputs:  make([]TXInput, len(tx.Inputs)),
		Outputs: make([]TXOutput, len(tx.Outputs)),
	}

	for i, input := range tx.Inputs {
		cloned.Inputs[i] = TXInput{
			TxID:      append([]byte(nil), input.TxID...),
			Out:       input.Out,
			Signature: append([]byte(nil), input.Signature...),
			PubKey:    append([]byte(nil), input.PubKey...),
		}
	}
	for i, output := range tx.Outputs {
		cloned.Outputs[i] = TXOutput{
			Value:      output.Value,
			PubKeyHash: append([]byte(nil), output.PubKeyHash...),
		}
	}

	return cloned
}

// TrimmedCopy returns the signing copy without signatures and runtime pubkeys.
func (tx Transaction) TrimmedCopy() Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	for _, input := range tx.Inputs {
		inputs = append(inputs, TXInput{
			TxID:      append([]byte(nil), input.TxID...),
			Out:       input.Out,
			Signature: nil,
			PubKey:    nil,
		})
	}
	for _, output := range tx.Outputs {
		outputs = append(outputs, TXOutput{
			Value:      output.Value,
			PubKeyHash: append([]byte(nil), output.PubKeyHash...),
		})
	}

	return Transaction{
		ID:      append([]byte(nil), tx.ID...),
		Inputs:  inputs,
		Outputs: outputs,
	}
}

// String formats the transaction for logging.
func (tx Transaction) String() string {
	return fmt.Sprintf("tx %s (inputs=%d outputs=%d)", tx.IDHex(), len(tx.Inputs), len(tx.Outputs))
}

// AmountString is a small helper for CLI formatting.
func (out TXOutput) AmountString() string {
	return strconv.Itoa(out.Value)
}

// UsesKey reports whether an input spends outputs belonging to the given key hash.
func (in TXInput) UsesKey(pubKeyHash []byte) bool {
	lockingHash := wallet.HashPublicKey(in.PubKey)
	return bytes.Equal(lockingHash, pubKeyHash)
}

// IsLockedWith reports whether the output belongs to the given public key hash.
func (out TXOutput) IsLockedWith(pubKeyHash []byte) bool {
	return bytes.Equal(out.PubKeyHash, pubKeyHash)
}

// TxIDHex returns the referenced transaction ID for CLI printing.
func (in TXInput) TxIDHex() string {
	return hex.EncodeToString(in.TxID)
}

// FromDisplay returns a human-readable sender hint for CLI output.
func (in TXInput) FromDisplay() string {
	if len(in.TxID) == 0 && in.Out == -1 {
		return string(in.PubKey)
	}
	if len(in.PubKey) == 0 {
		return ""
	}

	return wallet.AddressFromPubKey(in.PubKey)
}

// Address renders one output as a wallet address string.
func (out TXOutput) Address() string {
	return wallet.AddressFromPubKeyHash(out.PubKeyHash)
}
