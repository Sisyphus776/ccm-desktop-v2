package rpc

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"ccm-desktop-v2/backend/internal/mcp"
	"ccm-desktop-v2/backend/internal/memory"
	"ccm-desktop-v2/backend/internal/report"
	"ccm-desktop-v2/backend/internal/skills"
)

type AppSettings struct {
	ClaudeDir      string `json:"claudeDir"`
	HomeDir        string `json:"homeDir"`
	AutoStart      bool   `json:"autoStart"`
	StartMinimized bool   `json:"startMinimized"`
}

type DashboardSummary struct {
	SkillsCount  int `json:"skillsCount"`
	MemoryCount  int `json:"memoryCount"`
	MCPServers   int `json:"mcpServers"`
	ErrorCount   int `json:"errorCount"`
	WarningCount int `json:"warningCount"`
}

func checkAutoStart() bool {
	exePath, err := os.Executable()
	if err != nil {
		return false
	}
	out, err := exec.Command("reg", "query",
		`HKCU\Software\Microsoft\Windows\CurrentVersion\Run`,
		"/v", "CCM").Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(out), exePath)
}

func enableAutoStart() error {
	exePath, err := os.Executable()
	if err != nil {
		return err
	}
	return exec.Command("reg", "add",
		`HKCU\Software\Microsoft\Windows\CurrentVersion\Run`,
		"/v", "CCM", "/t", "REG_SZ", "/d", exePath, "/f").Run()
}

func disableAutoStart() error {
	return exec.Command("reg", "delete",
		`HKCU\Software\Microsoft\Windows\CurrentVersion\Run`,
		"/v", "CCM", "/f").Run()
}

func getSettings(ctx *AppContext) (any, error) {
	return AppSettings{
		ClaudeDir:      ctx.Cfg.ClaudeDir,
		HomeDir:        ctx.Cfg.HomeDir,
		AutoStart:      checkAutoStart(),
		StartMinimized: true,
	}, nil
}

func setAutoStart(ctx *AppContext, enabled bool) (any, error) {
	if enabled {
		if err := enableAutoStart(); err != nil {
			return nil, fmt.Errorf("failed to enable: %w", err)
		}
		return "Auto-start enabled", nil
	}
	if err := disableAutoStart(); err != nil {
		return nil, fmt.Errorf("failed to disable: %w", err)
	}
	return "Auto-start disabled", nil
}

func getDashboardSummary(ctx *AppContext) (any, error) {
	skillList, _ := skills.List(ctx.Cfg)
	mems, _ := memory.ListAll(ctx.Cfg)
	mcpList, _ := mcp.List(ctx.Cfg)

	errCount, warnCount := 0, 0
	for _, iss := range skills.Validate(ctx.Cfg) {
		if iss.Severity == report.Error {
			errCount++
		} else if iss.Severity == report.Warning {
			warnCount++
		}
	}
	for _, iss := range memory.Validate(ctx.Cfg) {
		if iss.Severity == report.Error {
			errCount++
		} else if iss.Severity == report.Warning {
			warnCount++
		}
	}
	for _, iss := range mcp.Validate(ctx.Cfg) {
		if iss.Severity == report.Error {
			errCount++
		} else if iss.Severity == report.Warning {
			warnCount++
		}
	}

	return DashboardSummary{
		SkillsCount:  len(skillList),
		MemoryCount:  len(mems),
		MCPServers:   len(mcpList),
		ErrorCount:   errCount,
		WarningCount: warnCount,
	}, nil
}
