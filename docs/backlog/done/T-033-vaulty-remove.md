# T-033: `vaulty remove <name>` command

**Epic:** 3 — CLI Commands — Vault Management
**Status:** done
**Priority:** P0

## Description

Implement `internal/cli/remove.go` — remove a secret.

## Acceptance Criteria

- [ ] Prompts for confirmation (with `-y` to skip)
- [ ] Removes from vault file and `vaulty.toml`
- [ ] Errors if secret doesn't exist
