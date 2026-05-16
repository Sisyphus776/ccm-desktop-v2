package rpc

// translateAll is a stub that signals translation has started.
// Full implementation requires wiring the translate package to scan all skills and plugins,
// then emit translation-ready notifications for each un-cached English description.
// For now, the background goroutine simply completes immediately.
func translateAll(ctx *AppContext, h *Handler) {
	// TODO: iterate skills.List(ctx.Cfg) and listPlugins(ctx),
	// call translate.TranslateDescription for each un-cached description,
	// then h.Notify("translation-ready", ...) per item.
}
