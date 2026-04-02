# T-012: In-memory secret store with zeroing

**Epic:** 1 — Vault & Encryption
**Status:** done
**Priority:** P0

## Description

Implement `internal/vault/memory.go` — holds decrypted secrets in `[]byte` slices with explicit zeroing on close.

## Acceptance Criteria

- [ ] Store secrets as `[]byte`, not `string`
- [ ] `Zero()` method overwrites all byte slices with zeros
- [ ] Unit test confirms memory is zeroed after close
