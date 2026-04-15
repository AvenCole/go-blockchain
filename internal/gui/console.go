package gui

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go-blockchain/internal/cli"
)

func (s *Service) ExecuteCLI(commandLine string) (CommandResult, error) {
	args, err := splitCommandLine(commandLine)
	if err != nil {
		return CommandResult{}, err
	}
	if len(args) == 0 {
		return CommandResult{}, fmt.Errorf("命令不能为空")
	}

	switch args[0] {
	case "startnode":
		return s.executeStartNodeFromConsole(args)
	case "stopnode":
		return s.executeStopNodeFromConsole(args)
	case "connectnode":
		return s.executeConnectNodeFromConsole(args)
	case "nodeinit":
		return s.executeNodeInitFromConsole(args)
	case "nodesend":
		return s.executeNodeSendFromConsole(args)
	case "nodemine":
		return s.executeNodeMineFromConsole(args)
	case "runnetdemo":
		return s.executeRunNetworkDemoFromConsole(args)
	case "nodes":
		return s.executeNodesFromConsole(args)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	app := cli.NewApp(s.cfg, &stdout, &stderr)
	exitCode := app.Run(args)

	return CommandResult{
		Command:  commandLine,
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode,
	}, nil
}

func (s *Service) executeStartNodeFromConsole(args []string) (CommandResult, error) {
	if len(args) < 2 || len(args) > 4 {
		return CommandResult{}, fmt.Errorf("startnode requires: <addr> [seed] [miner]")
	}

	addr := args[1]
	seed := ""
	miner := ""
	if len(args) >= 3 {
		seed = args[2]
	}
	if len(args) == 4 {
		miner = args[3]
	}

	actualAddress, err := s.StartNode(addr, seed, miner)
	if err != nil {
		return CommandResult{}, err
	}

	return CommandResult{
		Command:  strings.Join(args, " "),
		Stdout:   fmt.Sprintf("node started: %s\n", actualAddress),
		ExitCode: 0,
	}, nil
}

func (s *Service) executeStopNodeFromConsole(args []string) (CommandResult, error) {
	if len(args) != 2 {
		return CommandResult{}, fmt.Errorf("stopnode requires: <addr>")
	}

	if err := s.StopNode(args[1]); err != nil {
		return CommandResult{}, err
	}

	return CommandResult{
		Command:  strings.Join(args, " "),
		Stdout:   fmt.Sprintf("node stopped: %s\n", args[1]),
		ExitCode: 0,
	}, nil
}

func (s *Service) executeConnectNodeFromConsole(args []string) (CommandResult, error) {
	if len(args) != 3 {
		return CommandResult{}, fmt.Errorf("connectnode requires: <addr> <seed>")
	}

	if err := s.ConnectNode(args[1], args[2]); err != nil {
		return CommandResult{}, err
	}

	return CommandResult{
		Command:  strings.Join(args, " "),
		Stdout:   fmt.Sprintf("node connected: %s -> %s\n", args[1], args[2]),
		ExitCode: 0,
	}, nil
}

func (s *Service) executeNodeInitFromConsole(args []string) (CommandResult, error) {
	if len(args) != 2 && len(args) != 3 {
		return CommandResult{}, fmt.Errorf("nodeinit requires: <node-addr> [reward-address]")
	}

	reward := ""
	if len(args) == 3 {
		reward = args[2]
	}
	if err := s.InitializeNodeBlockchain(args[1], reward); err != nil {
		return CommandResult{}, err
	}

	return CommandResult{
		Command:  strings.Join(args, " "),
		Stdout:   fmt.Sprintf("node chain ready: %s\n", args[1]),
		ExitCode: 0,
	}, nil
}

func (s *Service) executeNodeSendFromConsole(args []string) (CommandResult, error) {
	if len(args) != 5 && len(args) != 6 {
		return CommandResult{}, fmt.Errorf("nodesend requires: <node-addr> <from> <to> <amount> [fee]")
	}

	amount, err := strconv.Atoi(args[4])
	if err != nil {
		return CommandResult{}, fmt.Errorf("parse amount: %w", err)
	}
	fee := 0
	if len(args) == 6 {
		fee, err = strconv.Atoi(args[5])
		if err != nil {
			return CommandResult{}, fmt.Errorf("parse fee: %w", err)
		}
	}

	txid, err := s.SubmitNodeTransaction(args[1], args[2], args[3], amount, fee)
	if err != nil {
		return CommandResult{}, err
	}

	return CommandResult{
		Command:  strings.Join(args, " "),
		Stdout:   fmt.Sprintf("node transaction queued: %s\n", txid),
		ExitCode: 0,
	}, nil
}

func (s *Service) executeNodeMineFromConsole(args []string) (CommandResult, error) {
	if len(args) != 2 {
		return CommandResult{}, fmt.Errorf("nodemine requires: <node-addr>")
	}

	hash, err := s.MineNodePending(args[1])
	if err != nil {
		return CommandResult{}, err
	}

	return CommandResult{
		Command:  strings.Join(args, " "),
		Stdout:   fmt.Sprintf("node mined block: %s\n", hash),
		ExitCode: 0,
	}, nil
}

func (s *Service) executeRunNetworkDemoFromConsole(args []string) (CommandResult, error) {
	if len(args) != 1 {
		return CommandResult{}, fmt.Errorf("runnetdemo does not accept extra arguments")
	}

	result, err := s.RunNetworkQuickDemo()
	if err != nil {
		return CommandResult{}, err
	}

	return CommandResult{
		Command: strings.Join(args, " "),
		Stdout: fmt.Sprintf(
			"network demo ready\nsource=%s\npeer=%s\nminer=%s\nreceiver=%s\ntxid=%s\nblock=%s\npeerHeight=%d\ntipAnnounced=%t\n",
			result.SourceNode,
			result.PeerNode,
			result.MinerAddress,
			result.ReceiverAddress,
			result.TxID,
			result.BlockHash,
			result.PeerHeight,
			result.TipAnnounced,
		),
		ExitCode: 0,
	}, nil
}

func (s *Service) executeNodesFromConsole(args []string) (CommandResult, error) {
	if len(args) != 1 {
		return CommandResult{}, fmt.Errorf("nodes does not accept extra arguments")
	}

	nodes, err := s.Nodes()
	if err != nil {
		return CommandResult{}, err
	}
	if len(nodes) == 0 {
		return CommandResult{
			Command:  strings.Join(args, " "),
			Stdout:   "no GUI-managed nodes are running\n",
			ExitCode: 0,
		}, nil
	}

	var stdout strings.Builder
	for _, node := range nodes {
		fmt.Fprintf(
			&stdout,
			"address=%s initialized=%t height=%d mempool=%d miner=%s peers=%s\n",
			node.Address,
			node.Initialized,
			node.Height,
			node.MempoolCount,
			fallbackText(node.MinerAddress, "(none)"),
			strings.Join(node.Peers, ","),
		)
	}

	return CommandResult{
		Command:  strings.Join(args, " "),
		Stdout:   stdout.String(),
		ExitCode: 0,
	}, nil
}

func splitCommandLine(input string) ([]string, error) {
	var args []string
	var current strings.Builder
	inQuotes := false

	for _, r := range input {
		switch {
		case r == '"':
			inQuotes = !inQuotes
		case (r == ' ' || r == '\t') && !inQuotes:
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}
	}

	if inQuotes {
		return nil, fmt.Errorf("存在未闭合的引号")
	}
	if current.Len() > 0 {
		args = append(args, current.String())
	}
	return args, nil
}

func contextFromBackground() (context.Context, context.CancelFunc) {
	return context.WithCancel(context.Background())
}

func shortWait() {
	time.Sleep(200 * time.Millisecond)
}

func fallbackText(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
