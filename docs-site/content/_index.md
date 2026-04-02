---
title: "Vaulty"
type: docs
---

# Vaulty

**Your AI agent's bouncer for secrets.** 🔐

Vaulty is a local-only CLI daemon that acts as a secrets proxy for AI coding agents. Your agent gets *capabilities*, not *credentials* — so even if a prompt injection tries to steal your API keys, there's nothing to steal.

## The Problem (It's You, Probably)

Let's be honest: you've probably pasted an API key into a `.env` file, added it to your MCP config, or — and we're not judging — directly into a chat with an AI agent. We've all been there. The problem is that your AI assistant can *see* those secrets, and anything that can see them can leak them.

A prompt injection, a rogue MCP tool, or even a confused agent could exfiltrate your credentials faster than you can say "rotate my Stripe key."

## The Fix (It's Vaulty, Obviously)

Vaulty sits between your AI agent and the outside world. Your agent says *"hey, make this API call with my Stripe key"* and Vaulty says *"sure, I'll handle the authentication part, you just worry about the JSON."*

The agent never touches the raw secret. It only knows the secret *exists* and what it's *called*. Like a VIP list at a club, except the bouncer also redacts any photos of the VIP.

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
│  1. Validate policy              │
│  2. Inject secret into request   │
│  3. Execute & redact output      │
│  4. Log to audit trail           │
└──────────────────────────────────┘
```

## Quick Start (60 Seconds, We Timed It)

```bash
# Install
go install github.com/djtouchette/vaulty/cmd/vaulty@latest

# Create a vault (you'll pick a passphrase)
vaulty init

# Add a secret with a domain policy
vaulty set STRIPE_SECRET_KEY --domains "api.stripe.com"

# Start the daemon (loads secrets into memory)
vaulty start

# Make an authenticated API call — agent never sees the key
vaulty proxy POST https://api.stripe.com/v1/charges \
  --secret STRIPE_SECRET_KEY \
  --header "Content-Type: application/x-www-form-urlencoded" \
  --body "amount=2000&currency=usd"

# Done? Stop the daemon (zeroes secrets from memory)
vaulty stop
```

That's it. No Docker, no cloud accounts, no runtime dependencies. Just a single Go binary that keeps your secrets safe from overly curious AI agents.

## What's Inside

{{< columns >}}

### [Getting Started]({{< relref "/docs/getting-started" >}})
Installation, your first vault, and your first proxied request. All in about 5 minutes (even if you type slowly).

<--->

### [Guides]({{< relref "/docs/guides" >}})
MCP integration, team sharing, policy templates, framework imports, and other things that make your life easier.

<--->

### [Reference]({{< relref "/docs/reference" >}})
Every CLI command, every config option, every flag. The boring-but-essential stuff.

{{< /columns >}}

{{< columns >}}

### [Architecture]({{< relref "/docs/architecture" >}})
How Vaulty works under the hood. For the curious, the skeptical, and the security auditors.

<--->

### [Templates]({{< relref "/docs/guides/templates" >}})
Pre-built policies for Stripe, OpenAI, AWS, and more. Copy, paste, done.

<--->

&nbsp;

{{< /columns >}}
