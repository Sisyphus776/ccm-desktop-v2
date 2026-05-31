package main

import (
	"log"
	"os"
	"time"

	"ccm-desktop-v2/backend/internal/config"
	"ccm-desktop-v2/backend/internal/fsutil"
	"ccm-desktop-v2/backend/rpc"
)

func main() {
	cfg := config.Resolve("")
	ctx := &rpc.AppContext{Cfg: cfg}
	h := rpc.NewHandler(make(chan map[string]any, 64))
	rpc.RegisterAll(h, ctx)

	// Start file change poller (every 3 seconds)
	go watchDirs(cfg, h)

	h.RunLoop()
}

// watchDirs polls watched directories and notifies the frontend when files change.
func watchDirs(cfg *config.Config, h *rpc.Handler) {
	type snapshot struct {
		modTime time.Time
	}
	prev := map[string]snapshot{}

	check := func(dir string, domain string) {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return
		}
		for _, e := range entries {
			info, err := e.Info()
			if err != nil {
				continue
			}
			key := dir + "/" + e.Name()
			s, ok := prev[key]
			if !ok || !info.ModTime().Equal(s.modTime) {
				prev[key] = snapshot{modTime: info.ModTime()}
				if ok {
					// File modified since last check
					h.Notify("config-changed", map[string]any{"domain": domain})
					return // one notification per domain per cycle is enough
				}
			}
		}
	}

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if fsutil.PathExists(cfg.SkillsDir) {
			check(cfg.SkillsDir, "skills")
		}
		pluginsDir := cfg.ClaudeDir + "\\plugins"
		if fsutil.PathExists(pluginsDir) {
			check(pluginsDir, "plugins")
		}
		// Check claude.json for MCP changes
		if fsutil.PathExists(cfg.ClaudeJSON) {
			check(filepathDir(cfg.ClaudeJSON), "mcp")
		}
	}

	log.Println("watchDirs: unexpected exit from ticker loop")
}

func filepathDir(p string) string {
	for i := len(p) - 1; i >= 0; i-- {
		if p[i] == '\\' || p[i] == '/' {
			return p[:i]
		}
	}
	return p
}

