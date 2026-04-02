# T-000: Initialize Go module and project structure

**Epic:** 0 — Project Bootstrap
**Status:** done
**Priority:** P0

## Description

Set up `go.mod`, directory layout (`cmd/vaulty/`, `internal/`), and a minimal `main.go` with cobra root command that prints version.

## Acceptance Criteria

- [ ] `go build ./cmd/vaulty` compiles
- [ ] `vaulty --version` prints version string
- [ ] Directory structure matches PRD module layout
