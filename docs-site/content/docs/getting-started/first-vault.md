---
title: "Your First Vault"
weight: 2
---

# Your First Vault

A vault is where your secrets live, encrypted at rest with [age](https://age-encryption.org/) encryption. Think of it as a password manager, but specifically designed so AI agents can *use* your secrets without *seeing* them.

## Initialize

```bash
vaulty init
```

This does three things:
1. Asks you for a passphrase (pick a good one — this is the master key to your secrets)
2. Creates an encrypted vault file at `~/.config/vaulty/vault.age`
3. Creates a starter config at `~/.config/vaulty/vaulty.toml`

**Pro tip:** Vaulty will offer to save your passphrase in your OS keychain (macOS Keychain, GNOME Keyring, Windows Credential Manager). Say yes unless you enjoy typing passphrases every time you start the daemon. Your fingers will thank you.

### Per-Project Vaults

Working on a project with its own secrets? Use `--local`:

```bash
cd my-project
vaulty init --local
```

This creates `.vaulty/vault.age` and `.vaulty/vaulty.toml` in your project directory. The vault file is gitignored (obviously), but the config file is safe to commit since it only contains policies, never secret values.

## Add Your First Secret

Let's add a Stripe API key:

```bash
vaulty set STRIPE_SECRET_KEY --domains "api.stripe.com"
```

Vaulty will prompt you to paste the secret value. It uses terminal raw mode, so your key won't echo to the screen — no one looking over your shoulder will see it, and it won't end up in your shell history. Paranoia is a feature, not a bug.

You can also pipe it in:

```bash
echo "sk_test_..." | vaulty set STRIPE_SECRET_KEY --domains "api.stripe.com"
```

The `--domains` flag creates a policy: this secret can *only* be used for requests to `api.stripe.com`. If your AI agent tries to send it to `evil-server.com`, Vaulty says no. Politely, but firmly.

### What About CLI Tools?

For secrets that get injected as environment variables (database URLs, AWS credentials), use `--commands` instead:

```bash
vaulty set DATABASE_URL --commands "psql,prisma,drizzle-kit"
```

This means `DATABASE_URL` can only be injected into those specific commands. Your agent can run `prisma migrate deploy` with the database URL, but it can't `curl` it somewhere sketchy.

## Check Your Work

```bash
vaulty list
```

You'll see your secret names and their policies. Never the values — Vaulty doesn't show those to anyone, not even you (okay, you can decrypt the vault file directly with `age` if you really need to, but Vaulty won't help you do it).

## Start the Daemon

```bash
vaulty start
```

This decrypts your vault into memory and starts listening for requests. The daemon serves on both a Unix socket (`/tmp/vaulty.sock`) and a localhost HTTP port (`localhost:19876`).

Now you're ready to proxy some requests. Head to [Your First Request]({{< relref "/docs/getting-started/first-request" >}}).
