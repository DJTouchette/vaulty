# T-020: TOML config parser

**Epic:** 2 — Config & Policy
**Status:** done
**Priority:** P0

## Description

Implement `internal/policy/config.go` — parse `vaulty.toml` into Go structs. Handle vault settings and per-secret policy definitions.

## Acceptance Criteria

- [ ] Parses the example `vaulty.toml` from PRD
- [ ] Supports `[vault]` section (path, idle_timeout, socket, http_port)
- [ ] Supports `[secrets.<NAME>]` sections (allowed_domains, allowed_commands, inject_as, header_name, also_inject)
- [ ] Returns typed config struct
- [ ] Unit tests with sample TOML
