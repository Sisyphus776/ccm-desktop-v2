package skills

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"ccm-desktop-v2/internal/config"
	"ccm-desktop-v2/internal/fsutil"
	"ccm-desktop-v2/internal/parser"
	"ccm-desktop-v2/internal/report"
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

// List discovers all installed skills recursively.
func List(cfg *config.Config) ([]Skill, error) {
	var skills []Skill

	if !fsutil.PathExists(cfg.SkillsDir) {
		return skills, nil
	}

	// First, scan top-level entries for symlinks (WalkDir doesn't follow them).
	topEntries, _ := os.ReadDir(cfg.SkillsDir)
	processedDirs := map[string]bool{}

	for _, entry := range topEntries {
		entryPath := filepath.Join(cfg.SkillsDir, entry.Name())

		// Symlinks — WalkDir won't enter them, process here
		if isSymlink(entryPath) {
			skill := detectSkill(entryPath, entry)
			skills = append(skills, skill)
			if entry.IsDir() {
				processedDirs[entryPath] = true
			}
		}
	}

	// Now walk the directory tree for directory-type and standalone-file skills.
	filepath.WalkDir(cfg.SkillsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		name := d.Name()

		// Skip hidden files/dirs
		if name != "" && name[0] == '.' && !strings.HasSuffix(name, ".md") && !strings.HasSuffix(name, ".yml") && !strings.HasSuffix(name, ".yaml") && !strings.HasSuffix(name, ".disabled") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip symlinks already processed above
		if processedDirs[path] || isSymlink(path) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// SKILL.md / SKILL.yml inside a directory = skill root
		if isSkillFile(name) {
			dir := filepath.Dir(path)
			for _, existing := range skills {
				if existing.Path == dir {
					return nil
				}
			}
			skill := detectSkill(dir, nil)
			skills = append(skills, skill)
			return filepath.SkipDir
		}

		// Standalone .md / .yml / .yaml file (not named SKILL)
		ext := filepath.Ext(name)
		lowerName := strings.ToLower(name)
		if (ext == ".md" || ext == ".yml" || ext == ".yaml") && !strings.HasPrefix(lowerName, "skill.") {
			// Also handle .disabled suffix
			if !d.IsDir() {
				skill := detectSkill(path, d)
				skills = append(skills, skill)
				return nil
			}
		}

		return nil
	})

	return skills, nil
}

// isSkillFile returns true if the filename represents a skill definition file.
// Recognizes: SKILL.md, SKILL.yml, SKILL.yaml (and .disabled variants).
func isSkillFile(name string) bool {
	if strings.HasSuffix(name, ".disabled") {
		name = strings.TrimSuffix(name, ".disabled")
	}
	switch name {
	case "SKILL.md", "SKILL.yml", "SKILL.yaml":
		return true
	default:
		return false
	}
}

func detectSkill(path string, entry os.DirEntry) Skill {
	// If called from WalkDir, entry may be nil — get info from path
	if entry == nil {
		info, err := os.Stat(path)
		if err != nil {
			return Skill{Name: filepath.Base(path), Path: path}
		}
		entry = fs.FileInfoToDirEntry(info)
	}
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
		if target != "" && !skill.IsBroken {
			skillMD := findSkillFile(target)
			if skillMD != "" {
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

	// Directory: check for SKILL.md / SKILL.yml inside
	if entry.IsDir() {
		skillMD := findSkillFile(path)
		if skillMD != "" {
			disabled := strings.HasSuffix(filepath.Base(skillMD), ".disabled")
			skill := Skill{
				Name:       name,
				Path:       path,
				Type:       SkillTypeDir,
				SkillMD:    skillMD,
				IsDisabled: disabled,
			}
			if fm, err := parseSkillFrontmatter(skillMD); err == nil {
				skill.Frontmatter = fm
				skill.Name = fm.Name
			}
			return skill
		}
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
		skillMD := findSkillFile(target)
		if skillMD != "" {
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

	// Standalone .md / .yml / .yaml file — any markdown/yaml file
	ext := strings.ToLower(filepath.Ext(name))
	if ext == ".md" || ext == ".yml" || ext == ".yaml" {
		disabled := false
		baseName := name
		if strings.HasSuffix(baseName, ".disabled") {
			baseName = strings.TrimSuffix(baseName, ".disabled")
			disabled = true
		}
		skill := Skill{
			Name:       strings.TrimSuffix(baseName, ext),
			Path:       path,
			Type:       SkillTypeStandaloneMD,
			SkillMD:    path,
			IsDisabled: disabled,
		}
		if fm, err := parseSkillFrontmatter(path); err == nil {
			skill.Frontmatter = fm
			skill.Name = fm.Name
		}
		return skill
	}

	return Skill{Name: name, Path: path}
}

// findSkillFile searches a directory for any recognized skill definition file.
// Priority: SKILL.md > SKILL.yml > SKILL.yaml (and their .disabled variants).
func findSkillFile(dir string) string {
	for _, name := range []string{"SKILL.md", "SKILL.yml", "SKILL.yaml"} {
		p := filepath.Join(dir, name)
		if fsutil.PathExists(p) {
			return p
		}
		p = filepath.Join(dir, name+".disabled")
		if fsutil.PathExists(p) {
			return p
		}
	}
	return ""
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
	// Git Bash text symlinks contain an absolute path (Unix / or Windows drive letter)
	if strings.HasPrefix(content, "/") || (len(content) >= 2 && content[1] == ':') {
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
