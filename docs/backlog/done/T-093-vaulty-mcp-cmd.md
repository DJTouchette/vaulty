# T-093: `vaulty mcp` CLI command

**Epic:** 9 — MCP Server
**Status:** done
**Priority:** P0

## Description

Implement `internal/cli/mcp.go` — start vaulty in MCP server mode.

## Acceptance Criteria

- [ ] `vaulty mcp` starts MCP server on stdio
- [ ] Decrypts vault on startup (prompts for passphrase or reads from env)
- [ ] Works when added to Claude Code / Cursor MCP config
