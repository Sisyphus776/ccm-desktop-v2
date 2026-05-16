package main

import (
	"ccm-desktop-v2/backend/internal/config"
	"ccm-desktop-v2/backend/rpc"
)

func main() {
	cfg := config.Resolve("")
	ctx := &rpc.AppContext{Cfg: cfg}
	h := rpc.NewHandler(make(chan map[string]any, 64))
	rpc.RegisterAll(h, ctx)
	h.RunLoop()
}
