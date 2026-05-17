package rpc

import (
	"fmt"
	"os"
	"path/filepath"

	"ccm-desktop-v2/backend/internal/claudemd"
	"ccm-desktop-v2/backend/internal/fsutil"
)

type ClaudeMDItem struct {
	Path       string   `json:"path"`
	Level      string   `json:"level"`
	Size       int64    `json:"size"`
	References []string `json:"references"`
}

func listClaudeMD(ctx *AppContext) (any, error) {
	mds := claudemd.FindAll(ctx.Cfg, nil)
	var items []ClaudeMDItem
	for _, md := range mds {
		items = append(items, ClaudeMDItem{
			Path:       md.Path,
			Level:      md.Level,
			Size:       md.Size,
			References: md.References,
		})
	}
	return items, nil
}

func getClaudeMDContent(ctx *AppContext, path string) (any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", nil
	}
	return string(data), nil
}

func createClaudeMD(ctx *AppContext, path, content string) (any, error) {
	if path == "" {
		return nil, fmt.Errorf("路径不能为空")
	}
	if content == "" {
		return nil, fmt.Errorf("内容不能为空")
	}
	if fsutil.PathExists(path) {
		return nil, fmt.Errorf("文件已存在: %s", path)
	}
	// Ensure parent directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("无法创建目录 %s: %w", dir, err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return nil, fmt.Errorf("写入失败: %w", err)
	}
	return fmt.Sprintf("已创建: %s", path), nil
}

func updateClaudeMD(ctx *AppContext, path, content string) (any, error) {
	if path == "" {
		return nil, fmt.Errorf("路径不能为空")
	}
	if !fsutil.PathExists(path) {
		return nil, fmt.Errorf("文件不存在: %s", path)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return nil, fmt.Errorf("写入失败: %w", err)
	}
	return fmt.Sprintf("已更新: %s", path), nil
}

func deleteClaudeMD(ctx *AppContext, path string) (any, error) {
	if path == "" {
		return nil, fmt.Errorf("路径不能为空")
	}
	if !fsutil.PathExists(path) {
		return nil, fmt.Errorf("文件不存在: %s", path)
	}
	if err := os.Remove(path); err != nil {
		return nil, fmt.Errorf("删除失败: %w", err)
	}
	return fmt.Sprintf("已删除: %s", path), nil
}

func validateClaudeMD(ctx *AppContext) (any, error) {
	var items []IssueItem
	mds := claudemd.FindAll(ctx.Cfg, nil)
	for _, md := range mds {
		for _, iss := range claudemd.Validate(ctx.Cfg, md) {
			items = append(items, issueToItem(iss))
		}
	}
	return items, nil
}
