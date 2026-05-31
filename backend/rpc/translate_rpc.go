package rpc

import (
	"ccm-desktop-v2/internal/skills"
	"ccm-desktop-v2/internal/translate"
)

// setTranslateConfig stores Baidu API credentials.
func setTranslateConfig(appID, secretKey string) string {
	translate.SetAPIConfig(appID, secretKey)
	return "百度翻译 API 配置已保存"
}

// TranslateAll scans all skills and plugins, translates un-cached English descriptions,
// and emits translation-ready notifications as each completes.
func TranslateAll(ctx *AppContext, h *Handler) {
	// Translate skills
	skillList, err := skills.List(ctx.Cfg)
	if err == nil {
		for _, sk := range skillList {
			if sk.Frontmatter == nil || sk.Frontmatter.Description == "" {
				continue
			}
			if translate.IsMostlyChinese(sk.Frontmatter.Description) {
				continue
			}
			cn := translate.TranslateDescription(sk.Frontmatter.Description)
			if cn != "" {
				h.Notify("translation-ready", map[string]any{
					"domain":        "skills",
					"name":          sk.Frontmatter.Name,
					"descriptionCN": cn,
				})
			}
		}
	}

	// Translate plugin skills
	pluginsRaw, err := listPlugins(ctx)
	if err == nil && pluginsRaw != nil {
		plugins, ok := pluginsRaw.([]PluginItem)
		if ok {
			for _, p := range plugins {
			for _, s := range p.Skills {
				if s.Description == "" || translate.IsMostlyChinese(s.Description) {
					continue
				}
				cn := translate.TranslateDescription(s.Description)
				if cn != "" {
					h.Notify("translation-ready", map[string]any{
						"domain":        "plugins",
						"pluginName":    p.Name,
						"name":          s.Name,
						"descriptionCN": cn,
					})
				}
				}
			}
		}
	}
}
