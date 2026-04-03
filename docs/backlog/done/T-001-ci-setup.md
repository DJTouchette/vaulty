# T-001: CI setup (lint, test, build)

**Epic:** 0 — Project Bootstrap
**Status:** done
**Priority:** P1

## Description

Add a `Makefile` or `justfile` with `build`, `test`, `lint` targets. Optionally add GitHub Actions workflow.

## Acceptance Criteria

- [ ] `make build` produces a binary
- [ ] `make test` runs tests
- [ ] `make lint` runs `golangci-lint`
