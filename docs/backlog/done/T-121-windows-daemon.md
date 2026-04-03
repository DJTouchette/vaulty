# T-121: Windows daemon support (HTTP-only, no Unix socket)

**Epic:** 12 — Platform Compatibility
**Status:** done
**Priority:** P1

## Description

On Windows, Unix sockets are not available. The daemon should fall back to HTTP-only mode. Signal handling needs to use `os.Interrupt` instead of `syscall.SIGTERM`.

Split `internal/daemon/daemon.go` into platform-specific files using build tags:
- `daemon_unix.go` — Unix socket listener, SIGTERM handling
- `daemon_windows.go` — HTTP-only, os.Interrupt handling, named mutex for PID liveness

## Acceptance Criteria

- [ ] `vaulty start` works on Windows using HTTP listener only
- [ ] `vaulty stop` works on Windows (sends interrupt or uses named mutex)
- [ ] Socket-related code gated behind `//go:build !windows`
- [ ] No compilation errors on `GOOS=windows`
