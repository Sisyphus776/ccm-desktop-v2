package memory

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"ccm-desktop/internal/config"
	"ccm-desktop/internal/fsutil"
	"ccm-desktop/internal/parser"
	"ccm-desktop/internal/report"
)

type MemoryIndexEntry struct {
	Title    string
	FileName string
	Desc     string
}

type MemoryFile struct {
	Path        string
	Project     string
	Frontmatter *parser.MemoryFrontmatter
	Body        string
	ModTime     time.Time
}

// ParseIndex reads MEMORY.md and returns the list of entries.
func ParseIndex(indexPath string) ([]MemoryIndexEntry, error) {
	data, err := os.ReadFile(indexPath)
	if err != nil {
		return nil, err
	}

	var entries []MemoryIndexEntry
	// Format: - [Title](filename.md) -- description
	re := regexp.MustCompile(`-\s+\[([^\]]+)\]\(([^)]+)\)\s*(?:--\s*(.*))?`)
	for _, m := range re.FindAllStringSubmatch(string(data), -1) {
		e := MemoryIndexEntry{Title: m[1], FileName: m[2]}
		if len(m) > 3 {
			e.Desc = strings.TrimSpace(m[3])
		}
		entries = append(entries, e)
	}
	return entries, nil
}

// ListAll discovers all memory files across all projects.
func ListAll(cfg *config.Config) ([]MemoryFile, error) {
	var mems []MemoryFile

	if !fsutil.PathExists(cfg.ProjectsDir) {
		return mems, nil
	}

	projEntries, err := os.ReadDir(cfg.ProjectsDir)
	if err != nil {
		return nil, err
	}

	for _, proj := range projEntries {
		if !proj.IsDir() {
			continue
		}
		memoryDir := filepath.Join(cfg.ProjectsDir, proj.Name(), "memory")
		if !fsutil.PathExists(memoryDir) {
			continue
		}

		// Read MEMORY.md index
		indexPath := filepath.Join(memoryDir, "MEMORY.md")
		entries, _ := ParseIndex(indexPath)

		indexedFiles := map[string]bool{}
		for _, e := range entries {
			indexedFiles[e.FileName] = true
			memPath := filepath.Join(memoryDir, e.FileName)
			if fsutil.PathExists(memPath) {
				mf := readMemoryFile(memPath, proj.Name())
				mems = append(mems, mf)
			} else {
				// Orphaned index entry
				mems = append(mems, MemoryFile{
					Path:    memPath,
					Project: proj.Name(),
				})
			}
		}

		// Find orphaned .md files (not in index)
		mdFiles, _ := filepath.Glob(filepath.Join(memoryDir, "*.md"))
		for _, f := range mdFiles {
			base := filepath.Base(f)
			if base == "MEMORY.md" || indexedFiles[base] {
				continue
			}
			mems = append(mems, MemoryFile{
				Path:    f,
				Project: proj.Name(),
			})
		}
	}

	return mems, nil
}

func readMemoryFile(path, project string) MemoryFile {
	mf := MemoryFile{Path: path, Project: project}

	data, err := os.ReadFile(path)
	if err != nil {
		return mf
	}

	if fi, err := os.Stat(path); err == nil {
		mf.ModTime = fi.ModTime()
	}

	fm, body, err := parser.ExtractFrontmatter(data)
	if err != nil {
		return mf
	}
	mf.Body = string(body)

	parsed, err := parser.ParseMemoryFrontmatter(fm)
	if err != nil {
		return mf
	}
	mf.Frontmatter = parsed
	return mf
}

// Stats returns memory statistics.
type Stats struct {
	TotalCount  int
	ByType      map[string]int
	ByProject   map[string]int
	Oldest      time.Time
	Newest      time.Time
	OrphanCount int
}

func GetStats(mems []MemoryFile) Stats {
	s := Stats{
		ByType:    map[string]int{},
		ByProject: map[string]int{},
	}

	for _, m := range mems {
		s.TotalCount++

		if m.Frontmatter != nil {
			t := m.Frontmatter.Type
			if t == "" {
				t = "unknown"
			}
			s.ByType[t]++
		} else {
			s.ByType["unknown"]++
		}

		s.ByProject[m.Project]++

		if !m.ModTime.IsZero() {
			if s.Newest.IsZero() || m.ModTime.After(s.Newest) {
				s.Newest = m.ModTime
			}
			if s.Oldest.IsZero() || m.ModTime.Before(s.Oldest) {
				s.Oldest = m.ModTime
			}
		}

		if m.Frontmatter == nil {
			s.OrphanCount++
		}
	}

	return s
}

// Validate checks memory files and reports issues.
func Validate(cfg *config.Config) []report.Issue {
	var issues []report.Issue

	mems, err := ListAll(cfg)
	if err != nil {
		issues = append(issues, report.Issue{
			Severity: report.Error,
			Domain:   "memory",
			Message:  fmt.Sprintf("Failed to list memory files: %v", err),
		})
		return issues
	}

	for _, m := range mems {
		base := filepath.Base(m.Path)

		// Missing file but in index
		if !fsutil.PathExists(m.Path) && m.Frontmatter == nil && m.Body == "" {
			issues = append(issues, report.Issue{
				Severity: report.Warning,
				Domain:   fmt.Sprintf("memory/%s", base),
				Message:  "Memory file referenced in MEMORY.md but does not exist",
				FixSuggestion: fmt.Sprintf("Create %s or remove from MEMORY.md", m.Path),
			})
			continue
		}

		// Orphaned file (exists but no frontmatter and not in index)
		if m.Frontmatter == nil && m.Body == "" {
			continue // skip files we couldn't parse, likely not memory files
		}

		if m.Frontmatter == nil {
			issues = append(issues, report.Issue{
				Severity: report.Warning,
				Domain:   fmt.Sprintf("memory/%s", base),
				Message:  "Missing or invalid YAML frontmatter",
			})
			continue
		}

		// Validate type
		if m.Frontmatter.Type != "" && !parser.ValidateMemoryType(m.Frontmatter.Type) {
			issues = append(issues, report.Issue{
				Severity: report.Warning,
				Domain:   fmt.Sprintf("memory/%s", base),
				Message:  fmt.Sprintf("Unknown memory type: %s", m.Frontmatter.Type),
				Detail:   "Valid types are: user, feedback, project, reference",
			})
		}

		// Check name field
		if m.Frontmatter.Name == "" {
			issues = append(issues, report.Issue{
				Severity: report.Warning,
				Domain:   fmt.Sprintf("memory/%s", base),
				Message:  "Memory file missing 'name' in frontmatter",
			})
		}
	}

	return issues
}
