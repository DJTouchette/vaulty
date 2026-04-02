---
title: "Configuration"
weight: 2
---

# Configuration Reference

Vaulty uses a TOML config file (`vaulty.toml`) that stores **policies only** — never secret values. It's safe to commit to version control.

## Config File Location

Vaulty searches for config in this order:

1. `./vaulty.toml` — project root
2. `./.vaulty/vaulty.toml` — project `.vaulty` directory
3. `~/.config/vaulty/vaulty.toml` — user config directory

The first one found wins. Use `--config <path>` to override.

## Full Reference

```toml
# ─── Vault Settings ───────────────────────────────────

[vault]
# Path to the encrypted vault file
path = "~/.config/vaulty/vault.age"

# Auto-lock daemon after this idle period (0 = never)
idle_timeout = "8h"

# Unix socket path for daemon communication
socket = "/tmp/vaulty.sock"

# Localhost HTTP port (0 to disable HTTP listener)
http_port = 19876

# Desktop notifications on policy denials
notifications = false


# ─── Secret Policies ──────────────────────────────────

# Each [secrets.<NAME>] block defines access policy for a secret.
# The actual secret value lives in the encrypted vault, NOT here.

[secrets.STRIPE_SECRET_KEY]
# Human-readable purpose (shown in vaulty list and MCP vaulty_list)
description = "Stripe live API key"

# Domains this secret can be sent to (empty = any domain)
allowed_domains = ["api.stripe.com"]

# Commands this secret can be injected into (empty = any command)
# allowed_commands = []

# How to inject the secret:
#   "bearer" → Authorization: Bearer <secret>
#   "basic"  → Authorization: Basic <base64(secret)>
#   "header" → Custom header (set header_name)
#   "query"  → Append ?key=<secret> to URL
#   "env"    → Set as environment variable
inject_as = "bearer"

# Custom header name (only used when inject_as = "header")
# header_name = "X-API-Key"

# Other secrets to inject alongside this one
# also_inject = ["COMPANION_SECRET"]

# Which named vault this secret belongs to (for multi-vault setups)
# vault = "production"

# Skip manual approval in MCP mode (default: false)
# auto_approve = false

# Cloud backend this secret was pulled from (informational)
# backend = "aws_prod"


[secrets.DATABASE_URL]
description = "Postgres connection string"
allowed_commands = ["psql", "prisma", "drizzle-kit", "pg_dump"]
inject_as = "env"


# ─── Cloud Backends ───────────────────────────────────

[backends.aws_prod]
type = "aws"                    # aws | gcp | hashicorp | onepassword
region = "us-east-1"
profile = "production"

[backends.gcp_dev]
type = "gcp"
project = "my-gcp-project"

[backends.vault_corp]
type = "hashicorp"
addr = "https://vault.company.com"
mount = "secret"

[backends.onepass]
type = "onepassword"
```

## Secret Policy Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `description` | string | `""` | Human-readable purpose |
| `allowed_domains` | string[] | `[]` | Domain allowlist for HTTP proxy (empty = any) |
| `allowed_commands` | string[] | `[]` | Command allowlist for exec (empty = any) |
| `inject_as` | string | `"bearer"` | Injection mode: `bearer`, `basic`, `header`, `query`, `env` |
| `header_name` | string | `""` | Custom header name (when `inject_as = "header"`) |
| `also_inject` | string[] | `[]` | Companion secrets to inject alongside |
| `vault` | string | `""` | Named vault (for multi-vault setups) |
| `auto_approve` | bool | `false` | Skip MCP approval prompt |
| `backend` | string | `""` | Cloud backend name (informational) |

## Injection Modes Explained

### `bearer`
Adds an `Authorization: Bearer <secret>` header. The most common mode for REST APIs.

### `basic`
Adds an `Authorization: Basic <base64(secret)>` header. For APIs using HTTP Basic auth (like Twilio).

### `header`
Adds a custom header with the secret as the value. Requires `header_name` to be set. For APIs that use non-standard auth headers (like `X-API-Key`).

### `query`
Appends `?key=<secret>` to the URL. For APIs that accept keys as query parameters (like some Google APIs). Use sparingly — query params can leak into server logs.

### `env`
Injects the secret as an environment variable in the child process. Used with `vaulty exec` for CLI tools that read configuration from env vars.

## Policy Enforcement Rules

- **Empty `allowed_domains`** = secret can be used with *any* domain (be careful)
- **Empty `allowed_commands`** = secret can be used with *any* command (be careful)
- Both empty = no restrictions. Fine for development, risky for production.
- Domain matching is exact (no wildcards, no subdomains). `api.stripe.com` does not match `evil.api.stripe.com`.
- Command matching checks if any `allowed_commands` entry appears as a substring of the command being run. `prisma` matches `npx prisma migrate deploy`.

## Environment Variables

| Variable | Description |
|----------|-------------|
| `VAULTY_PASSPHRASE` | Vault passphrase (avoids interactive prompt) |
| `VAULTY_IDENTITY` | Path to age identity file (for team mode) |
| `VAULTY_CONFIG` | Path to config file (overrides search order) |
