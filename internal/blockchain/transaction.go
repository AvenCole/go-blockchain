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
	"sort"
	"strconv"

	"go-blockchain/internal/wallet"
)

const coinbaseInputSource = "COINBASE"
const subsidy = 50
const txVersionScriptVM = 2

// Transaction is the signed UTXO-style transaction model used in Plan 6.
type Transaction struct {
	Version int
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
	ScriptSig Script
}

// TXOutput locks one value to the recipient public key hash.
type TXOutput struct {
	Value        int
	PubKeyHash   []byte
	ScriptPubKey Script
}

// CachedUTXO keeps the original transaction output index alongside one output.
type CachedUTXO struct {
	Index  int
	Output TXOutput
}

// NewCoinbaseTransaction creates the prototype mining reward transaction.
func NewCoinbaseTransaction(to, data string) Transaction {
	return NewCoinbaseTransactionWithReward(to, data, subsidy)
}

// NewCoinbaseTransactionWithReward creates a mining reward transaction with a custom amount.
func NewCoinbaseTransactionWithReward(to, data string, reward int) Transaction {
	if data == "" {
		data = fmt.Sprintf("reward to %s", to)
	}

	output, err := NewTXOutput(reward, to)
	if err != nil {
		panic(err)
	}

	tx := Transaction{
		Version: txVersionScriptVM,
		Inputs: []TXInput{
			{
				TxID:      nil,
				Out:       -1,
				Signature: nil,
				PubKey:    []byte(data),
				ScriptSig: NewCoinbaseScript(data),
			},
		},
		Outputs: []TXOutput{output},
	}
	tx.ID = tx.Hash()
	return tx
}

// NewUTXOTransaction creates a signed UTXO transaction from spendable outputs.
func NewUTXOTransaction(fromWallet *wallet.Wallet, to string, amount int, fee int, bc *Blockchain) (Transaction, error) {
	mainOutput, err := NewTXOutput(amount, to)
	if err != nil {
		return Transaction{}, err
	}
	return buildSpendTransaction(fromWallet, mainOutput, amount, fee, bc, txVersionScriptVM)
}

// NewP2PKTransaction creates a script-VM transaction whose main output is a P2PK script.
func NewP2PKTransaction(fromWallet *wallet.Wallet, toWallet *wallet.Wallet, amount int, fee int, bc *Blockchain) (Transaction, error) {
	if fromWallet == nil {
		return Transaction{}, fmt.Errorf("from wallet must not be nil")
	}
	if toWallet == nil {
		return Transaction{}, fmt.Errorf("to wallet must not be nil")
	}
	mainOutput, err := NewP2PKOutput(amount, toWallet.PublicKey)
	if err != nil {
		return Transaction{}, err
	}
	return buildSpendTransaction(fromWallet, mainOutput, amount, fee, bc, txVersionScriptVM)
}

func buildSpendTransaction(fromWallet *wallet.Wallet, mainOutput TXOutput, amount int, fee int, bc *Blockchain, version int) (Transaction, error) {
	if fromWallet == nil {
		return Transaction{}, fmt.Errorf("from wallet must not be nil")
	}
	if amount <= 0 {
		return Transaction{}, fmt.Errorf("amount must be positive")
	}
	if fee < 0 {
		return Transaction{}, fmt.Errorf("fee must not be negative")
	}

	fromAddress := fromWallet.Address()
	fromPubKeyHash := wallet.HashPublicKey(fromWallet.PublicKey)
	required := amount + fee
	accumulated, spendable, err := bc.FindSpendableOutputs(fromPubKeyHash, required)
	if err != nil {
		return Transaction{}, err
	}
	if accumulated < required {
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

	outputs := []TXOutput{mainOutput}
	if accumulated > required {
		changeOutput, err := NewTXOutput(accumulated-required, fromAddress)
		if err != nil {
			return Transaction{}, err
		}
		outputs = append(outputs, changeOutput)
	}

	tx := Transaction{
		Version: version,
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
		Value:        value,
		PubKeyHash:   pubKeyHash,
		ScriptPubKey: NewP2PKHLockingScript(pubKeyHash),
	}, nil
}

// NewP2PKOutput creates one output locked directly to a public key.
func NewP2PKOutput(value int, pubKey []byte) (TXOutput, error) {
	if value <= 0 {
		return TXOutput{}, fmt.Errorf("value must be positive")
	}
	if len(pubKey) == 0 {
		return TXOutput{}, fmt.Errorf("pubkey must not be empty")
	}

	return TXOutput{
		Value:        value,
		PubKeyHash:   wallet.HashPublicKey(pubKey),
		ScriptPubKey: NewP2PKLockingScript(pubKey),
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

	for inID := range tx.Inputs {
		digest := tx.signatureHash(inID, prevTXs)
		if len(digest) == 0 {
			return fmt.Errorf("sign input %d: empty digest", inID)
		}

		r, s, err := ecdsa.Sign(rand.Reader, &privKey, digest)
		if err != nil {
			return fmt.Errorf("sign input %d: %w", inID, err)
		}

		signature := append(r.Bytes(), s.Bytes()...)
		tx.Inputs[inID].Signature = signature
		if tx.UsesScriptVM() {
			pubKey := append([]byte(nil), tx.Inputs[inID].PubKey...)
			if len(pubKey) == 0 {
				pubKey = serializePublicKey(&privKey.PublicKey)
				tx.Inputs[inID].PubKey = append([]byte(nil), pubKey...)
			}
			prevTx := prevTXs[tx.Inputs[inID].TxIDHex()]
			prevOutput := prevTx.Outputs[tx.Inputs[inID].Out]
			if _, ok := ExtractP2PKPubKey(prevOutput.EffectiveScriptPubKey()); ok {
				tx.Inputs[inID].ScriptSig = NewP2PKUnlockingScript(signature)
			} else {
				tx.Inputs[inID].ScriptSig = NewP2PKHUnlockingScript(signature, pubKey)
			}
		}
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

	for inID, input := range tx.Inputs {
		if !tx.verifyInput(inID, input, prevTXs) {
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
		Version: tx.Version,
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
			ScriptSig: input.ScriptSig.Clone(),
		}
	}
	for i, output := range tx.Outputs {
		cloned.Outputs[i] = TXOutput{
			Value:        output.Value,
			PubKeyHash:   append([]byte(nil), output.PubKeyHash...),
			ScriptPubKey: output.ScriptPubKey.Clone(),
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
			ScriptSig: Script{},
		})
	}
	for _, output := range tx.Outputs {
		outputs = append(outputs, TXOutput{
			Value:        output.Value,
			PubKeyHash:   append([]byte(nil), output.PubKeyHash...),
			ScriptPubKey: output.ScriptPubKey.Clone(),
		})
	}

	return Transaction{
		Version: tx.Version,
		ID:      append([]byte(nil), tx.ID...),
		Inputs:  inputs,
		Outputs: outputs,
	}
}

// String formats the transaction for logging.
func (tx Transaction) String() string {
	return fmt.Sprintf("tx %s (inputs=%d outputs=%d)", tx.IDHex(), len(tx.Inputs), len(tx.Outputs))
}

// Fee computes the fee using referenced previous outputs.
func (tx Transaction) Fee(prevTXs map[string]Transaction) int {
	if tx.IsCoinbase() {
		return 0
	}

	inputTotal := 0
	for _, input := range tx.Inputs {
		prevTx := prevTXs[input.TxIDHex()]
		if input.Out >= 0 && input.Out < len(prevTx.Outputs) {
			inputTotal += prevTx.Outputs[input.Out].Value
		}
	}

	outputTotal := 0
	for _, output := range tx.Outputs {
		outputTotal += output.Value
	}

	return inputTotal - outputTotal
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
	return bytes.Equal(out.effectivePubKeyHash(), pubKeyHash)
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
	return wallet.AddressFromPubKeyHash(out.effectivePubKeyHash())
}

func (tx Transaction) UsesScriptVM() bool {
	return tx.Version >= txVersionScriptVM
}

func (tx Transaction) signatureHash(inputIndex int, prevTXs map[string]Transaction) []byte {
	if !tx.UsesScriptVM() {
		return tx.legacySignatureHash(inputIndex, prevTXs)
	}

	prevTx := prevTXs[tx.Inputs[inputIndex].TxIDHex()]
	prevOutput := prevTx.Outputs[tx.Inputs[inputIndex].Out]

	txCopy := tx.TrimmedCopy()
	txCopy.Inputs[inputIndex].ScriptSig = prevOutput.EffectiveScriptPubKey()
	return txCopy.Hash()
}

func (tx Transaction) legacySignatureHash(inputIndex int, prevTXs map[string]Transaction) []byte {
	prevTx := prevTXs[tx.Inputs[inputIndex].TxIDHex()]
	txCopy := tx.TrimmedCopy()
	txCopy.Inputs[inputIndex].PubKey = append([]byte(nil), prevTx.Outputs[tx.Inputs[inputIndex].Out].effectivePubKeyHash()...)
	txCopy.ID = txCopy.Hash()
	txCopy.Inputs[inputIndex].PubKey = nil
	return append([]byte(nil), txCopy.ID...)
}

func (tx Transaction) verifyInput(inID int, input TXInput, prevTXs map[string]Transaction) bool {
	prevTx := prevTXs[input.TxIDHex()]
	if tx.UsesScriptVM() {
		prevOutput := prevTx.Outputs[input.Out]
		unlocking := input.EffectiveScriptSig()
		locking := prevOutput.EffectiveScriptPubKey()

		if pubKey, ok := ExtractP2PKPubKey(locking); ok {
			signature, ok := ExtractP2PKSignature(unlocking)
			if !ok {
				return false
			}
			if len(input.Signature) > 0 && !bytes.Equal(input.Signature, signature) {
				return false
			}
			if len(input.PubKey) > 0 && !bytes.Equal(input.PubKey, pubKey) {
				return false
			}
			return VerifyScripts(unlocking, locking, tx.signatureHash(inID, prevTXs))
		}

		signature, pubKey, ok := ExtractP2PKHUnlockingData(unlocking)
		if !ok {
			return false
		}
		if len(input.Signature) > 0 && !bytes.Equal(input.Signature, signature) {
			return false
		}
		if len(input.PubKey) > 0 && !bytes.Equal(input.PubKey, pubKey) {
			return false
		}
		return VerifyScripts(unlocking, locking, tx.signatureHash(inID, prevTXs))
	}

	if !bytes.Equal(wallet.HashPublicKey(input.PubKey), prevTx.Outputs[input.Out].effectivePubKeyHash()) {
		return false
	}

	return verifyECDSASignature(input.PubKey, input.Signature, tx.signatureHash(inID, prevTXs))
}

func (in TXInput) EffectiveScriptSig() Script {
	if !in.ScriptSig.IsEmpty() {
		return in.ScriptSig.Clone()
	}
	if len(in.Signature) == 0 && len(in.PubKey) == 0 {
		return Script{}
	}
	if len(in.Signature) > 0 && len(in.PubKey) == 0 {
		return NewP2PKUnlockingScript(in.Signature)
	}
	return NewP2PKHUnlockingScript(in.Signature, in.PubKey)
}

func (out TXOutput) EffectiveScriptPubKey() Script {
	if !out.ScriptPubKey.IsEmpty() {
		return out.ScriptPubKey.Clone()
	}
	if len(out.PubKeyHash) == 0 {
		return Script{}
	}
	return NewP2PKHLockingScript(out.PubKeyHash)
}

func (out TXOutput) effectivePubKeyHash() []byte {
	if len(out.PubKeyHash) > 0 {
		return append([]byte(nil), out.PubKeyHash...)
	}
	if pubKey, ok := ExtractP2PKPubKey(out.EffectiveScriptPubKey()); ok {
		return wallet.HashPublicKey(pubKey)
	}
	pubKeyHash, ok := ExtractP2PKHPubKeyHash(out.EffectiveScriptPubKey())
	if !ok {
		return nil
	}
	return pubKeyHash
}

func serializePublicKey(publicKey *ecdsa.PublicKey) []byte {
	if publicKey == nil {
		return nil
	}
	return elliptic.Marshal(publicKey.Curve, publicKey.X, publicKey.Y)
}

func encodeOutputs(outputs []TXOutput) ([]byte, error) {
	var result bytes.Buffer

	if err := gob.NewEncoder(&result).Encode(outputs); err != nil {
		return nil, err
	}

	return result.Bytes(), nil
}

func decodeOutputs(data []byte) ([]TXOutput, error) {
	var outputs []TXOutput

	if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&outputs); err != nil {
		return nil, err
	}

	return outputs, nil
}

func encodeCachedUTXOs(outputs []CachedUTXO) ([]byte, error) {
	var result bytes.Buffer

	if err := gob.NewEncoder(&result).Encode(outputs); err != nil {
		return nil, err
	}

	return result.Bytes(), nil
}

func decodeCachedUTXOs(data []byte) ([]CachedUTXO, error) {
	var outputs []CachedUTXO

	if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&outputs); err != nil {
		return nil, err
	}

	return outputs, nil
}
