# T-141: Vault Switching in Daemon

**Epic:** 14 — Multi-Vault Support
**Status:** done
**Priority:** P1

## Description

Extend the daemon to hold multiple vaults simultaneously. Requests can target a specific vault via the `vault` field in the request JSON, or fall back to the vault specified in the secret's policy, or use the default vault.

## Acceptance Criteria

- [x] `Request` struct has `Vault` field
- [x] `Daemon` holds `map[string]*vault.Vault` instead of single vault
- [x] `SecretPolicy` has `Vault` field for policy-level vault binding
- [x] `handleProxy`, `handleExec`, `handleList` resolve vault per-request
- [x] `--vaults` flag on `start` command loads additional named vaults
- [x] Combined redactor covers secrets from all loaded vaults
- [x] Tests cover multi-vault list, exec, and missing vault error
