package rpc

import (
	"fmt"
	"os"
	"strings"

	"ccm-desktop-v2/internal/claudemd"
	"ccm-desktop-v2/internal/mcp"
	"ccm-desktop-v2/internal/report"
)

type PortabilityResult struct {
	Issues   []PortabilityIssue `json:"issues"`
	Critical int                `json:"critical"`
	Warning  int                `json:"warning"`
	Info     int                `json:"info"`
}

type PortabilityIssue struct {
	Severity string `json:"severity"`
	Domain   string `json:"domain"`
	Message  string `json:"message"`
	Fix      string `json:"fix"`
}

func getPortabilityReport(ctx *AppContext) (any, error) {
	var items []PortabilityIssue
	crit, warn, info := 0, 0, 0

	add := func(iss []report.Issue) {
		for _, i := range iss {
			items = append(items, PortabilityIssue{
				Severity: string(i.Severity),
				Domain:   i.Domain,
				Message:  i.Message,
				Fix:      i.FixSuggestion,
			})
			switch i.Severity {
			case "critical":
				crit++
			case "warning":
				warn++
			case "info":
				info++
			}
		}
	}

	add(mcp.CheckPortability(ctx.Cfg))
	add(claudemd.CheckPortability(ctx.Cfg))

	return PortabilityResult{Issues: items, Critical: crit, Warning: warn, Info: info}, nil
}

func fixPath(ctx *AppContext, h *Handler, oldPath, newPath string) (any, error) {
	if oldPath == "" || newPath == "" {
		return nil, fmt.Errorf("路径不能为空")
	}
	count := 0
	files := []string{ctx.Cfg.SettingsJSON, ctx.Cfg.SettingsLocal, ctx.Cfg.ClaudeMD}
	for _, fp := range files {
		if data, err := os.ReadFile(fp); err == nil {
			newData := strings.ReplaceAll(string(data), oldPath, newPath)
			if string(data) != newData {
				os.WriteFile(fp, []byte(newData), 0644)
				count++
			}
		}
	}
	if count > 0 {
		h.Notify("config-changed", map[string]any{"domain": "settings"})
		return fmt.Sprintf("已在 %d 个文件中将 %s 替换为 %s", count, oldPath, newPath), nil
	}
	return "未找到需要替换的路径", nil
}
