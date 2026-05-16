package rpc

import (
	"os"

	"ccm-desktop-v2/backend/internal/claudemd"
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
