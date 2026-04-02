---
title: "Framework Integrations"
weight: 5
---

# Framework Integrations

Already have secrets in `.env` files, Rails credentials, Docker Compose, or Kubernetes manifests? Vaulty can import them — and export back when you need to.

## .env Files

The universal language of "I need this secret in my app."

### Import

```bash
vaulty import-env .env
```

This reads each `KEY=VALUE` pair from the file and stores them in your vault. You'll be prompted for policy configuration (domains/commands) for each one, or you can add policies to `vaulty.toml` afterward.

### Export

```bash
vaulty export-env > .env.local
```

Exports all secrets in standard `.env` format. Useful for tools that absolutely insist on reading from a file. 

**Heads up:** The exported file contains raw secret values. Treat it like the vault itself — don't commit it, don't leave it lying around.

## Rails Credentials

For Ruby on Rails apps that use `credentials.yml.enc`:

### Import

```bash
vaulty import-rails
```

Reads your Rails encrypted credentials and imports them into Vaulty. You'll need your Rails master key available (in `config/master.key` or `RAILS_MASTER_KEY` env var).

### Export

```bash
vaulty export-rails
```

Exports secrets in Rails YAML format, ready to pipe into `rails credentials:edit`.

## Docker Compose

For Docker setups with secrets defined in `docker-compose.yml`:

### Import

```bash
vaulty import-docker docker-compose.yml
```

### Export

```bash
vaulty export-docker
```

Outputs secrets in a format compatible with Docker Compose's secrets configuration.

## Kubernetes

For K8s `Secret` manifests:

### Import

```bash
vaulty import-k8s secret.yaml
```

Imports secrets from a Kubernetes Secret YAML file. Handles base64 decoding automatically.

### Export

```bash
vaulty export-k8s
```

Generates a Kubernetes-compatible Secret YAML manifest with base64-encoded values.

## Vault Transfer

For moving vaults between machines or making backups:

```bash
# Export (creates an encrypted snapshot)
vaulty export > vault-backup.age

# Import on another machine
vaulty import < vault-backup.age
```

The export is encrypted with your passphrase, so it's safe to send over less-than-perfectly-secure channels (but maybe don't post it on Twitter).
