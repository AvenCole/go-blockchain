package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"go-blockchain/internal/blockchain"
	"go-blockchain/internal/config"
	"go-blockchain/internal/network"
	"go-blockchain/internal/wallet"
)

const version = "0.1.0"
const perfOutputDirEnv = "GO_BLOCKCHAIN_PERF_DIR"

// App provides the minimal CLI skeleton approved in Plan 1.
type App struct {
	cfg    config.Config
	stdout io.Writer
	stderr io.Writer
}

// NewApp creates the initialization-phase CLI entrypoint.
func NewApp(cfg config.Config, stdout io.Writer, stderr io.Writer) App {
	return App{
		cfg:    cfg,
		stdout: stdout,
		stderr: stderr,
	}
}

// Run executes one command and returns a process exit code.
func (a App) Run(args []string) int {
	if len(args) == 0 || isHelpArg(args[0]) {
		a.printHelp()
		return 0
	}

	switch strings.ToLower(args[0]) {
	case "help":
		a.printHelp()
		return 0
	case "version", "about":
		a.printVersion()
		return 0
	case "doctor":
		a.printDoctor()
		return 0
	case "createblockchain":
		return a.createBlockchain(args[1:])
	case "addblock":
		return a.addBlock(args[1:])
	case "printchain":
		return a.printChain(args[1:])
	case "send":
		return a.send(args[1:])
	case "getbalance":
		return a.getBalance(args[1:])
	case "createwallet":
		return a.createWallet(args[1:])
	case "listaddresses":
		return a.listAddresses(args[1:])
	case "reindexutxo":
		return a.reindexUTXO(args[1:])
	case "showscript":
		return a.showScript(args[1:])
	case "mine":
		return a.mine(args[1:])
	case "printmempool":
		return a.printMempool(args[1:])
	case "startnode":
		return a.startNode(args[1:])
	case "simdouble":
		return a.simulateDoubleSpend(args[1:])
	case "runperf":
		return a.runPerformance(args[1:])
	default:
		fmt.Fprintf(a.stderr, "unknown command: %s\n\n", args[0])
		a.printHelp()
		return 1
	}
}

func isHelpArg(arg string) bool {
	switch strings.ToLower(arg) {
	case "-h", "--help", "/?":
		return true
	default:
		return false
	}
}

func (a App) printHelp() {
	fmt.Fprintf(a.stdout, "%s blockchain simulation CLI\n\n", a.cfg.ProjectName)
	fmt.Fprintln(a.stdout, "Usage:")
	fmt.Fprintf(a.stdout, "  %s [command]\n", a.cfg.ProjectName)
	fmt.Fprintln(a.stdout)
	fmt.Fprintln(a.stdout, "Available commands:")
	fmt.Fprintln(a.stdout, "  help      Show command help")
	fmt.Fprintln(a.stdout, "  version   Show project version information")
	fmt.Fprintln(a.stdout, "  about     Show project version information")
	fmt.Fprintln(a.stdout, "  doctor    Check initialization-stage readiness")
	fmt.Fprintln(a.stdout, "  createblockchain [genesis-address]  Create a new blockchain")
	fmt.Fprintln(a.stdout, "  addblock <label>                    Append a debug coinbase-style block")
	fmt.Fprintln(a.stdout, "  printchain                       Print the blockchain from tip to genesis")
	fmt.Fprintln(a.stdout, "  send <from> <to> <amount> [fee]     Queue a signed transaction into the mempool")
	fmt.Fprintln(a.stdout, "  getbalance <address>                Show the current UTXO balance for one address")
	fmt.Fprintln(a.stdout, "  createwallet                     Create a new wallet and save it")
	fmt.Fprintln(a.stdout, "  listaddresses                    List all saved wallet addresses")
	fmt.Fprintln(a.stdout, "  reindexutxo                      Rebuild the cached UTXO set")
	fmt.Fprintln(a.stdout, "  showscript <address>             Show the standard P2PKH locking script for one address")
	fmt.Fprintln(a.stdout, "  startnode <addr> [seed] [miner]  Start one local network node")
	fmt.Fprintln(a.stdout, "  mine <miner-address>             Mine all pending transactions into a block")
	fmt.Fprintln(a.stdout, "  printmempool                     List pending transaction IDs")
	fmt.Fprintln(a.stdout, "  simdouble <from> <to1> <to2> <amount> [fee]  Demonstrate double-spend rejection")
	fmt.Fprintln(a.stdout, "  runperf [lookups]                Run cache-vs-scan performance comparison")
}

func (a App) printVersion() {
	fmt.Fprintf(a.stdout, "%s version %s\n", a.cfg.ProjectName, version)
	fmt.Fprintf(a.stdout, "default data dir: %s\n", a.cfg.DataDir)
	fmt.Fprintf(a.stdout, "default port: %d\n", a.cfg.DefaultPort)
}

func (a App) printDoctor() {
	initialized, err := blockchain.ChainExists(a.cfg.DataDir)
	chainStatus := "unknown"
	if err == nil {
		if initialized {
			chainStatus = "ready"
		} else {
			chainStatus = "not_initialized"
		}
	}

	fmt.Fprintf(a.stdout, "doctor: blockchain transaction demo is ready\n")
	fmt.Fprintf(a.stdout, "project=%s\n", a.cfg.ProjectName)
	fmt.Fprintf(a.stdout, "data_dir=%s\n", a.cfg.DataDir)
	fmt.Fprintf(a.stdout, "default_port=%d\n", a.cfg.DefaultPort)
	fmt.Fprintf(a.stdout, "log_level=%s\n", a.cfg.LogLevel)
	fmt.Fprintf(a.stdout, "network_mode=%s\n", a.cfg.NetworkMode)
	fmt.Fprintf(a.stdout, "chain_status=%s\n", chainStatus)
	fmt.Fprintln(a.stdout, "next_step=prepare final report and presentation")
}

func (a App) createBlockchain(args []string) int {
	genesisAddress := "miner"
	if len(args) > 0 {
		genesisAddress = strings.Join(args, " ")
	}
	if !wallet.ValidateAddress(genesisAddress) {
		fmt.Fprintln(a.stderr, "invalid genesis address")
		return 1
	}

	chain, err := blockchain.CreateBlockchain(a.cfg.DataDir, genesisAddress)
	if err != nil {
		if errors.Is(err, blockchain.ErrBlockchainAlreadyExists) {
			fmt.Fprintln(a.stderr, "blockchain already exists")
			return 1
		}

		fmt.Fprintf(a.stderr, "create blockchain: %v\n", err)
		return 1
	}
	defer chain.Close()

	fmt.Fprintf(a.stdout, "created blockchain with genesis reward address: %s\n", genesisAddress)
	return 0
}

func (a App) addBlock(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(a.stderr, "addblock requires block data")
		return 1
	}

	chain, err := blockchain.OpenBlockchain(a.cfg.DataDir)
	if err != nil {
		if errors.Is(err, blockchain.ErrBlockchainNotInitialized) {
			fmt.Fprintln(a.stderr, "blockchain not initialized; run createblockchain first")
			return 1
		}

		fmt.Fprintf(a.stderr, "open blockchain: %v\n", err)
		return 1
	}
	defer chain.Close()

	wallets, err := wallet.NewWallets(a.cfg.DataDir)
	if err != nil {
		fmt.Fprintf(a.stderr, "load wallets: %v\n", err)
		return 1
	}
	addresses := wallets.Addresses()
	if len(addresses) == 0 {
		fmt.Fprintln(a.stderr, "addblock requires at least one wallet address; run createwallet first")
		return 1
	}

	noteTx := blockchain.NewCoinbaseTransaction(addresses[0], strings.Join(args, " "))
	block, err := chain.AddBlock([]blockchain.Transaction{noteTx})
	if err != nil {
		fmt.Fprintf(a.stderr, "add block: %v\n", err)
		return 1
	}

	fmt.Fprintf(a.stdout, "added block height=%d hash=%s\n", block.Height, block.HashHex())
	return 0
}

func (a App) printChain(args []string) int {
	if len(args) > 0 {
		fmt.Fprintln(a.stderr, "printchain does not accept extra arguments")
		return 1
	}

	chain, err := blockchain.OpenBlockchain(a.cfg.DataDir)
	if err != nil {
		if errors.Is(err, blockchain.ErrBlockchainNotInitialized) {
			fmt.Fprintln(a.stderr, "blockchain not initialized; run createblockchain first")
			return 1
		}

		fmt.Fprintf(a.stderr, "open blockchain: %v\n", err)
		return 1
	}
	defer chain.Close()

	blocks, err := chain.Blocks()
	if err != nil {
		fmt.Fprintf(a.stderr, "print blockchain: %v\n", err)
		return 1
	}

	for _, block := range blocks {
		fmt.Fprintf(a.stdout, "Height: %d\n", block.Height)
		fmt.Fprintf(a.stdout, "Timestamp: %s\n", time.Unix(block.Timestamp, 0).UTC().Format(time.RFC3339))
		fmt.Fprintf(a.stdout, "Hash: %s\n", block.HashHex())
		fmt.Fprintf(a.stdout, "PrevHash: %s\n", block.PrevHashHex())
		fmt.Fprintf(a.stdout, "MerkleRoot: %s\n", block.MerkleRootHex())
		fmt.Fprintf(a.stdout, "Difficulty: %d\n", block.Difficulty)
		fmt.Fprintf(a.stdout, "Nonce: %d\n", block.Nonce)
		fmt.Fprintf(a.stdout, "PoWValid: %t\n", block.VerifyProofOfWork())
		fmt.Fprintf(a.stdout, "Transactions: %d\n", len(block.Transactions))
		for _, tx := range block.Transactions {
			fmt.Fprintf(a.stdout, "  TxID: %s\n", tx.IDHex())
			for _, input := range tx.Inputs {
				source := input.FromDisplay()
				fmt.Fprintf(a.stdout, "    Input: txid=%s out=%d source=%s\n", input.TxIDHex(), input.Out, source)
				fmt.Fprintf(a.stdout, "      ScriptSig: %s\n", input.EffectiveScriptSig().String())
			}
			for _, output := range tx.Outputs {
				fmt.Fprintf(a.stdout, "    Output: to=%s value=%d\n", output.Address(), output.Value)
				fmt.Fprintf(a.stdout, "      ScriptPubKey: %s\n", output.EffectiveScriptPubKey().String())
			}
		}
		fmt.Fprintln(a.stdout)
	}

	return 0
}

func (a App) send(args []string) int {
	if len(args) != 3 && len(args) != 4 {
		fmt.Fprintln(a.stderr, "send requires: <from> <to> <amount> [fee]")
		return 1
	}
	if !wallet.ValidateAddress(args[0]) || !wallet.ValidateAddress(args[1]) {
		fmt.Fprintln(a.stderr, "send requires valid from/to wallet addresses")
		return 1
	}

	amount, err := strconv.Atoi(args[2])
	if err != nil {
		fmt.Fprintf(a.stderr, "parse amount: %v\n", err)
		return 1
	}
	fee := 0
	if len(args) == 4 {
		fee, err = strconv.Atoi(args[3])
		if err != nil {
			fmt.Fprintf(a.stderr, "parse fee: %v\n", err)
			return 1
		}
	}

	chain, err := blockchain.OpenBlockchain(a.cfg.DataDir)
	if err != nil {
		if errors.Is(err, blockchain.ErrBlockchainNotInitialized) {
			fmt.Fprintln(a.stderr, "blockchain not initialized; run createblockchain first")
			return 1
		}

		fmt.Fprintf(a.stderr, "open blockchain: %v\n", err)
		return 1
	}
	defer chain.Close()

	wallets, err := wallet.NewWallets(a.cfg.DataDir)
	if err != nil {
		fmt.Fprintf(a.stderr, "load wallets: %v\n", err)
		return 1
	}

	fromWallet, ok := wallets.GetWallet(args[0])
	if !ok {
		fmt.Fprintln(a.stderr, "sender wallet not found")
		return 1
	}

	tx, err := chain.SendTransaction(fromWallet, args[1], amount, fee)
	if err != nil {
		fmt.Fprintf(a.stderr, "send transaction: %v\n", err)
		return 1
	}

	fmt.Fprintf(a.stdout, "queued transaction txid=%s fee=%d\n", tx.IDHex(), fee)
	return 0
}

func (a App) getBalance(args []string) int {
	if len(args) != 1 {
		fmt.Fprintln(a.stderr, "getbalance requires: <address>")
		return 1
	}

	chain, err := blockchain.OpenBlockchain(a.cfg.DataDir)
	if err != nil {
		if errors.Is(err, blockchain.ErrBlockchainNotInitialized) {
			fmt.Fprintln(a.stderr, "blockchain not initialized; run createblockchain first")
			return 1
		}

		fmt.Fprintf(a.stderr, "open blockchain: %v\n", err)
		return 1
	}
	defer chain.Close()

	balance, err := chain.BalanceOf(args[0])
	if err != nil {
		fmt.Fprintf(a.stderr, "get balance: %v\n", err)
		return 1
	}

	fmt.Fprintf(a.stdout, "balance[%s]=%d\n", args[0], balance)
	return 0
}

func (a App) createWallet(args []string) int {
	if len(args) != 0 {
		fmt.Fprintln(a.stderr, "createwallet does not accept extra arguments")
		return 1
	}

	wallets, err := wallet.NewWallets(a.cfg.DataDir)
	if err != nil {
		fmt.Fprintf(a.stderr, "load wallets: %v\n", err)
		return 1
	}

	address, err := wallets.CreateWallet()
	if err != nil {
		fmt.Fprintf(a.stderr, "create wallet: %v\n", err)
		return 1
	}

	if err := wallets.SaveFile(a.cfg.DataDir); err != nil {
		fmt.Fprintf(a.stderr, "save wallets: %v\n", err)
		return 1
	}

	fmt.Fprintf(a.stdout, "created wallet address=%s\n", address)
	return 0
}

func (a App) listAddresses(args []string) int {
	if len(args) != 0 {
		fmt.Fprintln(a.stderr, "listaddresses does not accept extra arguments")
		return 1
	}

	wallets, err := wallet.NewWallets(a.cfg.DataDir)
	if err != nil {
		fmt.Fprintf(a.stderr, "load wallets: %v\n", err)
		return 1
	}

	addresses := wallets.Addresses()
	if len(addresses) == 0 {
		fmt.Fprintln(a.stdout, "no wallet addresses found")
		return 0
	}

	for _, address := range addresses {
		fmt.Fprintln(a.stdout, address)
	}

	return 0
}

func (a App) reindexUTXO(args []string) int {
	if len(args) != 0 {
		fmt.Fprintln(a.stderr, "reindexutxo does not accept extra arguments")
		return 1
	}

	chain, err := blockchain.OpenBlockchain(a.cfg.DataDir)
	if err != nil {
		if errors.Is(err, blockchain.ErrBlockchainNotInitialized) {
			fmt.Fprintln(a.stderr, "blockchain not initialized; run createblockchain first")
			return 1
		}

		fmt.Fprintf(a.stderr, "open blockchain: %v\n", err)
		return 1
	}
	defer chain.Close()

	if err := chain.ReindexUTXO(); err != nil {
		fmt.Fprintf(a.stderr, "reindex utxo: %v\n", err)
		return 1
	}

	fmt.Fprintln(a.stdout, "utxo set reindexed")
	return 0
}

func (a App) showScript(args []string) int {
	if len(args) != 1 {
		fmt.Fprintln(a.stderr, "showscript requires: <address>")
		return 1
	}

	pubKeyHash, err := wallet.PublicKeyHashFromAddress(args[0])
	if err != nil {
		fmt.Fprintf(a.stderr, "decode address: %v\n", err)
		return 1
	}

	script := blockchain.NewP2PKHLockingScript(pubKeyHash)
	fmt.Fprintf(a.stdout, "address=%s\n", args[0])
	fmt.Fprintf(a.stdout, "pubKeyHash=%x\n", pubKeyHash)
	fmt.Fprintf(a.stdout, "scriptPubKey=%s\n", script.String())
	return 0
}

func (a App) mine(args []string) int {
	if len(args) != 1 {
		fmt.Fprintln(a.stderr, "mine requires: <miner-address>")
		return 1
	}
	if !wallet.ValidateAddress(args[0]) {
		fmt.Fprintln(a.stderr, "invalid miner address")
		return 1
	}

	chain, err := blockchain.OpenBlockchain(a.cfg.DataDir)
	if err != nil {
		if errors.Is(err, blockchain.ErrBlockchainNotInitialized) {
			fmt.Fprintln(a.stderr, "blockchain not initialized; run createblockchain first")
			return 1
		}
		fmt.Fprintf(a.stderr, "open blockchain: %v\n", err)
		return 1
	}
	defer chain.Close()

	block, mined, err := chain.MineMempool(args[0])
	if err != nil {
		fmt.Fprintf(a.stderr, "mine mempool: %v\n", err)
		return 1
	}

	fmt.Fprintf(a.stdout, "mined block height=%d txs=%d hash=%s\n", block.Height, mined, block.HashHex())
	return 0
}

func (a App) printMempool(args []string) int {
	if len(args) != 0 {
		fmt.Fprintln(a.stderr, "printmempool does not accept extra arguments")
		return 1
	}

	chain, err := blockchain.OpenBlockchain(a.cfg.DataDir)
	if err != nil {
		if errors.Is(err, blockchain.ErrBlockchainNotInitialized) {
			fmt.Fprintln(a.stderr, "blockchain not initialized; run createblockchain first")
			return 1
		}
		fmt.Fprintf(a.stderr, "open blockchain: %v\n", err)
		return 1
	}
	defer chain.Close()

	txs, err := chain.PendingTransactions()
	if err != nil {
		fmt.Fprintf(a.stderr, "read mempool: %v\n", err)
		return 1
	}
	if len(txs) == 0 {
		fmt.Fprintln(a.stdout, "mempool empty")
		return 0
	}
	for _, tx := range txs {
		fmt.Fprintln(a.stdout, tx.IDHex())
	}
	return 0
}

func (a App) startNode(args []string) int {
	if len(args) < 1 || len(args) > 3 {
		fmt.Fprintln(a.stderr, "startnode requires: <addr> [seed] [miner-address]")
		return 1
	}

	address := args[0]
	seed := ""
	miner := ""
	if len(args) >= 2 {
		seed = args[1]
	}
	if len(args) == 3 {
		miner = args[2]
		if !wallet.ValidateAddress(miner) {
			fmt.Fprintln(a.stderr, "invalid miner address")
			return 1
		}
	}

	node := network.NewNode(address, a.cfg.DataDir, miner)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if seed != "" {
		go func() {
			time.Sleep(200 * time.Millisecond)
			_ = node.Connect(seed)
		}()
	}

	fmt.Fprintf(a.stdout, "starting node at %s\n", address)
	if seed != "" {
		fmt.Fprintf(a.stdout, "seed=%s\n", seed)
	}
	if miner != "" {
		fmt.Fprintf(a.stdout, "miner=%s\n", miner)
	}

	if err := node.Listen(ctx); err != nil {
		fmt.Fprintf(a.stderr, "start node: %v\n", err)
		return 1
	}

	return 0
}

func (a App) simulateDoubleSpend(args []string) int {
	if len(args) != 4 && len(args) != 5 {
		fmt.Fprintln(a.stderr, "simdouble requires: <from> <to1> <to2> <amount> [fee]")
		return 1
	}
	if !wallet.ValidateAddress(args[0]) || !wallet.ValidateAddress(args[1]) || !wallet.ValidateAddress(args[2]) {
		fmt.Fprintln(a.stderr, "simdouble requires valid wallet addresses")
		return 1
	}

	amount, err := strconv.Atoi(args[3])
	if err != nil {
		fmt.Fprintf(a.stderr, "parse amount: %v\n", err)
		return 1
	}
	fee := 0
	if len(args) == 5 {
		fee, err = strconv.Atoi(args[4])
		if err != nil {
			fmt.Fprintf(a.stderr, "parse fee: %v\n", err)
			return 1
		}
	}

	chain, err := blockchain.OpenBlockchain(a.cfg.DataDir)
	if err != nil {
		if errors.Is(err, blockchain.ErrBlockchainNotInitialized) {
			fmt.Fprintln(a.stderr, "blockchain not initialized; run createblockchain first")
			return 1
		}
		fmt.Fprintf(a.stderr, "open blockchain: %v\n", err)
		return 1
	}
	defer chain.Close()

	wallets, err := wallet.NewWallets(a.cfg.DataDir)
	if err != nil {
		fmt.Fprintf(a.stderr, "load wallets: %v\n", err)
		return 1
	}
	fromWallet, ok := wallets.GetWallet(args[0])
	if !ok {
		fmt.Fprintln(a.stderr, "sender wallet not found")
		return 1
	}

	firstTx, err := chain.SendTransaction(fromWallet, args[1], amount, fee)
	if err != nil {
		fmt.Fprintf(a.stderr, "queue first tx: %v\n", err)
		return 1
	}
	secondTx, err := blockchain.NewUTXOTransaction(fromWallet, args[2], amount, fee, chain)
	if err != nil {
		fmt.Fprintf(a.stderr, "build second tx: %v\n", err)
		return 1
	}
	secondErr := chain.AddToMempool(secondTx)
	if !errors.Is(secondErr, blockchain.ErrDoubleSpend) && (secondErr == nil || !strings.Contains(secondErr.Error(), "mempool conflict")) {
		fmt.Fprintf(a.stderr, "double-spend simulation expected rejection, got: %v\n", secondErr)
		return 1
	}

	fmt.Fprintf(a.stdout, "first_tx=%s queued\n", firstTx.IDHex())
	fmt.Fprintf(a.stdout, "second_tx=%s rejected=%v\n", secondTx.IDHex(), secondErr != nil)
	return 0
}

func (a App) runPerformance(args []string) int {
	lookups := 20
	if len(args) > 1 {
		fmt.Fprintln(a.stderr, "runperf accepts at most one argument: [lookups]")
		return 1
	}
	if len(args) == 1 {
		n, err := strconv.Atoi(args[0])
		if err != nil || n <= 0 {
			fmt.Fprintf(a.stderr, "parse lookups: %v\n", err)
			return 1
		}
		lookups = n
	}

	wallets, err := wallet.NewWallets(a.cfg.DataDir)
	if err != nil {
		fmt.Fprintf(a.stderr, "load wallets: %v\n", err)
		return 1
	}
	addresses := wallets.Addresses()
	if len(addresses) == 0 {
		fmt.Fprintln(a.stderr, "runperf requires at least one wallet address")
		return 1
	}

	chain, err := blockchain.OpenBlockchain(a.cfg.DataDir)
	if err != nil {
		if errors.Is(err, blockchain.ErrBlockchainNotInitialized) {
			fmt.Fprintln(a.stderr, "blockchain not initialized; run createblockchain first")
			return 1
		}
		fmt.Fprintf(a.stderr, "open blockchain: %v\n", err)
		return 1
	}
	defer chain.Close()

	result, err := chain.RunBalanceBenchmark(addresses, lookups)
	if err != nil {
		fmt.Fprintf(a.stderr, "run benchmark: %v\n", err)
		return 1
	}
	outputDir := os.Getenv(perfOutputDirEnv)
	if outputDir == "" {
		outputDir = "docs/perf"
	}
	if err := blockchain.WriteBalanceBenchmark(result, outputDir); err != nil {
		fmt.Fprintf(a.stderr, "write benchmark files: %v\n", err)
		return 1
	}

	fmt.Fprintf(a.stdout, "cached_ms=%.3f\n", result.CachedDurationMs)
	fmt.Fprintf(a.stdout, "scan_ms=%.3f\n", result.FullScanDurationMs)
	fmt.Fprintf(a.stdout, "speedup=%.3fx\n", result.Speedup)
	fmt.Fprintf(a.stdout, "report=%s\n", filepath.Join(outputDir, "latest.md"))
	return 0
}
