# T-172: Docker/Compose secrets support

**Epic:** 17 — Framework Secret Files
**Status:** done
**Priority:** P1

## Description

Add Docker Compose environment variable parsing and export to Vaulty. Includes compose override generation, Docker secret file writing, and `import-docker`/`export-docker` CLI commands.

## Acceptance Criteria

- [x] ParseComposeEnv extracts environment variables from docker-compose.yml (mapping and list syntax)
- [x] WriteComposeOverride generates docker-compose.override.yml fragments
- [x] WriteSecretFiles writes each secret as a separate file
- [x] `vaulty import-docker <compose-file>` CLI command
- [x] `vaulty export-docker` CLI command with --out, --service, --secrets-dir flags
- [x] Comprehensive unit tests
- [x] Example project in examples/docker-compose/
