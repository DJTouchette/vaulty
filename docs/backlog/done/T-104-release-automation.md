# T-104: Release automation

**Epic:** 10 — Distribution & Docs
**Status:** done
**Priority:** P1

## Description

GitHub Actions workflow for goreleaser — builds binaries for linux/mac/windows on tag push.

## Acceptance Criteria

- [ ] `.goreleaser.yml` config
- [ ] GitHub Actions workflow triggers on `v*` tags
- [ ] Produces binaries for linux/darwin/windows (amd64/arm64)
- [ ] Attaches binaries to GitHub release
