package main

import (
	"context"
	"log"

	"github.com/fast-facts/jev-mcp/internal/config"
	mcpserver "github.com/fast-facts/jev-mcp/internal/mcp"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	if err := mcpserver.New(cfg).Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}
