---
title: "Audit Logging"
weight: 3
---

# Audit Logging

Every action Vaulty takes is logged. Every request proxied, every command executed, every policy denial. If something goes wrong (or something goes right and you want to prove it), the audit log has your back.

## Log Format

The audit log is an append-only JSONL file — one JSON object per line. Simple to parse, simple to grep, simple to ship to your log aggregator if you're into that sort of thing.

**Location:** `~/.config/vaulty/audit.log`

### Successful Proxy Request

```json
{
  "ts": "2026-04-01T14:23:01Z",
  "action": "proxy",
  "secret": "STRIPE_SECRET_KEY",
  "target": "https://api.stripe.com/v1/charges",
  "method": "POST",
  "status": 200
}
```

### Successful Command Execution

```json
{
  "ts": "2026-04-01T14:23:05Z",
  "action": "exec",
  "secret": "DATABASE_URL",
  "command": "npx prisma migrate deploy",
  "exit_code": 0
}
```

### Policy Denial

```json
{
  "ts": "2026-04-01T14:23:10Z",
  "action": "denied",
  "secret": "STRIPE_SECRET_KEY",
  "target": "https://evil.com/steal",
  "reason": "domain not in allowlist"
}
```

## What's NOT in the Log

Secret values. Ever. The log records *which* secret was used and *where* it was sent, but never the actual value. You can safely share audit logs with your security team, pipe them to a SIEM, or include them in incident reports.

## Why JSONL?

- **Append-only** — new entries are appended, old entries are never modified. Good for integrity.
- **Line-oriented** — one entry per line means you can use `grep`, `jq`, `tail -f`, or any line-based tool.
- **Structured** — JSON means you can parse it programmatically without regex acrobatics.
- **Streamable** — no opening/closing brackets to worry about. The file is always valid even if the process crashes mid-write.

## Useful Queries

```bash
# See all requests from the last hour
jq 'select(.action == "proxy")' ~/.config/vaulty/audit.log

# Find all policy denials
jq 'select(.action == "denied")' ~/.config/vaulty/audit.log

# Count requests per secret
jq -r '.secret' ~/.config/vaulty/audit.log | sort | uniq -c | sort -rn

# Watch the log in real time
tail -f ~/.config/vaulty/audit.log | jq .
```
