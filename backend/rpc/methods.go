package rpc

import (
	"encoding/json"
	"ccm-desktop-v2/backend/internal/config"
)

type AppContext struct {
	Cfg *config.Config
}

func RegisterAll(h *Handler, ctx *AppContext) {
	h.Register("skills.list", func(p json.RawMessage) (any, error) {
		return listSkills(ctx)
	})
	h.Register("skills.validate", func(p json.RawMessage) (any, error) {
		return validateSkills(ctx)
	})
	h.Register("skills.toggle", func(p json.RawMessage) (any, error) {
		var params struct{ Name string }
		json.Unmarshal(p, &params)
		return toggleSkill(ctx, params.Name)
	})
	h.Register("skills.create", func(p json.RawMessage) (any, error) {
		var params struct{ Name, Description string }
		json.Unmarshal(p, &params)
		return createSkill(ctx, params.Name, params.Description)
	})
	h.Register("skills.delete", func(p json.RawMessage) (any, error) {
		var params struct{ Name string }
		json.Unmarshal(p, &params)
		return deleteSkill(ctx, params.Name)
	})
	h.Register("skills.import", func(p json.RawMessage) (any, error) {
		var params struct{ URL string }
		json.Unmarshal(p, &params)
		return importSkill(ctx, params.URL)
	})
	h.Register("skills.get_errors", func(p json.RawMessage) (any, error) {
		var params struct{ Name string }
		json.Unmarshal(p, &params)
		return getSkillErrors(ctx, params.Name)
	})
	h.Register("plugins.list", func(p json.RawMessage) (any, error) {
		return listPlugins(ctx)
	})
	h.Register("plugins.toggle_plugin", func(p json.RawMessage) (any, error) {
		var params struct{ Name, InstallPath string }
		json.Unmarshal(p, &params)
		return togglePlugin(ctx, params.Name, params.InstallPath)
	})
	h.Register("plugins.toggle_skill", func(p json.RawMessage) (any, error) {
		var params struct{ PluginName, SkillName, InstallPath string }
		json.Unmarshal(p, &params)
		return togglePluginSkill(ctx, params.PluginName, params.SkillName, params.InstallPath)
	})
	h.Register("plugins.disable_all", func(p json.RawMessage) (any, error) {
		return disableAllPlugins(ctx)
	})
	h.Register("plugins.enable_all", func(p json.RawMessage) (any, error) {
		return enableAllPlugins(ctx)
	})
	h.Register("memory.list", func(p json.RawMessage) (any, error) { return listMemory(ctx) })
	h.Register("memory.stats", func(p json.RawMessage) (any, error) { return getMemoryStats(ctx) })
	h.Register("memory.validate", func(p json.RawMessage) (any, error) { return validateMemory(ctx) })
	h.Register("memory.create", func(p json.RawMessage) (any, error) {
		var params struct{ Name, Type, Description, Content string }
		json.Unmarshal(p, &params)
		return createMemory(ctx, params.Name, params.Type, params.Description, params.Content)
	})
	h.Register("memory.get_content", func(p json.RawMessage) (any, error) {
		var params struct{ File string }
		json.Unmarshal(p, &params)
		return getMemoryContent(ctx, params.File)
	})
	h.Register("memory.delete", func(p json.RawMessage) (any, error) {
		var params struct{ File string }
		json.Unmarshal(p, &params)
		return deleteMemory(ctx, params.File)
	})
	h.Register("mcp.list", func(p json.RawMessage) (any, error) { return listMCP(ctx) })
	h.Register("mcp.validate", func(p json.RawMessage) (any, error) { return validateMCP(ctx) })
	h.Register("mcp.toggle", func(p json.RawMessage) (any, error) {
		var params struct{ Name string }
		json.Unmarshal(p, &params)
		return toggleMCP(ctx, params.Name)
	})
	h.Register("claudemd.list", func(p json.RawMessage) (any, error) { return listClaudeMD(ctx) })
	h.Register("claudemd.get_content", func(p json.RawMessage) (any, error) {
		var params struct{ Path string }
		json.Unmarshal(p, &params)
		return getClaudeMDContent(ctx, params.Path)
	})
	h.Register("portability.report", func(p json.RawMessage) (any, error) { return getPortabilityReport(ctx) })
	h.Register("portability.fix_path", func(p json.RawMessage) (any, error) {
		var params struct{ OldPath, NewPath string }
		json.Unmarshal(p, &params)
		return fixPath(ctx, params.OldPath, params.NewPath)
	})
	h.Register("secrets.scan", func(p json.RawMessage) (any, error) { return scanSecrets(ctx) })
	h.Register("backup.create", func(p json.RawMessage) (any, error) {
		var params struct{ OutputPath string }
		json.Unmarshal(p, &params)
		return createBackup(ctx, params.OutputPath)
	})
	h.Register("backup.restore", func(p json.RawMessage) (any, error) {
		var params struct{ ZipPath string; Force bool }
		json.Unmarshal(p, &params)
		return restoreBackup(ctx, params.ZipPath, params.Force)
	})
	h.Register("translate.batch", func(p json.RawMessage) (any, error) {
		go translateAll(ctx, h)
		return "started", nil
	})
	h.Register("settings.get", func(p json.RawMessage) (any, error) { return getSettings(ctx) })
	h.Register("settings.set_autostart", func(p json.RawMessage) (any, error) {
		var params struct{ Enabled bool }
		json.Unmarshal(p, &params)
		return setAutoStart(ctx, params.Enabled)
	})
	h.Register("dashboard.summary", func(p json.RawMessage) (any, error) { return getDashboardSummary(ctx) })
}
