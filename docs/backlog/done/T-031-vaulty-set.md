# T-031: `vaulty set <name>` command

**Epic:** 3 — CLI Commands — Vault Management
**Status:** done
**Priority:** P0

## Description

Implement `internal/cli/set.go` — add/update secrets.

## Acceptance Criteria

- [ ] Interactive mode: prompts for value (no echo) and policy (domains, commands)
- [ ] Flag mode: `--value`, `--domains`, `--commands` flags
- [ ] Pipe mode: reads value from stdin
- [ ] Updates `vaulty.toml` with policy (never the secret value)
