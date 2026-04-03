package policy

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "vaulty.toml")

	content := `
[vault]
path = "~/.config/vaulty/vault.age"
idle_timeout = "4h"
socket = "/tmp/vaulty-test.sock"
http_port = 9999

[secrets.STRIPE_SECRET_KEY]
allowed_domains = ["api.stripe.com"]
inject_as = "bearer"

[secrets.DATABASE_URL]
allowed_commands = ["psql", "prisma"]
inject_as = "env"

[secrets.CUSTOM_KEY]
allowed_domains = ["api.example.com"]
inject_as = "header"
header_name = "X-API-Key"
`
	os.WriteFile(configPath, []byte(content), 0644)

	cfg, err := LoadOrDefault(configPath)
	if err != nil {
		t.Fatalf("LoadOrDefault: %v", err)
	}

	if cfg.Vault.IdleTimeout != "4h" {
		t.Errorf("IdleTimeout = %q, want 4h", cfg.Vault.IdleTimeout)
	}
	if cfg.Vault.HTTPPort != 9999 {
		t.Errorf("HTTPPort = %d, want 9999", cfg.Vault.HTTPPort)
	}

	stripe := cfg.GetSecretPolicy("STRIPE_SECRET_KEY")
	if len(stripe.AllowedDomains) != 1 || stripe.AllowedDomains[0] != "api.stripe.com" {
		t.Errorf("STRIPE AllowedDomains = %v", stripe.AllowedDomains)
	}
	if stripe.InjectAs != "bearer" {
		t.Errorf("STRIPE InjectAs = %q, want bearer", stripe.InjectAs)
	}

	db := cfg.GetSecretPolicy("DATABASE_URL")
	if len(db.AllowedCommands) != 2 {
		t.Errorf("DATABASE_URL AllowedCommands = %v", db.AllowedCommands)
	}

	custom := cfg.GetSecretPolicy("CUSTOM_KEY")
	if custom.HeaderName != "X-API-Key" {
		t.Errorf("CUSTOM_KEY HeaderName = %q, want X-API-Key", custom.HeaderName)
	}
}

func TestDefaultConfig(t *testing.T) {
	// Load from nonexistent directory — should return defaults
	cfg, err := LoadOrDefault("")
	if err != nil {
		t.Fatalf("LoadOrDefault: %v", err)
	}

	if cfg.Vault.Path != "~/.config/vaulty/vault.age" {
		t.Errorf("default Path = %q", cfg.Vault.Path)
	}
	if cfg.Vault.IdleTimeout != "8h" {
		t.Errorf("default IdleTimeout = %q", cfg.Vault.IdleTimeout)
	}
	if cfg.Vault.HTTPPort != 19876 {
		t.Errorf("default HTTPPort = %d", cfg.Vault.HTTPPort)
	}
}

func TestLoadOrDefaultFindsLocalVaultyDir(t *testing.T) {
	// Create a temp dir and chdir into it so the search path resolves locally.
	dir := t.TempDir()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(orig) })
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	// Create .vaulty/vaulty.toml
	vaultyDir := filepath.Join(dir, ".vaulty")
	if err := os.MkdirAll(vaultyDir, 0700); err != nil {
		t.Fatal(err)
	}

	content := `
[vault]
path = ".vaulty/vault.age"
idle_timeout = "2h"
`
	if err := os.WriteFile(filepath.Join(vaultyDir, "vaulty.toml"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadOrDefault("")
	if err != nil {
		t.Fatalf("LoadOrDefault: %v", err)
	}

	if cfg.Vault.Path != ".vaulty/vault.age" {
		t.Errorf("Vault.Path = %q, want .vaulty/vault.age", cfg.Vault.Path)
	}
	if cfg.Vault.IdleTimeout != "2h" {
		t.Errorf("IdleTimeout = %q, want 2h", cfg.Vault.IdleTimeout)
	}
	if cfg.Path() != filepath.Join(".vaulty", "vaulty.toml") {
		t.Errorf("config path = %q, want .vaulty/vaulty.toml", cfg.Path())
	}
}

func TestSetAndRemoveSecretPolicy(t *testing.T) {
	cfg := defaultConfig()

	cfg.SetSecretPolicy("MY_KEY", SecretPolicy{
		AllowedDomains: []string{"example.com"},
		InjectAs:       "bearer",
	})

	sp := cfg.GetSecretPolicy("MY_KEY")
	if len(sp.AllowedDomains) != 1 {
		t.Errorf("after Set: AllowedDomains = %v", sp.AllowedDomains)
	}

	cfg.RemoveSecretPolicy("MY_KEY")
	sp = cfg.GetSecretPolicy("MY_KEY")
	if len(sp.AllowedDomains) != 0 {
		t.Errorf("after Remove: AllowedDomains should be empty, got %v", sp.AllowedDomains)
	}
}

func TestLoadConfigWithBackends(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "vaulty.toml")

	content := `
[vault]
path = "~/.config/vaulty/vault.age"

[backends.aws-prod]
type = "aws-secrets-manager"
region = "us-east-1"
profile = "production"
ttl = "10m"

[backends.gcp-staging]
type = "gcp-secret-manager"
project = "my-gcp-project"

[backends.hcv]
type = "hashicorp-vault"
addr = "https://vault.example.com"
mount = "kv"

[backends.onepass]
type = "1password"
op_vault = "Engineering"
ttl = "15m"

[secrets.API_KEY]
allowed_domains = ["api.example.com"]
inject_as = "bearer"
backend = "aws-prod"
`
	os.WriteFile(configPath, []byte(content), 0644)

	cfg, err := LoadOrDefault(configPath)
	if err != nil {
		t.Fatalf("LoadOrDefault: %v", err)
	}

	if len(cfg.Backends) != 4 {
		t.Fatalf("expected 4 backends, got %d", len(cfg.Backends))
	}

	aws := cfg.Backends["aws-prod"]
	if aws.Type != "aws-secrets-manager" {
		t.Errorf("aws-prod Type = %q", aws.Type)
	}
	if aws.Region != "us-east-1" {
		t.Errorf("aws-prod Region = %q", aws.Region)
	}
	if aws.Profile != "production" {
		t.Errorf("aws-prod Profile = %q", aws.Profile)
	}
	if aws.TTL != "10m" {
		t.Errorf("aws-prod TTL = %q", aws.TTL)
	}

	gcp := cfg.Backends["gcp-staging"]
	if gcp.Type != "gcp-secret-manager" {
		t.Errorf("gcp-staging Type = %q", gcp.Type)
	}
	if gcp.Project != "my-gcp-project" {
		t.Errorf("gcp-staging Project = %q", gcp.Project)
	}

	hcv := cfg.Backends["hcv"]
	if hcv.Addr != "https://vault.example.com" {
		t.Errorf("hcv Addr = %q", hcv.Addr)
	}
	if hcv.Mount != "kv" {
		t.Errorf("hcv Mount = %q", hcv.Mount)
	}

	op := cfg.Backends["onepass"]
	if op.OpVault != "Engineering" {
		t.Errorf("onepass OpVault = %q", op.OpVault)
	}

	apiKey := cfg.GetSecretPolicy("API_KEY")
	if apiKey.Backend != "aws-prod" {
		t.Errorf("API_KEY Backend = %q, want aws-prod", apiKey.Backend)
	}
}

func TestWriteAndReloadWithBackends(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "vaulty.toml")

	cfg := defaultConfig()
	cfg.path = configPath
	cfg.Backends = map[string]BackendConfig{
		"my-aws": {
			Type:   "aws-secrets-manager",
			Region: "eu-west-1",
			TTL:    "3m",
		},
	}
	cfg.SetSecretPolicy("DB_PASS", SecretPolicy{
		Backend: "my-aws",
	})

	if err := cfg.Write(); err != nil {
		t.Fatalf("Write: %v", err)
	}

	cfg2, err := LoadOrDefault(configPath)
	if err != nil {
		t.Fatalf("Reload: %v", err)
	}

	if len(cfg2.Backends) != 1 {
		t.Fatalf("expected 1 backend after reload, got %d", len(cfg2.Backends))
	}

	aws := cfg2.Backends["my-aws"]
	if aws.Region != "eu-west-1" {
		t.Errorf("after reload: Region = %q", aws.Region)
	}

	sp := cfg2.GetSecretPolicy("DB_PASS")
	if sp.Backend != "my-aws" {
		t.Errorf("after reload: Backend = %q", sp.Backend)
	}
}

func TestLoadYAMLConfig(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "vaulty.yaml")

	content := `
vault:
  path: "~/.config/vaulty/vault.age"
  idle_timeout: "4h"
  socket: "/tmp/vaulty-test.sock"
  http_port: 9999

secrets:
  STRIPE_SECRET_KEY:
    allowed_domains:
      - api.stripe.com
    inject_as: bearer
  DATABASE_URL:
    allowed_commands:
      - psql
      - prisma
    inject_as: env
  CUSTOM_KEY:
    allowed_domains:
      - api.example.com
    inject_as: header
    header_name: X-API-Key
`
	os.WriteFile(configPath, []byte(content), 0644)

	cfg, err := LoadOrDefault(configPath)
	if err != nil {
		t.Fatalf("LoadOrDefault YAML: %v", err)
	}

	if cfg.Vault.IdleTimeout != "4h" {
		t.Errorf("IdleTimeout = %q, want 4h", cfg.Vault.IdleTimeout)
	}
	if cfg.Vault.HTTPPort != 9999 {
		t.Errorf("HTTPPort = %d, want 9999", cfg.Vault.HTTPPort)
	}

	stripe := cfg.GetSecretPolicy("STRIPE_SECRET_KEY")
	if len(stripe.AllowedDomains) != 1 || stripe.AllowedDomains[0] != "api.stripe.com" {
		t.Errorf("STRIPE AllowedDomains = %v", stripe.AllowedDomains)
	}
	if stripe.InjectAs != "bearer" {
		t.Errorf("STRIPE InjectAs = %q, want bearer", stripe.InjectAs)
	}

	db := cfg.GetSecretPolicy("DATABASE_URL")
	if len(db.AllowedCommands) != 2 {
		t.Errorf("DATABASE_URL AllowedCommands = %v", db.AllowedCommands)
	}

	custom := cfg.GetSecretPolicy("CUSTOM_KEY")
	if custom.HeaderName != "X-API-Key" {
		t.Errorf("CUSTOM_KEY HeaderName = %q, want X-API-Key", custom.HeaderName)
	}
}

func TestLoadYMLExtension(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "vaulty.yml")

	content := `
vault:
  path: "~/.config/vaulty/vault.age"
  idle_timeout: "6h"
`
	os.WriteFile(configPath, []byte(content), 0644)

	cfg, err := LoadOrDefault(configPath)
	if err != nil {
		t.Fatalf("LoadOrDefault .yml: %v", err)
	}

	if cfg.Vault.IdleTimeout != "6h" {
		t.Errorf("IdleTimeout = %q, want 6h", cfg.Vault.IdleTimeout)
	}
}

func TestYAMLWriteAndReload(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "vaulty.yaml")

	cfg := defaultConfig()
	cfg.path = configPath
	cfg.SetSecretPolicy("TEST_KEY", SecretPolicy{
		AllowedDomains: []string{"api.test.com"},
		InjectAs:       "bearer",
	})

	if err := cfg.Write(); err != nil {
		t.Fatalf("Write YAML: %v", err)
	}

	// Verify the file is valid YAML
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("YAML file is empty")
	}

	cfg2, err := LoadOrDefault(configPath)
	if err != nil {
		t.Fatalf("Reload YAML: %v", err)
	}

	sp := cfg2.GetSecretPolicy("TEST_KEY")
	if len(sp.AllowedDomains) != 1 || sp.AllowedDomains[0] != "api.test.com" {
		t.Errorf("after YAML reload: AllowedDomains = %v", sp.AllowedDomains)
	}
	if sp.InjectAs != "bearer" {
		t.Errorf("after YAML reload: InjectAs = %q, want bearer", sp.InjectAs)
	}
}

func TestYAMLWithBackends(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "vaulty.yaml")

	content := `
vault:
  path: "~/.config/vaulty/vault.age"

backends:
  aws-prod:
    type: aws-secrets-manager
    region: us-east-1
    profile: production
    ttl: "10m"

secrets:
  API_KEY:
    allowed_domains:
      - api.example.com
    inject_as: bearer
    backend: aws-prod
`
	os.WriteFile(configPath, []byte(content), 0644)

	cfg, err := LoadOrDefault(configPath)
	if err != nil {
		t.Fatalf("LoadOrDefault YAML with backends: %v", err)
	}

	if len(cfg.Backends) != 1 {
		t.Fatalf("expected 1 backend, got %d", len(cfg.Backends))
	}

	aws := cfg.Backends["aws-prod"]
	if aws.Type != "aws-secrets-manager" {
		t.Errorf("aws-prod Type = %q", aws.Type)
	}
	if aws.Region != "us-east-1" {
		t.Errorf("aws-prod Region = %q", aws.Region)
	}

	apiKey := cfg.GetSecretPolicy("API_KEY")
	if apiKey.Backend != "aws-prod" {
		t.Errorf("API_KEY Backend = %q, want aws-prod", apiKey.Backend)
	}
}

func TestSearchFindsYAMLBeforeFallback(t *testing.T) {
	dir := t.TempDir()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(orig) })
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	// Create vaulty.yaml (no vaulty.toml)
	content := `
vault:
  path: "test-vault.age"
  idle_timeout: "3h"
`
	if err := os.WriteFile("vaulty.yaml", []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadOrDefault("")
	if err != nil {
		t.Fatalf("LoadOrDefault: %v", err)
	}

	if cfg.Vault.Path != "test-vault.age" {
		t.Errorf("Vault.Path = %q, want test-vault.age", cfg.Vault.Path)
	}
	if cfg.Vault.IdleTimeout != "3h" {
		t.Errorf("IdleTimeout = %q, want 3h", cfg.Vault.IdleTimeout)
	}
}

func TestTOMLTakesPrecedenceOverYAML(t *testing.T) {
	dir := t.TempDir()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(orig) })
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	// Create both vaulty.toml and vaulty.yaml
	os.WriteFile("vaulty.toml", []byte(`
[vault]
idle_timeout = "1h"
`), 0644)
	os.WriteFile("vaulty.yaml", []byte(`
vault:
  idle_timeout: "2h"
`), 0644)

	cfg, err := LoadOrDefault("")
	if err != nil {
		t.Fatalf("LoadOrDefault: %v", err)
	}

	// TOML should win
	if cfg.Vault.IdleTimeout != "1h" {
		t.Errorf("IdleTimeout = %q, want 1h (TOML should take precedence)", cfg.Vault.IdleTimeout)
	}
}

func TestWriteAndReload(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "vaulty.toml")

	cfg := defaultConfig()
	cfg.path = configPath
	cfg.SetSecretPolicy("TEST_KEY", SecretPolicy{
		AllowedDomains: []string{"api.test.com"},
		InjectAs:       "bearer",
	})

	if err := cfg.Write(); err != nil {
		t.Fatalf("Write: %v", err)
	}

	cfg2, err := LoadOrDefault(configPath)
	if err != nil {
		t.Fatalf("Reload: %v", err)
	}

	sp := cfg2.GetSecretPolicy("TEST_KEY")
	if len(sp.AllowedDomains) != 1 || sp.AllowedDomains[0] != "api.test.com" {
		t.Errorf("after reload: AllowedDomains = %v", sp.AllowedDomains)
	}
}
