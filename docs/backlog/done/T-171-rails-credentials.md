# T-171: Rails credentials support

**Epic:** 17 — Framework Secret Files
**Status:** done
**Priority:** P1

## Description

Add Rails encrypted credentials parsing to Vaulty. Includes YAML flattening/unflattening, AES-256-GCM decryption, and `import-rails`/`export-rails` CLI commands.

## Acceptance Criteria

- [x] ParseRailsCredentials parses YAML and flattens to SECTION_KEY format
- [x] FlattenYAML recursively flattens nested maps
- [x] WriteRailsCredentials unflattens keys back to nested YAML
- [x] DecryptRailsCredentials handles Rails 7+ AES-256-GCM format
- [x] `vaulty import-rails` CLI command with --env flag
- [x] `vaulty export-rails` CLI command with --out flag
- [x] Supports RAILS_MASTER_KEY env var and config/master.key
- [x] Comprehensive unit tests including encryption round-trip
- [x] Example project in examples/rails-credentials/
