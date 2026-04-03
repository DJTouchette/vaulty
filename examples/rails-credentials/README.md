# Vaulty + Rails Credentials

This example shows how to use Vaulty with Rails encrypted credentials.

## Rails credentials format

Rails stores encrypted secrets in `config/credentials.yml.enc`. The encryption key is either:

- `config/master.key` (a 64-character hex string, never committed)
- The `RAILS_MASTER_KEY` environment variable

The YAML structure typically looks like:

```yaml
aws:
  access_key_id: AKIAIOSFODNN7EXAMPLE
  secret_access_key: wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY

secret_key_base: abc123def456ghi789

stripe:
  publishable_key: pk_test_example
  secret_key: sk_test_example
```

## Walkthrough

### 1. Import Rails credentials

```bash
vaulty import-rails
```

This reads `config/credentials.yml.enc`, decrypts with `config/master.key`, flattens the YAML, and stores secrets in the vault:

- `AWS_ACCESS_KEY_ID` = `AKIAIOSFODNN7EXAMPLE`
- `AWS_SECRET_ACCESS_KEY` = `wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY`
- `SECRET_KEY_BASE` = `abc123def456ghi789`
- `STRIPE_PUBLISHABLE_KEY` = `pk_test_example`
- `STRIPE_SECRET_KEY` = `sk_test_example`

### 2. Per-environment credentials

```bash
vaulty import-rails --env production
# Reads config/credentials/production.yml.enc
```

### 3. Using RAILS_MASTER_KEY

```bash
RAILS_MASTER_KEY=abc123... vaulty import-rails
```

### 4. Export back to Rails YAML format

```bash
vaulty export-rails
```

Output:
```yaml
aws:
  access_key_id: AKIAIOSFODNN7EXAMPLE
  secret_access_key: wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
secret_key_base: abc123def456ghi789
stripe:
  publishable_key: pk_test_example
  secret_key: sk_test_example
```

### Alternative: pipe from Rails CLI

If direct decryption doesn't work for your setup:

```bash
rails credentials:show | vaulty import-env /dev/stdin
```

## Sample credentials YAML

See `sample_credentials.yml` for the YAML structure that would be inside `credentials.yml.enc`.
