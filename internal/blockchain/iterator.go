package blockchain

import (
	"github.com/cockroachdb/pebble"
)

// Iterator walks blocks from the chain tip backwards to genesis.
type Iterator struct {
	currentHash []byte
	db          *pebble.DB
}

// Next returns the current block and advances to the previous one.
func (it *Iterator) Next() (*Block, error) {
	if len(it.currentHash) == 0 {
		return nil, nil
	}

	data, err := loadValue(it.db, it.currentHash)
	if err != nil {
		return nil, err
	}

	block, err := DeserializeBlock(data)
	if err != nil {
		return nil, err
	}

	it.currentHash = append([]byte(nil), block.PrevBlockHash...)

	return block, nil
}
