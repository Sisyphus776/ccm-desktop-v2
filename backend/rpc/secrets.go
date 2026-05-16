package rpc

import (
	"os"
	"path/filepath"

	"ccm-desktop-v2/backend/internal/parser"
)

type SecretItem struct {
	Pattern  string `json:"pattern"`
	Line     int    `json:"line"`
	Match    string `json:"match"`
	FilePath string `json:"filePath"`
}

func scanSecrets(ctx *AppContext) (any, error) {
	var items []SecretItem
	for _, path := range []string{ctx.Cfg.SettingsJSON, ctx.Cfg.SettingsLocal, ctx.Cfg.ClaudeMD} {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		for _, r := range parser.ScanForSecrets(data) {
			items = append(items, SecretItem{
				Pattern:  r.Pattern,
				Line:     r.Line,
				Match:    r.Redacted,
				FilePath: filepath.Base(path),
			})
		}
	}
	return items, nil
}
