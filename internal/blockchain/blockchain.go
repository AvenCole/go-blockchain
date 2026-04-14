package blockchain

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cockroachdb/pebble"
)

var (
	// ErrBlockchainAlreadyExists reports that the chain has already been created.
	ErrBlockchainAlreadyExists = errors.New("blockchain already exists")
	// ErrBlockchainNotInitialized reports that the chain has not been created yet.
	ErrBlockchainNotInitialized = errors.New("blockchain not initialized")
	// ErrInsufficientFunds reports that the sender cannot cover the requested amount.
	ErrInsufficientFunds = errors.New("insufficient funds")
)

var lastHashKey = []byte("lh")

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
	if err := batch.Commit(pebble.Sync); err != nil {
		db.Close()
		return nil, err
	}

	return &Blockchain{
		tip: genesis.Hash,
		db:  db,
	}, nil
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

	return &Blockchain{
		tip: tip,
		db:  db,
	}, nil
}

// AddBlock appends a new block to the tip of the chain.
func (bc *Blockchain) AddBlock(transactions []Transaction) (*Block, error) {
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
	if err := batch.Commit(pebble.Sync); err != nil {
		return nil, err
	}

	bc.tip = newBlock.Hash

	return newBlock, nil
}

// SendTransaction creates a UTXO-style unsigned transaction and stores it in a new block.
func (bc *Blockchain) SendTransaction(from, to string, amount int) (*Block, Transaction, error) {
	accumulated, spendable, err := bc.FindSpendableOutputs(from, amount)
	if err != nil {
		return nil, Transaction{}, err
	}

	tx, err := NewUTXOTransaction(from, to, amount, spendable, accumulated)
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

// BalanceOf computes the balance by summing unspent outputs.
func (bc *Blockchain) BalanceOf(address string) (int, error) {
	utxos, err := bc.FindUTXO(address)
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
func (bc *Blockchain) FindSpendableOutputs(address string, amount int) (int, map[string][]int, error) {
	blocks, err := bc.Blocks()
	if err != nil {
		return 0, nil, err
	}

	spentTXOs := make(map[string]map[int]bool)
	unspent := make(map[string][]int)
	accumulated := 0

	for _, block := range blocks {
		for _, tx := range block.Transactions {
			txID := tx.IDHex()

			for outIdx, output := range tx.Outputs {
				if spentTXOs[txID] != nil && spentTXOs[txID][outIdx] {
					continue
				}
				if !output.IsLockedWith(address) {
					continue
				}

				accumulated += output.Value
				unspent[txID] = append(unspent[txID], outIdx)
				if accumulated >= amount {
					return accumulated, unspent, nil
				}
			}

			if tx.IsCoinbase() {
				continue
			}

			for _, input := range tx.Inputs {
				if !input.UsesKey(address) {
					continue
				}

				inputTxID := input.TxIDHex()
				if spentTXOs[inputTxID] == nil {
					spentTXOs[inputTxID] = make(map[int]bool)
				}
				spentTXOs[inputTxID][input.Out] = true
			}
		}
	}

	return accumulated, unspent, nil
}

// FindUTXO returns all currently unspent outputs for one address.
func (bc *Blockchain) FindUTXO(address string) ([]TXOutput, error) {
	blocks, err := bc.Blocks()
	if err != nil {
		return nil, err
	}

	spentTXOs := make(map[string]map[int]bool)
	var utxos []TXOutput

	for _, block := range blocks {
		for _, tx := range block.Transactions {
			txID := tx.IDHex()

			for outIdx, output := range tx.Outputs {
				if spentTXOs[txID] != nil && spentTXOs[txID][outIdx] {
					continue
				}
				if output.IsLockedWith(address) {
					utxos = append(utxos, output)
				}
			}

			if tx.IsCoinbase() {
				continue
			}

			for _, input := range tx.Inputs {
				if !input.UsesKey(address) {
					continue
				}

				inputTxID := input.TxIDHex()
				if spentTXOs[inputTxID] == nil {
					spentTXOs[inputTxID] = make(map[int]bool)
				}
				spentTXOs[inputTxID][input.Out] = true
			}
		}
	}

	return utxos, nil
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

type quietLogger struct{}

func (quietLogger) Infof(string, ...interface{}) {}

func (quietLogger) Fatalf(format string, args ...interface{}) {
	panic(fmt.Sprintf(format, args...))
}
