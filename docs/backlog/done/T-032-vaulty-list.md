# T-032: `vaulty list` command

**Epic:** 3 — CLI Commands — Vault Management
**Status:** done
**Priority:** P0

## Description

Implement `internal/cli/list.go` — list secret names and policies.

## Acceptance Criteria

- [ ] Shows name, allowed domains, allowed commands for each secret
- [ ] Never shows secret values
- [ ] Clean tabular output
- [ ] Works without daemon running (reads vault directly)
