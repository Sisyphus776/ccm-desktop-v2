package portability

import (
	"fmt"

	"ccm-desktop/internal/claudemd"
	"ccm-desktop/internal/config"
	"ccm-desktop/internal/mcp"
	"ccm-desktop/internal/report"
)

// Run performs a full cross-computer portability analysis.
func Run(cfg *config.Config) {
	report.Header("Portability Analysis")

	var allIssues []report.Issue

	// 1. CLAUDE.md portability
	mdIssues := claudemd.CheckPortability(cfg)
	allIssues = append(allIssues, mdIssues...)

	// 2. MCP portability
	mcpIssues := mcp.CheckPortability(cfg)
	allIssues = append(allIssues, mcpIssues...)

	// 3. Settings files - scan for hardcoded paths
	settingIssues := scanSettingsPaths(cfg)
	allIssues = append(allIssues, settingIssues...)

	// Print issues by severity
	critical := filterBySeverity(allIssues, "critical")
	warnings := filterBySeverity(allIssues, "warning")
	info := filterBySeverity(allIssues, "info")

	report.InfoLine(fmt.Sprintf("Found %d path references across all config files.", len(allIssues)))
	report.InfoLine(fmt.Sprintf("  %d CRITICAL - will break on other computers", len(critical)))
	report.InfoLine(fmt.Sprintf("  %d WARNING  - may break depending on setup", len(warnings)))
	report.InfoLine(fmt.Sprintf("  %d INFO     - informational", len(info)))

	if len(critical) > 0 {
		report.SubHeader("CRITICAL Issues")
		for _, issue := range critical {
			report.PrintIssue(issue)
		}
	}

	if len(warnings) > 0 {
		report.SubHeader("WARNING Issues")
		for _, issue := range warnings {
			report.PrintIssue(issue)
		}
	}

	// Migration checklist
	report.SubHeader("Migration Checklist")
	checklist := generateChecklist(allIssues)
	for i, item := range checklist {
		report.InfoLine(fmt.Sprintf("  [%d] %s", i+1, item))
	}
}

func filterBySeverity(issues []report.Issue, severity string) []report.Issue {
	var result []report.Issue
	for _, issue := range issues {
		if string(issue.Severity) == severity {
			result = append(result, issue)
		}
	}
	return result
}

func scanSettingsPaths(cfg *config.Config) []report.Issue {
	var issues []report.Issue

	// Use fsutil to scan settings files
	// This is handled within claudemd for the path detection
	// For settings.json specifically, we check env var paths
	_ = cfg // placeholder - detailed impl reads settings.json for env var values with paths

	return issues
}

func generateChecklist(issues []report.Issue) []string {
	var checklist []string
	seen := map[string]bool{}

	for _, issue := range issues {
		if issue.FixSuggestion != "" && !seen[issue.FixSuggestion] {
			seen[issue.FixSuggestion] = true
			checklist = append(checklist, issue.FixSuggestion)
		}
	}

	if len(checklist) == 0 {
		checklist = append(checklist, "No specific actions needed - config appears portable")
	}

	// Add standard items
	standardItems := []string{
		"Ensure Claude Code is installed on target machine",
		"Copy .claude/ directory to target machine",
		"Verify MCP dependencies (Python, Node.js, etc.) are installed",
	}
	checklist = append(standardItems, checklist...)

	return checklist
}
