# T-070: `vaulty proxy` command

**Epic:** 7 — CLI Commands — Proxy & Exec
**Status:** done
**Priority:** P0

## Description

Implement `internal/cli/proxy.go` — CLI interface for HTTP proxy.

## Acceptance Criteria

- [ ] `vaulty proxy <METHOD> <URL> --secret <NAME> [--header K:V] [--body <data>]`
- [ ] Sends request through daemon (via socket/HTTP)
- [ ] Prints response body to stdout (redacted)
- [ ] Prints status code to stderr
- [ ] Errors clearly if daemon not running
