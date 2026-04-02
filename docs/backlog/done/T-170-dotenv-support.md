# T-170: .env file support

**Epic:** 17 — Framework Secret Files
**Status:** done
**Priority:** P1

## Description

Add .env file parsing and writing to Vaulty. Includes `import-env` and `export-env` CLI commands for importing secrets from .env files into the vault and exporting vault secrets in .env format.

## Acceptance Criteria

- [x] ParseDotenv handles KEY=value, comments, blank lines, quoted values, export prefix
- [x] WriteDotenv outputs sorted keys with redaction support
- [x] `vaulty import-env <file>` CLI command with --prefix flag
- [x] `vaulty export-env` CLI command with --out, --reveal, --format flags
- [x] .gitignore warning when importing
- [x] Comprehensive unit tests
- [x] Example project in examples/dotenv-node/
