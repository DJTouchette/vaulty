# T-021: Policy validator

**Epic:** 2 — Config & Policy
**Status:** done
**Priority:** P0

## Description

Implement `internal/policy/validator.go` — validate a request against a secret's policy.

## Acceptance Criteria

- [ ] `ValidateDomain(secretName, targetURL)` checks URL host against allowed_domains
- [ ] `ValidateCommand(secretName, command)` checks command against allowed_commands
- [ ] Wildcard policy (empty allowlist) permits all
- [ ] Returns clear error message on denial
- [ ] Unit tests for allow, deny, and wildcard cases
