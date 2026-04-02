# T-042: Localhost HTTP listener

**Epic:** 4 — Daemon
**Status:** done
**Priority:** P0

## Description

Implement `internal/daemon/http.go` — fallback HTTP listener on localhost.

## Acceptance Criteria

- [ ] Listens on `127.0.0.1:<port>` (configurable, default 19876)
- [ ] Binds only to loopback (never `0.0.0.0`)
- [ ] Same request/response format as Unix socket
- [ ] Can be disabled (`http_port = 0`)
