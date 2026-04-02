# T-140: Vault Namespaces

**Epic:** 14 — Multi-Vault Support
**Status:** done
**Priority:** P1

## Description

Add a `--vault <name>` global flag that lets users target named vaults. Named vaults are stored under `~/.config/vaulty/vaults/<name>.age` (global) or `.vaulty/vaults/<name>.age` (per-project). The default vault (no name) continues to work as today.

## Acceptance Criteria

- [x] `ResolveVaultPath(name, basePath)` returns correct paths for named and default vaults
- [x] `--vault` / `-V` persistent flag added to root command
- [x] All CLI commands (set, list, remove, rotate, start, mcp) resolve vault path via the flag
- [x] Tests cover vault path resolution and named vault CRUD
