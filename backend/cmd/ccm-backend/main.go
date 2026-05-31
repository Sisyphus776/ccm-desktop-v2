package main

import (
	"encoding/json"
	"os"

	"ccm-desktop-v2/internal/config"
	"ccm-desktop-v2/backend/rpc"
)

func main() {
	cfg := config.Resolve("")
	ctx := &rpc.AppContext{Cfg: cfg}

	// stdout-based notify for Electron mode
	emitFn := func(method string, params map[string]any) {
		b, _ := json.Marshal(map[string]any{
			"jsonrpc": "2.0",
			"method":  method,
			"params":  params,
		})
		os.Stdout.Write(b)
		os.Stdout.Write([]byte{'\n'})
	}

	h := rpc.NewHandler(emitFn)
	rpc.RegisterAll(h, ctx)
	h.RunLoop()
}
