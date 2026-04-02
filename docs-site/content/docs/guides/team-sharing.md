---
title: "Team Sharing"
weight: 3
---

# Team Sharing

Solo vaults are great, but eventually you'll need to share secrets with your team. The obvious approach — sharing a passphrase over Slack — makes security people cry. Vaulty uses a better approach: age identity keys.

## How It Works

Instead of everyone knowing the same passphrase, each team member gets their own cryptographic keypair:

- **Public key** — shared openly, used to encrypt the vault for that person
- **Private key** — stays on their machine, used to decrypt

The vault gets encrypted for *all* recipients. Anyone with their private key can decrypt it, but no one needs to know anyone else's credentials. It's like a safety deposit box with multiple keyholes.

## Setup (Team Lead)

### 1. Generate Keypairs

Each team member generates their own age keypair:

```bash
# Install age if you haven't
# brew install age (macOS) / apt install age (Ubuntu)

age-keygen -o ~/age-key.txt
# Output: Public key: age1abc123...
```

Share the public key (the `age1...` part). Keep the private key file safe.

### 2. Add Recipients

The team lead adds each member's public key:

```bash
vaulty team add age1abc123...   # Alice
vaulty team add age1def456...   # Bob
vaulty team add age1ghi789...   # Charlie
```

From this point on, the vault is re-encrypted for all recipients every time a secret is added or updated.

### 3. Commit the Shared Files

These are safe to commit:

- `vaulty.toml` — policies only, no secrets
- `.vaulty/recipients` — public keys (they're *public*, that's the whole point)

This is **not** safe to commit:
- `vault.age` — encrypted, but still, keep it in `.gitignore`
- Anyone's private key file — obviously

## Usage (Team Members)

### Decrypt with Identity File

```bash
# Use the -i flag
vaulty list -i ~/age-key.txt
vaulty start -i ~/age-key.txt

# Or set the env var to skip the flag
export VAULTY_IDENTITY=~/age-key.txt
vaulty list
vaulty start
```

### MCP Config for Team Members

```json
{
  "mcpServers": {
    "vaulty": {
      "command": "vaulty",
      "args": ["mcp", "-i", "/home/you/age-key.txt"]
    }
  }
}
```

No passphrase needed — the identity file does the work.

## Managing Recipients

```bash
# See who has access
vaulty team list

# Add someone new (vault gets re-encrypted)
vaulty team add age1newperson...

# Remove someone (vault gets re-encrypted without them)
vaulty team remove age1formerperson...
```

When you remove a recipient, the vault is re-encrypted without their key. They can still decrypt any *copies* they already have, so you should rotate any sensitive secrets after removing someone. (This is true of any secret-sharing system, not just Vaulty.)

## Switching Modes

- **No recipients configured** → passphrase mode (default, single-user)
- **Recipients configured** → identity key mode (team)

You can switch between modes by adding your first recipient (`vaulty team add`) or removing all recipients.

## FAQ

**Q: Can I still use a passphrase with team mode?**
A: Team mode uses age identity keys instead of a passphrase. The two are mutually exclusive — when recipients are configured, Vaulty uses X25519 encryption instead of scrypt.

**Q: What if someone loses their private key?**
A: They lose access to the vault. Another team member who still has access can re-add them with a new public key. This is why it's good to have at least two people with access.

**Q: Can I use this with CI/CD?**
A: Technically yes — generate a keypair for your CI system, add the public key as a recipient, and provide the private key as a CI secret. But for CI, you might be better off using Vaulty's [cloud backends]({{< relref "/docs/guides/backends" >}}) to pull secrets from AWS Secrets Manager or similar.
