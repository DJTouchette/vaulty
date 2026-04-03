# T-052: Redaction engine

**Epic:** 5 — HTTP Proxy & Secret Injection
**Status:** done
**Priority:** P0

## Description

Implement `internal/proxy/redactor.go` — redact secret values from output.

## Acceptance Criteria

- [ ] Replaces raw secret value with `[VAULTY:SECRET_NAME]`
- [ ] Replaces base64-encoded secret with `[VAULTY:SECRET_NAME:b64]`
- [ ] Replaces URL-encoded secret with `[VAULTY:SECRET_NAME:url]`
- [ ] Handles multiple secrets in same output
- [ ] Streaming-capable (for stdout/stderr pipes)
- [ ] Unit tests with various encodings and edge cases
