# T-030: `vaulty init` command

**Epic:** 3 — CLI Commands — Vault Management
**Status:** done
**Priority:** P0

## Description

Implement `internal/cli/init.go` — create a new vault.

## Acceptance Criteria

- [ ] Prompts for passphrase (no echo)
- [ ] Creates encrypted vault file at configured path
- [ ] Creates default `vaulty.toml` if none exists
- [ ] Errors if vault already exists (with `--force` override)
