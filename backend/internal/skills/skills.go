package skills

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"ccm-desktop/internal/config"
	"ccm-desktop/internal/fsutil"
	"ccm-desktop/internal/parser"
	"ccm-desktop/internal/report"
)

type SkillType string

const (
	SkillTypeDir          SkillType = "directory"
	SkillTypeSymlink      SkillType = "symlink"
	SkillTypeStandaloneMD SkillType = "standalone-md"
)

type Skill struct {
	Name          string
	Path          string
	Type          SkillType
	SymlinkTarget string
	IsBroken      bool
IsDisabled    bool
	SkillMD       string // path to SKILL.md (for directory type)
	Frontmatter   *parser.SkillFrontmatter
}

// List discovers all installed skills.
func List(cfg *config.Config) ([]Skill, error) {
	var skills []Skill

	if !fsutil.PathExists(cfg.SkillsDir) {
		return skills, nil
	}

	entries, err := os.ReadDir(cfg.SkillsDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		entryPath := filepath.Join(cfg.SkillsDir, entry.Name())
		skill := detectSkill(entryPath, entry)
		skills = append(skills, skill)
	}

	return skills, nil
}

func detectSkill(path string, entry os.DirEntry) Skill {
	name := entry.Name()

	// Check if it's a symlink (Readlink is more reliable than ModeSymlink on Windows)
	if isSymlink(path) {
		target, err := fsutil.ReadSymlink(path)
		skill := Skill{
			Name:          name,
			Path:          path,
			Type:          SkillTypeSymlink,
			SymlinkTarget: target,
		}
		if err != nil || !fsutil.PathExists(target) {
			skill.IsBroken = true
		}
		// Try to read SKILL.md from the target
		if target != "" {
			skillMD := filepath.Join(target, "SKILL.md")
			if fsutil.PathExists(skillMD) {
				skill.SkillMD = skillMD
				if fm, err := parseSkillFrontmatter(skillMD); err == nil {
					skill.Frontmatter = fm
					skill.Name = fm.Name
				}
			} else {
				skill.IsBroken = true
			}
		}
		return skill
	}

	// Directory with SKILL.md
	if entry.IsDir() {
		skillMD := filepath.Join(path, "SKILL.md")
		if fsutil.PathExists(skillMD) {
			skill := Skill{
				Name:    name,
				Path:    path,
				Type:    SkillTypeDir,
				SkillMD: skillMD,
			}
			if fm, err := parseSkillFrontmatter(skillMD); err == nil {
				skill.Frontmatter = fm
				skill.Name = fm.Name
			}
			return skill
		}
		if fsutil.PathExists(skillMD+".disabled") {
			skill := Skill{
				Name:       name,
				Path:       path,
				Type:       SkillTypeDir,
				IsDisabled: true,
			}
			if fm, err := parseSkillFrontmatter(skillMD+".disabled"); err == nil {
				skill.Frontmatter = fm
				skill.Name = fm.Name
			}
			return skill
		}
		// Might be a directory with symlinked contents
		return Skill{Name: name, Path: path, Type: SkillTypeDir}
	}

	// Git Bash text symlink (small file containing a target path)
	if target := readGitBashSymlink(path); target != "" {
		skill := Skill{
			Name:          name,
			Path:          path,
			Type:          SkillTypeSymlink,
			SymlinkTarget: target,
		}
		skillMD := filepath.Join(target, "SKILL.md")
		if fsutil.PathExists(skillMD) {
			skill.SkillMD = skillMD
			if fm, err := parseSkillFrontmatter(skillMD); err == nil {
				skill.Frontmatter = fm
				skill.Name = fm.Name
			}
		} else {
			skill.IsBroken = true
		}
		return skill
	}

	// Standalone .md file (or .md.disabled)
	if strings.HasSuffix(name, ".md") {
		skill := Skill{
			Name:    strings.TrimSuffix(name, ".md"),
			Path:    path,
			Type:    SkillTypeStandaloneMD,
			SkillMD: path,
		}
		if fm, err := parseSkillFrontmatter(path); err == nil {
			skill.Frontmatter = fm
			skill.Name = fm.Name
		}
		return skill
	}
	if strings.HasSuffix(name, ".md.disabled") {
		skill := Skill{
			Name:       strings.TrimSuffix(name, ".md.disabled"),
			Path:       path,
			Type:       SkillTypeStandaloneMD,
			IsDisabled: true,
		}
		if fm, err := parseSkillFrontmatter(path); err == nil {
			skill.Frontmatter = fm
			skill.Name = fm.Name
		}
		return skill
	}

	return Skill{Name: name, Path: path}
}

// isSymlink checks if path is a symlink/reparse point (works on Windows junctions too).
func isSymlink(path string) bool {
	_, err := os.Readlink(path)
	return err == nil
}

// readGitBashSymlink detects Git Bash text symlinks (small files containing a target path).
func readGitBashSymlink(path string) string {
	fi, err := os.Stat(path)
	if err != nil || fi.IsDir() || fi.Size() > 512 {
		return ""
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	content := strings.TrimSpace(string(data))
	if content == "" {
		return ""
	}
	// Git Bash text symlinks contain a path starting with / or drive letter
	if strings.HasPrefix(content, "/") || strings.HasPrefix(content, "C:") || strings.HasPrefix(content, "D:") {
		return fsutil.NormalizePath(content)
	}
	return ""
}

func parseSkillFrontmatter(mdPath string) (*parser.SkillFrontmatter, error) {
	data, err := os.ReadFile(mdPath)
	if err != nil {
		return nil, err
	}
	fm, _, err := parser.ExtractFrontmatter(data)
	if err != nil {
		return nil, err
	}
	return parser.ParseSkillFrontmatter(fm)
}

// Validate checks all skills and reports issues.
func Validate(cfg *config.Config) []report.Issue {
	var issues []report.Issue

	skills, err := List(cfg)
	if err != nil {
		issues = append(issues, report.Issue{
			Severity: report.Error,
			Domain:   "skills",
			Message:  fmt.Sprintf("Failed to list skills: %v", err),
		})
		return issues
	}

	names := map[string]bool{}
	for _, s := range skills {
		// Check broken symlinks
		if s.IsBroken {
			issues = append(issues, report.Issue{
				Severity: report.Error,
				Domain:   fmt.Sprintf("skills/%s", s.Name),
				Message:  "Broken symlink",
				Detail:   fmt.Sprintf("Target: %s", s.SymlinkTarget),
				FixSuggestion: fmt.Sprintf("Recreate the symlink: ln -s <target> %s", s.Path),
			})
		}

		// Check for SKILL.md validity
		if s.SkillMD != "" && s.Frontmatter == nil {
			issues = append(issues, report.Issue{
				Severity: report.Warning,
				Domain:   fmt.Sprintf("skills/%s", s.Name),
				Message:  "Invalid or missing YAML frontmatter in SKILL.md",
			})
		}

		// Check required fields
		if s.Frontmatter != nil {
			if s.Frontmatter.Name == "" {
				issues = append(issues, report.Issue{
					Severity: report.Warning,
					Domain:   fmt.Sprintf("skills/%s", s.Name),
					Message:  "SKILL.md frontmatter missing 'name' field",
				})
			}
			if s.Frontmatter.Description == "" {
				issues = append(issues, report.Issue{
					Severity: report.Warning,
					Domain:   fmt.Sprintf("skills/%s", s.Name),
					Message:  "SKILL.md frontmatter missing 'description' field",
				})
			}
		}

		// Check duplicate names
		if s.Frontmatter != nil && s.Frontmatter.Name != "" {
			if names[s.Frontmatter.Name] {
				issues = append(issues, report.Issue{
					Severity: report.Warning,
					Domain:   "skills",
					Message:  fmt.Sprintf("Duplicate skill name: %s", s.Frontmatter.Name),
				})
			}
			names[s.Frontmatter.Name] = true
		}
	}

	return issues
}

// SkillUsage holds the usage instructions for a skill.
type SkillUsage struct {
	Name        string   `json:"name"`
	Invocation  string   `json:"invocation"`  // e.g. "/docx"
	Description string   `json:"description"`
	Triggers    []string `json:"triggers"`    // extracted trigger keywords
	Deps        string   `json:"deps,omitempty"` // install command if any
}

// GetUsage returns structured usage information for a skill by name.
func GetUsage(cfg *config.Config, name string) (*SkillUsage, error) {
	skillList, err := List(cfg)
	if err != nil {
		return nil, err
	}

	for _, s := range skillList {
		if s.Frontmatter == nil || s.Frontmatter.Name != name {
			continue
		}

		u := &SkillUsage{
			Name:        s.Frontmatter.Name,
			Invocation:  "/" + s.Frontmatter.Name,
			Description: s.Frontmatter.Description,
			Triggers:    extractTriggers(s.Frontmatter.Description),
		}

		// Try to extract install deps from SKILL.md body
		if s.SkillMD != "" {
			if data, err := os.ReadFile(s.SkillMD); err == nil {
				_, body, _ := parser.ExtractFrontmatter(data)
				u.Deps = extractDeps(string(body))
			}
		}

		return u, nil
	}
	return nil, fmt.Errorf("skill not found: %s", name)
}

// GetAllUsage returns usage info for all installed skills.
func GetAllUsage(cfg *config.Config) []SkillUsage {
	var result []SkillUsage
	skillList, _ := List(cfg)
	for _, s := range skillList {
		if s.Frontmatter == nil {
			continue
		}
		u := SkillUsage{
			Name:        s.Frontmatter.Name,
			Invocation:  "/" + s.Frontmatter.Name,
			Description: s.Frontmatter.Description,
			Triggers:    extractTriggers(s.Frontmatter.Description),
		}
		if s.SkillMD != "" {
			if data, err := os.ReadFile(s.SkillMD); err == nil {
				_, body, _ := parser.ExtractFrontmatter(data)
				u.Deps = extractDeps(string(body))
			}
		}
		result = append(result, u)
	}
	return result
}

// extractTriggers parses trigger keywords from a skill description.
// Common patterns: "Triggers include: ...", "Triggers on: ...", "Trigger whenever ..."
func extractTriggers(desc string) []string {
	if desc == "" {
		return nil
	}

	// Try to find explicit trigger lists
	patterns := []string{
		`Triggers include:`,
		`Triggers on:`,
		`Trigger whenever`,
		`Triggers when`,
	}

	for _, p := range patterns {
		idx := strings.Index(strings.ToLower(desc), strings.ToLower(p))
		if idx < 0 {
			continue
		}
		rest := desc[idx+len(p):]
		// Take the first sentence (up to . or newline or next keyword)
		if end := strings.IndexAny(rest, ".\n"); end > 0 {
			rest = rest[:end]
		}
		// Extract quoted or comma-separated keywords
		return extractKeywords(rest)
	}

	return nil
}

func extractKeywords(s string) []string {
	var keywords []string
	seen := map[string]bool{}

	addKW := func(kw string) {
		kw = strings.Trim(kw, `,;. "'`+"`")
		kw = strings.TrimSpace(kw)
		kw = strings.TrimLeft(kw, "-•*")
		kw = strings.TrimSpace(kw)
		if kw != "" && !seen[kw] {
			seen[kw] = true
			keywords = append(keywords, kw)
		}
	}

	// Quoted strings
	for _, q := range []string{`"`, `'`, "`"} {
		reStart := 0
		for {
			start := strings.Index(s[reStart:], q)
			if start < 0 {
				break
			}
			start += reStart + 1
			end := strings.Index(s[start:], q)
			if end < 0 {
				break
			}
			addKW(s[start : start+end])
			reStart = start + end + 1
		}
	}

	// Comma-separated fallback
	if len(keywords) == 0 {
		for _, part := range strings.Split(s, ",") {
			addKW(part)
		}
	}

	return keywords
}

// extractDeps finds install commands (npm/pip/apt) from SKILL.md body.
func extractDeps(body string) string {
	lines := strings.Split(body, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		for _, prefix := range []string{"npm install", "pip install", "apt install", "go install", "cargo install"} {
			if strings.HasPrefix(trimmed, prefix) {
				return trimmed
			}
		}
	}
	return ""
}
