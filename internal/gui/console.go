package gui

import (
	"bytes"
	"context"
	"fmt"
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
		fmt.Fprintf(&stdout, "address=%s height=%d miner=%s peers=%s\n", node.Address, node.Height, fallbackText(node.MinerAddress, "(none)"), strings.Join(node.Peers, ","))
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
