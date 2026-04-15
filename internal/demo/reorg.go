package demo

import (
	"fmt"

	"go-blockchain/internal/blockchain"
	"go-blockchain/internal/wallet"
)

type ReorgRecoveryResult struct {
	MinedBlockHash    string
	MinedBlockHeight  int
	ReorgTxID         string
	Restored          bool
	MempoolSize       int
	BalanceAfterReorg int
	OldHeight         int
	OldTip            string
	NewHeight         int
	NewTip            string
}

func RunReorgMempoolRecovery(dataDir string, from *wallet.Wallet, to string, amount int, advance int) (ReorgRecoveryResult, error) {
	chain, err := blockchain.OpenBlockchain(dataDir)
	if err != nil {
		return ReorgRecoveryResult{}, err
	}

	tx, err := chain.SendTransaction(from, to, amount, 0)
	if err != nil {
		_ = chain.Close()
		return ReorgRecoveryResult{}, err
	}
	block, _, err := chain.MineMempool(from.Address())
	if err != nil {
		_ = chain.Close()
		return ReorgRecoveryResult{}, err
	}
	if err := chain.Close(); err != nil {
		return ReorgRecoveryResult{}, err
	}

	result, err := ForceReorgToLongerGenesisFork(dataDir, from.Address(), tx.IDHex(), to, advance)
	if err != nil {
		return ReorgRecoveryResult{}, err
	}
	result.MinedBlockHash = block.HashHex()
	result.MinedBlockHeight = block.Height
	return result, nil
}

func ForceReorgToLongerGenesisFork(dataDir string, minerAddress string, watchedTxID string, receiverAddress string, advance int) (ReorgRecoveryResult, error) {
	chain, err := blockchain.OpenBlockchain(dataDir)
	if err != nil {
		return ReorgRecoveryResult{}, err
	}
	defer chain.Close()

	current, err := chain.CurrentBlock()
	if err != nil {
		return ReorgRecoveryResult{}, err
	}
	beforeHeight := current.Height
	beforeTip := current.HashHex()

	blocks, err := chain.Blocks()
	if err != nil {
		return ReorgRecoveryResult{}, err
	}
	genesis := blocks[len(blocks)-1]

	targetHeight := current.Height + advance
	prevHash := append([]byte(nil), genesis.Hash...)
	for height := 1; height <= targetHeight; height++ {
		forkBlock := blockchain.NewBlock(
			[]blockchain.Transaction{blockchain.NewCoinbaseTransaction(minerAddress, fmt.Sprintf("reorg fork height %d", height))},
			prevHash,
			height,
		)
		if err := chain.ImportBlock(forkBlock); err != nil {
			return ReorgRecoveryResult{}, err
		}
		prevHash = append([]byte(nil), forkBlock.Hash...)
	}

	after, err := chain.CurrentBlock()
	if err != nil {
		return ReorgRecoveryResult{}, err
	}
	pending, err := chain.PendingTransactions()
	if err != nil {
		return ReorgRecoveryResult{}, err
	}
	balance, err := chain.BalanceOf(receiverAddress)
	if err != nil {
		return ReorgRecoveryResult{}, err
	}

	result := ReorgRecoveryResult{
		ReorgTxID:         watchedTxID,
		MempoolSize:       len(pending),
		BalanceAfterReorg: balance,
		OldHeight:         beforeHeight,
		OldTip:            beforeTip,
		NewHeight:         after.Height,
		NewTip:            after.HashHex(),
	}
	for _, candidate := range pending {
		if candidate.IDHex() == watchedTxID {
			result.Restored = true
			break
		}
	}

	status, err := chain.LastReorgStatus()
	if err == nil && status != nil {
		result.OldHeight = status.OldHeight
		result.OldTip = status.OldTip
		result.NewHeight = status.NewHeight
		result.NewTip = status.NewTip
	}

	return result, nil
}
