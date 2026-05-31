package settings

import (
	"encoding/json"
	"fmt"
	"os"

	"ccm-desktop-v2/internal/config"
	"ccm-desktop-v2/internal/parser"
	"ccm-desktop-v2/internal/report"
)

type SettingsInfo struct {
	Path    string
	Content []byte
	Parsed  interface{}
	Valid   bool
	Err     error
}

// Load reads and parses a settings JSON file.
func Load(path string) (*SettingsInfo, error) {
	info := &SettingsInfo{Path: path}

	data, err := os.ReadFile(path)
	if err != nil {
		info.Err = err
		return info, err
	}
	info.Content = data

	// Try to parse as JSON
	var parsed interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		info.Err = err
		return info, err
	}
	info.Parsed = parsed
	info.Valid = true
	return info, nil
}

// Validate checks settings files and reports issues.
func Validate(cfg *config.Config) []report.Issue {
	var issues []report.Issue

	for _, path := range []string{cfg.SettingsJSON, cfg.SettingsLocal} {
		info, err := Load(path)
		if err != nil {
			issues = append(issues, report.Issue{
				Severity: report.Error,
				Domain:   "settings",
				Message:  fmt.Sprintf("Failed to load %s: %v", path, err),
			})
			continue
		}

		if !info.Valid {
			issues = append(issues, report.Issue{
				Severity: report.Error,
				Domain:   "settings",
				Message:  fmt.Sprintf("Invalid JSON in %s: %v", path, info.Err),
			})
			continue
		}

		// Scan for secrets
		findings := parser.ScanForSecrets(info.Content)
		for _, f := range findings {
			issues = append(issues, report.Issue{
				Severity: report.Warning,
				Domain:   "settings",
				Message:  fmt.Sprintf("Potential secret found in %s: %s", path, f.Pattern),
				Detail:   fmt.Sprintf("Matched: %s", f.Redacted),
				FixSuggestion: "Use credential manager or environment variables instead of hardcoded keys",
			})
		}
	}

	return issues
}
