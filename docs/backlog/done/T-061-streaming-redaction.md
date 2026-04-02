# T-061: Streaming output with redaction

**Epic:** 6 — Command Executor
**Status:** done
**Priority:** P0

## Description

Implement `internal/executor/stream.go` — pipe stdout/stderr through redactor.

## Acceptance Criteria

- [ ] Streams stdout and stderr separately
- [ ] Applies redaction filter to both streams
- [ ] If agent runs `echo $SECRET`, output shows `[VAULTY:SECRET_NAME]`
- [ ] Unit test for streaming redaction
