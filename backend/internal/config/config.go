package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
)

type Config struct {
	HomeDir       string
	ClaudeDir     string
	ClaudeJSON    string
	ClaudeMD      string
	SettingsJSON  string
	SettingsLocal string
	SkillsDir     string
	ProjectsDir   string
	PluginsDir    string
}

// ClaudeJSONState represents the structure of ~/.claude.json
type ClaudeJSONState struct {
	Projects map[string]ClaudeProject `json:"projects"`
}

type ClaudeProject struct {
	MCPServers map[string]MCPServerConfig `json:"mcpServers"`
}

type MCPServerConfig struct {
	Type    string            `json:"type"`
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env"`
}

func Resolve(claudeDirOverride string) *Config {
	home := homeDir()
	claudeDir := filepath.Join(home, ".claude")
	if claudeDirOverride != "" {
		claudeDir = claudeDirOverride
	}

	return &Config{
		HomeDir:       home,
		ClaudeDir:     claudeDir,
		ClaudeJSON:    filepath.Join(home, ".claude.json"),
		ClaudeMD:      filepath.Join(claudeDir, "CLAUDE.md"),
		SettingsJSON:  filepath.Join(claudeDir, "settings.json"),
		SettingsLocal: filepath.Join(claudeDir, "settings.local.json"),
		SkillsDir:     filepath.Join(claudeDir, "skills"),
		ProjectsDir:   filepath.Join(claudeDir, "projects"),
		PluginsDir:    filepath.Join(claudeDir, "plugins"),
	}
}

func (c *Config) LoadClaudeJSON() (*ClaudeJSONState, error) {
	data, err := os.ReadFile(c.ClaudeJSON)
	if err != nil {
		return nil, err
	}
	var state ClaudeJSONState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	return &state, nil
}

func homeDir() string {
	if runtime.GOOS == "windows" {
		if d := os.Getenv("USERPROFILE"); d != "" {
			return d
		}
		return os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
	}
	return os.Getenv("HOME")
}
