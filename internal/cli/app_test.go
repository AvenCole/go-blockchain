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
	if !strings.Contains(stdout.String(), "next_step=implement gui system") {
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
	address := strings.TrimPrefix(strings.TrimSpace(stdout.String()), "created wallet address=")
	stdout.Reset()
	stderr.Reset()

	if code := app.Run([]string{"listaddresses"}); code != 0 {
		t.Fatalf("listaddresses exit code = %d, stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), address) {
		t.Fatalf("listaddresses output = %q, want %q", stdout.String(), address)
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
	if !strings.Contains(output, "MerkleRoot: ") {
		t.Fatalf("printchain output missing MerkleRoot: %q", output)
	}
	if !strings.Contains(output, "Difficulty: ") || !strings.Contains(output, "Nonce: ") || !strings.Contains(output, "PoWValid: true") {
		t.Fatalf("printchain output missing PoW fields: %q", output)
	}
	if !strings.Contains(output, "Output: to="+address+" value=50") {
		t.Fatalf("printchain output missing coinbase output: %q", output)
	}
}

func TestRunSendMineAndGetBalance(t *testing.T) {
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

	if code := app.Run([]string{"send", miner, alice, "20", "2"}); code != 0 {
		t.Fatalf("send exit code = %d, stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "queued transaction") {
		t.Fatalf("send output = %q", stdout.String())
	}
	stdout.Reset()
	stderr.Reset()

	if code := app.Run([]string{"printmempool"}); code != 0 {
		t.Fatalf("printmempool exit code = %d, stderr=%q", code, stderr.String())
	}
	if strings.TrimSpace(stdout.String()) == "" || strings.Contains(stdout.String(), "mempool empty") {
		t.Fatalf("printmempool output = %q, want pending tx", stdout.String())
	}
	stdout.Reset()
	stderr.Reset()

	if code := app.Run([]string{"mine", miner}); code != 0 {
		t.Fatalf("mine exit code = %d, stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "mined block") {
		t.Fatalf("mine output = %q", stdout.String())
	}
	stdout.Reset()
	stderr.Reset()

	if code := app.Run([]string{"printmempool"}); code != 0 {
		t.Fatalf("printmempool after mine exit code = %d, stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "mempool empty") {
		t.Fatalf("printmempool after mine output = %q, want empty", stdout.String())
	}
	stdout.Reset()
	stderr.Reset()

	if code := app.Run([]string{"getbalance", alice}); code != 0 {
		t.Fatalf("getbalance alice exit code = %d, stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "balance["+alice+"]=20") {
		t.Fatalf("getbalance alice output = %q", stdout.String())
	}
	stdout.Reset()
	stderr.Reset()

	if code := app.Run([]string{"getbalance", miner}); code != 0 {
		t.Fatalf("getbalance miner exit code = %d, stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "balance["+miner+"]=80") {
		t.Fatalf("getbalance miner output = %q", stdout.String())
	}
}

func TestMempoolStaysEmptyAfterReopenFollowingMine(t *testing.T) {
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

	if code := app.Run([]string{"send", miner, alice, "20", "1"}); code != 0 {
		t.Fatalf("send exit code = %d, stderr=%q", code, stderr.String())
	}
	stdout.Reset()
	stderr.Reset()

	if code := app.Run([]string{"mine", miner}); code != 0 {
		t.Fatalf("mine exit code = %d, stderr=%q", code, stderr.String())
	}
	stdout.Reset()
	stderr.Reset()

	reopened := NewApp(cfg, &stdout, &stderr)
	if code := reopened.Run([]string{"printmempool"}); code != 0 {
		t.Fatalf("printmempool reopen exit code = %d, stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "mempool empty") {
		t.Fatalf("printmempool reopen output = %q, want empty", stdout.String())
	}
	stdout.Reset()
	stderr.Reset()

	if code := reopened.Run([]string{"mine", miner}); code != 1 {
		t.Fatalf("mine reopen exit code = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "mempool is empty") {
		t.Fatalf("mine reopen stderr = %q, want empty mempool", stderr.String())
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

func TestSimDoubleSpendRejectsSecondTransaction(t *testing.T) {
	cfg := config.Default()
	cfg.DataDir = filepath.Join(t.TempDir(), "data")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	app := NewApp(cfg, &stdout, &stderr)

	if code := app.Run([]string{"createwallet"}); code != 0 {
		t.Fatalf("createwallet source exit code = %d, stderr=%q", code, stderr.String())
	}
	from := strings.TrimPrefix(strings.TrimSpace(stdout.String()), "created wallet address=")
	stdout.Reset()
	stderr.Reset()

	if code := app.Run([]string{"createwallet"}); code != 0 {
		t.Fatalf("createwallet to1 exit code = %d, stderr=%q", code, stderr.String())
	}
	to1 := strings.TrimPrefix(strings.TrimSpace(stdout.String()), "created wallet address=")
	stdout.Reset()
	stderr.Reset()

	if code := app.Run([]string{"createwallet"}); code != 0 {
		t.Fatalf("createwallet to2 exit code = %d, stderr=%q", code, stderr.String())
	}
	to2 := strings.TrimPrefix(strings.TrimSpace(stdout.String()), "created wallet address=")
	stdout.Reset()
	stderr.Reset()

	if code := app.Run([]string{"createblockchain", from}); code != 0 {
		t.Fatalf("createblockchain exit code = %d, stderr=%q", code, stderr.String())
	}
	stdout.Reset()
	stderr.Reset()

	if code := app.Run([]string{"simdouble", from, to1, to2, "20"}); code != 0 {
		t.Fatalf("simdouble exit code = %d, stderr=%q", code, stderr.String())
	}

	if !strings.Contains(stdout.String(), "rejected=true") {
		t.Fatalf("simdouble output = %q, want rejected second tx", stdout.String())
	}
}
