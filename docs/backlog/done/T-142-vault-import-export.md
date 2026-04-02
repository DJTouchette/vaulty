# T-142: Vault Import/Export

**Epic:** 14 — Multi-Vault Support
**Status:** done
**Priority:** P1

## Description

Add `vaulty export` and `vaulty import` CLI commands for transferring vault secrets between environments. Export creates an encrypted snapshot; import merges secrets into an existing vault.

## Acceptance Criteria

- [x] `vault.Export()` serializes and encrypts vault secrets
- [x] `vault.Import()` decrypts and returns a vault (supports passphrase and identity file)
- [x] `vault.MergeVaults()` copies secrets with optional overwrite
- [x] `vaulty export --out <file>` command works
- [x] `vaulty import --from <file> [--overwrite]` command works
- [x] Tests cover round-trip export/import, identity-based import, wrong passphrase, merge with/without overwrite, empty merge
