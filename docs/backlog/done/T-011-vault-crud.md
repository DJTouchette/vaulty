# T-011: Vault CRUD operations

**Epic:** 1 — Vault & Encryption
**Status:** done
**Priority:** P0

## Description

Implement `internal/vault/vault.go` — create, read, write vault file. Vault stores a JSON map of `name -> secret_value`, encrypted with age.

## Acceptance Criteria

- [ ] `Create(path, passphrase)` creates new empty vault file
- [ ] `Open(path, passphrase)` decrypts and returns secrets map
- [ ] `Save(path, passphrase, secrets)` encrypts and writes vault file
- [ ] `Set(name, value)`, `Get(name)`, `Remove(name)`, `List()` on open vault
- [ ] Unit tests for all CRUD operations
