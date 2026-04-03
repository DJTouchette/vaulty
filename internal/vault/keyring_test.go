package vault

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/zalando/go-keyring"
)

func init() {
	// Use the mock keyring backend for tests so they work in CI
	// and don't pollute the real OS keychain.
	keyring.MockInit()
}

func TestSaveAndGetPassphrase(t *testing.T) {
	service := "vaulty-test"
	account := "vault:/tmp/test.age"

	if err := SavePassphrase(service, account, "s3cret"); err != nil {
		t.Fatalf("SavePassphrase: %v", err)
	}

	got, err := GetPassphrase(service, account)
	if err != nil {
		t.Fatalf("GetPassphrase: %v", err)
	}
	if got != "s3cret" {
		t.Fatalf("GetPassphrase = %q, want %q", got, "s3cret")
	}
}

func TestGetPassphraseNotFound(t *testing.T) {
	service := "vaulty-test"
	account := "vault:/nonexistent"

	_, err := GetPassphrase(service, account)
	if err == nil {
		t.Fatal("GetPassphrase should fail for missing entry")
	}
}

func TestDeletePassphrase(t *testing.T) {
	service := "vaulty-test"
	account := "vault:/tmp/delete-test.age"

	if err := SavePassphrase(service, account, "todelete"); err != nil {
		t.Fatalf("SavePassphrase: %v", err)
	}

	if err := DeletePassphrase(service, account); err != nil {
		t.Fatalf("DeletePassphrase: %v", err)
	}

	_, err := GetPassphrase(service, account)
	if err == nil {
		t.Fatal("GetPassphrase should fail after DeletePassphrase")
	}
}

func TestHasPassphrase(t *testing.T) {
	service := "vaulty-test"
	account := "vault:/tmp/has-test.age"

	if HasPassphrase(service, account) {
		t.Fatal("HasPassphrase should be false before saving")
	}

	if err := SavePassphrase(service, account, "exists"); err != nil {
		t.Fatalf("SavePassphrase: %v", err)
	}

	if !HasPassphrase(service, account) {
		t.Fatal("HasPassphrase should be true after saving")
	}
}

func TestKeyringAccount(t *testing.T) {
	path := filepath.Join("home", "user", ".config", "vaulty", "vault.age")
	account := KeyringAccount(path)
	if !strings.HasPrefix(account, "vault:") {
		t.Fatalf("KeyringAccount = %q, want prefix vault:", account)
	}
	if !strings.Contains(account, "vault.age") {
		t.Fatalf("KeyringAccount = %q, should contain vault.age", account)
	}
}

func TestSaveOverwrite(t *testing.T) {
	service := "vaulty-test"
	account := "vault:/tmp/overwrite.age"

	if err := SavePassphrase(service, account, "first"); err != nil {
		t.Fatalf("SavePassphrase first: %v", err)
	}
	if err := SavePassphrase(service, account, "second"); err != nil {
		t.Fatalf("SavePassphrase second: %v", err)
	}

	got, err := GetPassphrase(service, account)
	if err != nil {
		t.Fatalf("GetPassphrase: %v", err)
	}
	if got != "second" {
		t.Fatalf("GetPassphrase = %q, want %q", got, "second")
	}
}
