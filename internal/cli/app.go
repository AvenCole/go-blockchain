package cli

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"go-blockchain/internal/blockchain"
	"go-blockchain/internal/config"
	"go-blockchain/internal/wallet"
)

const version = "0.1.0"

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
	fmt.Fprintln(a.stdout, "  send <from> <to> <amount>           Add an unsigned UTXO-style transaction block")
	fmt.Fprintln(a.stdout, "  getbalance <address>                Show the current UTXO balance for one address")
	fmt.Fprintln(a.stdout, "  createwallet                     Create a new wallet and save it")
	fmt.Fprintln(a.stdout, "  listaddresses                    List all saved wallet addresses")
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
	fmt.Fprintln(a.stdout, "next_step=implement transaction signatures")
}

func (a App) createBlockchain(args []string) int {
	genesisAddress := "miner"
	if len(args) > 0 {
		genesisAddress = strings.Join(args, " ")
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

	noteTx := blockchain.NewCoinbaseTransaction("system", strings.Join(args, " "))
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
		fmt.Fprintf(a.stdout, "Transactions: %d\n", len(block.Transactions))
		for _, tx := range block.Transactions {
			fmt.Fprintf(a.stdout, "  TxID: %s\n", tx.IDHex())
			for _, input := range tx.Inputs {
				fmt.Fprintf(a.stdout, "    Input: txid=%s out=%d from=%s\n", input.TxIDHex(), input.Out, input.From)
			}
			for _, output := range tx.Outputs {
				fmt.Fprintf(a.stdout, "    Output: to=%s value=%d\n", output.To, output.Value)
			}
		}
		fmt.Fprintln(a.stdout)
	}

	return 0
}

func (a App) send(args []string) int {
	if len(args) != 3 {
		fmt.Fprintln(a.stderr, "send requires: <from> <to> <amount>")
		return 1
	}

	amount, err := strconv.Atoi(args[2])
	if err != nil {
		fmt.Fprintf(a.stderr, "parse amount: %v\n", err)
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

	block, tx, err := chain.SendTransaction(args[0], args[1], amount)
	if err != nil {
		fmt.Fprintf(a.stderr, "send transaction: %v\n", err)
		return 1
	}

	fmt.Fprintf(a.stdout, "sent transaction txid=%s in block height=%d\n", tx.IDHex(), block.Height)
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
