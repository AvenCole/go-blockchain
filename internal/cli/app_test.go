package cli

import (
	"bytes"
	"os"
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
	if !strings.Contains(stdout.String(), "next_step=prepare final report and presentation") {
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
	if !strings.Contains(output, "ScriptPubKey: OP_DUP OP_HASH160") {
		t.Fatalf("printchain output missing ScriptPubKey: %q", output)
	}
	if !strings.Contains(output, "Output: to="+address+" value=50") {
		t.Fatalf("printchain output missing coinbase output: %q", output)
	}
	stdout.Reset()
	stderr.Reset()

	if code := app.Run([]string{"showevents", "5"}); code != 0 {
		t.Fatalf("showevents exit code = %d, stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "kind=main_block") {
		t.Fatalf("showevents output = %q, want main_block event", stdout.String())
	}
}

func TestRunShowScript(t *testing.T) {
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

	if code := app.Run([]string{"showscript", address}); code != 0 {
		t.Fatalf("showscript exit code = %d, stderr=%q", code, stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "scriptPubKey=OP_DUP OP_HASH160") {
		t.Fatalf("showscript output = %q, want P2PKH script", output)
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

func TestRunPerf(t *testing.T) {
	cfg := config.Default()
	cfg.DataDir = filepath.Join(t.TempDir(), "data")
	perfDir := filepath.Join(t.TempDir(), "perf")
	t.Setenv(perfOutputDirEnv, perfDir)

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
	stdout.Reset()
	stderr.Reset()

	if code := app.Run([]string{"mine", miner}); code != 0 {
		t.Fatalf("mine exit code = %d, stderr=%q", code, stderr.String())
	}
	stdout.Reset()
	stderr.Reset()

	if code := app.Run([]string{"runperf", "5"}); code != 0 {
		t.Fatalf("runperf exit code = %d, stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "speedup=") {
		t.Fatalf("runperf output = %q, want speedup", stdout.String())
	}
	if _, err := os.Stat(filepath.Join(perfDir, "latest.json")); err != nil {
		t.Fatalf("latest.json missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(perfDir, "latest.md")); err != nil {
		t.Fatalf("latest.md missing: %v", err)
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

func TestSimForkSwitchesToLongerBranch(t *testing.T) {
	cfg := config.Default()
	cfg.DataDir = filepath.Join(t.TempDir(), "data")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	app := NewApp(cfg, &stdout, &stderr)

	if code := app.Run([]string{"createwallet"}); code != 0 {
		t.Fatalf("createwallet exit code = %d, stderr=%q", code, stderr.String())
	}
	miner := strings.TrimPrefix(strings.TrimSpace(stdout.String()), "created wallet address=")
	stdout.Reset()
	stderr.Reset()

	if code := app.Run([]string{"createblockchain", miner}); code != 0 {
		t.Fatalf("createblockchain exit code = %d, stderr=%q", code, stderr.String())
	}
	stdout.Reset()
	stderr.Reset()

	if code := app.Run([]string{"addblock", "main-1"}); code != 0 {
		t.Fatalf("addblock main-1 exit code = %d, stderr=%q", code, stderr.String())
	}
	stdout.Reset()
	stderr.Reset()

	if code := app.Run([]string{"simfork", miner, "2"}); code != 0 {
		t.Fatalf("simfork exit code = %d, stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "switched=true") {
		t.Fatalf("simfork output = %q, want switched=true", stdout.String())
	}
}

func TestSimReorgRestoresTransactionToMempool(t *testing.T) {
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

	if code := app.Run([]string{"simreorg", miner, alice, "20", "1"}); code != 0 {
		t.Fatalf("simreorg exit code = %d, stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "restored=true") {
		t.Fatalf("simreorg output = %q, want restored=true", stdout.String())
	}
	if !strings.Contains(stdout.String(), "balance_after_reorg["+alice+"]=0") {
		t.Fatalf("simreorg output = %q, want alice balance reset", stdout.String())
	}
	stdout.Reset()
	stderr.Reset()

	if code := app.Run([]string{"showreorg"}); code != 0 {
		t.Fatalf("showreorg exit code = %d, stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "restored_tx=1") {
		t.Fatalf("showreorg output = %q, want restored_tx=1", stdout.String())
	}
	stdout.Reset()
	stderr.Reset()

	if code := app.Run([]string{"showevents", "3"}); code != 0 {
		t.Fatalf("showevents exit code = %d, stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "kind=reorg") {
		t.Fatalf("showevents output = %q, want reorg event", stdout.String())
	}
}
