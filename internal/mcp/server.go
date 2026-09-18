// Package mcp registers the Jev MCP tools.
package mcp

import (
	"github.com/fast-facts/jev-mcp/internal/config"
	"github.com/fast-facts/jev-mcp/internal/jev"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const Version = "0.1.0"

type handlers struct {
	client *jev.Client
}

func New(cfg config.Config) *mcp.Server {
	h := &handlers{client: jev.New(cfg.BaseURL, cfg.APIKey, cfg.Model)}
	s := mcp.NewServer(&mcp.Implementation{Name: "jev-mcp", Version: Version}, nil)
	mcp.AddTool(s, &mcp.Tool{Name: "jev_ask", Description: askDesc}, h.ask)
	mcp.AddTool(s, &mcp.Tool{Name: "jev_verify", Description: verifyDesc}, h.verify)
	mcp.AddTool(s, &mcp.Tool{Name: "jev_screen", Description: screenDesc}, h.screen)
	mcp.AddTool(s, &mcp.Tool{Name: "jev_find", Description: findDesc}, h.find)
	mcp.AddTool(s, &mcp.Tool{Name: "jev_pick", Description: pickDesc}, h.pick)
	mcp.AddTool(s, &mcp.Tool{Name: "jev_extract", Description: extractDesc}, h.extract)
	mcp.AddTool(s, &mcp.Tool{Name: "jev_classify", Description: classifyDesc}, h.classify)
	return s
}
