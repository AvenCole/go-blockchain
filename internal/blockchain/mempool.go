package blockchain

import "fmt"

// MempoolSize reports the number of queued transactions.
func (bc *Blockchain) MempoolSize() (int, error) {
	txs, err := bc.PendingTransactions()
	if err != nil {
		return 0, err
	}
	return len(txs), nil
}

// ValidateMempoolNotEmpty returns an error when no pending transactions exist.
func (bc *Blockchain) ValidateMempoolNotEmpty() error {
	size, err := bc.MempoolSize()
	if err != nil {
		return err
	}
	if size == 0 {
		return fmt.Errorf("mempool is empty")
	}
	return nil
}
