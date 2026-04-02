# Vaulty + Next.js Environment Variables

This example shows how to use Vaulty with Next.js environment variables, including the `NEXT_PUBLIC_` convention.

## Next.js env conventions

- `NEXT_PUBLIC_*` variables are embedded into the browser bundle at build time and are visible to users
- All other variables are server-side only and never reach the browser
- Next.js loads from `.env`, `.env.local`, `.env.development`, `.env.production`

## Walkthrough

### 1. Create a `.env.local` file

```bash
cp .env.local.example .env.local
# Edit with your real values
```

### 2. Import into Vaulty

```bash
vaulty import-env .env.local
```

### 3. Export as Next.js env files

```bash
vaulty export-env --format nextjs --reveal
```

This generates:
- `.env.development` — for `next dev`
- `.env.production` — for `next build`

Both contain the same secrets. Vaulty will warn about any `NEXT_PUBLIC_` variables since they'll be exposed to browsers.

Example warning:
```
Warning: 2 NEXT_PUBLIC_ variable(s) will be exposed to the browser:
  - NEXT_PUBLIC_API_URL
  - NEXT_PUBLIC_SITE_NAME
```

### 4. Export redacted (for review)

```bash
vaulty export-env --format nextjs
```

Writes redacted values (`****`) so you can verify which variables are set without revealing secrets.

### 5. View just the public/private split

```bash
vaulty export-env
```

Output:
```
DATABASE_URL=****
NEXT_PUBLIC_API_URL=****
NEXT_PUBLIC_SITE_NAME=****
NEXTAUTH_SECRET=****
STRIPE_SECRET_KEY=****
```

## Security note

`NEXT_PUBLIC_` variables are intentionally public, but you should still manage them through Vaulty to maintain a single source of truth and avoid accidentally committing `.env.local` files.
