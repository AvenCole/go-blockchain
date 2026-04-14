package blockchain

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go-blockchain/internal/wallet"
)

// BalanceBenchmarkResult captures one cache-vs-scan comparison run.
type BalanceBenchmarkResult struct {
	GeneratedAt        string         `json:"generatedAt"`
	DataDir            string         `json:"dataDir"`
	Height             int            `json:"height"`
	AddressesTested    int            `json:"addressesTested"`
	LookupsPerAddress  int            `json:"lookupsPerAddress"`
	TotalLookups       int            `json:"totalLookups"`
	CachedDurationMs   float64        `json:"cachedDurationMs"`
	FullScanDurationMs float64        `json:"fullScanDurationMs"`
	Speedup            float64        `json:"speedup"`
	SampleBalances     map[string]int `json:"sampleBalances"`
}

// BalanceOfByScan computes the balance using a full chain scan without the UTXO cache.
func (bc *Blockchain) BalanceOfByScan(address string) (int, error) {
	pubKeyHash, err := wallet.PublicKeyHashFromAddress(address)
	if err != nil {
		return 0, err
	}

	utxos, err := bc.findUTXOByScan(pubKeyHash)
	if err != nil {
		return 0, err
	}

	total := 0
	for _, output := range utxos {
		total += output.Value
	}
	return total, nil
}

// RunBalanceBenchmark compares cached balance lookups against full-chain scans.
func (bc *Blockchain) RunBalanceBenchmark(addresses []string, lookupsPerAddress int) (BalanceBenchmarkResult, error) {
	if lookupsPerAddress <= 0 {
		lookupsPerAddress = 20
	}

	height, err := bc.Height()
	if err != nil {
		return BalanceBenchmarkResult{}, err
	}

	result := BalanceBenchmarkResult{
		GeneratedAt:       time.Now().UTC().Format(time.RFC3339),
		DataDir:           bc.dataDir,
		Height:            height,
		AddressesTested:   len(addresses),
		LookupsPerAddress: lookupsPerAddress,
		TotalLookups:      len(addresses) * lookupsPerAddress,
		SampleBalances:    make(map[string]int),
	}

	startCached := time.Now()
	for _, address := range addresses {
		for i := 0; i < lookupsPerAddress; i++ {
			balance, err := bc.BalanceOf(address)
			if err != nil {
				return BalanceBenchmarkResult{}, err
			}
			if i == 0 {
				result.SampleBalances[address] = balance
			}
		}
	}
	result.CachedDurationMs = float64(time.Since(startCached).Microseconds()) / 1000.0

	startScan := time.Now()
	for _, address := range addresses {
		for i := 0; i < lookupsPerAddress; i++ {
			if _, err := bc.BalanceOfByScan(address); err != nil {
				return BalanceBenchmarkResult{}, err
			}
		}
	}
	result.FullScanDurationMs = float64(time.Since(startScan).Microseconds()) / 1000.0

	if result.CachedDurationMs > 0 {
		result.Speedup = result.FullScanDurationMs / result.CachedDurationMs
	}

	return result, nil
}

// WriteBalanceBenchmark writes the benchmark result to JSON and Markdown files.
func WriteBalanceBenchmark(result BalanceBenchmarkResult, outputDir string) error {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return err
	}

	jsonPath := filepath.Join(outputDir, "latest.json")
	mdPath := filepath.Join(outputDir, "latest.md")

	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(jsonPath, jsonBytes, 0o644); err != nil {
		return err
	}

	var builder strings.Builder
	builder.WriteString("# Performance Benchmark\n\n")
	builder.WriteString(fmt.Sprintf("- GeneratedAt: %s\n", result.GeneratedAt))
	builder.WriteString(fmt.Sprintf("- Height: %d\n", result.Height))
	builder.WriteString(fmt.Sprintf("- AddressesTested: %d\n", result.AddressesTested))
	builder.WriteString(fmt.Sprintf("- LookupsPerAddress: %d\n", result.LookupsPerAddress))
	builder.WriteString(fmt.Sprintf("- TotalLookups: %d\n", result.TotalLookups))
	builder.WriteString(fmt.Sprintf("- CachedDurationMs: %.3f\n", result.CachedDurationMs))
	builder.WriteString(fmt.Sprintf("- FullScanDurationMs: %.3f\n", result.FullScanDurationMs))
	builder.WriteString(fmt.Sprintf("- Speedup: %.3fx\n", result.Speedup))
	builder.WriteString("\n## Sample Balances\n\n")
	for address, balance := range result.SampleBalances {
		builder.WriteString(fmt.Sprintf("- %s: %d\n", address, balance))
	}

	return os.WriteFile(mdPath, []byte(builder.String()), 0o644)
}

func (bc *Blockchain) findUTXOByScan(pubKeyHash []byte) ([]TXOutput, error) {
	blocks, err := bc.Blocks()
	if err != nil {
		return nil, err
	}

	spentTXOs := make(map[string]map[int]bool)
	var utxos []TXOutput

	for _, block := range blocks {
		for _, tx := range block.Transactions {
			txID := tx.IDHex()

			for outIdx, output := range tx.Outputs {
				if spentTXOs[txID] != nil && spentTXOs[txID][outIdx] {
					continue
				}
				if output.IsLockedWith(pubKeyHash) {
					utxos = append(utxos, output)
				}
			}

			if tx.IsCoinbase() {
				continue
			}

			for _, input := range tx.Inputs {
				if !input.UsesKey(pubKeyHash) {
					continue
				}
				inputTxID := input.TxIDHex()
				if spentTXOs[inputTxID] == nil {
					spentTXOs[inputTxID] = make(map[int]bool)
				}
				spentTXOs[inputTxID][input.Out] = true
			}
		}
	}

	return utxos, nil
}
