package gui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"go-blockchain/internal/config"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const guiDataDirEnv = "GO_BLOCKCHAIN_GUI_DATA_DIR"

type Service struct {
	ctx    context.Context
	cfg    config.Config
	mu     sync.Mutex
	nodeMu sync.Mutex
	nodes  map[string]*nodeSession
	emit   func(ctx context.Context, eventName string, optionalData ...interface{})
}

func NewService() *Service {
	cfg := config.Default()
	cfg.DataDir = resolveGUIDataDir(cfg.DataDir)

	return &Service{
		cfg:   cfg,
		nodes: make(map[string]*nodeSession),
		emit: func(ctx context.Context, eventName string, optionalData ...interface{}) {
			if ctx == nil {
				return
			}
			runtime.EventsEmit(ctx, eventName, optionalData...)
		},
	}
}

func (s *Service) Startup(ctx context.Context) {
	s.ctx = ctx
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
