# T-161: GCP Secret Manager Backend

**Epic:** 16 — Cloud Provider Integrations
**Status:** done
**Priority:** P1

## Description

Add a GCP Secret Manager backend that shells out to the `gcloud` CLI to list and retrieve secrets. Extracts secret names from full resource paths.

## Acceptance Criteria

- [x] `GCPBackend` struct implements `SecretBackend` interface
- [x] Uses `exec.Command` to call `gcloud secrets list` and `gcloud secrets versions access`
- [x] Supports `--project` flag
- [x] Parses full resource paths to extract secret names
- [x] Validates `gcloud` CLI availability with actionable error
- [x] Command argument construction is tested
- [x] CLI availability error is tested
