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
const subsidy = 50

// Transaction is the unsigned UTXO-style transaction model used in Plan 5.
type Transaction struct {
	ID      []byte
	Inputs  []TXInput
	Outputs []TXOutput
}

// TXInput references one previous transaction output.
type TXInput struct {
	TxID []byte
	Out  int
	From string
}

// TXOutput locks one value to a destination address.
type TXOutput struct {
	Value int
	To    string
}

// NewCoinbaseTransaction creates the prototype mining reward transaction.
func NewCoinbaseTransaction(to, data string) Transaction {
	if data == "" {
		data = fmt.Sprintf("reward to %s", to)
	}

	tx := Transaction{
		Inputs: []TXInput{
			{
				TxID: nil,
				Out:  -1,
				From: data,
			},
		},
		Outputs: []TXOutput{
			{
				To:    to,
				Value: subsidy,
			},
		},
	}
	tx.ID = tx.Hash()
	return tx
}

// NewUTXOTransaction creates an unsigned transaction from spendable outputs.
func NewUTXOTransaction(from, to string, amount int, spendable map[string][]int, accumulated int) (Transaction, error) {
	if from == "" || to == "" {
		return Transaction{}, fmt.Errorf("from and to must not be empty")
	}
	if amount <= 0 {
		return Transaction{}, fmt.Errorf("amount must be positive")
	}
	if accumulated < amount {
		return Transaction{}, ErrInsufficientFunds
	}

	var inputs []TXInput
	for txIDHex, indexes := range spendable {
		txID, err := hex.DecodeString(txIDHex)
		if err != nil {
			return Transaction{}, fmt.Errorf("decode txid %s: %w", txIDHex, err)
		}

		for _, index := range indexes {
			inputs = append(inputs, TXInput{
				TxID: append([]byte(nil), txID...),
				Out:  index,
				From: from,
			})
		}
	}

	outputs := []TXOutput{
		{
			To:    to,
			Value: amount,
		},
	}
	if accumulated > amount {
		outputs = append(outputs, TXOutput{
			To:    from,
			Value: accumulated - amount,
		})
	}

	tx := Transaction{
		Inputs:  inputs,
		Outputs: outputs,
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
			TxID: append([]byte(nil), input.TxID...),
			Out:  input.Out,
			From: input.From,
		}
	}
	copy(cloned.Outputs, tx.Outputs)

	return cloned
}

// String formats the transaction for logging.
func (tx Transaction) String() string {
	return fmt.Sprintf("tx %s (inputs=%d outputs=%d)", tx.IDHex(), len(tx.Inputs), len(tx.Outputs))
}

// AmountString is a small helper for CLI formatting.
func (out TXOutput) AmountString() string {
	return strconv.Itoa(out.Value)
}

// UsesKey reports whether an input spends outputs belonging to the given address.
func (in TXInput) UsesKey(address string) bool {
	return in.From == address
}

// IsLockedWith reports whether the output belongs to the given address.
func (out TXOutput) IsLockedWith(address string) bool {
	return out.To == address
}

// TxIDHex returns the referenced transaction ID for CLI printing.
func (in TXInput) TxIDHex() string {
	return hex.EncodeToString(in.TxID)
}
