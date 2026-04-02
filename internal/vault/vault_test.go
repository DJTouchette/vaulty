package vault

import (
	"os"
	"path/filepath"
	"testing"
)

func TestVaultCRUD(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.age")
	pass := "test-pass"

	// Create
	if err := Create(path, pass); err != nil {
		t.Fatalf("Create: %v", err)
	}

	if !Exists(path) {
		t.Fatal("vault should exist after Create")
	}

	// Open empty vault
	v, err := Open(path, pass)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	if len(v.List()) != 0 {
		t.Fatalf("new vault should be empty, got %d secrets", len(v.List()))
	}

	// Set secrets
	v.Set("API_KEY", "sk_live_abc123")
	v.Set("DB_URL", "postgres://localhost/db")

	if !v.Has("API_KEY") {
		t.Fatal("Has(API_KEY) should be true")
	}

	val, ok := v.Get("API_KEY")
	if !ok || val != "sk_live_abc123" {
		t.Fatalf("Get(API_KEY) = %q, %v; want sk_live_abc123, true", val, ok)
	}

	if len(v.List()) != 2 {
		t.Fatalf("List() should return 2 secrets, got %d", len(v.List()))
	}

	// Save and reopen
	if err := v.Save(path, pass); err != nil {
		t.Fatalf("Save: %v", err)
	}
	v.Zero()

	v2, err := Open(path, pass)
	if err != nil {
		t.Fatalf("Open after save: %v", err)
	}
	defer v2.Zero()

	val, ok = v2.Get("API_KEY")
	if !ok || val != "sk_live_abc123" {
		t.Fatalf("after reopen: Get(API_KEY) = %q, %v", val, ok)
	}

	// Remove
	v2.Remove("API_KEY")
	if v2.Has("API_KEY") {
		t.Fatal("Has(API_KEY) should be false after Remove")
	}
	if len(v2.List()) != 1 {
		t.Fatalf("List() should return 1 after Remove, got %d", len(v2.List()))
	}
}

func TestVaultZero(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.age")
	pass := "test-pass"

	if err := Create(path, pass); err != nil {
		t.Fatalf("Create: %v", err)
	}

	v, err := Open(path, pass)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	v.Set("SECRET", "mysecretvalue")

	// Keep a reference to the underlying bytes
	raw := v.secrets["SECRET"]

	v.Zero()

	// All bytes should be zeroed
	for i, b := range raw {
		if b != 0 {
			t.Fatalf("byte %d not zeroed: got %d", i, b)
		}
	}

	if len(v.List()) != 0 {
		t.Fatal("List() should be empty after Zero")
	}
}

func TestVaultWrongPassphrase(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.age")

	if err := Create(path, "correct"); err != nil {
		t.Fatalf("Create: %v", err)
	}

	_, err := Open(path, "wrong")
	if err == nil {
		t.Fatal("Open with wrong passphrase should fail")
	}
}

func TestResolveVaultPath(t *testing.T) {
	// Empty name returns basePath unchanged
	base := filepath.Join("home", "user", ".config", "vaulty", "vault.age")
	if got := ResolveVaultPath("", base); got != base {
		t.Errorf("ResolveVaultPath('', base) = %q, want %q", got, base)
	}

	// Named vault returns vaults/<name>.age under the same directory
	got := ResolveVaultPath("staging", base)
	want := filepath.Join("home", "user", ".config", "vaulty", "vaults", "staging.age")
	if got != want {
		t.Errorf("ResolveVaultPath('staging', base) = %q, want %q", got, want)
	}

	// Works with per-project paths too
	projectBase := filepath.Join("projects", "myapp", ".vaulty", "vault.age")
	got = ResolveVaultPath("prod", projectBase)
	want = filepath.Join("projects", "myapp", ".vaulty", "vaults", "prod.age")
	if got != want {
		t.Errorf("ResolveVaultPath('prod', projectBase) = %q, want %q", got, want)
	}
}

func TestNamedVaultCRUD(t *testing.T) {
	dir := t.TempDir()
	basePath := filepath.Join(dir, "vault.age")
	pass := "test-pass"

	// Resolve named vault path
	namedPath := ResolveVaultPath("staging", basePath)
	wantPath := filepath.Join(dir, "vaults", "staging.age")
	if namedPath != wantPath {
		t.Fatalf("namedPath = %q, want %q", namedPath, wantPath)
	}

	// Create and use named vault
	if err := Create(namedPath, pass); err != nil {
		t.Fatalf("Create named vault: %v", err)
	}

	if !Exists(namedPath) {
		t.Fatal("named vault should exist after Create")
	}

	v, err := Open(namedPath, pass)
	if err != nil {
		t.Fatalf("Open named vault: %v", err)
	}

	v.Set("STAGING_KEY", "staging-secret")
	if err := v.Save(namedPath, pass); err != nil {
		t.Fatalf("Save named vault: %v", err)
	}
	v.Zero()

	v2, err := Open(namedPath, pass)
	if err != nil {
		t.Fatalf("Reopen named vault: %v", err)
	}
	defer v2.Zero()

	val, ok := v2.Get("STAGING_KEY")
	if !ok || val != "staging-secret" {
		t.Fatalf("Get(STAGING_KEY) = %q, %v; want staging-secret, true", val, ok)
	}
}

func TestExists(t *testing.T) {
	if Exists("/nonexistent/path/vault.age") {
		t.Fatal("Exists should return false for nonexistent path")
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "test.age")
	os.WriteFile(path, []byte("data"), 0600)

	if !Exists(path) {
		t.Fatal("Exists should return true for existing path")
	}
}
