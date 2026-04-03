# T-160: AWS Secrets Manager Backend

**Epic:** 16 — Cloud Provider Integrations
**Status:** done
**Priority:** P1

## Description

Add an AWS Secrets Manager backend that shells out to the `aws` CLI to list and retrieve secrets. Part of the backend abstraction layer for external secret providers.

## Acceptance Criteria

- [x] `AWSBackend` struct implements `SecretBackend` interface
- [x] Uses `exec.Command` to call `aws secretsmanager list-secrets` and `get-secret-value`
- [x] Supports `--region` and `--profile` flags
- [x] Validates `aws` CLI availability via `exec.LookPath` with actionable error
- [x] Command argument construction is tested
- [x] CLI availability error is tested
