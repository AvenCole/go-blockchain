package gui

import (
	"time"

	"go-blockchain/internal/blockchain"
)

func (s *Service) Blocks() ([]BlockView, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	bc, err := s.openBlockchain()
	if err == blockchain.ErrBlockchainNotInitialized {
		return []BlockView{}, nil
	}
	if err != nil {
		return nil, err
	}
	defer bc.Close()
	blocks, err := bc.Blocks()
	if err != nil {
		return nil, err
	}

	views := make([]BlockView, 0, len(blocks))
	for _, block := range blocks {
		txViews := make([]TransactionView, 0, len(block.Transactions))
		for _, tx := range block.Transactions {
			prevTXs := make(map[string]blockchain.Transaction)
			for _, input := range tx.Inputs {
				prev, err := bc.FindTransaction(input.TxID)
				if err == nil {
					prevTXs[input.TxIDHex()] = prev
				}
			}

			inputs := make([]InputView, 0, len(tx.Inputs))
			for _, input := range tx.Inputs {
				inputs = append(inputs, InputView{
					TxID:      input.TxIDHex(),
					Out:       input.Out,
					Source:    input.FromDisplay(),
					ScriptSig: input.EffectiveScriptSig().String(),
				})
			}

			outputs := make([]OutputView, 0, len(tx.Outputs))
			for _, output := range tx.Outputs {
				outputs = append(outputs, OutputView{
					To:           output.Address(),
					Value:        output.Value,
					ScriptPubKey: output.EffectiveScriptPubKey().String(),
				})
			}

			txViews = append(txViews, TransactionView{
				ID:           tx.IDHex(),
				Version:      tx.Version,
				Fee:          tx.Fee(prevTXs),
				UsesScriptVM: tx.UsesScriptVM(),
				Inputs:       inputs,
				Outputs:      outputs,
			})
		}

		views = append(views, BlockView{
			Height:           block.Height,
			Hash:             block.HashHex(),
			PrevHash:         block.PrevHashHex(),
			MerkleRoot:       block.MerkleRootHex(),
			Difficulty:       block.Difficulty,
			Nonce:            block.Nonce,
			PoWValid:         block.VerifyProofOfWork(),
			Timestamp:        time.Unix(block.Timestamp, 0).UTC().Format(time.RFC3339),
			TransactionCount: len(block.Transactions),
			Transactions:     txViews,
		})
	}

	return views, nil
}
