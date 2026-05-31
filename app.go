package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"ccm-desktop-v2/internal/config"
	"ccm-desktop-v2/internal/fsutil"
	"ccm-desktop-v2/backend/rpc"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx     context.Context
	cfg     *config.Config
	handler *rpc.Handler
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.cfg = config.Resolve("")
	appCtx := &rpc.AppContext{Cfg: a.cfg}

	// Wails event-based emit func
	emitFn := func(method string, params map[string]any) {
		runtime.EventsEmit(ctx, method, params)
	}

	a.handler = rpc.NewHandler(emitFn)
	rpc.RegisterAll(a.handler, appCtx)

	// Start file change poller
	go a.watchDirs()
}

func (a *App) shutdown(ctx context.Context) {}

// Call dispatches a JSON-RPC request and returns the JSON-RPC response.
// This is the main bridge — the frontend's rpc-client.ts calls this.
func (a *App) Call(method string, paramsJSON string) string {
	var params json.RawMessage
	if paramsJSON != "" && paramsJSON != "{}" {
		params = json.RawMessage(paramsJSON)
	} else {
		params = json.RawMessage("{}")
	}
	req := rpc.Request{
		JSONRPC: "2.0",
		ID:      1,
		Method:  method,
		Params:  params,
	}
	line, _ := json.Marshal(req)
	resp := a.handler.HandleRaw(line)
	if resp == nil {
		return `{"jsonrpc":"2.0","id":1,"result":null}`
	}
	return string(resp)
}

// TranslateBatch triggers async translation of all skill/plugin descriptions.
func (a *App) TranslateBatch() {
	go rpc.TranslateAll(&rpc.AppContext{Cfg: a.cfg}, a.handler)
}

// watchDirs polls skill/plugin/config directories for changes.
func (a *App) watchDirs() {
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
					runtime.EventsEmit(a.ctx, "config-changed", map[string]any{"domain": domain})
					return
				}
			}
		}
	}

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if a.ctx == nil {
			continue
		}
		if fsutil.PathExists(a.cfg.SkillsDir) {
			check(a.cfg.SkillsDir, "skills")
		}
		pluginsDir := a.cfg.ClaudeDir + "\\plugins"
		if fsutil.PathExists(pluginsDir) {
			check(pluginsDir, "plugins")
		}
	}
	log.Println("watchDirs: exited ticker loop")
}
