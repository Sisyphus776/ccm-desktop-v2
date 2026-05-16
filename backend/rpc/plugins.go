package rpc

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"ccm-desktop-v2/backend/internal/fsutil"
	"ccm-desktop-v2/backend/internal/parser"
	"ccm-desktop-v2/backend/internal/translate"
)

type PluginItem struct {
	Name        string      `json:"name"`
	Version     string      `json:"version"`
	Skills      []SkillItem `json:"skills"`
	InstallPath string      `json:"installPath"`
	Disabled    bool        `json:"disabled"`
}

func listPlugins(ctx *AppContext) (any, error) {
	pluginsFile := filepath.Join(ctx.Cfg.ClaudeDir, "plugins", "installed_plugins.json")
	data, err := os.ReadFile(pluginsFile)
	if err != nil {
		return nil, nil
	}
	type pluginEntry struct {
		Version int `json:"version"`
		Plugins map[string][]struct {
			Scope       string `json:"scope"`
			InstallPath string `json:"installPath"`
			Version     string `json:"version"`
		} `json:"plugins"`
	}
	var registry pluginEntry
	if err := json.Unmarshal(data, &registry); err != nil {
		return nil, nil
	}
	var items []PluginItem
	for key, installs := range registry.Plugins {
		name := key
		if idx := strings.Index(key, "@"); idx >= 0 {
			name = key[:idx]
		}
		for _, inst := range installs {
			item := PluginItem{Name: name, Version: inst.Version, InstallPath: inst.InstallPath}
			skillsDir := filepath.Join(inst.InstallPath, "skills")
			if entries, err := os.ReadDir(skillsDir); err == nil {
				for _, e := range entries {
					if e.IsDir() {
						skillMD := filepath.Join(skillsDir, e.Name(), "SKILL.md")
						isDisabled := false
						if !fsutil.PathExists(skillMD) && fsutil.PathExists(skillMD+".disabled") {
							skillMD = skillMD + ".disabled"
							isDisabled = true
						}
						if fsutil.PathExists(skillMD) {
							si := SkillItem{Name: e.Name(), Type: "plugin", Status: "ok", Disabled: isDisabled}
							if data, err := os.ReadFile(skillMD); err == nil {
								if fmBytes, _, err := parser.ExtractFrontmatter(data); err == nil {
									if fm, err := parser.ParseSkillFrontmatter(fmBytes); err == nil {
										si.Description = fm.Description
										si.Name = fm.Name
										si.Invocation = "/" + fm.Name
										if !translate.IsMostlyChinese(fm.Description) {
											si.DescriptionCN = translate.TranslateDescription(fm.Description)
										}
									}
								}
							}
							item.Skills = append(item.Skills, si)
						}
					}
				}
			}
			allDisabled := len(item.Skills) > 0
			for _, s := range item.Skills {
				if !s.Disabled {
					allDisabled = false
					break
				}
			}
			item.Disabled = allDisabled
			items = append(items, item)
		}
	}
	return items, nil
}

func togglePluginSkill(ctx *AppContext, pluginName, skillName, installPath string) (any, error) {
	skillsDir := filepath.Join(installPath, "skills", skillName)
	enabled, err := toggleSkillFile(skillsDir)
	if err != nil {
		return nil, fmt.Errorf("插件 Skill 不存在")
	}
	if enabled {
		return "已启用: " + skillName, nil
	}
	return "已禁用: " + skillName, nil
}

func togglePlugin(ctx *AppContext, h *Handler, pluginName, installPath string) (any, error) {
	skillsDir := filepath.Join(installPath, "skills")
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		return nil, fmt.Errorf("无法读取插件目录")
	}
	hasEnabled := false
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if fsutil.PathExists(filepath.Join(skillsDir, e.Name(), "SKILL.md")) {
			hasEnabled = true
			break
		}
	}
	count := 0
	if hasEnabled {
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			if _, err := toggleSkillFile(filepath.Join(skillsDir, e.Name())); err == nil {
				count++
			}
		}
		h.Notify("config-changed", map[string]any{"domain": "skills"})
		return fmt.Sprintf("已禁用 %s 的 %d 个 Skill", pluginName, count), nil
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if _, err := toggleSkillFile(filepath.Join(skillsDir, e.Name())); err == nil {
			count++
		}
	}
	h.Notify("config-changed", map[string]any{"domain": "skills"})
	return fmt.Sprintf("已启用 %s 的 %d 个 Skill", pluginName, count), nil
}

// forEachPluginSkill reads installed_plugins.json and calls fn for each skill directory.
func forEachPluginSkill(ctx *AppContext, fn func(installPath, skillDirName string)) {
	pluginsFile := filepath.Join(ctx.Cfg.ClaudeDir, "plugins", "installed_plugins.json")
	data, err := os.ReadFile(pluginsFile)
	if err != nil {
		return
	}
	type pluginEntry struct {
		Plugins map[string][]struct {
			InstallPath string `json:"installPath"`
		} `json:"plugins"`
	}
	var registry pluginEntry
	if err := json.Unmarshal(data, &registry); err != nil {
		return
	}
	for _, installs := range registry.Plugins {
		for _, inst := range installs {
			skillsDir := filepath.Join(inst.InstallPath, "skills")
			entries, err := os.ReadDir(skillsDir)
			if err != nil {
				continue
			}
			for _, e := range entries {
				if e.IsDir() {
					fn(inst.InstallPath, e.Name())
				}
			}
		}
	}
}

func disableAllPlugins(ctx *AppContext, h *Handler) (any, error) {
	count := 0
	forEachPluginSkill(ctx, func(installPath, skillDirName string) {
		skillsDir := filepath.Join(installPath, "skills", skillDirName)
		if _, err := toggleSkillFile(skillsDir); err == nil {
			count++
		}
	})
	h.Notify("config-changed", map[string]any{"domain": "skills"})
	return fmt.Sprintf("已禁用 %d 个插件 Skill", count), nil
}

func enableAllPlugins(ctx *AppContext, h *Handler) (any, error) {
	count := 0
	forEachPluginSkill(ctx, func(installPath, skillDirName string) {
		skillsDir := filepath.Join(installPath, "skills", skillDirName)
		if _, err := toggleSkillFile(skillsDir); err == nil {
			count++
		}
	})
	h.Notify("config-changed", map[string]any{"domain": "skills"})
	return fmt.Sprintf("已启用 %d 个插件 Skill", count), nil
}
