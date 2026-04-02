---
title: "Cloud Backends"
weight: 6
---

# Cloud Backends

Already using a cloud secret manager? Vaulty can pull secrets from it so you don't have to copy-paste values between services like it's 2015.

## Supported Backends

| Backend | Service | Config Type |
|---------|---------|------------|
| `aws` | AWS Secrets Manager | Region + profile |
| `gcp` | GCP Secret Manager | Project ID |
| `hashicorp` | HashiCorp Vault | Address + mount |
| `onepassword` | 1Password | Via CLI bridge |

## Configuration

Add backends to your `vaulty.toml`:

```toml
[backends.aws_prod]
type = "aws"
region = "us-east-1"
profile = "production"

[backends.gcp_dev]
type = "gcp"
project = "my-gcp-project"

[backends.vault_corp]
type = "hashicorp"
addr = "https://vault.company.com"
mount = "secret"

[backends.onepass]
type = "onepassword"
```

## Usage

### Browse Available Secrets

```bash
vaulty backend list                    # list configured backends
vaulty backend secrets aws_prod        # browse secrets in AWS
```

### Pull a Secret

```bash
vaulty backend pull aws_prod STRIPE_SECRET_KEY
```

This fetches the secret value from AWS Secrets Manager and stores it in your local Vaulty vault with the same name. From there, it behaves like any other Vaulty secret — policies, redaction, the whole deal.

### Why Not Just Use the Cloud Backend Directly?

Good question. You *could* just use AWS Secrets Manager directly. But then:

- Your AI agent needs AWS credentials (which are... also secrets)
- You lose Vaulty's policy enforcement (domain/command allowlists)
- You lose output redaction
- You lose the audit trail
- You're making network calls to AWS for every secret access

Vaulty's approach: pull once from the cloud, encrypt locally, use locally with full policy enforcement. Best of both worlds.

## Caching

Backend responses are cached locally to avoid unnecessary API calls. The cache respects TTLs from the backend provider. You can force a fresh pull with:

```bash
vaulty backend pull --refresh aws_prod STRIPE_SECRET_KEY
```

## Linking Secrets to Backends

You can link a secret to a backend in your config, so you remember where it came from:

```toml
[secrets.STRIPE_SECRET_KEY]
description = "Stripe API key (from AWS prod)"
allowed_domains = ["api.stripe.com"]
inject_as = "bearer"
backend = "aws_prod"
```

The `backend` field is informational — it doesn't auto-sync. It's just a breadcrumb for future you (or future your teammate) to remember where the value lives upstream.
