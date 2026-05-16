package rpc

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"ccm-desktop-v2/backend/internal/fsutil"
	"ccm-desktop-v2/backend/internal/memory"
)

type MemoryFileItem struct {
	File        string `json:"file"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Content     string `json:"content,omitempty"`
}

type MemoryStats struct {
	Total       int            `json:"total"`
	ByType      map[string]int `json:"byType"`
	ByProject   map[string]int `json:"byProject"`
	OrphanCount int            `json:"orphanCount"`
	Oldest      string         `json:"oldest"`
	Newest      string         `json:"newest"`
}

func extractProjectName(homeDir string) string {
	dir := strings.TrimPrefix(homeDir, "C:\\Users\\")
	dir = strings.TrimPrefix(dir, "/c/Users/")
	dir = strings.TrimPrefix(dir, "/Users/")
	dir = strings.ReplaceAll(dir, "\\", "--")
	dir = strings.ReplaceAll(dir, "/", "--")
	return "C--" + dir
}

func listMemory(ctx *AppContext) (any, error) {
	mems, _ := memory.ListAll(ctx.Cfg)
	var items []MemoryFileItem
	for _, m := range mems {
		item := MemoryFileItem{File: filepath.Base(m.Path)}
		if m.Frontmatter != nil {
			if m.Frontmatter.Name != "" {
				item.Name = m.Frontmatter.Name
			}
			if m.Frontmatter.Type != "" {
				item.Type = m.Frontmatter.Type
			} else if m.Frontmatter.Metadata.Type != "" {
				item.Type = m.Frontmatter.Metadata.Type
			}
			item.Description = m.Frontmatter.Description
		}
		items = append(items, item)
	}
	return items, nil
}

func getMemoryStats(ctx *AppContext) (any, error) {
	mems, _ := memory.ListAll(ctx.Cfg)
	stats := memory.GetStats(mems)
	result := MemoryStats{
		Total:       stats.TotalCount,
		ByType:      stats.ByType,
		ByProject:   stats.ByProject,
		OrphanCount: stats.OrphanCount,
	}
	if !stats.Oldest.IsZero() {
		result.Oldest = stats.Oldest.Format("2006-01-02")
	}
	if !stats.Newest.IsZero() {
		result.Newest = stats.Newest.Format("2006-01-02")
	}
	return result, nil
}

func validateMemory(ctx *AppContext) (any, error) {
	var items []IssueItem
	for _, iss := range memory.Validate(ctx.Cfg) {
		items = append(items, issueToItem(iss))
	}
	return items, nil
}

func createMemory(ctx *AppContext, h *Handler, name, memType, description, content string) (any, error) {
	if name == "" || memType == "" {
		return nil, fmt.Errorf("名称和类型不能为空")
	}
	projName := extractProjectName(ctx.Cfg.HomeDir)
	memoryDir := filepath.Join(ctx.Cfg.ProjectsDir, projName, "memory")
	if err := os.MkdirAll(memoryDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create memory dir: %w", err)
	}

	slug := strings.ToLower(strings.ReplaceAll(name, " ", "-"))
	fileName := slug + ".md"
	yaml := fmt.Sprintf(`---
name: %s
description: "%s"
metadata:
  type: %s
---

%s
`, name, description, memType, content)

	filePath := filepath.Join(memoryDir, fileName)
	if err := os.WriteFile(filePath, []byte(yaml), 0644); err != nil {
		return nil, fmt.Errorf("failed to write memory file: %w", err)
	}

	// Update MEMORY.md index
	indexPath := filepath.Join(memoryDir, "MEMORY.md")
	entry := fmt.Sprintf("- [%s](%s) -- %s\n", name, fileName, description)
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		os.WriteFile(indexPath, []byte(entry+"\n"), 0644)
	} else if data, err := os.ReadFile(indexPath); err == nil {
		content := string(data)
		if !strings.Contains(content, fileName) {
			content = strings.TrimRight(content, "\n") + "\n" + entry + "\n"
			os.WriteFile(indexPath, []byte(content), 0644)
		}
	}

	h.Notify("config-changed", map[string]any{"domain": "memory"})
	return "Memory created: " + filePath, nil
}

func getMemoryContent(ctx *AppContext, file string) (any, error) {
	projName := extractProjectName(ctx.Cfg.HomeDir)
	memoryDir := filepath.Join(ctx.Cfg.ProjectsDir, projName, "memory")
	filePath := filepath.Join(memoryDir, file)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", nil
	}
	return string(data), nil
}

func deleteMemory(ctx *AppContext, h *Handler, file string) (any, error) {
	projName := extractProjectName(ctx.Cfg.HomeDir)
	memoryDir := filepath.Join(ctx.Cfg.ProjectsDir, projName, "memory")
	filePath := filepath.Join(memoryDir, file)
	if !fsutil.PathExists(filePath) {
		return nil, fmt.Errorf("文件不存在")
	}
	// Remove from MEMORY.md index
	indexPath := filepath.Join(memoryDir, "MEMORY.md")
	if data, err := os.ReadFile(indexPath); err == nil {
		lines := strings.Split(string(data), "\n")
		var newLines []string
		for _, line := range lines {
			if !strings.Contains(line, file) {
				newLines = append(newLines, line)
			}
		}
		os.WriteFile(indexPath, []byte(strings.Join(newLines, "\n")), 0644)
	}
	if err := os.Remove(filePath); err != nil {
		return nil, fmt.Errorf("failed to delete: %w", err)
	}
	h.Notify("config-changed", map[string]any{"domain": "memory"})
	return "已删除: " + file, nil
}
