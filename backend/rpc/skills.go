package rpc

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"ccm-desktop-v2/backend/internal/fsutil"
	"ccm-desktop-v2/backend/internal/parser"
	"ccm-desktop-v2/backend/internal/skills"
	"ccm-desktop-v2/backend/internal/translate"
)

type SkillItem struct {
	Name          string   `json:"name"`
	Type          string   `json:"type"`
	Status        string   `json:"status"`
	Disabled      bool     `json:"disabled"`
	Invocation    string   `json:"invocation"`
	Description   string   `json:"description"`
	DescriptionCN string   `json:"descriptionCN"`
	Triggers      []string `json:"triggers"`
	Target        string   `json:"target"`
}

type SkillDetailItem struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Status      string   `json:"status"`
	Invocation  string   `json:"invocation"`
	Description string   `json:"description"`
	Triggers    []string `json:"triggers"`
	Target      string   `json:"target"`
	Errors      []string `json:"errors"`
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func toggleSkillFile(skillsDir string) (enabled bool, err error) {
	skillMD := filepath.Join(skillsDir, "SKILL.md")
	disabledMD := skillMD + ".disabled"
	if fsutil.PathExists(disabledMD) {
		return true, os.Rename(disabledMD, skillMD)
	}
	if fsutil.PathExists(skillMD) {
		return false, os.Rename(skillMD, disabledMD)
	}
	return false, fmt.Errorf("SKILL.md not found: %s", skillsDir)
}

func toggleStandaloneSkillFile(basePath string) (enabled bool, err error) {
	mdFile := basePath + ".md"
	disabledFile := basePath + ".md.disabled"
	if fsutil.PathExists(disabledFile) {
		return true, os.Rename(disabledFile, mdFile)
	}
	if fsutil.PathExists(mdFile) {
		return false, os.Rename(mdFile, disabledFile)
	}
	return false, fmt.Errorf("skill file not found: %s", basePath)
}

func extractRepoName(url string) string {
	url = strings.TrimSuffix(url, "/")
	url = strings.TrimSuffix(url, ".git")
	if idx := strings.LastIndex(url, "/"); idx >= 0 {
		return url[idx+1:]
	}
	return ""
}

func mirrorURL(url string) string {
	return strings.Replace(url, "https://github.com/", "https://ghproxy.com/https://github.com/", 1)
}

func firstMeaningfulLine(body string) string {
	lines := strings.Split(strings.TrimSpace(body), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			if strings.HasPrefix(line, "#") {
				return strings.TrimLeft(line, "# ")
			}
			continue
		}
		return line
	}
	return ""
}

// ---------------------------------------------------------------------------
// RPC methods
// ---------------------------------------------------------------------------

func listSkills(ctx *AppContext) (any, error) {
	skillList, err := skills.List(ctx.Cfg)
	if err != nil {
		return nil, nil
	}
	usages := skills.GetAllUsage(ctx.Cfg)
	usageMap := map[string]skills.SkillUsage{}
	for _, u := range usages {
		usageMap[u.Name] = u
	}

	var items []SkillItem
	for _, sk := range skillList {
		item := SkillItem{
			Name:   sk.Name,
			Type:   string(sk.Type),
			Target: sk.SymlinkTarget,
		}
		item.Disabled = sk.IsDisabled
		if sk.Frontmatter != nil {
			item.Description = sk.Frontmatter.Description
			if translate.IsMostlyChinese(sk.Frontmatter.Description) {
				item.DescriptionCN = ""
			} else {
				item.DescriptionCN = translate.TranslateDescription(sk.Frontmatter.Description)
			}
			item.Invocation = "/" + sk.Frontmatter.Name
			item.Name = sk.Frontmatter.Name
			if u, ok := usageMap[sk.Frontmatter.Name]; ok {
				item.Triggers = u.Triggers
			}
		} else if sk.SkillMD != "" {
			item.Disabled = sk.IsDisabled
			if data, err := os.ReadFile(sk.SkillMD); err == nil {
				_, body, _ := parser.ExtractFrontmatter(data)
				line := firstMeaningfulLine(string(body))
				if line != "" {
					item.Description = line
					if translate.IsMostlyChinese(line) {
						item.DescriptionCN = ""
					} else {
						item.DescriptionCN = translate.TranslateDescription(line)
					}
				}
			}
		}
		if sk.IsBroken {
			item.Status = "broken"
		} else if sk.Frontmatter != nil {
			item.Status = "ok"
		} else {
			item.Status = "warning"
		}
		items = append(items, item)
	}
	return items, nil
}

func validateSkills(ctx *AppContext) (any, error) {
	var items []IssueItem
	for _, iss := range skills.Validate(ctx.Cfg) {
		items = append(items, issueToItem(iss))
	}
	return items, nil
}

func toggleSkill(ctx *AppContext, h *Handler, name string) (any, error) {
	skillDir := filepath.Join(ctx.Cfg.SkillsDir, name)
	if enabled, err := toggleSkillFile(skillDir); err == nil {
		h.Notify("config-changed", map[string]any{"domain": "skills"})
		if enabled {
			return "已启用: " + name, nil
		}
		return "已禁用: " + name, nil
	}
	// standalone .md file
	basePath := filepath.Join(ctx.Cfg.SkillsDir, name)
	if enabled, err := toggleStandaloneSkillFile(basePath); err == nil {
		h.Notify("config-changed", map[string]any{"domain": "skills"})
		if enabled {
			return "已启用: " + name, nil
		}
		return "已禁用: " + name, nil
	}
	return nil, fmt.Errorf("skill not found: %s", name)
}

func createSkill(ctx *AppContext, h *Handler, name, description string) (any, error) {
	if name == "" {
		return nil, fmt.Errorf("名称不能为空")
	}
	skillDir := filepath.Join(ctx.Cfg.SkillsDir, name)
	yaml := fmt.Sprintf(`---
name: %s
description: "%s"
---
# %s

## 功能

%s

## 调用方式

- 调用指令：/%s
- 触发关键词：待填写
`, name, description, name, description, name)

	if err := os.MkdirAll(skillDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}
	mdPath := filepath.Join(skillDir, "SKILL.md")
	if err := os.WriteFile(mdPath, []byte(yaml), 0644); err != nil {
		return nil, fmt.Errorf("failed to write SKILL.md: %w", err)
	}
	h.Notify("config-changed", map[string]any{"domain": "skills"})
	return "Skill created: " + mdPath, nil
}

func deleteSkill(ctx *AppContext, h *Handler, name string) (any, error) {
	skillDir := filepath.Join(ctx.Cfg.SkillsDir, name)
	skillMD := filepath.Join(ctx.Cfg.SkillsDir, name+".md")
	if fsutil.PathExists(skillDir) {
		os.RemoveAll(skillDir)
		h.Notify("config-changed", map[string]any{"domain": "skills"})
		return "已删除: " + skillDir, nil
	}
	if fsutil.PathExists(skillMD) {
		os.Remove(skillMD)
		h.Notify("config-changed", map[string]any{"domain": "skills"})
		return "已删除: " + skillMD, nil
	}
	return nil, fmt.Errorf("skill not found: %s", name)
}

func importSkill(ctx *AppContext, h *Handler, githubURL string) (any, error) {
	if githubURL == "" {
		return nil, fmt.Errorf("URL 不能为空")
	}
	repoName := extractRepoName(githubURL)
	if repoName == "" {
		return nil, fmt.Errorf("无法从 URL 提取仓库名: %s", githubURL)
	}
	destDir := filepath.Join(ctx.Cfg.SkillsDir, repoName)
	if fsutil.PathExists(destDir) {
		return nil, fmt.Errorf("skill already exists: %s", destDir)
	}
	urls := []string{githubURL, mirrorURL(githubURL)}
	var lastErr error
	for _, u := range urls {
		cmd := exec.Command("git", "clone", "--progress", u, destDir)
		cmd.Dir = ctx.Cfg.SkillsDir
		stdout, _ := cmd.StdoutPipe()
		stderr, _ := cmd.StderrPipe()
		cmd.Stdin = nil
		if err := cmd.Start(); err != nil {
			lastErr = fmt.Errorf("%s: %v", u, err)
			continue
		}
		go func() {
			buf := make([]byte, 256)
			for {
				n, err := stdout.Read(buf)
				if n > 0 {
					h.Notify("import-progress", map[string]any{"line": string(buf[:n])})
				}
				if err != nil {
					break
				}
			}
		}()
		go func() {
			buf := make([]byte, 256)
			for {
				n, err := stderr.Read(buf)
				if n > 0 {
					h.Notify("import-progress", map[string]any{"line": string(buf[:n])})
				}
				if err != nil {
					break
				}
			}
		}()
		waitErr := cmd.Wait()
		if waitErr == nil {
			skillMD := filepath.Join(destDir, "SKILL.md")
			if fsutil.PathExists(skillMD) {
				h.Notify("config-changed", map[string]any{"domain": "skills"})
				return "导入成功: " + destDir, nil
			}
			os.RemoveAll(destDir)
			return nil, fmt.Errorf("仓库不包含 SKILL.md 文件，不是合法的 Skill")
		}
		lastErr = fmt.Errorf("%s: %v", u, waitErr)
	}
	return nil, fmt.Errorf("克隆失败 (直连和镜像均失败): %v", lastErr)
}

func getSkillErrors(ctx *AppContext, name string) (any, error) {
	skillList, _ := skills.List(ctx.Cfg)
	for _, sk := range skillList {
		if sk.Frontmatter == nil || sk.Frontmatter.Name != name {
			continue
		}
		item := &SkillDetailItem{
			Name: sk.Frontmatter.Name,
			Type: string(sk.Type),
		}
		if sk.Frontmatter != nil {
			item.Description = sk.Frontmatter.Description
			item.Invocation = "/" + sk.Frontmatter.Name
		}
		if sk.SymlinkTarget != "" {
			item.Target = sk.SymlinkTarget
		}

		var errs []string
		if sk.IsBroken {
			errs = append(errs, "软链接已断开")
			if sk.SymlinkTarget != "" {
				if !fsutil.PathExists(sk.SymlinkTarget) {
					errs = append(errs, fmt.Sprintf("目标路径不存在: %s", sk.SymlinkTarget))
				}
			}
			if sk.SkillMD == "" {
				errs = append(errs, "目标目录中未找到 SKILL.md 文件")
			}
		}
		if sk.Frontmatter == nil {
			errs = append(errs, "SKILL.md 缺少有效的 YAML frontmatter")
		} else {
			if sk.Frontmatter.Name == "" {
				errs = append(errs, "frontmatter 缺少 'name' 字段，将使用目录名")
			}
			if sk.Frontmatter.Description == "" {
				errs = append(errs, "frontmatter 缺少 'description' 字段，AI 无法识别触发词")
			}
		}
		if sk.SkillMD != "" && sk.Frontmatter != nil {
			if data, err := os.ReadFile(sk.SkillMD); err == nil {
				if _, body, _ := parser.ExtractFrontmatter(data); len(strings.TrimSpace(string(body))) < 10 {
					errs = append(errs, "SKILL.md body 内容过短，缺少使用说明")
				}
			}
		}
		if len(errs) > 0 {
			item.Status = "broken"
		} else {
			item.Status = "ok"
		}
		item.Errors = errs
		return item, nil
	}
	return nil, nil
}
