# T-122: Windows command executor (`cmd /C` instead of `sh -c`)

**Epic:** 12 — Platform Compatibility
**Status:** done
**Priority:** P1

## Description

`internal/executor/executor.go` uses `exec.Command("sh", "-c", command)` which fails on Windows. Use build tags or runtime detection to use `cmd /C` on Windows.

## Acceptance Criteria

- [ ] `vaulty exec` works on Windows
- [ ] `executor.go` uses `cmd /C` on Windows, `sh -c` on Unix
- [ ] Tests pass on Windows (use `os.TempDir()` not `/tmp`)
