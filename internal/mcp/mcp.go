package mcp

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"ccm-desktop-v2/internal/config"
	"ccm-desktop-v2/internal/fsutil"
	"ccm-desktop-v2/internal/report"
)

type MCPServer struct {
	Name    string
	Project string
	Config  config.MCPServerConfig
}

// settingsFile is a partial struct for reading mcpServers from settings.json
type settingsFile struct {
	MCPServers map[string]config.MCPServerConfig `json:"mcpServers"`
}

// List discovers all MCP server configurations from all known sources.
func List(cfg *config.Config) ([]MCPServer, error) {
	var servers []MCPServer
	seen := map[string]bool{}

	addServer := func(name, project string, srv config.MCPServerConfig) {
		key := project + ":" + name
		if seen[key] {
			return
		}
		seen[key] = true
		servers = append(servers, MCPServer{Name: name, Project: project, Config: srv})
	}

	// 1. ~/.claude.json — project-scoped MCP servers
	if state, err := cfg.LoadClaudeJSON(); err == nil {
		for projPath, proj := range state.Projects {
			for name, srv := range proj.MCPServers {
				addServer(name, projPath, srv)
			}
		}
	}

	// 2. ~/.claude/settings.json — user-level MCP servers
	loadMCPFromSettings := func(path string, project string) {
		data, err := os.ReadFile(path)
		if err != nil {
			return
		}
		var sf settingsFile
		if err := json.Unmarshal(data, &sf); err != nil {
			return
		}
		for name, srv := range sf.MCPServers {
			addServer(name, project, srv)
		}
	}

	loadMCPFromSettings(cfg.SettingsJSON, "user")
	// 3. ~/.claude/settings.local.json — local overrides (take precedence)
	loadMCPFromSettings(cfg.SettingsLocal, "user-local")

	return servers, nil
}

// Validate checks MCP server configs and reports issues.
func Validate(cfg *config.Config) []report.Issue {
	var issues []report.Issue

	servers, err := List(cfg)
	if err != nil {
		issues = append(issues, report.Issue{
			Severity: report.Error,
			Domain:   "mcp",
			Message:  fmt.Sprintf("Failed to parse MCP configs: %v", err),
		})
		return issues
	}

	for _, srv := range servers {
		domain := fmt.Sprintf("mcp/%s", srv.Name)

		// Check type
		if srv.Config.Type != "stdio" {
			issues = append(issues, report.Issue{
				Severity: report.Warning,
				Domain:   domain,
				Message:  fmt.Sprintf("Unknown MCP server type: %s", srv.Config.Type),
			})
		}

		// Check command exists
		cmd := srv.Config.Command
		if filepath.IsAbs(cmd) {
			if !fsutil.PathExists(cmd) {
				issues = append(issues, report.Issue{
					Severity: report.Error,
					Domain:   domain,
					Message:  fmt.Sprintf("MCP command not found: %s", cmd),
					FixSuggestion: "Install the required binary or update the path in claude.json",
				})
			}
		} else {
			if _, err := exec.LookPath(cmd); err != nil {
				issues = append(issues, report.Issue{
					Severity: report.Warning,
					Domain:   domain,
					Message:  fmt.Sprintf("MCP command not in PATH: %s", cmd),
					FixSuggestion: "Install the required binary and ensure it's in PATH",
				})
			}
		}

		// Check args for file existence (Python scripts etc.)
		for _, arg := range srv.Config.Args {
			if filepath.IsAbs(arg) || filepath.Ext(arg) == ".py" || filepath.Ext(arg) == ".js" {
				checkPath := arg
				if !filepath.IsAbs(checkPath) {
					// Relative - check from home dir
					checkPath = filepath.Join(cfg.HomeDir, checkPath)
				}
				if !fsutil.PathExists(checkPath) {
					issues = append(issues, report.Issue{
						Severity: report.Error,
						Domain:   domain,
						Message:  fmt.Sprintf("MCP script not found: %s", arg),
						FixSuggestion: "Ensure the script file exists at the specified path",
					})
				}
			}
		}
	}

	return issues
}

// CheckPortability analyzes MCP configs for cross-machine issues.
func CheckPortability(cfg *config.Config) []report.Issue {
	var issues []report.Issue

	servers, err := List(cfg)
	if err != nil {
		return issues
	}

	for _, srv := range servers {
		domain := fmt.Sprintf("mcp/%s", srv.Name)

		// Check command path
		cmd := srv.Config.Command
		if filepath.IsAbs(cmd) {
			isPortable, reason := fsutil.IsPortable(cmd, cfg.HomeDir)
			if !isPortable {
				sev := fsutil.PortabilitySeverity(reason)
				issues = append(issues, report.Issue{
					Severity:      report.Severity(sev),
					Domain:        domain,
					Message:       fmt.Sprintf("Non-portable MCP command path: %s", cmd),
					Detail:        fmt.Sprintf("Reason: %s", reason),
					FixSuggestion: "Use command name from PATH instead of absolute path",
				})
			}
		}

		// Check arg paths
		for _, arg := range srv.Config.Args {
			if filepath.IsAbs(arg) {
				isPortable, reason := fsutil.IsPortable(arg, cfg.HomeDir)
				if !isPortable {
					sev := fsutil.PortabilitySeverity(reason)
					issues = append(issues, report.Issue{
						Severity:      report.Severity(sev),
						Domain:        domain,
						Message:       fmt.Sprintf("Non-portable MCP argument path: %s", arg),
						Detail:        fmt.Sprintf("Reason: %s", reason),
						FixSuggestion: "Use a relative or env-var-based path",
					})
				}
			}
		}
	}

	return issues
}
