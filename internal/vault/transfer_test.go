package vault

import (
	"os"
	"path/filepath"
	"testing"

	"filippo.io/age"
)

func TestExportImportRoundTrip(t *testing.T) {
	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "test.age")
	pass := "test-pass"

	// Create a vault with some secrets
	if err := Create(vaultPath, pass); err != nil {
		t.Fatalf("Create: %v", err)
	}
	v, err := Open(vaultPath, pass)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	v.Set("API_KEY", "sk_live_abc123")
	v.Set("DB_URL", "postgres://localhost/db")
	if err := v.Save(vaultPath, pass); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Export with passphrase only (no recipients)
	exported, err := Export(v, pass, nil)
	if err != nil {
		t.Fatalf("Export: %v", err)
	}
	v.Zero()

	if len(exported) == 0 {
		t.Fatal("exported data should not be empty")
	}

	// Import the exported data
	imported, err := Import(exported, pass, "")
	if err != nil {
		t.Fatalf("Import: %v", err)
	}
	defer imported.Zero()

	// Verify secrets
	val, ok := imported.Get("API_KEY")
	if !ok || val != "sk_live_abc123" {
		t.Errorf("Get(API_KEY) = %q, %v; want sk_live_abc123, true", val, ok)
	}

	val, ok = imported.Get("DB_URL")
	if !ok || val != "postgres://localhost/db" {
		t.Errorf("Get(DB_URL) = %q, %v; want postgres://localhost/db, true", val, ok)
	}

	if len(imported.List()) != 2 {
		t.Errorf("imported vault should have 2 secrets, got %d", len(imported.List()))
	}
}

func TestExportImportWithIdentity(t *testing.T) {
	dir := t.TempDir()

	// Generate an age key pair
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatalf("GenerateX25519Identity: %v", err)
	}
	recipient := identity.Recipient()

	// Write identity file
	identityPath := filepath.Join(dir, "key.txt")
	content := "# created by vaulty test\n"
	content += "# public key: " + recipient.String() + "\n"
	content += identity.String() + "\n"
	if err := os.WriteFile(identityPath, []byte(content), 0600); err != nil {
		t.Fatalf("writing identity file: %v", err)
	}

	// Create a vault with secrets
	vaultPath := filepath.Join(dir, "test.age")
	pass := "test-pass"
	if err := Create(vaultPath, pass); err != nil {
		t.Fatalf("Create: %v", err)
	}
	v, err := Open(vaultPath, pass)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	v.Set("SECRET", "myvalue")

	// Export with recipients
	exported, err := Export(v, pass, []age.Recipient{recipient})
	if err != nil {
		t.Fatalf("Export with recipients: %v", err)
	}
	v.Zero()

	// Import using identity file
	imported, err := Import(exported, "", identityPath)
	if err != nil {
		t.Fatalf("Import with identity: %v", err)
	}
	defer imported.Zero()

	val, ok := imported.Get("SECRET")
	if !ok || val != "myvalue" {
		t.Errorf("Get(SECRET) = %q, %v; want myvalue, true", val, ok)
	}
}

func TestImportWrongPassphrase(t *testing.T) {
	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "test.age")
	pass := "correct"

	if err := Create(vaultPath, pass); err != nil {
		t.Fatalf("Create: %v", err)
	}
	v, err := Open(vaultPath, pass)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	v.Set("KEY", "val")

	exported, err := Export(v, pass, nil)
	if err != nil {
		t.Fatalf("Export: %v", err)
	}
	v.Zero()

	_, err = Import(exported, "wrong", "")
	if err == nil {
		t.Fatal("Import with wrong passphrase should fail")
	}
}

func TestMergeVaultsNoOverwrite(t *testing.T) {
	dir := t.TempDir()
	pass := "test-pass"

	// Create dst vault
	dstPath := filepath.Join(dir, "dst.age")
	if err := Create(dstPath, pass); err != nil {
		t.Fatalf("Create dst: %v", err)
	}
	dst, err := Open(dstPath, pass)
	if err != nil {
		t.Fatalf("Open dst: %v", err)
	}
	dst.Set("SHARED", "dst-value")
	dst.Set("DST_ONLY", "dst-only")

	// Create src vault
	srcPath := filepath.Join(dir, "src.age")
	if err := Create(srcPath, pass); err != nil {
		t.Fatalf("Create src: %v", err)
	}
	src, err := Open(srcPath, pass)
	if err != nil {
		t.Fatalf("Open src: %v", err)
	}
	src.Set("SHARED", "src-value")
	src.Set("SRC_ONLY", "src-only")

	// Merge without overwrite
	count := MergeVaults(dst, src, false)

	// Should have added SRC_ONLY but not overwritten SHARED
	if count != 1 {
		t.Errorf("MergeVaults count = %d, want 1", count)
	}

	val, _ := dst.Get("SHARED")
	if val != "dst-value" {
		t.Errorf("SHARED = %q, want dst-value (should not be overwritten)", val)
	}

	val, ok := dst.Get("SRC_ONLY")
	if !ok || val != "src-only" {
		t.Errorf("SRC_ONLY = %q, %v; want src-only, true", val, ok)
	}

	val, _ = dst.Get("DST_ONLY")
	if val != "dst-only" {
		t.Errorf("DST_ONLY = %q, want dst-only", val)
	}

	dst.Zero()
	src.Zero()
}

func TestMergeVaultsWithOverwrite(t *testing.T) {
	dir := t.TempDir()
	pass := "test-pass"

	// Create dst vault
	dstPath := filepath.Join(dir, "dst.age")
	if err := Create(dstPath, pass); err != nil {
		t.Fatalf("Create dst: %v", err)
	}
	dst, err := Open(dstPath, pass)
	if err != nil {
		t.Fatalf("Open dst: %v", err)
	}
	dst.Set("SHARED", "dst-value")
	dst.Set("DST_ONLY", "dst-only")

	// Create src vault
	srcPath := filepath.Join(dir, "src.age")
	if err := Create(srcPath, pass); err != nil {
		t.Fatalf("Create src: %v", err)
	}
	src, err := Open(srcPath, pass)
	if err != nil {
		t.Fatalf("Open src: %v", err)
	}
	src.Set("SHARED", "src-value")
	src.Set("SRC_ONLY", "src-only")

	// Merge with overwrite
	count := MergeVaults(dst, src, true)

	if count != 2 {
		t.Errorf("MergeVaults count = %d, want 2", count)
	}

	val, _ := dst.Get("SHARED")
	if val != "src-value" {
		t.Errorf("SHARED = %q, want src-value (should be overwritten)", val)
	}

	val, ok := dst.Get("SRC_ONLY")
	if !ok || val != "src-only" {
		t.Errorf("SRC_ONLY = %q, %v; want src-only, true", val, ok)
	}

	val, _ = dst.Get("DST_ONLY")
	if val != "dst-only" {
		t.Errorf("DST_ONLY = %q, want dst-only", val)
	}

	dst.Zero()
	src.Zero()
}

func TestMergeVaultsEmpty(t *testing.T) {
	dir := t.TempDir()
	pass := "test-pass"

	dstPath := filepath.Join(dir, "dst.age")
	if err := Create(dstPath, pass); err != nil {
		t.Fatalf("Create dst: %v", err)
	}
	dst, err := Open(dstPath, pass)
	if err != nil {
		t.Fatalf("Open dst: %v", err)
	}
	dst.Set("EXISTING", "value")

	srcPath := filepath.Join(dir, "src.age")
	if err := Create(srcPath, pass); err != nil {
		t.Fatalf("Create src: %v", err)
	}
	src, err := Open(srcPath, pass)
	if err != nil {
		t.Fatalf("Open src: %v", err)
	}

	// Merge empty src into dst
	count := MergeVaults(dst, src, true)
	if count != 0 {
		t.Errorf("MergeVaults count = %d, want 0 for empty source", count)
	}

	if len(dst.List()) != 1 {
		t.Errorf("dst should still have 1 secret, got %d", len(dst.List()))
	}

	dst.Zero()
	src.Zero()
}
