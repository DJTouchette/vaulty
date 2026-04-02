---
title: "Redaction Engine"
weight: 2
---

# Redaction Engine

The redaction engine is Vaulty's last line of defense. Even if something goes wrong with policy enforcement (it shouldn't, but defense in depth is the name of the game), the redaction engine catches secret values before they reach the agent.

## What Gets Redacted

Every piece of output that flows back to the agent — HTTP response bodies, stdout, stderr — is scanned for secret values. The engine looks for three encodings of each secret:

### 1. Raw Values

The literal secret string. If your Stripe key is `sk_live_abc123`, any occurrence of that exact string is replaced.

### 2. Base64-Encoded Values

Secrets often appear base64-encoded in HTTP headers, JWT tokens, and log output. The engine computes the base64 encoding of each secret and scans for that too.

```
# Before redaction:
Authorization: Basic c2tfbGl2ZV9hYmMxMjM6

# After redaction:
Authorization: Basic [VAULTY:STRIPE_SECRET_KEY]
```

### 3. URL-Encoded Values

Secrets in query strings and form bodies get URL-encoded. The engine catches these as well.

```
# Before redaction:
callback_url=https://example.com?key=sk_live_abc123&action=charge

# After redaction:
callback_url=https://example.com?key=[VAULTY:STRIPE_SECRET_KEY]&action=charge
```

## Replacement Format

All redacted values are replaced with:

```
[VAULTY:SECRET_NAME]
```

This tells the agent (and anyone reading the output) that a secret was here, which secret it was, and that the value has been removed. It's informative without being leaky.

## Performance Considerations

The redaction engine scans every byte of output. For large responses, this adds some overhead. In practice:

- Most API responses are small (a few KB). The overhead is negligible.
- For large binary responses (file downloads, etc.), the overhead is proportional to the response size, but still fast — it's just string matching, not regex.
- The engine runs in a streaming fashion, processing output as it arrives rather than buffering everything in memory.

## Edge Cases

### What if a secret appears in a partial match across chunk boundaries?

The streaming redaction engine maintains a buffer to handle partial matches at chunk boundaries. A secret that spans two output chunks will still be caught.

### What if output *should* contain the secret string?

It will still be redacted. Vaulty has no way to distinguish "intentional" from "accidental" inclusion of secret values in output. The policy is: if it looks like the secret, it gets redacted. This is the conservative, secure choice.

### What about very short secrets?

A 4-character secret like `test` would get redacted everywhere it appears — including in words like "testing" or "contest." This is a good reason to use real, high-entropy secrets instead of short ones. (You should be doing that anyway.)

## Where Redaction Happens

Redaction runs in two contexts:

1. **HTTP Proxy** (`internal/proxy/redactor.go`) — Scans HTTP response bodies before returning them to the caller.
2. **Command Executor** (`internal/executor/stream.go`) — Scans stdout and stderr streams from child processes in real time.

Both use the same underlying redaction logic. The difference is just the I/O wrapping.
