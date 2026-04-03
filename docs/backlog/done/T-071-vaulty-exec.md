# T-071: `vaulty exec` command

**Epic:** 7 — CLI Commands — Proxy & Exec
**Status:** done
**Priority:** P0

## Description

Implement `internal/cli/exec.go` — CLI interface for command executor.

## Acceptance Criteria

- [ ] `vaulty exec --secret <NAME> [--secret <NAME2>] -- <command> [args...]`
- [ ] Runs command through daemon
- [ ] Streams redacted output to terminal
- [ ] Returns command's exit code as process exit code
