# T-041: Unix socket listener

**Epic:** 4 — Daemon
**Status:** done
**Priority:** P0

## Description

Implement `internal/daemon/socket.go` — listen on Unix socket for requests.

## Acceptance Criteria

- [ ] Listens on configurable socket path (default `/tmp/vaulty.sock`)
- [ ] Accepts JSON request/response over HTTP
- [ ] Cleans up socket file on shutdown
