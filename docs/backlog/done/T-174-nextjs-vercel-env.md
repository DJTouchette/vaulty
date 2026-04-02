# T-174: Next.js/Vercel env support

**Epic:** 17 — Framework Secret Files
**Status:** done
**Priority:** P2

## Description

Add Next.js environment variable classification to Vaulty. Detects NEXT_PUBLIC_ variables and warns about browser exposure. Integrates with export-env --format nextjs.

## Acceptance Criteria

- [x] IsPublicEnvVar detects NEXT_PUBLIC_ prefix
- [x] ClassifyNextJSEnv splits secrets into public and private sets
- [x] export-env --format nextjs generates .env.development and .env.production
- [x] Warns about NEXT_PUBLIC_ variables being exposed to browser
- [x] Comprehensive unit tests
- [x] Example project in examples/nextjs-env/
