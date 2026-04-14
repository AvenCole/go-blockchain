package cli

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"go-blockchain/internal/config"
)

func TestRunHelp(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	app := NewApp(config.Default(), &stdout, &stderr)
	code := app.Run([]string{"--help"})

	if code != 0 {
		t.Fatalf("Run(help) exit code = %d, want 0", code)
	}

	if !strings.Contains(stdout.String(), "Available commands:") {
		t.Fatalf("help output missing command list: %q", stdout.String())
	}
}

func TestRunVersion(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	app := NewApp(config.Default(), &stdout, &stderr)
	code := app.Run([]string{"version"})

	if code != 0 {
		t.Fatalf("Run(version) exit code = %d, want 0", code)
	}

	if !strings.Contains(stdout.String(), "go-blockchain version") {
		t.Fatalf("version output missing version banner: %q", stdout.String())
	}
}

func TestRunDoctor(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	app := NewApp(config.Default(), &stdout, &stderr)
	code := app.Run([]string{"doctor"})

	if code != 0 {
		t.Fatalf("Run(doctor) exit code = %d, want 0", code)
	}

	if !strings.Contains(stdout.String(), "next_step=implement merkle tree") {
		t.Fatalf("doctor output missing next step: %q", stdout.String())
	}
}

func TestCreateWalletAndListAddresses(t *testing.T) {
	cfg := config.Default()
	cfg.DataDir = filepath.Join(t.TempDir(), "data")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	app := NewApp(cfg, &stdout, &stderr)

	if code := app.Run([]string{"createwallet"}); code != 0 {
		t.Fatalf("createwallet exit code = %d, stderr=%q", code, stderr.String())
	}

	created := stdout.String()
	if !strings.Contains(created, "created wallet address=") {
		t.Fatalf("createwallet output = %q", created)
	}

	address := strings.TrimPrefix(strings.TrimSpace(created), "created wallet address=")
	stdout.Reset()
	stderr.Reset()

	if code := app.Run([]string{"listaddresses"}); code != 0 {
		t.Fatalf("listaddresses exit code = %d, stderr=%q", code, stderr.String())
	}

	if !strings.Contains(stdout.String(), address) {
		t.Fatalf("listaddresses output = %q, want %q", stdout.String(), address)
	}
}

func TestReindexUTXO(t *testing.T) {
	cfg := config.Default()
	cfg.DataDir = filepath.Join(t.TempDir(), "data")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	app := NewApp(cfg, &stdout, &stderr)

	if code := app.Run([]string{"createwallet"}); code != 0 {
		t.Fatalf("createwallet exit code = %d, stderr=%q", code, stderr.String())
	}
	address := strings.TrimPrefix(strings.TrimSpace(stdout.String()), "created wallet address=")
	stdout.Reset()
	stderr.Reset()

	if code := app.Run([]string{"createblockchain", address}); code != 0 {
		t.Fatalf("createblockchain exit code = %d, stderr=%q", code, stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	if code := app.Run([]string{"reindexutxo"}); code != 0 {
		t.Fatalf("reindexutxo exit code = %d, stderr=%q", code, stderr.String())
	}

	if !strings.Contains(stdout.String(), "utxo set reindexed") {
		t.Fatalf("reindexutxo output = %q", stdout.String())
	}
}

func TestCreateBlockchainAddBlockAndPrintChain(t *testing.T) {
	cfg := config.Default()
	cfg.DataDir = filepath.Join(t.TempDir(), "data")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	app := NewApp(cfg, &stdout, &stderr)

	if code := app.Run([]string{"createwallet"}); code != 0 {
		t.Fatalf("createwallet exit code = %d, stderr=%q", code, stderr.String())
	}
	address := strings.TrimPrefix(strings.TrimSpace(stdout.String()), "created wallet address=")
	stdout.Reset()
	stderr.Reset()

	if code := app.Run([]string{"createblockchain", address}); code != 0 {
		t.Fatalf("createblockchain exit code = %d, stderr=%q", code, stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	if code := app.Run([]string{"addblock", "block-1"}); code != 0 {
		t.Fatalf("addblock exit code = %d, stderr=%q", code, stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	if code := app.Run([]string{"printchain"}); code != 0 {
		t.Fatalf("printchain exit code = %d, stderr=%q", code, stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "Transactions: 1") {
		t.Fatalf("printchain output missing transaction count: %q", output)
	}
	if !strings.Contains(output, "Output: to="+address+" value=50") {
		t.Fatalf("printchain output missing debug coinbase output: %q", output)
	}
}

func TestRunSendAndGetBalance(t *testing.T) {
	cfg := config.Default()
	cfg.DataDir = filepath.Join(t.TempDir(), "data")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	app := NewApp(cfg, &stdout, &stderr)

	if code := app.Run([]string{"createwallet"}); code != 0 {
		t.Fatalf("createwallet miner exit code = %d, stderr=%q", code, stderr.String())
	}
	miner := strings.TrimPrefix(strings.TrimSpace(stdout.String()), "created wallet address=")
	stdout.Reset()
	stderr.Reset()

	if code := app.Run([]string{"createwallet"}); code != 0 {
		t.Fatalf("createwallet alice exit code = %d, stderr=%q", code, stderr.String())
	}
	alice := strings.TrimPrefix(strings.TrimSpace(stdout.String()), "created wallet address=")
	stdout.Reset()
	stderr.Reset()

	if code := app.Run([]string{"createblockchain", miner}); code != 0 {
		t.Fatalf("createblockchain exit code = %d, stderr=%q", code, stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	if code := app.Run([]string{"send", miner, alice, "20"}); code != 0 {
		t.Fatalf("send exit code = %d, stderr=%q", code, stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	if code := app.Run([]string{"getbalance", alice}); code != 0 {
		t.Fatalf("getbalance exit code = %d, stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "balance["+alice+"]=20") {
		t.Fatalf("getbalance output = %q, want alice balance", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	if code := app.Run([]string{"getbalance", miner}); code != 0 {
		t.Fatalf("getbalance miner exit code = %d, stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "balance["+miner+"]=30") {
		t.Fatalf("getbalance miner output = %q, want miner balance", stdout.String())
	}
}

func TestSendInsufficientFunds(t *testing.T) {
	cfg := config.Default()
	cfg.DataDir = filepath.Join(t.TempDir(), "data")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	app := NewApp(cfg, &stdout, &stderr)

	if code := app.Run([]string{"createwallet"}); code != 0 {
		t.Fatalf("createwallet miner exit code = %d, stderr=%q", code, stderr.String())
	}
	miner := strings.TrimPrefix(strings.TrimSpace(stdout.String()), "created wallet address=")
	stdout.Reset()
	stderr.Reset()

	if code := app.Run([]string{"createwallet"}); code != 0 {
		t.Fatalf("createwallet bob exit code = %d, stderr=%q", code, stderr.String())
	}
	bob := strings.TrimPrefix(strings.TrimSpace(stdout.String()), "created wallet address=")
	stdout.Reset()
	stderr.Reset()

	if code := app.Run([]string{"createblockchain", miner}); code != 0 {
		t.Fatalf("createblockchain exit code = %d, stderr=%q", code, stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	if code := app.Run([]string{"send", bob, miner, "60"}); code != 1 {
		t.Fatalf("send insufficient funds exit code = %d, want 1", code)
	}

	if !strings.Contains(stderr.String(), "insufficient funds") {
		t.Fatalf("stderr = %q, want insufficient funds", stderr.String())
	}
}

func TestRunUnknownCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	app := NewApp(config.Default(), &stdout, &stderr)
	code := app.Run([]string{"unknown"})

	if code != 1 {
		t.Fatalf("Run(unknown) exit code = %d, want 1", code)
	}

	if !strings.Contains(stderr.String(), "unknown command") {
		t.Fatalf("stderr missing unknown command error: %q", stderr.String())
	}
}
