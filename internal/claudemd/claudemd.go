package claudemd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"ccm-desktop-v2/internal/config"
	"ccm-desktop-v2/internal/fsutil"
	"ccm-desktop-v2/internal/parser"
	"ccm-desktop-v2/internal/report"
)

type ClaudeMDInfo struct {
	Path       string
	Level      string // "user" or "project"
	Size       int64
	References []string
	PathRefs   []fsutil.PathRef
	Content    string
}

// FindAll discovers all CLAUDE.md files (user-level + project-level).
func FindAll(cfg *config.Config, extraDirs []string) []ClaudeMDInfo {
	var results []ClaudeMDInfo

	// User-level CLAUDE.md
	if fsutil.PathExists(cfg.ClaudeMD) {
		info := analyzeFile(cfg.ClaudeMD, "user")
		results = append(results, info)
	}

	// Project-level: check directories from claude.json projects
	state, err := cfg.LoadClaudeJSON()
	if err == nil {
		for projPath := range state.Projects {
			mdPath := filepath.Join(projPath, "CLAUDE.md")
			if fsutil.PathExists(mdPath) {
				results = append(results, analyzeFile(mdPath, "project"))
			}
		}
	}

	// Extra directories provided by user
	for _, dir := range extraDirs {
		mdPath := filepath.Join(dir, "CLAUDE.md")
		if fsutil.PathExists(mdPath) {
			results = append(results, analyzeFile(mdPath, "project"))
		}
	}

	return results
}

func analyzeFile(path, level string) ClaudeMDInfo {
	info := ClaudeMDInfo{Path: path, Level: level}

	data, err := os.ReadFile(path)
	if err != nil {
		return info
	}
	info.Content = string(data)

	if fi, err := os.Stat(path); err == nil {
		info.Size = fi.Size()
	}

	info.References = findAtReferences(info.Content)
	info.PathRefs = fsutil.FindAbsolutePaths(info.Content)

	return info
}

var atRefRe = regexp.MustCompile(`@(\S+)`)

func findAtReferences(content string) []string {
	var refs []string
	seen := map[string]bool{}
	for _, m := range atRefRe.FindAllStringSubmatch(content, -1) {
		if len(m) > 1 && !seen[m[1]] {
			seen[m[1]] = true
			refs = append(refs, m[1])
		}
	}
	return refs
}

// Validate checks a CLAUDE.md file and reports issues.
func Validate(cfg *config.Config, md ClaudeMDInfo) []report.Issue {
	var issues []report.Issue

	// Check @ references resolve
	dir := filepath.Dir(md.Path)
	for _, ref := range md.References {
		refPath := filepath.Join(dir, ref)
		if !fsutil.PathExists(refPath) {
			// Also try relative to .claude/
			refPath2 := filepath.Join(cfg.ClaudeDir, ref)
			if !fsutil.PathExists(refPath2) {
				issues = append(issues, report.Issue{
					Severity:      report.Warning,
					Domain:        "claudemd",
					Message:       fmt.Sprintf("Broken @-reference: @%s", ref),
					Detail:        fmt.Sprintf("File not found at %s or %s", refPath, refPath2),
					FixSuggestion: fmt.Sprintf("Create %s or remove the @%s reference", ref, ref),
				})
			}
		}
	}

	// Check for hardcoded D: drive paths
	for _, pr := range md.PathRefs {
		isPortable, reason := fsutil.IsPortable(pr.Normalized, cfg.HomeDir)
		if !isPortable && reason == "d-drive" {
			issues = append(issues, report.Issue{
				Severity: report.Warning,
				Domain:   "claudemd",
				Message:  fmt.Sprintf("Non-portable path: %s", pr.Raw),
				Detail:   "This path references D: drive which may not exist on other machines",
			})
		}
	}

	// Check for secrets in content
	secretFindings := parser.ScanForSecrets([]byte(md.Content))
	for _, f := range secretFindings {
		issues = append(issues, report.Issue{
			Severity: report.Error,
			Domain:   "claudemd",
			Message:  fmt.Sprintf("Potential secret found: %s at line %d", f.Pattern, f.Line),
			Detail:   fmt.Sprintf("Matched: %s", f.Redacted),
			FixSuggestion: "Remove hardcoded secrets. Use environment variables instead.",
		})
	}

	// Check frontmatter in project CLAUDE.md
	if md.Level == "project" && strings.HasPrefix(md.Content, "---") {
		if _, _, err := parser.ExtractFrontmatter([]byte(md.Content)); err != nil {
			issues = append(issues, report.Issue{
				Severity: report.Warning,
				Domain:   "claudemd",
				Message:  fmt.Sprintf("Invalid YAML frontmatter in %s: %v", filepath.Base(md.Path), err),
			})
		}
	}

	return issues
}

// CheckPortability analyzes CLAUDE.md files for cross-machine issues.
func CheckPortability(cfg *config.Config) []report.Issue {
	var issues []report.Issue
	all := FindAll(cfg, nil)

	for _, md := range all {
		for _, pr := range md.PathRefs {
			isPortable, reason := fsutil.IsPortable(pr.Normalized, cfg.HomeDir)
			if !isPortable {
				sev := fsutil.PortabilitySeverity(reason)
				issues = append(issues, report.Issue{
					Severity:      report.Severity(sev),
					Domain:        fmt.Sprintf("claudemd/%s", filepath.Base(md.Path)),
					Message:       fmt.Sprintf("Non-portable path: %s", pr.Raw),
					Detail:        fmt.Sprintf("Reason: %s", reason),
					FixSuggestion: "Consider using relative paths or environment variables",
				})
			}
		}
	}

	return issues
}
