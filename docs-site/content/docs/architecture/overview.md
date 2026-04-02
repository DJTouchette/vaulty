---
title: "System Overview"
weight: 1
---

# System Overview

Vaulty is a single Go binary that wears several hats depending on how you invoke it. Here's the big picture.

## The Architecture (in One Diagram)

```
┌─────────────────────────────────────────────────────────┐
│                     AI Agent                            │
│  (Claude Code, Cursor, Windsurf, or anything MCP)       │
│                                                         │
│  Knows: secret names, policies, redacted output         │
│  Doesn't know: secret values (not even a little)        │
└──────────┬──────────────────────┬───────────────────────┘
           │ MCP (stdio)          │ CLI
           │                      │
┌──────────▼──────────┐  ┌───────▼────────────────────────┐
│    MCP Server        │  │     Daemon                     │
│  (vaulty mcp)        │  │  (vaulty start)                │
│                      │  │                                │
│  JSON-RPC 2.0        │  │  Unix Socket (/tmp/vaulty.sock)│
│  over stdin/stdout   │  │  HTTP (localhost:19876)        │
│                      │  │                                │
│  Inline vault        │  │  Long-running background       │
│  decryption          │  │  process                       │
└──────────┬──────────┘  └───────┬────────────────────────┘
           │                      │
           └──────────┬───────────┘
                      │
         ┌────────────▼────────────────┐
         │     Request Router          │
         │                             │
         │  1. Parse request           │
         │  2. Look up secret + policy │
         │  3. Validate policy         │
         │  4. Route to handler        │
         └──────┬──────────┬───────────┘
                │          │
    ┌───────────▼──┐  ┌───▼───────────────┐
    │  HTTP Proxy  │  │  Command Executor  │
    │              │  │                    │
    │  Inject auth │  │  Inject env vars   │
    │  Make request│  │  Spawn process     │
    │  Redact resp │  │  Redact output     │
    └──────┬───────┘  └────────┬──────────┘
           │                    │
    ┌──────▼────────────────────▼──────────┐
    │           Audit Logger               │
    │  Append-only JSONL                   │
    │  Never contains secret values        │
    └──────────────────────────────────────┘
```

## Two Modes, Same Core

### Daemon Mode (`vaulty start`)

The daemon is a long-running background process:
- Decrypts the vault once on startup
- Holds secrets in memory
- Listens for requests on a Unix socket and HTTP
- Used by the CLI commands (`vaulty proxy`, `vaulty exec`)
- Auto-locks after configurable idle timeout
- Explicitly zeroes secrets on stop

### MCP Mode (`vaulty mcp`)

The MCP server is a single-session process:
- Launched by the AI agent (via MCP config)
- Decrypts the vault inline on startup
- Communicates over stdin/stdout (JSON-RPC 2.0)
- Lives and dies with the agent session
- No separate daemon needed

Both modes share the same core: vault decryption, policy enforcement, secret injection, output redaction, and audit logging. The only difference is the transport layer.

## Package Layout

```
cmd/vaulty/main.go         # CLI entrypoint (Cobra)
internal/
├── cli/                    # Command implementations
├── daemon/                 # Daemon lifecycle + protocol
├── vault/                  # Age encryption + CRUD
├── policy/                 # TOML config + validation
├── proxy/                  # HTTP proxy + injection + redaction
├── executor/               # Child process + env injection
├── mcp/                    # MCP server (JSON-RPC)
├── audit/                  # Append-only logger
├── backend/                # Cloud provider integrations
└── framework/              # Import/export utilities
```

Each package has a single responsibility. There are no circular dependencies. The `internal/` prefix means these packages aren't importable by external code — Vaulty is a tool, not a library.

## Key Design Decisions

### Why a Daemon?

The daemon holds decrypted secrets in memory so you don't have to type your passphrase for every request. It also means multiple CLI invocations share the same decrypted state. The alternative — decrypting for every request — would be secure but annoying.

### Why Unix Socket + HTTP?

The Unix socket is the primary transport — it's faster and has OS-level permission enforcement (only your user can connect). The HTTP listener is a fallback for tools that can't talk to Unix sockets. Both use the same JSON protocol defined in `internal/daemon/protocol.go`.

### Why Age?

Age is:
- Simple (one file format, clear semantics)
- Audited (by Cure53)
- Written by Filippo Valsorda (Go crypto team lead)
- Not GPG (which is a feature, not a bug)

It supports both passphrase-based encryption (scrypt) and public-key encryption (X25519), which maps perfectly to Vaulty's single-user and team modes.

### Why `[]byte` Instead of `string`?

Go strings are immutable — when you "overwrite" a string, you just create a new one and the old one sits in memory until the garbage collector gets around to it. Byte slices can be explicitly zeroed in place. For secrets, that distinction matters.

### Why TOML for Config?

TOML is:
- Human-readable and writable
- Expressive enough for nested config (like per-secret policies)
- Not YAML (no implicit type coercion, no "Norway problem")
- Not JSON (allows comments)

The perfect middle ground for a config file you want humans to edit and version control to track.
