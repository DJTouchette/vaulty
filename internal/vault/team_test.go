package vault

import (
	"os"
	"path/filepath"
	"testing"

	"filippo.io/age"
)

// generateTestIdentity creates a fresh age X25519 identity for testing
// and returns the identity and its public key string.
func generateTestIdentity(t *testing.T) (*age.X25519Identity, string) {
	t.Helper()
	id, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatalf("generating identity: %v", err)
	}
	return id, id.Recipient().String()
}

func TestAddRecipient(t *testing.T) {
	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "vault.age")

	_, pubkey := generateTestIdentity(t)

	if err := AddRecipient(vaultPath, pubkey); err != nil {
		t.Fatalf("AddRecipient: %v", err)
	}

	recipients, err := ListRecipients(vaultPath)
	if err != nil {
		t.Fatalf("ListRecipients: %v", err)
	}

	if len(recipients) != 1 {
		t.Fatalf("expected 1 recipient, got %d", len(recipients))
	}
	if recipients[0] != pubkey {
		t.Fatalf("expected %s, got %s", pubkey, recipients[0])
	}
}

func TestAddRecipientDuplicate(t *testing.T) {
	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "vault.age")

	_, pubkey := generateTestIdentity(t)

	if err := AddRecipient(vaultPath, pubkey); err != nil {
		t.Fatalf("AddRecipient: %v", err)
	}

	err := AddRecipient(vaultPath, pubkey)
	if err == nil {
		t.Fatal("expected error adding duplicate recipient, got nil")
	}
}

func TestAddRecipientInvalidKey(t *testing.T) {
	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "vault.age")

	err := AddRecipient(vaultPath, "age1invalidkey")
	if err == nil {
		t.Fatal("expected error with invalid key, got nil")
	}
}

func TestAddRecipientFromFile(t *testing.T) {
	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "vault.age")

	id, pubkey := generateTestIdentity(t)

	// Write an identity file in the standard age format
	keyFile := filepath.Join(dir, "key.txt")
	content := "# created: 2024-01-01\n# public key: " + pubkey + "\n" + id.String() + "\n"
	if err := os.WriteFile(keyFile, []byte(content), 0600); err != nil {
		t.Fatalf("writing key file: %v", err)
	}

	if err := AddRecipient(vaultPath, keyFile); err != nil {
		t.Fatalf("AddRecipient from file: %v", err)
	}

	recipients, err := ListRecipients(vaultPath)
	if err != nil {
		t.Fatalf("ListRecipients: %v", err)
	}

	if len(recipients) != 1 {
		t.Fatalf("expected 1 recipient, got %d", len(recipients))
	}
	if recipients[0] != pubkey {
		t.Fatalf("expected %s, got %s", pubkey, recipients[0])
	}
}

func TestListRecipientsEmpty(t *testing.T) {
	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "vault.age")

	recipients, err := ListRecipients(vaultPath)
	if err != nil {
		t.Fatalf("ListRecipients: %v", err)
	}

	if len(recipients) != 0 {
		t.Fatalf("expected 0 recipients, got %d", len(recipients))
	}
}

func TestRemoveRecipient(t *testing.T) {
	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "vault.age")

	_, pubkey1 := generateTestIdentity(t)
	_, pubkey2 := generateTestIdentity(t)

	if err := AddRecipient(vaultPath, pubkey1); err != nil {
		t.Fatalf("AddRecipient: %v", err)
	}
	if err := AddRecipient(vaultPath, pubkey2); err != nil {
		t.Fatalf("AddRecipient: %v", err)
	}

	if err := RemoveRecipient(vaultPath, pubkey1); err != nil {
		t.Fatalf("RemoveRecipient: %v", err)
	}

	recipients, err := ListRecipients(vaultPath)
	if err != nil {
		t.Fatalf("ListRecipients: %v", err)
	}

	if len(recipients) != 1 {
		t.Fatalf("expected 1 recipient, got %d", len(recipients))
	}
	if recipients[0] != pubkey2 {
		t.Fatalf("expected %s, got %s", pubkey2, recipients[0])
	}
}

func TestRemoveRecipientNotFound(t *testing.T) {
	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "vault.age")

	_, pubkey := generateTestIdentity(t)

	if err := AddRecipient(vaultPath, pubkey); err != nil {
		t.Fatalf("AddRecipient: %v", err)
	}

	err := RemoveRecipient(vaultPath, "age1notexist")
	if err == nil {
		t.Fatal("expected error removing non-existent recipient, got nil")
	}
}

func TestRemoveLastRecipientDeletesFile(t *testing.T) {
	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "vault.age")

	_, pubkey := generateTestIdentity(t)

	if err := AddRecipient(vaultPath, pubkey); err != nil {
		t.Fatalf("AddRecipient: %v", err)
	}

	if err := RemoveRecipient(vaultPath, pubkey); err != nil {
		t.Fatalf("RemoveRecipient: %v", err)
	}

	// The recipients file should be removed
	path := recipientsFile(vaultPath)
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("expected recipients file to be removed")
	}
}

func TestLoadRecipients(t *testing.T) {
	dir := t.TempDir()
	vaultPath := filepath.Join(dir, "vault.age")

	_, pubkey1 := generateTestIdentity(t)
	_, pubkey2 := generateTestIdentity(t)

	if err := AddRecipient(vaultPath, pubkey1); err != nil {
		t.Fatalf("AddRecipient: %v", err)
	}
	if err := AddRecipient(vaultPath, pubkey2); err != nil {
		t.Fatalf("AddRecipient: %v", err)
	}

	recipients, err := LoadRecipients(vaultPath)
	if err != nil {
		t.Fatalf("LoadRecipients: %v", err)
	}

	if len(recipients) != 2 {
		t.Fatalf("expected 2 recipients, got %d", len(recipients))
	}
}

func TestEncryptMultiDecryptWithIdentity(t *testing.T) {
	id, pubkey := generateTestIdentity(t)

	recipient, err := age.ParseX25519Recipient(pubkey)
	if err != nil {
		t.Fatalf("ParseX25519Recipient: %v", err)
	}

	plaintext := []byte(`{"secrets":{"API_KEY":"sk_live_test"}}`)

	ciphertext, err := EncryptMulti("test-pass", []age.Recipient{recipient}, plaintext)
	if err != nil {
		t.Fatalf("EncryptMulti: %v", err)
	}

	if len(ciphertext) == 0 {
		t.Fatal("ciphertext is empty")
	}

	// Write identity to a temp file for DecryptWithIdentity
	dir := t.TempDir()
	keyFile := filepath.Join(dir, "key.txt")
	if err := os.WriteFile(keyFile, []byte(id.String()+"\n"), 0600); err != nil {
		t.Fatalf("writing key file: %v", err)
	}

	decrypted, err := DecryptWithIdentity(keyFile, ciphertext)
	if err != nil {
		t.Fatalf("DecryptWithIdentity: %v", err)
	}

	if string(decrypted) != string(plaintext) {
		t.Fatalf("roundtrip mismatch: got %q, want %q", decrypted, plaintext)
	}
}

func TestEncryptMultiNoRecipientsFallsBack(t *testing.T) {
	passphrase := "test-passphrase"
	plaintext := []byte("hello world")

	ciphertext, err := EncryptMulti(passphrase, nil, plaintext)
	if err != nil {
		t.Fatalf("EncryptMulti: %v", err)
	}

	// Should be decryptable with the passphrase
	decrypted, err := Decrypt(passphrase, ciphertext)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}

	if string(decrypted) != string(plaintext) {
		t.Fatalf("roundtrip mismatch: got %q, want %q", decrypted, plaintext)
	}
}

func TestEncryptMultiMultipleRecipients(t *testing.T) {
	id1, pubkey1 := generateTestIdentity(t)
	id2, pubkey2 := generateTestIdentity(t)

	r1, _ := age.ParseX25519Recipient(pubkey1)
	r2, _ := age.ParseX25519Recipient(pubkey2)

	plaintext := []byte("shared secret data")

	ciphertext, err := EncryptMulti("pass", []age.Recipient{r1, r2}, plaintext)
	if err != nil {
		t.Fatalf("EncryptMulti: %v", err)
	}

	dir := t.TempDir()

	// Both identities should be able to decrypt
	for i, id := range []*age.X25519Identity{id1, id2} {
		keyFile := filepath.Join(dir, "key"+string(rune('0'+i))+".txt")
		if err := os.WriteFile(keyFile, []byte(id.String()+"\n"), 0600); err != nil {
			t.Fatalf("writing key file %d: %v", i, err)
		}

		decrypted, err := DecryptWithIdentity(keyFile, ciphertext)
		if err != nil {
			t.Fatalf("DecryptWithIdentity (identity %d): %v", i, err)
		}

		if string(decrypted) != string(plaintext) {
			t.Fatalf("roundtrip mismatch for identity %d: got %q, want %q", i, decrypted, plaintext)
		}
	}
}
