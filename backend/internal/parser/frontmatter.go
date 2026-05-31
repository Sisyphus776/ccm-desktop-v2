package parser

import (
	"bytes"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// SkillFrontmatter is the YAML frontmatter from a SKILL.md file.
type SkillFrontmatter struct {
	Name         string   `yaml:"name"`
	Description  string   `yaml:"description"`
	License      string   `yaml:"license"`
	AllowedTools string   `yaml:"allowed-tools"`
	Hidden       bool     `yaml:"hidden"`
	Model        string   `yaml:"model"`
}

// MemoryFrontmatter is the YAML frontmatter from a memory .md file.
type MemoryFrontmatter struct {
	Name            string `yaml:"name"`
	Description     string `yaml:"description"`
	Type            string `yaml:"type"`
	OriginSessionID string `yaml:"originSessionId"`
	Metadata        struct {
		Type string `yaml:"type"`
	} `yaml:"metadata"`
}

// ExtractFrontmatter extracts YAML frontmatter (between --- delimiters) from markdown content.
// Returns the frontmatter bytes, the body bytes, and any error.
// Tolerates leading whitespace and blank lines before the opening ---.
func ExtractFrontmatter(content []byte) ([]byte, []byte, error) {
	// Strip BOM if present
	content = bytes.TrimPrefix(content, []byte{0xEF, 0xBB, 0xBF})

	s := string(content)

	// Skip leading whitespace and blank lines
	trimmed := strings.TrimLeft(s, "\r\n\t ")
	if strings.HasPrefix(trimmed, "---") {
		rest := trimmed[3:]
		// Allow space/tab after --- (tolerate non-standard delimiters like "---  ")
		rest = strings.TrimLeft(rest, " \t")

		// Find closing --- (must be on its own line)
		if idx := strings.Index(rest, "\n---"); idx >= 0 {
			fm := strings.TrimSpace(rest[:idx])
			body := rest[idx+4:] // skip \n---
			return []byte(fm), []byte(strings.TrimSpace(body)), nil
		}
		if idx := strings.Index(rest, "\r\n---"); idx >= 0 {
			fm := strings.TrimSpace(rest[:idx])
			body := rest[idx+5:] // skip \r\n---
			return []byte(fm), []byte(strings.TrimSpace(body)), nil
		}
		return nil, content, fmt.Errorf("unclosed frontmatter")
	}

	return nil, content, fmt.Errorf("no frontmatter found")
}

// ParseSkillFrontmatter parses a SKILL.md frontmatter.
func ParseSkillFrontmatter(fm []byte) (*SkillFrontmatter, error) {
	var sf SkillFrontmatter
	if err := yaml.Unmarshal(fm, &sf); err != nil {
		return nil, err
	}
	return &sf, nil
}

// ParseMemoryFrontmatter parses a memory file frontmatter.
func ParseMemoryFrontmatter(fm []byte) (*MemoryFrontmatter, error) {
	var mf MemoryFrontmatter
	if err := yaml.Unmarshal(fm, &mf); err != nil {
		return nil, err
	}
	// metadata.type takes precedence if set
	if mf.Metadata.Type != "" {
		mf.Type = mf.Metadata.Type
	}
	return &mf, nil
}

// ValidateMemoryType checks if the memory type is valid.
func ValidateMemoryType(t string) bool {
	switch t {
	case "user", "feedback", "project", "reference":
		return true
	}
	return false
}
