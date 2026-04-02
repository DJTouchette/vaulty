---
title: "MCP Setup"
weight: 1
---

# MCP Setup

This is where Vaulty gets *really* cool. MCP (Model Context Protocol) lets your AI agent discover and use Vaulty's tools automatically. No manual `vaulty proxy` commands — the agent just... knows how to use it.

## What the Agent Gets

When you configure Vaulty as an MCP server, your AI agent discovers three tools:

| Tool | What It Does |
|------|-------------|
| `vaulty_request` | Makes authenticated HTTP requests. The agent says "call Stripe," Vaulty handles the auth. |
| `vaulty_exec` | Runs commands with secrets injected as env vars. Output comes back redacted. |
| `vaulty_list` | Lists available secret names and their policies. Never values. Not even a hint. |

The agent sees secret *names* and *policies*. It knows "there's a `STRIPE_SECRET_KEY` that works with `api.stripe.com`" — but it never sees `sk_live_abc123...`. It's like giving someone a labeled remote control with no way to open the battery compartment.

## Claude Code

Add Vaulty to your Claude Code MCP config. This goes in `~/.claude/claude_desktop_config.json` or your project's `.mcp.json`:

```json
{
  "mcpServers": {
    "vaulty": {
      "command": "vaulty",
      "args": ["mcp"],
      "env": {
        "VAULTY_PASSPHRASE": "your-passphrase"
      }
    }
  }
}
```

**Wait, a passphrase in a config file?** Yeah, we know. There are a few options:

1. **Keychain** — Run `vaulty keychain save` first, then skip the `env` block entirely. Vaulty will grab the passphrase from your OS keychain.
2. **Team mode** — Use an age identity file instead (see [Team Sharing]({{< relref "/docs/guides/team-sharing" >}})):
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
3. **YOLO** — Put the passphrase in the env block. It's a local file on your machine. We won't tell anyone.

## Cursor

Same config, same format — Cursor uses the MCP standard:

```json
{
  "mcpServers": {
    "vaulty": {
      "command": "vaulty",
      "args": ["mcp"],
      "env": {
        "VAULTY_PASSPHRASE": "your-passphrase"
      }
    }
  }
}
```

## How It Works Under the Hood

When you start Vaulty in MCP mode (`vaulty mcp`), it:

1. Decrypts the vault inline (no separate daemon needed)
2. Listens on stdin/stdout using JSON-RPC 2.0
3. Responds to tool discovery requests from the agent
4. Processes tool calls with full policy enforcement
5. Returns redacted results

The MCP server is a single-connection, single-session thing — it lives and dies with the agent process. No background daemon required (though you can run one separately if you also want CLI access).

## Approval Workflow

By default, Vaulty asks for your approval before each MCP tool call. You'll see a prompt like:

```
vaulty_request: POST https://api.stripe.com/v1/charges
  secret: STRIPE_SECRET_KEY
  Allow? [y/n]
```

If you trust a particular secret for routine use, you can set `auto_approve = true` in your config:

```toml
[secrets.STRIPE_SECRET_KEY]
allowed_domains = ["api.stripe.com"]
inject_as = "bearer"
auto_approve = true
```

This skips the approval prompt for that specific secret. Use it for dev/test keys; maybe don't auto-approve your production database credentials on day one.

## Example: Agent Makes a Stripe Call

Here's what it looks like in practice. Your AI agent wants to create a Stripe charge:

1. Agent calls `vaulty_list` → learns that `STRIPE_SECRET_KEY` exists and works with `api.stripe.com`
2. Agent calls `vaulty_request` with:
   ```json
   {
     "method": "POST",
     "url": "https://api.stripe.com/v1/charges",
     "secret_name": "STRIPE_SECRET_KEY",
     "headers": {"Content-Type": "application/x-www-form-urlencoded"},
     "body": "amount=2000&currency=usd"
   }
   ```
3. Vaulty checks the policy ✅, injects the Bearer token, makes the request
4. Vaulty returns the response to the agent with any secret values redacted

The agent got its Stripe charge created. It still has no idea what the actual API key is. Everyone's happy.
