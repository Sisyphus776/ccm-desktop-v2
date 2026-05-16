package rpc

import (
	"fmt"
	"path/filepath"
	"time"

	"ccm-desktop-v2/backend/internal/backup"
)

func createBackup(ctx *AppContext, outputPath string) (any, error) {
	if outputPath == "" {
		outputPath = filepath.Join(ctx.Cfg.HomeDir, "ccm-backup-"+time.Now().Format("20060102-150405")+".zip")
	}
	if err := backup.CreateBackup(ctx.Cfg, outputPath); err != nil {
		return nil, fmt.Errorf("backup failed: %w", err)
	}
	return outputPath, nil
}

func restoreBackup(ctx *AppContext, zipPath string, force bool) (any, error) {
	if err := backup.RestoreBackup(ctx.Cfg, zipPath, force); err != nil {
		return nil, fmt.Errorf("restore failed: %w", err)
	}
	return "Restore completed", nil
}
