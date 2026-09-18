# jev-mcp

Go MCP server for TypeSafe Jev. Jev answers typed Choice, Score, and Noul questions about a state. It does not write prose. Policy lives in this server; Jev only returns probabilities.

## Run

Set `TYPESAFE_API_KEY`. The process speaks MCP on stdin and stdout.

Local:

```bash
go run ./cmd/jev-mcp
```

From GitHub, no clone:

```bash
go run github.com/fast-facts/jev-mcp/cmd/jev-mcp@latest
```

### MCP client

OpenCode (`opencode.json`):

```json
{
  "mcp": {
    "jev": {
      "type": "local",
      "command": ["go", "run", "github.com/fast-facts/jev-mcp/cmd/jev-mcp@latest"],
      "environment": {
        "TYPESAFE_API_KEY": "ts_..."
      }
    }
  }
}
```

Other clients (`mcpServers`):

```json
{
  "mcpServers": {
    "jev": {
      "command": "go",
      "args": ["run", "github.com/fast-facts/jev-mcp/cmd/jev-mcp@latest"],
      "env": {
        "TYPESAFE_API_KEY": "ts_..."
      }
    }
  }
}
```

Go must be on your `PATH`. First start compiles; later starts reuse the module cache.

### Binary

```bash
go build -o bin/jev-mcp ./cmd/jev-mcp
```

Then point the client at `bin/jev-mcp`, or put it on your `PATH` and use `"command": ["jev-mcp"]`.

## Tools

| Tool | Job |
| --- | --- |
| `jev_verify` | Check claims against evidence. Verdicts: verified, contradicted, unsupported. |
| `jev_screen` | Judge untrusted text before an agent reads it. Actions: pass, review, block, skip. |
| `jev_find` | Rank candidates for a query. Also says if any candidate answers it. |
| `jev_pick` | Pick at most one candidate, or none. Use when a ranked winner is not enough. |
| `jev_extract` | Pick one exact span from a shortlist. The span is copied, never invented. |
| `jev_classify` | Assign each item to one class from a shared list. |
| `jev_ask` | Send your own noul, choice, and score questions when no other tool fits. |

## Env

| Name | Default | Role |
| --- | --- | --- |
| `TYPESAFE_API_KEY` | none | Required. Bearer key for `https://api.typesafe.ai`. |
| `JEV_MCP_MODEL` | `jev-latest` | Model id. |
| `TYPESAFE_BASE_URL` | `https://api.typesafe.ai` | Origin only. The client adds `/v1/systemone`. |

## Tests

```bash
go test -race -shuffle=on -count=1 ./...
```

Unit tests need no API key. Live MCP tests skip without `TYPESAFE_API_KEY`.

## License

MIT
