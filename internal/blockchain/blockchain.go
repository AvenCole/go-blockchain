package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

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
	// ErrInvalidBlock reports one invalid received block.
	ErrInvalidBlock = errors.New("invalid block")
	// ErrDoubleSpend reports that one output is being spent twice.
	ErrDoubleSpend = errors.New("double spend detected")
	// ErrOrphanBlock reports that a block parent is unknown locally.
	ErrOrphanBlock = errors.New("orphan block")
)

var lastHashKey = []byte("lh")
var utxoPrefix = []byte("utxo-")
var mempoolPrefix = []byte("mempool-")

// Blockchain provides chain storage and append operations.
type Blockchain struct {
	dataDir string
	tip     []byte
	db      *pebble.DB
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
		dataDir: dataDir,
		tip:     genesis.Hash,
		db:      db,
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
		dataDir: dataDir,
		tip:     tip,
		db:      db,
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
	return bc.commitBlock(transactions, nil)
}

func (bc *Blockchain) commitBlock(transactions []Transaction, mutateBatch func(*pebble.Batch) error) (*Block, error) {
	for _, tx := range transactions {
		if err := bc.ValidateTransaction(tx); err != nil {
			return nil, err
		}
	}

	lastBlock, err := bc.CurrentBlock()
	if err != nil {
		return nil, err
	}

	newBlock := NewBlock(transactions, bc.tip, lastBlock.Height+1)
	if err := bc.ValidateBlock(newBlock); err != nil {
		return nil, err
	}
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
	if mutateBatch != nil {
		if err := mutateBatch(batch); err != nil {
			return nil, err
		}
	}
	if err := batch.Commit(pebble.Sync); err != nil {
		return nil, err
	}

	bc.tip = newBlock.Hash

	return newBlock, nil
}

// SendTransaction creates a signed UTXO-style transaction and stores it in the mempool.
func (bc *Blockchain) SendTransaction(fromWallet *wallet.Wallet, to string, amount int, fee int) (Transaction, error) {
	tx, err := NewUTXOTransaction(fromWallet, to, amount, fee, bc)
	if err != nil {
		return Transaction{}, err
	}

	if err := bc.AddToMempool(tx); err != nil {
		return Transaction{}, err
	}

	return tx, nil
}

// MineMempool mines the current mempool into one new block with a coinbase reward.
func (bc *Blockchain) MineMempool(minerAddress string) (*Block, int, error) {
	transactions, err := bc.PendingTransactions()
	if err != nil {
		return nil, 0, err
	}
	if len(transactions) == 0 {
		return nil, 0, fmt.Errorf("mempool is empty")
	}

	totalFees := 0
	for _, tx := range transactions {
		prevTXs := make(map[string]Transaction)
		for _, input := range tx.Inputs {
			prevTx, err := bc.FindTransaction(input.TxID)
			if err != nil {
				return nil, 0, err
			}
			prevTXs[input.TxIDHex()] = prevTx
		}
		totalFees += tx.Fee(prevTXs)
	}

	coinbase := NewCoinbaseTransactionWithReward(minerAddress, "mined block reward", subsidy+totalFees)
	blockTransactions := append([]Transaction{coinbase}, transactions...)
	block, err := bc.commitBlock(blockTransactions, func(batch *pebble.Batch) error {
		for _, tx := range transactions {
			if err := batch.Delete(mempoolKey(tx.ID), nil); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, 0, err
	}

	return block, len(transactions), nil
}

// CurrentBlock returns the tip block.
func (bc *Blockchain) CurrentBlock() (*Block, error) {
	data, err := loadValue(bc.db, bc.tip)
	if err != nil {
		return nil, err
	}

	return DeserializeBlock(data)
}

// Height returns the current chain height.
func (bc *Blockchain) Height() (int, error) {
	current, err := bc.CurrentBlock()
	if err != nil {
		return -1, err
	}
	return current.Height, nil
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

// BlocksFromHeight returns blocks with height greater than fromHeight in ascending order.
func (bc *Blockchain) BlocksFromHeight(fromHeight int) ([]*Block, error) {
	blocks, err := bc.Blocks()
	if err != nil {
		return nil, err
	}

	var filtered []*Block
	for i := len(blocks) - 1; i >= 0; i-- {
		if blocks[i].Height > fromHeight {
			filtered = append(filtered, blocks[i])
		}
	}

	return filtered, nil
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

// ValidateTransaction performs full transaction safety checks for the current chain state.
func (bc *Blockchain) ValidateTransaction(tx Transaction) error {
	if tx.IsCoinbase() {
		return nil
	}
	if !bc.VerifyTransaction(tx) {
		return ErrInvalidTransaction
	}

	seen := make(map[string]struct{})
	for _, input := range tx.Inputs {
		key := fmt.Sprintf("%s:%d", input.TxIDHex(), input.Out)
		if _, exists := seen[key]; exists {
			return ErrDoubleSpend
		}
		seen[key] = struct{}{}

		ok, err := bc.outputAvailable(input.TxID, input.Out)
		if err != nil {
			return err
		}
		if !ok {
			return ErrDoubleSpend
		}
	}

	return nil
}

// ValidateBlock checks block integrity and rejects duplicate spends inside one block.
func (bc *Blockchain) ValidateBlock(block *Block) error {
	if block == nil {
		return ErrInvalidBlock
	}
	if !block.VerifyMerkleRoot() || !block.VerifyProofOfWork() {
		return ErrInvalidBlock
	}
	if len(block.Transactions) == 0 {
		return ErrInvalidBlock
	}

	coinbaseCount := 0
	spent := make(map[string]struct{})
	for idx, tx := range block.Transactions {
		if tx.IsCoinbase() {
			coinbaseCount++
			if idx != 0 {
				return ErrInvalidBlock
			}
			continue
		}

		if err := bc.ValidateTransaction(tx); err != nil {
			return err
		}
		for _, input := range tx.Inputs {
			key := fmt.Sprintf("%s:%d", input.TxIDHex(), input.Out)
			if _, exists := spent[key]; exists {
				return ErrDoubleSpend
			}
			spent[key] = struct{}{}
		}
	}
	if coinbaseCount != 1 {
		return ErrInvalidBlock
	}

	return nil
}

// ImportBlock stores an externally received block if it cleanly extends the tip.
func (bc *Blockchain) ImportBlock(block *Block) error {
	if block == nil {
		return ErrInvalidBlock
	}
	if block.Height == 0 {
		return ErrInvalidBlock
	}

	parent, err := bc.blockByHash(block.PrevBlockHash)
	if err != nil {
		if errors.Is(err, pebble.ErrNotFound) {
			return ErrOrphanBlock
		}
		return err
	}
	if block.Height != parent.Height+1 {
		return ErrInvalidBlock
	}
	if err := bc.validateBlockOnBranch(block, parent.Hash); err != nil {
		return err
	}

	current, err := bc.CurrentBlock()
	if err != nil {
		return err
	}

	if _, err := bc.blockByHash(block.Hash); err == nil {
		if block.Height > current.Height {
			return bc.switchTip(block.Hash)
		}
		return nil
	} else if !errors.Is(err, pebble.ErrNotFound) {
		return err
	}

	encodedBlock, err := block.Serialize()
	if err != nil {
		return err
	}

	batch := bc.db.NewBatch()
	defer batch.Close()
	if err := batch.Set(block.Hash, encodedBlock, nil); err != nil {
		return err
	}
	if err := batch.Commit(pebble.Sync); err != nil {
		return err
	}

	if block.Height > current.Height {
		return bc.switchTip(block.Hash)
	}

	return nil
}

// AddToMempool stores one signed transaction in the mempool cache.
func (bc *Blockchain) AddToMempool(tx Transaction) error {
	if err := bc.ValidateTransaction(tx); err != nil {
		return err
	}
	if err := bc.ensureNoMempoolConflicts(tx); err != nil {
		return err
	}

	encoded, err := tx.Serialize()
	if err != nil {
		return err
	}

	return bc.db.Set(mempoolKey(tx.ID), encoded, pebble.Sync)
}

// PendingTransactions returns the mempool contents sorted by txid.
func (bc *Blockchain) PendingTransactions() ([]Transaction, error) {
	iter, err := bc.db.NewIter(&pebble.IterOptions{
		LowerBound: mempoolPrefix,
		UpperBound: prefixUpperBound(mempoolPrefix),
	})
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	var transactions []Transaction
	for iter.First(); iter.Valid(); iter.Next() {
		tx, err := DeserializeTransaction(iter.Value())
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, tx.Clone())
	}
	if err := iter.Error(); err != nil {
		return nil, err
	}

	sort.Slice(transactions, func(i, j int) bool {
		return bytes.Compare(transactions[i].ID, transactions[j].ID) < 0
	})

	return transactions, nil
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

func mempoolKey(txID []byte) []byte {
	return append(append([]byte(nil), mempoolPrefix...), txID...)
}

func deleteMempoolTransactions(batch *pebble.Batch, transactions []Transaction) error {
	for _, tx := range transactions {
		if tx.IsCoinbase() {
			continue
		}
		if err := batch.Delete(mempoolKey(tx.ID), nil); err != nil {
			return err
		}
	}
	return nil
}

func (bc *Blockchain) outputAvailable(txID []byte, out int) (bool, error) {
	encoded, err := loadValue(bc.db, utxoKey(txID))
	if err != nil {
		if errors.Is(err, pebble.ErrNotFound) {
			return false, nil
		}
		return false, err
	}

	cached, err := decodeCachedUTXOs(encoded)
	if err != nil {
		return false, err
	}
	for _, item := range cached {
		if item.Index == out {
			return true, nil
		}
	}
	return false, nil
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

func (bc *Blockchain) ensureNoMempoolConflicts(tx Transaction) error {
	pending, err := bc.PendingTransactions()
	if err != nil {
		return err
	}

	for _, existing := range pending {
		for _, existingInput := range existing.Inputs {
			for _, newInput := range tx.Inputs {
				if bytes.Equal(existingInput.TxID, newInput.TxID) && existingInput.Out == newInput.Out {
					return fmt.Errorf("mempool conflict on txid=%s out=%d", newInput.TxIDHex(), newInput.Out)
				}
			}
		}
	}

	return nil
}

func (bc *Blockchain) switchTip(newTip []byte) error {
	batch := bc.db.NewBatch()
	defer batch.Close()

	if err := batch.Set(lastHashKey, newTip, nil); err != nil {
		return err
	}
	if err := batch.Commit(pebble.Sync); err != nil {
		return err
	}

	bc.tip = append([]byte(nil), newTip...)
	return bc.ReindexUTXO()
}

func (bc *Blockchain) blockByHash(hash []byte) (*Block, error) {
	data, err := loadValue(bc.db, hash)
	if err != nil {
		return nil, err
	}
	return DeserializeBlock(data)
}

func (bc *Blockchain) blocksFromTipHash(tipHash []byte) ([]*Block, error) {
	currentHash := append([]byte(nil), tipHash...)
	var blocks []*Block

	for len(currentHash) > 0 {
		block, err := bc.blockByHash(currentHash)
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, block)
		currentHash = append([]byte(nil), block.PrevBlockHash...)
	}

	return blocks, nil
}

func (bc *Blockchain) validateBlockOnBranch(block *Block, parentHash []byte) error {
	if block == nil {
		return ErrInvalidBlock
	}
	if !block.VerifyMerkleRoot() || !block.VerifyProofOfWork() {
		return ErrInvalidBlock
	}
	if len(block.Transactions) == 0 {
		return ErrInvalidBlock
	}

	txIndex, utxoState, err := bc.branchState(parentHash)
	if err != nil {
		return err
	}

	coinbaseCount := 0
	for idx, tx := range block.Transactions {
		if tx.IsCoinbase() {
			coinbaseCount++
			if idx != 0 {
				return ErrInvalidBlock
			}
			if err := applyTransactionToState(tx, txIndex, utxoState); err != nil {
				return err
			}
			continue
		}

		if err := validateTransactionOnState(tx, txIndex, utxoState); err != nil {
			return err
		}
		if err := applyTransactionToState(tx, txIndex, utxoState); err != nil {
			return err
		}
	}
	if coinbaseCount != 1 {
		return ErrInvalidBlock
	}

	return nil
}

func (bc *Blockchain) branchState(parentHash []byte) (map[string]Transaction, map[string][]CachedUTXO, error) {
	txIndex := make(map[string]Transaction)
	utxoState := make(map[string][]CachedUTXO)

	if len(parentHash) == 0 {
		return txIndex, utxoState, nil
	}

	blocks, err := bc.blocksFromTipHash(parentHash)
	if err != nil {
		return nil, nil, err
	}

	for i := len(blocks) - 1; i >= 0; i-- {
		block := blocks[i]
		for _, tx := range block.Transactions {
			if err := applyTransactionToState(tx, txIndex, utxoState); err != nil {
				return nil, nil, err
			}
		}
	}

	return txIndex, utxoState, nil
}

func validateTransactionOnState(tx Transaction, prevTXs map[string]Transaction, utxoState map[string][]CachedUTXO) error {
	if tx.IsCoinbase() {
		return nil
	}
	if !tx.Verify(prevTXs) {
		return ErrInvalidTransaction
	}

	seen := make(map[string]struct{})
	for _, input := range tx.Inputs {
		prevTx, ok := prevTXs[input.TxIDHex()]
		if !ok {
			return ErrTransactionNotFound
		}
		if input.Out < 0 || input.Out >= len(prevTx.Outputs) {
			return ErrInvalidTransaction
		}

		key := fmt.Sprintf("%s:%d", input.TxIDHex(), input.Out)
		if _, exists := seen[key]; exists {
			return ErrDoubleSpend
		}
		seen[key] = struct{}{}

		if !outputAvailableInState(utxoState, input.TxIDHex(), input.Out) {
			return ErrDoubleSpend
		}
	}

	return nil
}

func applyTransactionToState(tx Transaction, txIndex map[string]Transaction, utxoState map[string][]CachedUTXO) error {
	if !tx.IsCoinbase() {
		for _, input := range tx.Inputs {
			txIDHex := input.TxIDHex()
			cached := utxoState[txIDHex]
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
				delete(utxoState, txIDHex)
			} else {
				utxoState[txIDHex] = remaining
			}
		}
	}

	outputs := make([]CachedUTXO, 0, len(tx.Outputs))
	for idx, output := range tx.Outputs {
		outputs = append(outputs, CachedUTXO{
			Index:  idx,
			Output: output,
		})
	}
	utxoState[tx.IDHex()] = outputs
	txIndex[tx.IDHex()] = tx.Clone()
	return nil
}

func outputAvailableInState(utxoState map[string][]CachedUTXO, txIDHex string, out int) bool {
	for _, item := range utxoState[txIDHex] {
		if item.Index == out {
			return true
		}
	}
	return false
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

// BestHeight returns the best local height for one data dir, or -1 when no chain exists yet.
func BestHeight(dataDir string) (int, error) {
	bc, err := OpenBlockchain(dataDir)
	if err != nil {
		if errors.Is(err, ErrBlockchainNotInitialized) {
			return -1, nil
		}
		return -1, err
	}
	defer bc.Close()
	return bc.Height()
}

// ImportBlockToDir imports a received block into one on-disk chain.
func ImportBlockToDir(dataDir string, block *Block) error {
	if block == nil {
		return ErrInvalidBlock
	}

	exists, err := ChainExists(dataDir)
	if err != nil {
		return err
	}

	if !exists {
		if block.Height != 0 || !block.VerifyMerkleRoot() || !block.VerifyProofOfWork() {
			return ErrInvalidBlock
		}
		if len(block.Transactions) != 1 || !block.Transactions[0].IsCoinbase() {
			return ErrInvalidBlock
		}
		db, err := openDB(dataDir)
		if err != nil {
			return err
		}
		defer db.Close()

		encodedBlock, err := block.Serialize()
		if err != nil {
			return err
		}
		batch := db.NewBatch()
		defer batch.Close()
		if err := batch.Set(block.Hash, encodedBlock, nil); err != nil {
			return err
		}
		if err := batch.Set(lastHashKey, block.Hash, nil); err != nil {
			return err
		}
		if err := applyUTXOChanges(db, batch, block); err != nil {
			return err
		}
		return batch.Commit(pebble.Sync)
	}

	bc, err := OpenBlockchain(dataDir)
	if err != nil {
		return err
	}
	defer bc.Close()
	if err := bc.ValidateBlock(block); err != nil {
		return err
	}
	return bc.ImportBlock(block)
}

type quietLogger struct{}

func (quietLogger) Infof(string, ...interface{}) {}

func (quietLogger) Fatalf(format string, args ...interface{}) {
	panic(fmt.Sprintf(format, args...))
}
