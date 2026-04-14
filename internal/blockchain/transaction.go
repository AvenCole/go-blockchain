package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"strconv"
)

const coinbaseInputSource = "COINBASE"

// Transaction is the minimal, unsigned transaction model used in Plan 3.
type Transaction struct {
	ID      []byte
	Inputs  []TXInput
	Outputs []TXOutput
}

// TXInput is a simplified input that only records a source address and amount.
type TXInput struct {
	From   string
	Amount int
}

// TXOutput is a simplified output that only records a destination address and amount.
type TXOutput struct {
	To     string
	Amount int
}

// NewCoinbaseTransaction creates the prototype mining reward transaction.
func NewCoinbaseTransaction(to, data string) Transaction {
	if data == "" {
		data = fmt.Sprintf("reward to %s", to)
	}

	tx := Transaction{
		Inputs: []TXInput{
			{
				From:   coinbaseInputSource,
				Amount: 0,
			},
		},
		Outputs: []TXOutput{
			{
				To:     to,
				Amount: 50,
			},
		},
	}
	tx.ID = tx.Hash()
	return tx
}

// NewTransaction creates the minimal unsigned transfer transaction.
func NewTransaction(from, to string, amount int) (Transaction, error) {
	if from == "" || to == "" {
		return Transaction{}, fmt.Errorf("from and to must not be empty")
	}
	if amount <= 0 {
		return Transaction{}, fmt.Errorf("amount must be positive")
	}

	tx := Transaction{
		Inputs: []TXInput{
			{
				From:   from,
				Amount: amount,
			},
		},
		Outputs: []TXOutput{
			{
				To:     to,
				Amount: amount,
			},
		},
	}
	tx.ID = tx.Hash()
	return tx, nil
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

// IsCoinbase reports whether a transaction is the prototype mining reward.
func (tx Transaction) IsCoinbase() bool {
	return len(tx.Inputs) == 1 && tx.Inputs[0].From == coinbaseInputSource
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

	copy(cloned.Inputs, tx.Inputs)
	copy(cloned.Outputs, tx.Outputs)

	return cloned
}

// String formats the transaction for logging.
func (tx Transaction) String() string {
	return fmt.Sprintf("tx %s (inputs=%d outputs=%d)", tx.IDHex(), len(tx.Inputs), len(tx.Outputs))
}

// AmountString is a small helper for CLI formatting.
func (out TXOutput) AmountString() string {
	return strconv.Itoa(out.Amount)
}
