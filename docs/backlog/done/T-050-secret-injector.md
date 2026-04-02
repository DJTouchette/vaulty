# T-050: Secret injector

**Epic:** 5 — HTTP Proxy & Secret Injection
**Status:** done
**Priority:** P0

## Description

Implement `internal/proxy/injector.go` — inject secrets into HTTP requests.

## Acceptance Criteria

- [ ] `bearer`: sets `Authorization: Bearer <secret>`
- [ ] `basic`: sets `Authorization: Basic <base64(secret)>`
- [ ] `header`: sets custom header to secret value
- [ ] `query`: appends secret as query parameter
- [ ] Unit tests for each injection mode
