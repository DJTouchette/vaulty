# T-060: Command executor

**Epic:** 6 — Command Executor
**Status:** done
**Priority:** P0

## Description

Implement `internal/executor/executor.go` — spawn child process with secrets in env.

## Acceptance Criteria

- [ ] Spawns command with specified secrets injected as env vars
- [ ] Inherits current env plus secret env vars
- [ ] Returns exit code
- [ ] Unit test with simple command
