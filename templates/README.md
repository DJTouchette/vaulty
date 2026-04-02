# Policy Templates

Pre-built policy configurations for common APIs and services. Each template contains a ready-to-use `[secrets.X]` block that you can copy into your `vaulty.toml`.

## How to use

1. Find the template for your service below
2. Open the `.toml` file and copy the `[secrets.X]` block
3. Paste it into your project's `vaulty.toml` (or `~/.config/vaulty/vaulty.toml`)
4. Store the secret value: `vaulty set SECRET_NAME`

That's it. The policy block defines *where* the secret can be used and *how* it gets injected. The secret value itself lives in your encrypted vault, never in the config file.

## Available templates

| Template | Service | Injection | File |
|----------|---------|-----------|------|
| Stripe | api.stripe.com | Bearer token | [stripe.toml](stripe.toml) |
| OpenAI | api.openai.com | Bearer token | [openai.toml](openai.toml) |
| AWS | aws, cdk, sst CLI | Env vars | [aws.toml](aws.toml) |
| Supabase | supabase.co | Bearer token | [supabase.toml](supabase.toml) |
| Resend | api.resend.com | Bearer token | [resend.toml](resend.toml) |
| Twilio | api.twilio.com | Basic auth | [twilio.toml](twilio.toml) |
| GitHub | api.github.com | Bearer token | [github.toml](github.toml) |
| Database | psql, prisma, etc. | Env var | [database.toml](database.toml) |
| Cloudflare | api.cloudflare.com | Bearer token | [cloudflare.toml](cloudflare.toml) |
| Vercel | api.vercel.com | Bearer token | [vercel.toml](vercel.toml) |

## Example

To add Stripe support to your project, copy from `templates/stripe.toml`:

```toml
[secrets.STRIPE_SECRET_KEY]
allowed_domains = ["api.stripe.com"]
inject_as = "bearer"
```

Then store the secret:

```bash
vaulty set STRIPE_SECRET_KEY
# paste your sk_live_... or sk_test_... key when prompted
```

## Customizing templates

Templates provide sensible defaults. You can adjust them after copying:

- **Narrow domains** -- replace a broad domain with your specific subdomain (e.g., `yourproject.supabase.co` instead of `supabase.co`)
- **Add commands** -- extend `allowed_commands` to include additional CLI tools in your workflow
- **Combine secrets** -- use `also_inject` to group related secrets that should always be provided together (see the AWS template for an example)

## Contributing a template

Add a new `.toml` file to this directory with:

1. A comment block explaining the service, auth method, and setup steps
2. A valid `[secrets.X]` block with sensible defaults
3. A link to the service's authentication docs
4. An entry in the table above
