#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

echo "=== Starting test services ==="
docker compose -f docker-compose.test.yml up -d

echo "=== Waiting for LocalStack ==="
for i in $(seq 1 30); do
    if curl -sf http://localhost:4566/_localstack/health > /dev/null 2>&1; then
        echo "LocalStack ready"
        break
    fi
    [ "$i" -eq 30 ] && { echo "LocalStack failed to start"; docker compose -f docker-compose.test.yml logs localstack; exit 1; }
    sleep 1
done

echo "=== Waiting for Vault ==="
for i in $(seq 1 30); do
    if curl -sf http://localhost:8200/v1/sys/health > /dev/null 2>&1; then
        echo "Vault ready"
        break
    fi
    [ "$i" -eq 30 ] && { echo "Vault failed to start"; docker compose -f docker-compose.test.yml logs vault; exit 1; }
    sleep 1
done

echo "=== Checking CLI dependencies ==="
missing=0
for cli in aws vault; do
    if ! command -v "$cli" &> /dev/null; then
        echo "WARNING: $cli CLI not found — some tests will be skipped"
        missing=1
    fi
done

echo "=== Running integration tests ==="
go test -tags=integration ./internal/backend/... -count=1 -v -timeout 120s
result=$?

echo "=== Tearing down test services ==="
docker compose -f docker-compose.test.yml down -v

exit $result
