# T-043: Request router

**Epic:** 4 — Daemon
**Status:** done
**Priority:** P0

## Description

Route incoming requests (from socket/HTTP) to the appropriate action executor, after policy validation.

## Acceptance Criteria

- [ ] Routes `proxy` requests to HTTP proxy
- [ ] Routes `exec` requests to command executor
- [ ] Routes `list` requests to secret lister
- [ ] Runs policy validation before executing
- [ ] Returns structured error on policy denial
