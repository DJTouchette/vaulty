# Vaulty

A local-only CLI daemon that acts as a secrets proxy for AI coding agents, so agents can make authenticated API calls without ever seeing raw credentials.

## Build & Test

```bash
go build ./cmd/vaulty          # build the binary
go test ./... -count=1         # run all tests
go test ./internal/vault/...   # run tests for a specific package
go vet ./...                   # static analysis
```

Single binary, no Docker, no runtime dependencies. The binary entrypoint is `cmd/vaulty/main.go`.

## Architecture

- **Go module:** `github.com/djtouchette/vaulty`
- **CLI framework:** cobra (`github.com/spf13/cobra`)
- **Encryption:** age (`filippo.io/age`) with scrypt passphrase + multi-recipient X25519
- **Config:** TOML (`github.com/pelletier/go-toml/v2`)
- **Keyring:** OS keychain (`github.com/zalando/go-keyring`)

### Package layout

```
cmd/vaulty/main.go         CLI entrypoint
internal/cli/               cobra command implementations (init, set, list, remove, rotate, start, stop, proxy, exec, mcp, keychain, team)
internal/daemon/             daemon lifecycle, socket/HTTP listeners, request router, client
internal/vault/              age encryption, vault CRUD, in-memory store with zeroing, keyring, team recipients
internal/policy/             TOML config parsing, domain/command policy validation
internal/proxy/              HTTP proxy, secret injection (bearer/basic/header/query), redaction engine
internal/executor/           child process execution with env injection + redacted output
internal/mcp/                MCP server (stdio JSON-RPC), tool definitions, request handlers
internal/audit/              append-only JSONL audit logger
```

### Key design decisions

- Secrets are stored as `[]byte` and explicitly zeroed on shutdown — never left dangling in memory.
- The redaction engine catches raw, base64-encoded, and URL-encoded secret values in all output streams.
- Policy enforcement happens before every action: domain allowlists for HTTP proxy, command allowlists for exec.
- The daemon serves requests over both Unix socket (primary) and localhost HTTP (fallback). Both use the same JSON request/response protocol defined in `internal/daemon/protocol.go`.
- The MCP server uses stdio transport (JSON-RPC over stdin/stdout) for direct integration with Claude Code, Cursor, etc.

## Config

Config file is `vaulty.toml` — searched in `./vaulty.toml`, `./.vaulty/vaulty.toml`, then `~/.config/vaulty/vaulty.toml`. It stores policies (allowed domains, allowed commands, injection mode) but **never secret values**. Safe to commit.

Vault file is `~/.config/vaulty/vault.age` (or `.vaulty/vault.age` for per-project vaults) — age-encrypted JSON containing the actual secrets. **Never commit this.**

Policy templates for common APIs live in `templates/` (Stripe, OpenAI, AWS, etc.).

## Backlog / Ticket Workflow

Tickets live in `docs/backlog/` as markdown files. Open tickets are organized by epic directory. Completed tickets are moved to `docs/backlog/done/`.

```
docs/backlog/
├── done/                          # completed tickets (moved here when done)
│   ├── T-000-init-go-module.md
│   ├── T-010-age-encryption.md
│   └── ...
├── epic-00-bootstrap/             # open tickets grouped by epic
│   └── T-001-ci-setup.md
├── epic-10-distribution/
│   ├── T-100-readme.md
│   └── ...
└── epic-11-post-mvp/
    ├── T-110-os-keychain.md
    └── ...
```

### Ticket format

Each ticket is a markdown file with this structure:

```markdown
# T-XXX: Title

**Epic:** N — Epic Name
**Status:** todo | in-progress | done
**Priority:** P0 | P1 | P2 | P3

## Description

What to build and why.

## Acceptance Criteria

- [ ] Criterion 1
- [ ] Criterion 2
```

### Working with tickets

1. **Pick a ticket** — browse open epic directories in `docs/backlog/` for `todo` tickets.
2. **Start work** — set `**Status:** in-progress` in the ticket file.
3. **Complete work** — set `**Status:** done`, ensure tests pass.
4. **Move to done/** — move the ticket file to `docs/backlog/done/`.
5. If an epic directory becomes empty after moving all its tickets, delete the empty directory.

- Before starting work, read the relevant ticket file to understand scope and acceptance criteria.
- If a ticket needs to be split, create new ticket files in the same epic directory.
- Ticket IDs follow the pattern `T-{epic}{seq}` (e.g., T-050 = epic 5, first ticket).
- The full PRD is at `docs/prd.md` for context on product decisions.

### Current state

All 33 tickets across epics 0–11 are complete. All tickets are in `docs/backlog/done/`.

## Style

- No unnecessary abstractions — three similar lines is better than a premature helper.
- Tests use `testing` stdlib only, no test frameworks.
- Error messages should be actionable (e.g., "wrong passphrase?" not just "decryption failed").
- Never log, print, or include secret values in error messages or stack traces.
