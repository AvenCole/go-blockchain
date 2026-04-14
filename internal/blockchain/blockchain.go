package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cockroachdb/pebble"

	"go-blockchain/internal/wallet"
)

var (
	// ErrBlockchainAlreadyExists reports that the chain has already been created.
	ErrBlockchainAlreadyExists = errors.New("blockchain already exists")
	// ErrBlockchainNotInitialized reports that the chain has not been created yet.
	ErrBlockchainNotInitialized = errors.New("blockchain not initialized")
	// ErrInsufficientFunds reports that the sender cannot cover the requested amount.
	ErrInsufficientFunds = errors.New("insufficient funds")
	// ErrTransactionNotFound reports that one referenced transaction cannot be found.
	ErrTransactionNotFound = errors.New("transaction not found")
	// ErrInvalidTransaction reports that signature verification failed.
	ErrInvalidTransaction = errors.New("invalid transaction")
)

var lastHashKey = []byte("lh")
var utxoPrefix = []byte("utxo-")

// Blockchain provides chain storage and append operations.
type Blockchain struct {
	tip []byte
	db  *pebble.DB
}

// DBPath resolves the storage location used for the chain database.
func DBPath(dataDir string) string {
	return filepath.Join(dataDir, "chain-pebble")
}

// ChainExists reports whether a chain has already been initialized.
func ChainExists(dataDir string) (bool, error) {
	if _, err := os.Stat(DBPath(dataDir)); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}

		return false, err
	}

	db, err := openDB(dataDir)
	if err != nil {
		return false, err
	}
	defer db.Close()

	return lastHashExists(db)
}

// CreateBlockchain initializes a chain with a genesis block.
func CreateBlockchain(dataDir string, genesisAddress string) (*Blockchain, error) {
	db, err := openDB(dataDir)
	if err != nil {
		return nil, err
	}

	exists, err := lastHashExists(db)
	if err != nil {
		db.Close()
		return nil, err
	}
	if exists {
		db.Close()
		return nil, ErrBlockchainAlreadyExists
	}

	genesis := NewGenesisBlock(NewCoinbaseTransaction(genesisAddress, "genesis block"))
	encodedGenesis, err := genesis.Serialize()
	if err != nil {
		db.Close()
		return nil, err
	}

	batch := db.NewBatch()
	defer batch.Close()

	if err := batch.Set(genesis.Hash, encodedGenesis, nil); err != nil {
		db.Close()
		return nil, err
	}
	if err := batch.Set(lastHashKey, genesis.Hash, nil); err != nil {
		db.Close()
		return nil, err
	}
	if err := applyUTXOChanges(db, batch, genesis); err != nil {
		db.Close()
		return nil, err
	}
	if err := batch.Commit(pebble.Sync); err != nil {
		db.Close()
		return nil, err
	}

	bc := &Blockchain{
		tip: genesis.Hash,
		db:  db,
	}

	return bc, nil
}

// OpenBlockchain loads an existing chain from disk.
func OpenBlockchain(dataDir string) (*Blockchain, error) {
	db, err := openDB(dataDir)
	if err != nil {
		return nil, err
	}

	tip, err := loadLastHash(db)
	if err != nil {
		db.Close()
		if errors.Is(err, pebble.ErrNotFound) {
			return nil, ErrBlockchainNotInitialized
		}

		return nil, err
	}

	bc := &Blockchain{
		tip: tip,
		db:  db,
	}
	ok, err := bc.hasUTXOEntries()
	if err != nil {
		bc.Close()
		return nil, err
	}
	if !ok {
		if err := bc.ReindexUTXO(); err != nil {
			bc.Close()
			return nil, err
		}
	}

	return bc, nil
}

// AddBlock appends a new block to the tip of the chain.
func (bc *Blockchain) AddBlock(transactions []Transaction) (*Block, error) {
	for _, tx := range transactions {
		if !bc.VerifyTransaction(tx) {
			return nil, ErrInvalidTransaction
		}
	}

	lastBlock, err := bc.CurrentBlock()
	if err != nil {
		return nil, err
	}

	newBlock := NewBlock(transactions, bc.tip, lastBlock.Height+1)
	encodedBlock, err := newBlock.Serialize()
	if err != nil {
		return nil, err
	}

	batch := bc.db.NewBatch()
	defer batch.Close()

	if err := batch.Set(newBlock.Hash, encodedBlock, nil); err != nil {
		return nil, err
	}
	if err := batch.Set(lastHashKey, newBlock.Hash, nil); err != nil {
		return nil, err
	}
	if err := applyUTXOChanges(bc.db, batch, newBlock); err != nil {
		return nil, err
	}
	if err := batch.Commit(pebble.Sync); err != nil {
		return nil, err
	}

	bc.tip = newBlock.Hash

	return newBlock, nil
}

// SendTransaction creates a signed UTXO-style transaction and stores it in a new block.
func (bc *Blockchain) SendTransaction(fromWallet *wallet.Wallet, to string, amount int) (*Block, Transaction, error) {
	tx, err := NewUTXOTransaction(fromWallet, to, amount, bc)
	if err != nil {
		return nil, Transaction{}, err
	}

	block, err := bc.AddBlock([]Transaction{tx})
	if err != nil {
		return nil, Transaction{}, err
	}

	return block, tx, nil
}

// CurrentBlock returns the tip block.
func (bc *Blockchain) CurrentBlock() (*Block, error) {
	data, err := loadValue(bc.db, bc.tip)
	if err != nil {
		return nil, err
	}

	return DeserializeBlock(data)
}

// Iterator returns a chain iterator starting at the tip.
func (bc *Blockchain) Iterator() *Iterator {
	return &Iterator{
		currentHash: append([]byte(nil), bc.tip...),
		db:          bc.db,
	}
}

// Blocks returns all blocks ordered from newest to oldest.
func (bc *Blockchain) Blocks() ([]*Block, error) {
	iterator := bc.Iterator()
	var blocks []*Block

	for {
		block, err := iterator.Next()
		if err != nil {
			return nil, err
		}
		if block == nil {
			break
		}

		blocks = append(blocks, block)
	}

	return blocks, nil
}

// FindTransaction returns one transaction by its ID.
func (bc *Blockchain) FindTransaction(id []byte) (Transaction, error) {
	blocks, err := bc.Blocks()
	if err != nil {
		return Transaction{}, err
	}

	for _, block := range blocks {
		for _, tx := range block.Transactions {
			if bytes.Equal(tx.ID, id) {
				return tx.Clone(), nil
			}
		}
	}

	return Transaction{}, ErrTransactionNotFound
}

// SignTransaction signs one transaction against the referenced previous outputs.
func (bc *Blockchain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) error {
	prevTXs := make(map[string]Transaction)
	for _, input := range tx.Inputs {
		prevTx, err := bc.FindTransaction(input.TxID)
		if err != nil {
			return err
		}
		prevTXs[input.TxIDHex()] = prevTx
	}

	return tx.Sign(privKey, prevTXs)
}

// VerifyTransaction checks one transaction signature set.
func (bc *Blockchain) VerifyTransaction(tx Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}

	prevTXs := make(map[string]Transaction)
	for _, input := range tx.Inputs {
		prevTx, err := bc.FindTransaction(input.TxID)
		if err != nil {
			return false
		}
		prevTXs[input.TxIDHex()] = prevTx
	}

	return tx.Verify(prevTXs)
}

// BalanceOf computes the balance by summing unspent outputs.
func (bc *Blockchain) BalanceOf(address string) (int, error) {
	pubKeyHash, err := wallet.PublicKeyHashFromAddress(address)
	if err != nil {
		return 0, err
	}

	utxos, err := bc.FindUTXO(pubKeyHash)
	if err != nil {
		return 0, err
	}

	balance := 0
	for _, output := range utxos {
		balance += output.Value
	}

	return balance, nil
}

// FindSpendableOutputs returns spendable outputs for one address.
func (bc *Blockchain) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int, error) {
	unspent := make(map[string][]int)
	accumulated := 0

	iter, err := bc.db.NewIter(&pebble.IterOptions{
		LowerBound: utxoPrefix,
		UpperBound: prefixUpperBound(utxoPrefix),
	})
	if err != nil {
		return 0, nil, err
	}
	defer iter.Close()

	for iter.First(); iter.Valid(); iter.Next() {
		txIDHex := hex.EncodeToString(iter.Key()[len(utxoPrefix):])
		outputs, err := decodeCachedUTXOs(iter.Value())
		if err != nil {
			return 0, nil, err
		}

		for _, cached := range outputs {
			if !cached.Output.IsLockedWith(pubKeyHash) {
				continue
			}

			accumulated += cached.Output.Value
			unspent[txIDHex] = append(unspent[txIDHex], cached.Index)
			if accumulated >= amount {
				return accumulated, unspent, nil
			}
		}
	}

	if err := iter.Error(); err != nil {
		return 0, nil, err
	}

	return accumulated, unspent, nil
}

// FindUTXO returns all currently unspent outputs for one address.
func (bc *Blockchain) FindUTXO(pubKeyHash []byte) ([]TXOutput, error) {
	iter, err := bc.db.NewIter(&pebble.IterOptions{
		LowerBound: utxoPrefix,
		UpperBound: prefixUpperBound(utxoPrefix),
	})
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	var utxos []TXOutput

	for iter.First(); iter.Valid(); iter.Next() {
		outputs, err := decodeCachedUTXOs(iter.Value())
		if err != nil {
			return nil, err
		}

		for _, cached := range outputs {
			if cached.Output.IsLockedWith(pubKeyHash) {
				utxos = append(utxos, cached.Output)
			}
		}
	}

	if err := iter.Error(); err != nil {
		return nil, err
	}

	return utxos, nil
}

// ReindexUTXO rebuilds the cached UTXO set from the full chain.
func (bc *Blockchain) ReindexUTXO() error {
	iter, err := bc.db.NewIter(&pebble.IterOptions{
		LowerBound: utxoPrefix,
		UpperBound: prefixUpperBound(utxoPrefix),
	})
	if err != nil {
		return err
	}
	defer iter.Close()

	batch := bc.db.NewBatch()
	defer batch.Close()

	for iter.First(); iter.Valid(); iter.Next() {
		if err := batch.Delete(append([]byte(nil), iter.Key()...), nil); err != nil {
			return err
		}
	}
	if err := iter.Error(); err != nil {
		return err
	}

	utxoMap, err := bc.findAllUTXOByScan()
	if err != nil {
		return err
	}

	for txIDHex, outputs := range utxoMap {
		txID, err := hex.DecodeString(txIDHex)
		if err != nil {
			return err
		}
		encoded, err := encodeCachedUTXOs(outputs)
		if err != nil {
			return err
		}
		if err := batch.Set(utxoKey(txID), encoded, nil); err != nil {
			return err
		}
	}

	return batch.Commit(pebble.Sync)
}

func (bc *Blockchain) findAllUTXOByScan() (map[string][]CachedUTXO, error) {
	blocks, err := bc.Blocks()
	if err != nil {
		return nil, err
	}

	spentTXOs := make(map[string]map[int]bool)
	utxoMap := make(map[string][]CachedUTXO)

	for _, block := range blocks {
		for _, tx := range block.Transactions {
			txID := tx.IDHex()

			for outIdx, output := range tx.Outputs {
				if spentTXOs[txID] != nil && spentTXOs[txID][outIdx] {
					continue
				}
				utxoMap[txID] = append(utxoMap[txID], CachedUTXO{
					Index:  outIdx,
					Output: output,
				})
			}

			if tx.IsCoinbase() {
				continue
			}

			for _, input := range tx.Inputs {
				inputTxID := input.TxIDHex()
				if spentTXOs[inputTxID] == nil {
					spentTXOs[inputTxID] = make(map[int]bool)
				}
				spentTXOs[inputTxID][input.Out] = true
			}
		}
	}

	return utxoMap, nil
}

// Close releases the underlying database handle.
func (bc *Blockchain) Close() error {
	if bc == nil || bc.db == nil {
		return nil
	}

	return bc.db.Close()
}

func openDB(dataDir string) (*pebble.DB, error) {
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}

	db, err := pebble.Open(DBPath(dataDir), &pebble.Options{
		Logger: quietLogger{},
	})
	if err != nil {
		return nil, fmt.Errorf("open pebble: %w", err)
	}

	return db, nil
}

func lastHashExists(db *pebble.DB) (bool, error) {
	_, err := loadLastHash(db)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, pebble.ErrNotFound) {
		return false, nil
	}

	return false, err
}

func loadLastHash(db *pebble.DB) ([]byte, error) {
	return loadValue(db, lastHashKey)
}

func loadValue(db *pebble.DB, key []byte) ([]byte, error) {
	value, closer, err := db.Get(key)
	if err != nil {
		return nil, err
	}
	defer closer.Close()

	copyValue := append([]byte(nil), value...)
	return copyValue, nil
}

func utxoKey(txID []byte) []byte {
	return append(append([]byte(nil), utxoPrefix...), txID...)
}

func prefixUpperBound(prefix []byte) []byte {
	upper := append([]byte(nil), prefix...)
	upper[len(upper)-1]++
	return upper
}

func (bc *Blockchain) hasUTXOEntries() (bool, error) {
	iter, err := bc.db.NewIter(&pebble.IterOptions{
		LowerBound: utxoPrefix,
		UpperBound: prefixUpperBound(utxoPrefix),
	})
	if err != nil {
		return false, err
	}
	defer iter.Close()

	return iter.First(), nil
}

func applyUTXOChanges(db *pebble.DB, batch *pebble.Batch, block *Block) error {
	for _, tx := range block.Transactions {
		if !tx.IsCoinbase() {
			for _, input := range tx.Inputs {
				key := utxoKey(input.TxID)
				encoded, err := loadValue(db, key)
				if err != nil {
					return err
				}
				cached, err := decodeCachedUTXOs(encoded)
				if err != nil {
					return err
				}

				var remaining []CachedUTXO
				found := false
				for _, item := range cached {
					if item.Index == input.Out {
						found = true
						continue
					}
					remaining = append(remaining, item)
				}
				if !found {
					return ErrInvalidTransaction
				}

				if len(remaining) == 0 {
					if err := batch.Delete(key, nil); err != nil {
						return err
					}
				} else {
					encodedRemaining, err := encodeCachedUTXOs(remaining)
					if err != nil {
						return err
					}
					if err := batch.Set(key, encodedRemaining, nil); err != nil {
						return err
					}
				}
			}
		}

		cached := make([]CachedUTXO, 0, len(tx.Outputs))
		for idx, output := range tx.Outputs {
			cached = append(cached, CachedUTXO{
				Index:  idx,
				Output: output,
			})
		}
		encodedOutputs, err := encodeCachedUTXOs(cached)
		if err != nil {
			return err
		}
		if err := batch.Set(utxoKey(tx.ID), encodedOutputs, nil); err != nil {
			return err
		}
	}

	return nil
}

type quietLogger struct{}

func (quietLogger) Infof(string, ...interface{}) {}

func (quietLogger) Fatalf(format string, args ...interface{}) {
	panic(fmt.Sprintf(format, args...))
}
