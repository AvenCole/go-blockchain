package gui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"go-blockchain/internal/blockchain"
	"go-blockchain/internal/config"
)

const guiDataDirEnv = "GO_BLOCKCHAIN_GUI_DATA_DIR"

type Service struct {
	ctx   context.Context
	cfg   config.Config
	mu    sync.Mutex
	chain *blockchain.Blockchain
}

func NewService() *Service {
	cfg := config.Default()
	cfg.DataDir = resolveGUIDataDir(cfg.DataDir)

	return &Service{
		cfg: cfg,
	}
}

func (s *Service) Startup(ctx context.Context) {
	s.ctx = ctx
}

func (s *Service) ensureChain() (*blockchain.Blockchain, error) {
	if s.chain != nil {
		return s.chain, nil
	}

	chain, err := blockchain.OpenBlockchain(s.cfg.DataDir)
	if err != nil {
		return nil, normalizeStorageError(err)
	}
	s.chain = chain
	return s.chain, nil
}

func resolveGUIDataDir(baseDataDir string) string {
	if override := os.Getenv(guiDataDirEnv); override != "" {
		return override
	}
	return filepath.Join(baseDataDir, "gui-desktop")
}

func normalizeStorageError(err error) error {
	if err == nil {
		return nil
	}
	if strings.Contains(err.Error(), "used by another process") || strings.Contains(err.Error(), "being used by another process") {
		return fmt.Errorf("GUI data directory is already in use by another process; close the other GUI instance or set %s", guiDataDirEnv)
	}
	return err
}
