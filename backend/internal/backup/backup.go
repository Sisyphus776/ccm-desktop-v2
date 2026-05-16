package backup

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"ccm-desktop/internal/config"
	"ccm-desktop/internal/report"
)

var excludeDirs = map[string]bool{
	"sessions":        true,
	"telemetry":       true,
	"temp":            true,
	"cache":           true,
	"shell-snapshots": true,
	"session-env":     true,
	"file-history":    true,
	"tasks":           true,
	"paste-cache":     true,
	"ide":             true,
}

type Manifest struct {
	Version      string    `json:"version"`
	Timestamp    time.Time `json:"timestamp"`
	SourceHost   string    `json:"source_host"`
	SourceUser   string    `json:"source_user"`
	CLAUDEmd     bool      `json:"has_claudemd"`
	SkillsCount  int       `json:"skills_count"`
	MemoryCount  int       `json:"memory_count"`
	MCPServers   int       `json:"mcp_servers"`
	SettingsSize int64     `json:"settings_size"`
}

// CreateBackup creates a zip backup of the .claude configuration directory.
func CreateBackup(cfg *config.Config, outputPath string) error {
	if outputPath == "" {
		outputPath = filepath.Join(cfg.HomeDir, fmt.Sprintf("ccm-backup-%s.zip", time.Now().Format("20060102-150405")))
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create backup file: %w", err)
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	defer zw.Close()

	manifest := Manifest{
		Version:    "1.0",
		Timestamp:  time.Now(),
		SourceHost: hostname(),
		SourceUser: os.Getenv("USERNAME"),
	}

	err = filepath.Walk(cfg.ClaudeDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip files we can't read
		}

		relPath, _ := filepath.Rel(cfg.ClaudeDir, path)
		parts := strings.Split(relPath, string(filepath.Separator))

		// Skip excluded directories
		if len(parts) > 0 && excludeDirs[parts[0]] {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Track manifest stats
		updateManifest(&manifest, relPath, info)

		if info.IsDir() {
			return nil
		}

		w, err := zw.Create(relPath)
		if err != nil {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		w.Write(data)

		return nil
	})

	if err != nil {
		return fmt.Errorf("backup failed: %w", err)
	}

	// Write manifest
	mw, _ := zw.Create("backup-manifest.json")
	manifestData, _ := json.MarshalIndent(manifest, "", "  ")
	mw.Write(manifestData)

	report.InfoLine(fmt.Sprintf("Backup created: %s", outputPath))
	report.InfoLine(fmt.Sprintf("  CLAUDE.md: %v, Skills: %d, Memories: %d, MCP servers: %d",
		manifest.CLAUDEmd, manifest.SkillsCount, manifest.MemoryCount, manifest.MCPServers))

	return nil
}

// RestoreBackup restores from a zip backup.
func RestoreBackup(cfg *config.Config, zipPath string, force bool) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("failed to open backup: %w", err)
	}
	defer r.Close()

	for _, f := range r.File {
		if f.Name == "backup-manifest.json" {
			continue
		}

		targetPath := filepath.Join(cfg.ClaudeDir, f.Name)

		// Check if file exists and we're not forcing
		if !force {
			if _, err := os.Stat(targetPath); err == nil {
				report.InfoLine(fmt.Sprintf("Skipping %s (already exists, use --force to overwrite)", f.Name))
				continue
			}
		}

		// Create directory if needed
		dir := filepath.Dir(targetPath)
		os.MkdirAll(dir, 0755)

		// Extract file
		rc, err := f.Open()
		if err != nil {
			continue
		}

		out, err := os.Create(targetPath)
		if err != nil {
			rc.Close()
			continue
		}

		io.Copy(out, rc)
		out.Close()
		rc.Close()

		report.InfoLine(fmt.Sprintf("Restored: %s", f.Name))
	}

	report.InfoLine("Restore complete.")
	return nil
}

func hostname() string {
	h, _ := os.Hostname()
	return h
}

func updateManifest(m *Manifest, relPath string, info os.FileInfo) {
	switch {
	case info.IsDir():
		return
	case relPath == "CLAUDE.md":
		m.CLAUDEmd = true
	case strings.HasPrefix(relPath, "skills/") && strings.HasSuffix(relPath, "SKILL.md"):
		m.SkillsCount++
	case strings.Contains(relPath, "/memory/") && strings.HasSuffix(relPath, ".md") && relPath != "MEMORY.md":
		m.MemoryCount++
		m.SettingsSize += info.Size()
	}
}
