# T-090: MCP server (stdio transport)

**Epic:** 9 — MCP Server
**Status:** done
**Priority:** P0

## Description

Implement `internal/mcp/server.go` — MCP server lifecycle and stdio JSON-RPC transport.

## Acceptance Criteria

- [ ] Reads JSON-RPC from stdin, writes to stdout
- [ ] Handles `initialize`, `tools/list`, `tools/call` methods
- [ ] Proper MCP protocol compliance (capabilities, protocol version)
- [ ] Unit test with mock stdin/stdout
