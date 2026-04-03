# T-080: Audit logger

**Epic:** 8 — Audit Logging
**Status:** done
**Priority:** P0

## Description

Implement `internal/audit/logger.go` — append-only audit log.

## Acceptance Criteria

- [ ] Writes JSONL format to configurable path (default `~/.config/vaulty/audit.log`)
- [ ] Logs: timestamp, action (proxy/exec/denied), secret_name, target/command, status/exit_code, reason (for denials)
- [ ] Never logs secret values
- [ ] File-append only, no overwrites
- [ ] Concurrent-write safe (mutex)
- [ ] Unit test for log entry format
