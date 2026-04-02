# T-040: Daemon lifecycle (start/stop/signal handling)

**Epic:** 4 — Daemon
**Status:** done
**Priority:** P0

## Description

Implement `internal/daemon/daemon.go` — daemon process management.

## Acceptance Criteria

- [ ] Daemonizes (or runs in foreground with `--foreground`)
- [ ] Writes PID file
- [ ] Handles SIGTERM/SIGINT gracefully (zeroes memory before exit)
- [ ] `vaulty stop` reads PID file and sends signal
- [ ] Idle timeout auto-lock (configurable, default 8h)
