# T-120: Cross-platform compatibility audit

**Epic:** 12 — Platform Compatibility
**Status:** done
**Priority:** P1

## Description

Audit and fix all features to work correctly on Linux, macOS, and Windows. Currently several subsystems assume Unix-only behavior.

## Known Issues

### Daemon (Unix socket + signals)
- `internal/daemon/daemon.go` uses Unix sockets (`net.Listen("unix", ...)`) — not available on Windows
- Signal handling uses `syscall.SIGTERM` / `syscall.SIGINT` — Windows doesn't support SIGTERM
- Socket cleanup (`os.Remove(socketPath)`) is Unix-specific
- `internal/cli/stop.go` sends `syscall.SIGTERM` to stop daemon

**Fix:** Use named pipes on Windows, or fall back to HTTP-only mode. Use `os.Interrupt` instead of `SIGTERM` on Windows. Gate Unix-specific code behind build tags.

### Executor (shell invocation)
- `internal/executor/executor.go` shells out via `sh -c` — not available on Windows
- Tests assume `/tmp` exists

**Fix:** Use `cmd /C` on Windows. Use `os.TempDir()` in tests.

### Vault (path expansion)
- `internal/vault/vault.go` `expandPath()` checks for `~/` prefix — Windows uses `%USERPROFILE%`
- `internal/audit/logger.go` has its own `expandPath()` with same issue

**Fix:** Both already call `os.UserHomeDir()` which is cross-platform, but the `~/` prefix detection is fine since Windows paths won't start with `~/`. Verify this works.

### File permissions
- Multiple `os.MkdirAll(..., 0700)` and `os.WriteFile(..., 0600)` calls — Windows ignores Unix permission bits
- `os.Chmod(socketPath, 0600)` is a no-op on Windows

**Fix:** Acceptable — these are no-ops on Windows but don't error. Document that Windows relies on NTFS ACLs instead.

### Keyring
- `go-keyring` supports all three platforms (macOS Keychain, GNOME Keyring/KWallet, Windows Credential Manager)
- Verify it works on each platform in CI

### PID file
- `internal/daemon/daemon.go` uses `os.FindProcess` + `process.Signal(nil)` to check liveness — behavior differs on Windows (FindProcess always succeeds)

**Fix:** On Windows, use a lock file or named mutex instead of PID + signal check.

## Acceptance Criteria

- [ ] Build tags separate Unix-specific and Windows-specific code
- [ ] `vaulty start` / `vaulty stop` work on Windows (HTTP-only mode, no Unix socket)
- [ ] `vaulty exec` works on Windows (uses `cmd /C` instead of `sh -c`)
- [ ] `vaulty proxy` works on all platforms (pure HTTP, should already work)
- [ ] `vaulty keychain` works on all platforms
- [ ] `vaulty mcp` works on all platforms (stdio, should already work)
- [ ] CI matrix tests on ubuntu, macos, windows
- [ ] Tests don't hardcode Unix paths (`/tmp`, etc.)
