//go:build integration

package backend

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// These tests require:
//   docker compose -f docker-compose.test.yml up -d
//   aws CLI installed
//   vault CLI installed
//
// Run with: go test -tags=integration ./internal/backend/... -count=1 -v
// Or use: ./scripts/test-integration.sh

func skipIfNoDocker(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker not available")
	}
}

func skipIfNoCLI(t *testing.T, name string) {
	t.Helper()
	if _, err := exec.LookPath(name); err != nil {
		t.Skipf("%s CLI not available", name)
	}
}

func waitForService(url string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := http.Get(url)
		if err == nil {
			resp.Body.Close()
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("service at %s not ready after %v", url, timeout)
}

// --- AWS / LocalStack integration tests ---

func seedAWSSecret(t *testing.T, endpoint, name, value string) {
	t.Helper()

	// Use the aws CLI to create the secret (handles signing correctly)
	cmd := exec.Command("aws", "secretsmanager", "create-secret",
		"--name", name,
		"--secret-string", value,
		"--endpoint-url", endpoint,
		"--region", "us-east-1",
	)
	cmd.Env = append(os.Environ(),
		"AWS_ACCESS_KEY_ID=test",
		"AWS_SECRET_ACCESS_KEY=test",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		// Ignore "already exists" errors
		if !strings.Contains(string(out), "ResourceExistsException") {
			t.Fatalf("seeding secret %s: %v: %s", name, err, out)
		}
	}
}

func TestIntegrationAWSList(t *testing.T) {
	skipIfNoDocker(t)
	skipIfNoCLI(t, "aws")

	endpoint := "http://localhost:4566"
	if err := waitForService(endpoint+"/_localstack/health", 30*time.Second); err != nil {
		t.Skip("LocalStack not running — start with: docker compose -f docker-compose.test.yml up -d")
	}

	// Seed test secrets
	seedAWSSecret(t, endpoint, "vaulty-test/api-key", "sk_test_12345")
	seedAWSSecret(t, endpoint, "vaulty-test/db-password", "supersecretpw")

	// Test via the real backend
	b, err := NewAWSBackendWithEndpoint("us-east-1", "", endpoint)
	if err != nil {
		t.Fatalf("NewAWSBackendWithEndpoint: %v", err)
	}

	// Set dummy AWS creds for LocalStack (it doesn't validate them)
	t.Setenv("AWS_ACCESS_KEY_ID", "test")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "test")

	names, err := b.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	found := map[string]bool{}
	for _, n := range names {
		found[n] = true
	}
	if !found["vaulty-test/api-key"] {
		t.Errorf("expected vaulty-test/api-key in list, got %v", names)
	}
	if !found["vaulty-test/db-password"] {
		t.Errorf("expected vaulty-test/db-password in list, got %v", names)
	}
}

func TestIntegrationAWSGet(t *testing.T) {
	skipIfNoDocker(t)
	skipIfNoCLI(t, "aws")

	endpoint := "http://localhost:4566"
	if err := waitForService(endpoint+"/_localstack/health", 30*time.Second); err != nil {
		t.Skip("LocalStack not running")
	}

	seedAWSSecret(t, endpoint, "vaulty-test/get-test", "my-secret-value-42")

	b, err := NewAWSBackendWithEndpoint("us-east-1", "", endpoint)
	if err != nil {
		t.Fatalf("NewAWSBackendWithEndpoint: %v", err)
	}

	t.Setenv("AWS_ACCESS_KEY_ID", "test")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "test")

	val, err := b.Get("vaulty-test/get-test")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if val != "my-secret-value-42" {
		t.Errorf("Get = %q, want my-secret-value-42", val)
	}
}

func TestIntegrationAWSGetNonexistent(t *testing.T) {
	skipIfNoDocker(t)
	skipIfNoCLI(t, "aws")

	endpoint := "http://localhost:4566"
	if err := waitForService(endpoint+"/_localstack/health", 30*time.Second); err != nil {
		t.Skip("LocalStack not running")
	}

	b, err := NewAWSBackendWithEndpoint("us-east-1", "", endpoint)
	if err != nil {
		t.Fatalf("NewAWSBackendWithEndpoint: %v", err)
	}

	t.Setenv("AWS_ACCESS_KEY_ID", "test")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "test")

	_, err = b.Get("vaulty-test/does-not-exist-" + fmt.Sprintf("%d", time.Now().UnixNano()))
	if err == nil {
		t.Fatal("expected error for nonexistent secret")
	}
}

func TestIntegrationAWSCachedBackend(t *testing.T) {
	skipIfNoDocker(t)
	skipIfNoCLI(t, "aws")

	endpoint := "http://localhost:4566"
	if err := waitForService(endpoint+"/_localstack/health", 30*time.Second); err != nil {
		t.Skip("LocalStack not running")
	}

	seedAWSSecret(t, endpoint, "vaulty-test/cached", "cached-value")

	b, err := NewAWSBackendWithEndpoint("us-east-1", "", endpoint)
	if err != nil {
		t.Fatalf("NewAWSBackendWithEndpoint: %v", err)
	}

	t.Setenv("AWS_ACCESS_KEY_ID", "test")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "test")

	cached := NewCachedBackend(b, 1*time.Minute)
	defer cached.Zero()

	// First call
	val, err := cached.Get("vaulty-test/cached")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if val != "cached-value" {
		t.Errorf("Get = %q, want cached-value", val)
	}

	// Second call should use cache (no way to verify without timing, but at least it shouldn't error)
	val2, err := cached.Get("vaulty-test/cached")
	if err != nil {
		t.Fatalf("cached Get: %v", err)
	}
	if val2 != val {
		t.Errorf("cached Get = %q, want %q", val2, val)
	}
}

// --- HashiCorp Vault integration tests ---

func seedVaultSecret(t *testing.T, addr, token, mount, path, key, value string) {
	t.Helper()

	payload := map[string]any{
		"data": map[string]string{key: value},
	}
	body, _ := json.Marshal(payload)

	url := fmt.Sprintf("%s/v1/%s/data/%s", addr, mount, path)
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		t.Fatalf("creating request: %v", err)
	}
	req.Header.Set("X-Vault-Token", token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("seeding vault secret %s: %v", path, err)
	}
	resp.Body.Close()
	if resp.StatusCode >= 400 {
		t.Fatalf("seeding vault secret %s: HTTP %d", path, resp.StatusCode)
	}
}

func TestIntegrationVaultList(t *testing.T) {
	skipIfNoDocker(t)
	skipIfNoCLI(t, "vault")

	addr := "http://localhost:8200"
	token := "test-root-token"

	if err := waitForService(addr+"/v1/sys/health", 30*time.Second); err != nil {
		t.Skip("Vault not running — start with: docker compose -f docker-compose.test.yml up -d")
	}

	// Seed secrets via HTTP API
	seedVaultSecret(t, addr, token, "secret", "vaulty-test/api-key", "value", "sk_live_abc")
	seedVaultSecret(t, addr, token, "secret", "vaulty-test/db-pass", "value", "pgpass123")

	// Test via the real backend
	t.Setenv("VAULT_TOKEN", token)

	b, err := NewHashiCorpBackend(addr, "secret")
	if err != nil {
		t.Fatalf("NewHashiCorpBackend: %v", err)
	}

	names, err := b.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	found := map[string]bool{}
	for _, n := range names {
		found[n] = true
	}
	if !found["vaulty-test/"] {
		t.Errorf("expected vaulty-test/ in list, got %v", names)
	}
}

func TestIntegrationVaultGet(t *testing.T) {
	skipIfNoDocker(t)
	skipIfNoCLI(t, "vault")

	addr := "http://localhost:8200"
	token := "test-root-token"

	if err := waitForService(addr+"/v1/sys/health", 30*time.Second); err != nil {
		t.Skip("Vault not running")
	}

	seedVaultSecret(t, addr, token, "secret", "vaulty-test/mykey", "value", "top-secret-42")

	t.Setenv("VAULT_TOKEN", token)

	b, err := NewHashiCorpBackend(addr, "secret")
	if err != nil {
		t.Fatalf("NewHashiCorpBackend: %v", err)
	}

	val, err := b.Get("vaulty-test/mykey")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if val != "top-secret-42" {
		t.Errorf("Get = %q, want top-secret-42", val)
	}
}

func TestIntegrationVaultGetMultiField(t *testing.T) {
	skipIfNoDocker(t)
	skipIfNoCLI(t, "vault")

	addr := "http://localhost:8200"
	token := "test-root-token"

	if err := waitForService(addr+"/v1/sys/health", 30*time.Second); err != nil {
		t.Skip("Vault not running")
	}

	// Seed a secret with multiple fields (no single "value" field)
	payload := map[string]any{
		"data": map[string]string{
			"username": "admin",
			"password": "s3cret",
		},
	}
	body, _ := json.Marshal(payload)
	url := fmt.Sprintf("%s/v1/secret/data/vaulty-test/multi", addr)
	req, _ := http.NewRequest("POST", url, bytes.NewReader(body))
	req.Header.Set("X-Vault-Token", token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("seeding: %v", err)
	}
	resp.Body.Close()

	t.Setenv("VAULT_TOKEN", token)

	b, err := NewHashiCorpBackend(addr, "secret")
	if err != nil {
		t.Fatalf("NewHashiCorpBackend: %v", err)
	}

	val, err := b.Get("vaulty-test/multi")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	// Should return JSON since there's no single "value" field
	var data map[string]string
	if err := json.Unmarshal([]byte(val), &data); err != nil {
		t.Fatalf("expected JSON response, got %q: %v", val, err)
	}
	if data["username"] != "admin" {
		t.Errorf("username = %q, want admin", data["username"])
	}
	if data["password"] != "s3cret" {
		t.Errorf("password = %q, want s3cret", data["password"])
	}
}

func TestIntegrationVaultCachedBackend(t *testing.T) {
	skipIfNoDocker(t)
	skipIfNoCLI(t, "vault")

	addr := "http://localhost:8200"
	token := "test-root-token"

	if err := waitForService(addr+"/v1/sys/health", 30*time.Second); err != nil {
		t.Skip("Vault not running")
	}

	seedVaultSecret(t, addr, token, "secret", "vaulty-test/cache-test", "value", "cache-me")

	t.Setenv("VAULT_TOKEN", token)

	b, err := NewHashiCorpBackend(addr, "secret")
	if err != nil {
		t.Fatalf("NewHashiCorpBackend: %v", err)
	}

	cached := NewCachedBackend(b, 1*time.Minute)
	defer cached.Zero()

	val, err := cached.Get("vaulty-test/cache-test")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if val != "cache-me" {
		t.Errorf("Get = %q, want cache-me", val)
	}

	// Second call — cached
	val2, err := cached.Get("vaulty-test/cache-test")
	if err != nil {
		t.Fatalf("cached Get: %v", err)
	}
	if val2 != val {
		t.Errorf("cached Get = %q, want %q", val2, val)
	}
}

// --- Factory integration test ---

func TestIntegrationFactory(t *testing.T) {
	skipIfNoDocker(t)
	skipIfNoCLI(t, "vault")

	addr := "http://localhost:8200"
	token := "test-root-token"

	if err := waitForService(addr+"/v1/sys/health", 30*time.Second); err != nil {
		t.Skip("Vault not running")
	}

	seedVaultSecret(t, addr, token, "secret", "vaulty-test/factory", "value", "from-factory")

	t.Setenv("VAULT_TOKEN", token)

	b, err := NewBackend(BackendConfig{
		Type:  "hashicorp-vault",
		Addr:  addr,
		Mount: "secret",
		TTL:   "30s",
	})
	if err != nil {
		t.Fatalf("NewBackend: %v", err)
	}

	val, err := b.Get("vaulty-test/factory")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if val != "from-factory" {
		t.Errorf("Get = %q, want from-factory", val)
	}
}

// --- End-to-end: full backend workflow ---

func TestIntegrationE2EWorkflow(t *testing.T) {
	skipIfNoDocker(t)
	skipIfNoCLI(t, "aws")
	skipIfNoCLI(t, "vault")

	awsEndpoint := "http://localhost:4566"
	vaultAddr := "http://localhost:8200"
	vaultToken := "test-root-token"

	if err := waitForService(awsEndpoint+"/_localstack/health", 30*time.Second); err != nil {
		t.Skip("LocalStack not running")
	}
	if err := waitForService(vaultAddr+"/v1/sys/health", 30*time.Second); err != nil {
		t.Skip("Vault not running")
	}

	t.Setenv("AWS_ACCESS_KEY_ID", "test")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "test")
	t.Setenv("VAULT_TOKEN", vaultToken)

	// Seed both backends
	ts := fmt.Sprintf("%d", time.Now().UnixNano())
	awsSecretName := "e2e-test/key-" + ts
	vaultSecretPath := "e2e-test/key-" + ts

	seedAWSSecret(t, awsEndpoint, awsSecretName, "aws-secret-"+ts)
	seedVaultSecret(t, vaultAddr, vaultToken, "secret", vaultSecretPath, "value", "vault-secret-"+ts)

	// AWS backend
	awsBackend, err := NewAWSBackendWithEndpoint("us-east-1", "", awsEndpoint)
	if err != nil {
		t.Fatalf("AWS backend: %v", err)
	}
	awsCached := NewCachedBackend(awsBackend, 1*time.Minute)
	defer awsCached.Zero()

	awsVal, err := awsCached.Get(awsSecretName)
	if err != nil {
		t.Fatalf("AWS Get: %v", err)
	}
	if awsVal != "aws-secret-"+ts {
		t.Errorf("AWS Get = %q, want aws-secret-%s", awsVal, ts)
	}

	// Vault backend
	vaultBackend, err := NewHashiCorpBackend(vaultAddr, "secret")
	if err != nil {
		t.Fatalf("Vault backend: %v", err)
	}
	vaultCached := NewCachedBackend(vaultBackend, 1*time.Minute)
	defer vaultCached.Zero()

	vaultVal, err := vaultCached.Get(vaultSecretPath)
	if err != nil {
		t.Fatalf("Vault Get: %v", err)
	}
	if vaultVal != "vault-secret-"+ts {
		t.Errorf("Vault Get = %q, want vault-secret-%s", vaultVal, ts)
	}

	// List both
	awsNames, err := awsCached.List()
	if err != nil {
		t.Fatalf("AWS List: %v", err)
	}
	if len(awsNames) == 0 {
		t.Error("AWS List returned no secrets")
	}

	vaultNames, err := vaultCached.List()
	if err != nil {
		t.Fatalf("Vault List: %v", err)
	}
	if len(vaultNames) == 0 {
		t.Error("Vault List returned no secrets")
	}

	// Verify Zero clears cache
	awsCached.Zero()
	vaultCached.Zero()
	if len(awsCached.cache) != 0 {
		t.Error("AWS cache not cleared after Zero")
	}
	if len(vaultCached.cache) != 0 {
		t.Error("Vault cache not cleared after Zero")
	}
}

func TestMain(m *testing.M) {
	// Check if we're running integration tests
	for _, arg := range os.Args {
		if arg == "-test.run" || arg == "-test.v" {
			break
		}
	}
	os.Exit(m.Run())
}
