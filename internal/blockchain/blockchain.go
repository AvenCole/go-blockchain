package blockchain

import (
	"bytes"
	"crypto/ecdsa"
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
				if !output.IsLockedWith(pubKeyHash) {
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
				if !input.UsesKey(pubKeyHash) {
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
func (bc *Blockchain) FindUTXO(pubKeyHash []byte) ([]TXOutput, error) {
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
				if output.IsLockedWith(pubKeyHash) {
					utxos = append(utxos, output)
				}
			}

			if tx.IsCoinbase() {
				continue
			}

			for _, input := range tx.Inputs {
				if !input.UsesKey(pubKeyHash) {
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
