# AGENTS.md

Go MCP server for TypeSafe Jev. Jev API key only.

## Commands

- `go test -race -shuffle=on -count=1 ./...`
- `go build -o bin/jev-mcp ./cmd/jev-mcp`

## Architecture

- `cmd/jev-mcp/main.go` — load env, run stdio MCP
- `internal/config` — `TYPESAFE_API_KEY`, `JEV_MCP_MODEL`, `TYPESAFE_BASE_URL`
- `internal/jev` — `POST /v1/systemone`
- `internal/policy` — thresholds and verdict mapping
- `internal/mcp` — MCP tools

## Tools

- `jev_ask` — raw noul / choice / score
- `jev_verify` — claims against evidence
- `jev_screen` — injection / substance / relevance
- `jev_find` — rank candidates plus exists check
- `jev_pick` — one candidate or none
- `jev_extract` — one verbatim span or none
- `jev_classify` — batch items against a class catalog

## Conventions

- `context.Context` first on every public I/O function
- Errors wrapped with `%w`
- 250 pure LOC ceiling per file
- No other LLM provider
