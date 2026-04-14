package cli

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"go-blockchain/internal/blockchain"
	"go-blockchain/internal/config"
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
	fmt.Fprintf(a.stdout, "%s initialization CLI\n\n", a.cfg.ProjectName)
	fmt.Fprintln(a.stdout, "Usage:")
	fmt.Fprintf(a.stdout, "  %s [command]\n", a.cfg.ProjectName)
	fmt.Fprintln(a.stdout)
	fmt.Fprintln(a.stdout, "Available commands:")
	fmt.Fprintln(a.stdout, "  help      Show command help")
	fmt.Fprintln(a.stdout, "  version   Show project version information")
	fmt.Fprintln(a.stdout, "  about     Show project version information")
	fmt.Fprintln(a.stdout, "  doctor    Check initialization-stage readiness")
	fmt.Fprintln(a.stdout, "  createblockchain [genesis-data]  Create a new blockchain")
	fmt.Fprintln(a.stdout, "  addblock <data>                  Append a block to the current chain")
	fmt.Fprintln(a.stdout, "  printchain                       Print the blockchain from tip to genesis")
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

	fmt.Fprintf(a.stdout, "doctor: blockchain skeleton is ready\n")
	fmt.Fprintf(a.stdout, "project=%s\n", a.cfg.ProjectName)
	fmt.Fprintf(a.stdout, "data_dir=%s\n", a.cfg.DataDir)
	fmt.Fprintf(a.stdout, "default_port=%d\n", a.cfg.DefaultPort)
	fmt.Fprintf(a.stdout, "log_level=%s\n", a.cfg.LogLevel)
	fmt.Fprintf(a.stdout, "network_mode=%s\n", a.cfg.NetworkMode)
	fmt.Fprintf(a.stdout, "chain_status=%s\n", chainStatus)
	fmt.Fprintln(a.stdout, "next_step=implement transaction model")
}

func (a App) createBlockchain(args []string) int {
	genesisData := "Genesis Block"
	if len(args) > 0 {
		genesisData = strings.Join(args, " ")
	}

	chain, err := blockchain.CreateBlockchain(a.cfg.DataDir, genesisData)
	if err != nil {
		if errors.Is(err, blockchain.ErrBlockchainAlreadyExists) {
			fmt.Fprintln(a.stderr, "blockchain already exists")
			return 1
		}

		fmt.Fprintf(a.stderr, "create blockchain: %v\n", err)
		return 1
	}
	defer chain.Close()

	fmt.Fprintf(a.stdout, "created blockchain with genesis block: %s\n", genesisData)
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

	block, err := chain.AddBlock(strings.Join(args, " "))
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
		fmt.Fprintf(a.stdout, "Data: %s\n", string(block.Data))
		fmt.Fprintf(a.stdout, "Hash: %s\n", block.HashHex())
		fmt.Fprintf(a.stdout, "PrevHash: %s\n", block.PrevHashHex())
		fmt.Fprintln(a.stdout)
	}

	return 0
}
