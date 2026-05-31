package rpc

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"ccm-desktop-v2/internal/mcp"
)

type MCPItem struct {
	Name     string   `json:"name"`
	Project  string   `json:"project"`
	Command  string   `json:"command"`
	Args     []string `json:"args"`
	Status   string   `json:"status"`
	Disabled bool     `json:"disabled"`
}

func getDisabledMCPPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".claude", "ccm-disabled-mcp.json")
}

func loadDisabledMCP() map[string]bool {
	disabled := map[string]bool{}
	data, err := os.ReadFile(getDisabledMCPPath())
	if err != nil {
		return disabled
	}
	json.Unmarshal(data, &disabled)
	return disabled
}

func saveDisabledMCP(disabled map[string]bool) {
	data, _ := json.MarshalIndent(disabled, "", "  ")
	os.WriteFile(getDisabledMCPPath(), data, 0644)
}

func commandExists(cmd string) bool {
	if cmd == "" {
		return false
	}
	_, err := exec.LookPath(cmd)
	return err == nil
}

func listMCP(ctx *AppContext) (any, error) {
	servers, _ := mcp.List(ctx.Cfg)
	disabled := loadDisabledMCP()
	var items []MCPItem
	for _, s := range servers {
		status := "ok"
		if s.Config.Command == "" {
			status = "warning"
		} else if !commandExists(s.Config.Command) {
			status = "missing"
		}
		items = append(items, MCPItem{
			Name:     s.Name,
			Project:  s.Project,
			Command:  s.Config.Command,
			Args:     s.Config.Args,
			Status:   status,
			Disabled: disabled[s.Name],
		})
	}
	return items, nil
}

func validateMCP(ctx *AppContext) (any, error) {
	var items []IssueItem
	for _, iss := range mcp.Validate(ctx.Cfg) {
		items = append(items, issueToItem(iss))
	}
	return items, nil
}

func toggleMCP(ctx *AppContext, name string) (any, error) {
	disabled := loadDisabledMCP()
	if disabled[name] {
		delete(disabled, name)
		saveDisabledMCP(disabled)
		return fmt.Sprintf("已启用: %s", name), nil
	}
	disabled[name] = true
	saveDisabledMCP(disabled)
	return fmt.Sprintf("已禁用: %s", name), nil
}
