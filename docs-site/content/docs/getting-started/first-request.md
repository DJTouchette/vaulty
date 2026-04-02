---
title: "Your First Request"
weight: 3
---

# Your First Request

You've got a vault, you've got a secret, you've got a running daemon. Time to actually *use* this thing.

## Proxy an HTTP Request

Let's make an authenticated API call to Stripe:

```bash
vaulty proxy POST https://api.stripe.com/v1/charges \
  --secret STRIPE_SECRET_KEY \
  --header "Content-Type: application/x-www-form-urlencoded" \
  --body "amount=2000&currency=usd"
```

Here's what just happened:

1. **Policy check** — Vaulty verified that `STRIPE_SECRET_KEY` is allowed to talk to `api.stripe.com`. ✅
2. **Secret injection** — Vaulty grabbed the secret from memory and added an `Authorization: Bearer sk_live_...` header. You didn't have to know the key, Vaulty just did its thing.
3. **Request execution** — The actual HTTP request went out with your real credentials attached.
4. **Response redaction** — If the response body happened to contain your secret value (it shouldn't, but Stripe has had weird days), Vaulty would replace it with `[VAULTY:STRIPE_SECRET_KEY]`.
5. **Audit logging** — The request was logged to `~/.config/vaulty/audit.log` (without the secret value, naturally).

## Execute a Command

For CLI tools, use `exec` instead of `proxy`:

```bash
vaulty exec --secret DATABASE_URL -- npx prisma migrate deploy
```

This:
1. Checks that `prisma` is in the `allowed_commands` for `DATABASE_URL`
2. Spawns `npx prisma migrate deploy` with `DATABASE_URL=postgres://...` in its environment
3. Streams stdout/stderr back to you, but redacts any occurrence of the database URL
4. Logs the execution to the audit trail

The `--` separator is important — everything after it is the command to run. Without it, Vaulty might get confused about what's a flag and what's a command. Vaulty is smart, but not *that* smart.

## What If a Policy Blocks You?

Try requesting a secret for a domain that's not in its allowlist:

```bash
vaulty proxy GET https://definitely-not-stripe.com/steal-secrets \
  --secret STRIPE_SECRET_KEY
```

You'll get a nice, clear error:

```
Error: domain "definitely-not-stripe.com" not in allowed domains for STRIPE_SECRET_KEY
       allowed: [api.stripe.com]
```

No secret was sent. No request was made. The audit log recorded the denial. Your Stripe key remains safe from `definitely-not-stripe.com` (which, honestly, sounds like it was trying its best).

## What's Next?

If you're using an AI agent like Claude Code or Cursor, you'll want to set up the [MCP integration]({{< relref "/docs/guides/mcp-setup" >}}) so the agent can use Vaulty's tools directly.

If you want to fine-tune your policies, check out the [Policy Templates]({{< relref "/docs/guides/templates" >}}) for pre-built configs for popular APIs.

Or if you're the "give me all the commands" type, head to the [CLI Reference]({{< relref "/docs/reference/cli" >}}).
