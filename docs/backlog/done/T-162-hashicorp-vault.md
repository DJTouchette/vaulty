# T-162: HashiCorp Vault Backend

**Epic:** 16 — Cloud Provider Integrations
**Status:** done
**Priority:** P1

## Description

Add a HashiCorp Vault backend that shells out to the `vault` CLI to list and retrieve secrets. Supports KV v2 with configurable mount path and VAULT_ADDR.

## Acceptance Criteria

- [x] `HashiCorpBackend` struct implements `SecretBackend` interface
- [x] Uses `exec.Command` to call `vault kv list` and `vault kv get`
- [x] Sets `VAULT_ADDR` environment variable for commands
- [x] Defaults mount to "secret" if not specified
- [x] Tries `-field=value` first, falls back to full JSON for multi-field secrets
- [x] Validates `vault` CLI availability with actionable error
- [x] Command argument construction is tested
- [x] CLI availability error is tested
