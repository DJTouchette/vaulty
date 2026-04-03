# Vaulty + Docker Compose

This example shows how to use Vaulty to manage secrets for Docker Compose services.

## Walkthrough

### 1. Import existing compose environment variables

```bash
vaulty import-docker docker-compose.yml
```

This extracts all `environment:` variables from every service and stores them in the vault.

### 2. Generate a compose override with real secrets

```bash
vaulty export-docker --service web --out docker-compose.override.yml
```

This creates `docker-compose.override.yml` with your vault secrets as environment variables for the `web` service. Docker Compose automatically merges override files.

### 3. Export as Docker secret files

```bash
vaulty export-docker --secrets-dir ./secrets
```

This writes each secret as a separate file under `./secrets/`:
```
secrets/
  API_KEY
  DATABASE_URL
  REDIS_URL
```

These can be mounted as Docker secrets in swarm mode or bind-mounted into containers.

### 4. Run with compose

```bash
docker compose up
# docker-compose.override.yml is automatically picked up
```

## Service name

The `--service` flag controls which service gets the environment block in the override file. Defaults to `app` if not specified.

```bash
vaulty export-docker --service db --out docker-compose.override.yml
```
