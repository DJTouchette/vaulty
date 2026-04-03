# Vaulty

A local-only CLI daemon that acts as a secrets proxy for AI coding agents, so agents can make authenticated API calls without ever seeing raw credentials.

**The problem:** AI coding agents (Claude Code, Cursor, Windsurf) need API keys to call external services. Today, developers put secrets in env vars, `.env` files, or MCP configs — all readable by the agent. A prompt injection or malicious tool can exfiltrate them.

**Vaulty's approach:** Agents get capabilities, not credentials. Vaulty holds secrets in memory, proxies authenticated requests, and redacts any secret values from output. The agent never sees the raw key.

## Quick Start

```bash
# Install
go install github.com/djtouchette/vaulty/cmd/vaulty@latest

# Create a vault
vaulty init

# Add a secret with a domain policy
vaulty set STRIPE_SECRET_KEY --domains "api.stripe.com"

# Add a secret with a command policy
vaulty set DATABASE_URL --commands "psql,prisma,drizzle-kit"

# Start the daemon
vaulty start

# Make an authenticated API call (agent never sees the key)
vaulty proxy POST https://api.stripe.com/v1/charges \
  --secret STRIPE_SECRET_KEY \
  --header "Content-Type: application/x-www-form-urlencoded" \
  --body "amount=2000&currency=usd"

# Run a command with secrets in env (output is redacted)
vaulty exec --secret DATABASE_URL -- npx prisma migrate deploy

# List secrets (names and policies only, never values)
vaulty list

# Stop the daemon
vaulty stop
```

## MCP Server (Claude Code / Cursor)

Vaulty runs as an MCP server for direct integration with AI agents. Claude Code / Cursor spawns the process automatically — you don't need to run the server yourself.

### Quick setup

```bash
# In your project directory:
vaulty mcp init

# Or target a specific project:
vaulty mcp init --dir ~/work/my-project
```

This writes a `.mcp.json` file that Claude Code and Cursor discover automatically.

### Manual setup

Add to your project's `.mcp.json`:

```json
{
  "mcpServers": {
    "vaulty": {
      "command": "vaulty",
      "args": ["mcp", "start"]
    }
  }
}
```

If your passphrase is saved in the OS keychain (`vaulty keychain save`), no env vars are needed. Otherwise, add `VAULTY_PASSPHRASE`:

```json
{
  "mcpServers": {
    "vaulty": {
      "command": "vaulty",
      "args": ["mcp", "start"],
      "env": {
        "VAULTY_PASSPHRASE": "your-passphrase"
      }
    }
  }
}
```

For team mode, pass the identity file:

```json
{
  "mcpServers": {
    "vaulty": {
      "command": "vaulty",
      "args": ["mcp", "start", "-i", "/home/you/age-key.txt"]
    }
  }
}
```

The agent gets three tools:

| Tool | Description |
|------|-------------|
| `vaulty_request` | Make an authenticated HTTP request. Vaulty injects the credential. |
| `vaulty_exec` | Run a command with secrets as env vars. Output is redacted. |
| `vaulty_list` | List available secret names and policies (never values). |

## How It Works

```
┌──────────────────────────────────┐
│           AI Agent               │
│  Sees: secret names, policies,   │
│        redacted output           │
│  Never sees: raw secret values   │
└──────────┬───────────────────────┘
           │
┌──────────▼───────────────────────┐
│         Vaulty Daemon            │
│                                  │
│  1. Validate policy (domains,    │
│     commands)                    │
│  2. Inject secret into request   │
│  3. Execute & redact output      │
│  4. Log to audit trail           │
└──────────────────────────────────┘
```

- **Policy enforcement** — each secret has a domain allowlist (for HTTP) or command allowlist (for exec). Requests to unauthorized targets are denied.
- **Secret injection** — supports Bearer token, Basic auth, custom header, query parameter, or environment variable.
- **Output redaction** — raw, base64-encoded, and URL-encoded secret values are replaced with `[VAULTY:SECRET_NAME]` in all output.
- **Audit logging** — every request is logged (timestamp, action, target, status) to `~/.config/vaulty/audit.log`. Secret values are never logged.

## Config Reference

Vaulty uses a TOML config file (`vaulty.toml`) that stores policies but never secret values. Safe to commit.

Searched in: `./vaulty.toml`, `./.vaulty/vaulty.toml`, then `~/.config/vaulty/vaulty.toml`.

```toml
[vault]
path = "~/.config/vaulty/vault.age"
idle_timeout = "8h"
socket = "/tmp/vaulty.sock"
http_port = 19876                      # 0 to disable
notifications = true                   # desktop notifications on policy denials

[secrets.STRIPE_SECRET_KEY]
description = "Stripe live API key"    # shown to agents via vaulty_list
allowed_domains = ["api.stripe.com"]
inject_as = "bearer"                   # bearer | basic | header | query | env

[secrets.DATABASE_URL]
description = "Postgres connection string for the app database"
allowed_commands = ["psql", "prisma"]
inject_as = "env"
```

See [`vaulty.example.toml`](vaulty.example.toml) for a full annotated example.

## Team Sharing

Share a vault across a team using age identity keys instead of a shared passphrase. Each team member gets their own private key — no one needs to know anyone else's credentials.

### Setup (team lead)

```bash
# Each team member generates an age keypair
age-keygen -o ~/age-key.txt
# Outputs: Public key: age1abc...

# Lead adds each member's public key
vaulty team add age1abc...   # Alice
vaulty team add age1def...   # Bob

# From this point, the vault is encrypted for all recipients.
# The vault file, recipients list, and vaulty.toml can all be committed to git.
# Only each member's private key stays on their machine.
```

### Usage (team member)

```bash
# Use --identity (-i) flag to decrypt with your private key
vaulty list -i ~/age-key.txt
vaulty start -i ~/age-key.txt --foreground

# Or set the env var so you don't need the flag every time
export VAULTY_IDENTITY=~/age-key.txt
vaulty list
vaulty start
```

### MCP config (team member)

```json
{
  "mcpServers": {
    "vaulty": {
      "command": "vaulty",
      "args": ["mcp", "start", "-i", "/home/you/age-key.txt"]
    }
  }
}
```

### Managing recipients

```bash
vaulty team list              # show all public keys
vaulty team add age1xyz...    # add a new member
vaulty team remove age1xyz... # revoke access
```

When recipients are configured, the vault is encrypted using age X25519 keys. When no recipients exist, it falls back to passphrase-based encryption (the default single-user mode).

## CLI Reference

| Command | Description |
|---------|-------------|
| `vaulty init` | Create a new encrypted vault |
| `vaulty set <name>` | Add or update a secret |
| `vaulty list` | List secret names and policies |
| `vaulty remove <name>` | Remove a secret |
| `vaulty rotate <name>` | Rotate a secret's value |
| `vaulty start` | Start the daemon (decrypts vault into memory) |
| `vaulty stop` | Stop the daemon (zeroes secrets from memory) |
| `vaulty proxy <METHOD> <URL>` | Make an authenticated HTTP request |
| `vaulty exec -- <command>` | Run a command with secrets injected |
| `vaulty mcp start` | Start as an MCP server (stdio) |
| `vaulty mcp init` | Write `.mcp.json` for Claude Code / Cursor |
| `vaulty keychain save\|delete\|status` | Manage passphrase in OS keychain |
| `vaulty team add\|list\|remove` | Manage team recipients |
| `vaulty backend list\|secrets\|pull` | Manage cloud secret backends |
| `vaulty export --out <file>` | Export vault as encrypted snapshot |
| `vaulty import --from <file>` | Import secrets from encrypted snapshot |
| `vaulty export-env` | Export secrets as `.env` format |
| `vaulty import-env --from <file>` | Import secrets from a `.env` file |
| `vaulty export-docker` | Export secrets for Docker/Compose |
| `vaulty import-docker --from <file>` | Import env vars from `docker-compose.yml` |
| `vaulty export-k8s` | Export secrets as Kubernetes Secret manifest |
| `vaulty import-k8s --from <file>` | Import secrets from Kubernetes Secret |
| `vaulty export-rails` | Export secrets as Rails credentials YAML |
| `vaulty import-rails` | Import secrets from Rails encrypted credentials |

### Global flags

| Flag | Description |
|------|-------------|
| `-i, --identity <file>` | Age private key file for team vault decryption |
| `-V, --vault <name>` | Named vault to use (stored in `vaults/<name>.age`) |

## Security Model

- **Zero cloud.** Everything runs on your machine. No accounts, no SaaS, no telemetry.
- **Zero trust of agents.** Agents get capabilities, not credentials.
- **Encryption at rest.** Vault file is encrypted with [age](https://age-encryption.org/) (scrypt passphrase).
- **Encryption in memory.** Secrets stored as `[]byte`, explicitly zeroed on daemon stop.
- **Policy enforcement.** Every request validated against domain/command allowlists before execution.
- **Output redaction.** All stdout/stderr/response bodies filtered for secret values (raw, base64, URL-encoded).

## Install

```bash
# Go
go install github.com/djtouchette/vaulty/cmd/vaulty@latest

# From source
git clone https://github.com/djtouchette/vaulty.git
cd vaulty
make build
```

## License

MIT
