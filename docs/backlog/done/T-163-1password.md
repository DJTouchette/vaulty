# T-163: 1Password Backend

**Epic:** 16 — Cloud Provider Integrations
**Status:** done
**Priority:** P1

## Description

Add a 1Password backend that shells out to the `op` CLI to list and retrieve secrets. Uses `op://` URI scheme for reading, with fallback to `op item get`.

## Acceptance Criteria

- [x] `OnePasswordBackend` struct implements `SecretBackend` interface
- [x] Uses `exec.Command` to call `op item list` and `op read`
- [x] Supports vault name filtering
- [x] Falls back to `op item get --fields password` if `op read` fails
- [x] Validates `op` CLI availability with actionable error
- [x] Command argument construction is tested
- [x] CLI availability error is tested
