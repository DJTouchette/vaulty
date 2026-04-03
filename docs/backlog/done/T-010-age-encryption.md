# T-010: Age encryption/decryption module

**Epic:** 1 — Vault & Encryption
**Status:** done
**Priority:** P0

## Description

Implement `internal/vault/encrypt.go` — passphrase-based age encryption and decryption of arbitrary `[]byte` payloads using `filippo.io/age`.

## Acceptance Criteria

- [ ] `Encrypt(passphrase, plaintext) -> ciphertext` works
- [ ] `Decrypt(passphrase, ciphertext) -> plaintext` roundtrips correctly
- [ ] Wrong passphrase returns clear error
- [ ] Unit tests for encrypt/decrypt roundtrip and wrong-passphrase case
