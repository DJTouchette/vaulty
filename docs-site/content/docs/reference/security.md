---
title: "Security Model"
weight: 3
---

# Security Model

This is the page you show to your security team, your CISO, or that one friend who always asks "but is it *really* secure?" Yes. Here's how.

## Threat Model

Vaulty protects against one specific, increasingly common threat:

> **An AI coding agent (or tool it invokes) exfiltrates secrets that it has access to.**

This can happen via:
- **Prompt injection** — malicious content in a webpage, repo, or API response that instructs the agent to leak secrets
- **Rogue MCP tools** — a malicious or compromised tool that reads environment variables and sends them somewhere
- **Accidental exposure** — the agent logs, prints, or includes secret values in generated code

Vaulty's answer: **the agent never has the secret in the first place.** It has a name and a policy, not a value.

## Defense in Depth

### 1. Zero Knowledge for Agents

The AI agent interacts with Vaulty through three tools:
- `vaulty_list` — returns secret names and policies, never values
- `vaulty_request` — Vaulty makes the HTTP call, the agent gets a redacted response
- `vaulty_exec` — Vaulty runs the command, the agent gets redacted output

At no point does the agent receive, handle, or process raw secret material.

### 2. Policy Enforcement

Every action is validated before execution:

- **Domain allowlists** — A secret configured for `api.stripe.com` cannot be sent to `evil.com`. The request is denied before any network activity occurs.
- **Command allowlists** — A database URL configured for `psql,prisma` cannot be injected into `curl`. The command is rejected before execution.
- **Empty allowlists** = no restrictions (configurable, but not recommended for production secrets).

### 3. Output Redaction

All output streams (HTTP response bodies, stdout, stderr) are scanned for secret values before being returned to the agent. The redaction engine catches:

- **Raw values** — the literal secret string
- **Base64-encoded** — because secrets often appear base64'd in headers and logs
- **URL-encoded** — because secrets can appear in query strings and form bodies

Matches are replaced with `[VAULTY:SECRET_NAME]`, which tells the agent "there was a secret here, but you don't get to see it."

### 4. Encryption at Rest

The vault file uses [age](https://age-encryption.org/) encryption:

- **Single-user mode** — scrypt-derived key from your passphrase. Industry-standard KDF with high work factor.
- **Team mode** — X25519 public-key encryption with multiple recipients. Each team member has their own key pair.

Age is a modern, audited encryption tool designed by Filippo Valsorda (Go crypto team lead). It's the spiritual successor to GPG, without the complexity or footguns.

### 5. Memory Safety

Secrets in memory are handled with care:

- Stored as `[]byte` slices (not strings — Go strings are immutable and can linger in memory)
- **Explicitly zeroed** on daemon stop, idle timeout, or secret rotation
- Never included in error messages, log output, or stack traces
- Constant-time comparison where applicable

When you run `vaulty stop`, the secrets aren't just garbage collected "eventually" — they're overwritten with zeros immediately.

### 6. Audit Trail

Every action is logged to an append-only JSONL audit file:

```json
{"ts":"2026-04-01T14:23:01Z","action":"proxy","secret":"STRIPE_SECRET_KEY","target":"https://api.stripe.com/v1/charges","method":"POST","status":200}
{"ts":"2026-04-01T14:23:05Z","action":"denied","secret":"STRIPE_SECRET_KEY","target":"https://evil.com/steal","reason":"domain not in allowlist"}
```

The audit log never contains secret values. It records what happened, when, and whether it was allowed — useful for incident investigation and compliance.

## What Vaulty Does NOT Protect Against

Let's be honest about the limits:

- **Compromised host** — If an attacker has root on your machine, they can read the daemon's memory. Vaulty protects your secrets from *agents*, not from *root-level attackers*.
- **Social engineering** — If you tell the agent your passphrase, Vaulty can't help you.
- **Side-channel attacks** — We don't claim to be side-channel resistant. We're a developer tool, not a HSM.
- **Already-leaked secrets** — If a secret was previously exposed (in a `.env` file committed to git, for example), adding it to Vaulty now doesn't un-leak it. Rotate first, then add to Vaulty.

## Zero Cloud

Vaulty runs entirely on your machine:

- No cloud accounts
- No SaaS dependencies
- No telemetry
- No phone-home behavior
- No auto-update mechanism

Your secrets never leave your machine except through explicitly policy-approved requests that you configured.

## Comparison with Alternatives

| | Vaulty | .env files | MCP env vars | Cloud secret managers |
|---|---|---|---|---|
| Agent sees raw secret | **No** | Yes | Yes | Depends |
| Policy enforcement | **Yes** | No | No | IAM only |
| Output redaction | **Yes** | No | No | No |
| Works offline | **Yes** | Yes | Yes | No |
| Audit trail | **Yes** | No | No | Yes |
| Setup complexity | Low | None | None | High |
