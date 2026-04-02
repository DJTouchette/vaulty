---
title: "Policy Templates"
weight: 2
---

# Policy Templates

Nobody wants to figure out the auth scheme for every API from scratch. We've done the boring part for you — just copy, paste, and store the secret.

## How to Use a Template

1. Find the template for your service below
2. Copy the TOML block into your `vaulty.toml`
3. Store the secret: `vaulty set SECRET_NAME`
4. That's it. Go do something more interesting.

## Available Templates

### Stripe

The classic. Bearer token auth, locked to `api.stripe.com`.

```toml
[secrets.STRIPE_SECRET_KEY]
description = "Stripe API key"
allowed_domains = ["api.stripe.com"]
inject_as = "bearer"
```

```bash
vaulty set STRIPE_SECRET_KEY
# Paste your sk_live_... or sk_test_... key
```

### OpenAI

For when your AI agent needs to call... another AI. It's AIs all the way down.

```toml
[secrets.OPENAI_API_KEY]
description = "OpenAI API key"
allowed_domains = ["api.openai.com"]
inject_as = "bearer"
```

### AWS

AWS is special because it uses multiple env vars. Vaulty handles companion secrets with `also_inject`:

```toml
[secrets.AWS_SECRET_ACCESS_KEY]
description = "AWS secret key"
allowed_commands = ["aws", "cdk", "sst"]
inject_as = "env"
also_inject = ["AWS_ACCESS_KEY_ID"]

[secrets.AWS_ACCESS_KEY_ID]
description = "AWS access key ID"
allowed_commands = ["aws", "cdk", "sst"]
inject_as = "env"
```

```bash
vaulty set AWS_ACCESS_KEY_ID
vaulty set AWS_SECRET_ACCESS_KEY
```

Now `vaulty exec --secret AWS_SECRET_ACCESS_KEY -- aws s3 ls` injects both keys. Teamwork.

### Database

For all your database connection string needs:

```toml
[secrets.DATABASE_URL]
description = "Database connection string"
allowed_commands = ["psql", "prisma", "drizzle-kit", "pg_dump"]
inject_as = "env"
```

### GitHub

```toml
[secrets.GITHUB_TOKEN]
description = "GitHub personal access token"
allowed_domains = ["api.github.com"]
inject_as = "bearer"
```

### Supabase

```toml
[secrets.SUPABASE_SERVICE_ROLE_KEY]
description = "Supabase service role key"
allowed_domains = ["supabase.co"]
inject_as = "bearer"
```

**Pro tip:** Narrow the domain to your specific project: `["your-project.supabase.co"]`

### Resend

```toml
[secrets.RESEND_API_KEY]
description = "Resend email API key"
allowed_domains = ["api.resend.com"]
inject_as = "header"
header_name = "X-API-Key"
```

### Twilio

One of the few APIs still using Basic auth in 2026. Respect for the classics.

```toml
[secrets.TWILIO_AUTH_TOKEN]
description = "Twilio auth token"
allowed_domains = ["api.twilio.com"]
inject_as = "basic"
```

### Cloudflare

```toml
[secrets.CLOUDFLARE_API_TOKEN]
description = "Cloudflare API token"
allowed_domains = ["api.cloudflare.com"]
inject_as = "bearer"
```

### Vercel

```toml
[secrets.VERCEL_TOKEN]
description = "Vercel access token"
allowed_domains = ["api.vercel.com"]
inject_as = "bearer"
```

## Customizing Templates

These are sensible defaults, not sacred texts. Feel free to:

- **Narrow domains** — Replace `supabase.co` with `your-project.supabase.co`
- **Add commands** — Extend `allowed_commands` for your specific CLI tools
- **Combine secrets** — Use `also_inject` to group secrets that travel together
- **Add descriptions** — The `description` field shows up in `vaulty list` and `vaulty_list` MCP tool, helping your AI agent understand what each secret is for

## Don't See Your Service?

Adding a template is easy — it's just a `[secrets.X]` TOML block. Check the service's API docs for the auth method:

- **Bearer token?** → `inject_as = "bearer"`
- **API key in header?** → `inject_as = "header"` + `header_name = "X-Whatever-They-Call-It"`
- **Basic auth?** → `inject_as = "basic"`
- **Query parameter?** → `inject_as = "query"` (looking at you, Google Maps)
- **Env var for CLI?** → `inject_as = "env"` + `allowed_commands`
