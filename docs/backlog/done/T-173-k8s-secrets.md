# T-173: Kubernetes secrets support

**Epic:** 17 — Framework Secret Files
**Status:** done
**Priority:** P1

## Description

Add Kubernetes Secret manifest parsing and generation to Vaulty. Includes base64 encoding/decoding and `import-k8s`/`export-k8s` CLI commands.

## Acceptance Criteria

- [x] ParseK8sSecret parses K8s Secret YAML and base64-decodes values
- [x] WriteK8sSecret generates K8s Secret YAML with base64-encoded values
- [x] `vaulty import-k8s <manifest.yaml>` CLI command
- [x] `vaulty export-k8s` CLI command with --name, --namespace, --out flags
- [x] Round-trip test (write then parse back)
- [x] Comprehensive unit tests
