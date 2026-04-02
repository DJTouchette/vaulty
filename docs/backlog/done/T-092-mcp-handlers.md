# T-092: MCP request handlers

**Epic:** 9 — MCP Server
**Status:** done
**Priority:** P0

## Description

Implement `internal/mcp/handler.go` — handle tool calls by delegating to proxy/executor/vault.

## Acceptance Criteria

- [ ] `vaulty_request` -> HTTP proxy with policy check
- [ ] `vaulty_exec` -> command executor with policy check
- [ ] `vaulty_list` -> returns secret names + policies
- [ ] Returns structured MCP tool results
- [ ] Errors are MCP-compliant
