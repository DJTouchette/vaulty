# T-152: MCP Resource Support

**Epic:** 15 — MCP Enhancements
**Status:** done
**Priority:** P2

## Description

Implement MCP resources endpoint so clients can read Vaulty metadata (secret names, policy config, audit log) through the standard MCP resources protocol.

## Acceptance Criteria

- [x] ResourceHandler with ListResources and ReadResource
- [x] vaulty://secrets resource returns secret names (no values)
- [x] vaulty://policy resource returns active policy as TOML
- [x] vaulty://audit resource returns last 50 lines of audit log
- [x] Server advertises resources capability in initialize response
- [x] Server dispatches resources/list and resources/read
- [x] CLI updated to pass ResourceHandler to NewServer
- [x] Tests for resource list, read each type, unknown URI error
