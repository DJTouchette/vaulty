# T-151: MCP Secret Discovery Tools

**Epic:** 15 — MCP Enhancements
**Status:** done
**Priority:** P2

## Description

Expose richer MCP tools for agents to discover available secrets and policies without seeing secret values.

## Acceptance Criteria

- [x] vaulty_list_services tool returns configured service names, domains, injection modes
- [x] vaulty_check_access tool takes a URL and returns which secrets can access it
- [x] vaulty_secret_metadata tool returns secret names with descriptions and policies
- [x] No secret values exposed in any discovery tool output
- [x] Tests for list services, check access (matching and non-matching), metadata
