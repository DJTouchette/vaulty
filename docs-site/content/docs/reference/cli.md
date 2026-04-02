---
title: "CLI Reference"
weight: 1
---

# CLI Reference

Every command Vaulty offers, in alphabetical order. For when you need the exact syntax and not just the general idea.

## Global Flags

These work with any command:

| Flag | Description |
|------|-------------|
| `-i, --identity <file>` | Age private key file for team vault decryption |
| `--config <path>` | Path to config file (default: auto-detected) |
| `-h, --help` | Help for any command |

## Commands

### `vaulty init`

Create a new encrypted vault and config file.

```bash
vaulty init [flags]
```

| Flag | Description |
|------|-------------|
| `--local` | Create vault in `.vaulty/` (project-local) instead of `~/.config/vaulty/` |
| `--force` | Overwrite existing vault and config |

**What it does:**
- Prompts for a passphrase
- Creates the encrypted vault file
- Creates a starter `vaulty.toml`
- Offers to save the passphrase in your OS keychain

---

### `vaulty set <name>`

Add or update a secret.

```bash
vaulty set SECRET_NAME [flags]
```

| Flag | Description |
|------|-------------|
| `--value <string>` | Secret value (if not provided, prompts interactively) |
| `--domains <list>` | Comma-separated domain allowlist |
| `--commands <list>` | Comma-separated command allowlist |
| `--description <text>` | Human-readable description |

**Examples:**

```bash
# Interactive (prompts for value)
vaulty set STRIPE_SECRET_KEY --domains "api.stripe.com"

# Non-interactive
vaulty set API_KEY --value "sk_live_..." --domains "api.example.com"

# Piped
echo "my-secret-value" | vaulty set API_KEY --domains "api.example.com"

# CLI tool secret
vaulty set DATABASE_URL --commands "psql,prisma" --description "Production DB"
```

---

### `vaulty list`

List all secrets with their names and policies. Never shows values.

```bash
vaulty list
```

Output looks like:

```
STRIPE_SECRET_KEY
  domains: api.stripe.com
  inject:  bearer

DATABASE_URL
  commands: psql, prisma
  inject:   env
  desc:     Production database
```

---

### `vaulty remove <name>`

Remove a secret from the vault.

```bash
vaulty remove SECRET_NAME
```

Prompts for confirmation before removing. Because mistakes happen and we care about your feelings.

---

### `vaulty rotate <name>`

Replace a secret's value with a new one.

```bash
vaulty rotate SECRET_NAME
```

Prompts for the new value. Policies stay the same — just the value changes. The old value is zeroed from memory.

---

### `vaulty start`

Start the Vaulty daemon.

```bash
vaulty start [flags]
```

| Flag | Description |
|------|-------------|
| `--foreground` | Run in foreground (don't daemonize) |
| `--vaults <list>` | Load multiple named vaults |

**What it does:**
- Decrypts the vault into memory
- Starts a Unix socket listener at `/tmp/vaulty.sock`
- Starts an HTTP listener at `localhost:19876`
- Writes a PID file for `vaulty stop`

---

### `vaulty stop`

Stop the daemon and zero all secrets from memory.

```bash
vaulty stop
```

The secrets don't just get garbage collected — they're explicitly overwritten with zeros before the process exits. Paranoia as a service.

---

### `vaulty proxy <METHOD> <URL>`

Make an authenticated HTTP request through the daemon.

```bash
vaulty proxy METHOD URL [flags]
```

| Flag | Description |
|------|-------------|
| `--secret <name>` | Which secret to inject |
| `--header <key: value>` | Additional request headers (repeatable) |
| `--body <string>` | Request body |

**Examples:**

```bash
# GET request with Bearer token
vaulty proxy GET https://api.stripe.com/v1/balance \
  --secret STRIPE_SECRET_KEY

# POST with body
vaulty proxy POST https://api.stripe.com/v1/charges \
  --secret STRIPE_SECRET_KEY \
  --header "Content-Type: application/x-www-form-urlencoded" \
  --body "amount=2000&currency=usd"
```

---

### `vaulty exec -- <command>`

Run a command with secrets injected as environment variables.

```bash
vaulty exec [--secret NAME]... -- command [args...]
```

| Flag | Description |
|------|-------------|
| `--secret <name>` | Secret to inject as env var (repeatable) |

**Examples:**

```bash
# Single secret
vaulty exec --secret DATABASE_URL -- npx prisma migrate deploy

# Multiple secrets
vaulty exec --secret DATABASE_URL --secret REDIS_URL -- npm run start

# AWS (with companion secrets)
vaulty exec --secret AWS_SECRET_ACCESS_KEY -- aws s3 ls
```

The `--` is required to separate Vaulty flags from the command.

---

### `vaulty mcp`

Start Vaulty as an MCP server (stdio JSON-RPC).

```bash
vaulty mcp [flags]
```

| Flag | Description |
|------|-------------|
| `-i, --identity <file>` | Age identity file (for team mode) |

Designed to be called by AI agents — not directly by humans. See [MCP Setup]({{< relref "/docs/guides/mcp-setup" >}}).

---

### `vaulty keychain`

Manage passphrase storage in the OS keychain.

```bash
vaulty keychain save      # store passphrase
vaulty keychain status    # check if stored
vaulty keychain delete    # remove stored passphrase
```

---

### `vaulty team`

Manage team recipients for shared vault access.

```bash
vaulty team add <public-key>      # add a recipient
vaulty team list                   # show all recipients
vaulty team remove <public-key>   # revoke access
```

---

### `vaulty backend`

Manage cloud secret backends.

```bash
vaulty backend list                    # list configured backends
vaulty backend secrets <backend>       # browse secrets
vaulty backend pull <backend> <name>   # import a secret
```

---

### `vaulty import-env <file>`

Import secrets from a `.env` file.

```bash
vaulty import-env .env
vaulty import-env .env.production
```

### `vaulty export-env`

Export secrets as `.env` format.

```bash
vaulty export-env > .env.local
```

---

### `vaulty import-rails` / `vaulty export-rails`

Import/export Rails encrypted credentials.

### `vaulty import-docker` / `vaulty export-docker`

Import/export Docker Compose secrets.

### `vaulty import-k8s` / `vaulty export-k8s`

Import/export Kubernetes Secret manifests.

---

### `vaulty export` / `vaulty import`

Transfer vaults between machines.

```bash
vaulty export > backup.age
vaulty import < backup.age
```
