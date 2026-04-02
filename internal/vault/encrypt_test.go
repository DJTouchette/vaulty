package vault

import (
	"testing"
)

func TestEncryptDecryptRoundtrip(t *testing.T) {
	passphrase := "test-passphrase-123"
	plaintext := []byte(`{"secrets":{"API_KEY":"sk_live_abc123"}}`)

	ciphertext, err := Encrypt(passphrase, plaintext)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	if len(ciphertext) == 0 {
		t.Fatal("ciphertext is empty")
	}

	decrypted, err := Decrypt(passphrase, ciphertext)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}

	if string(decrypted) != string(plaintext) {
		t.Fatalf("roundtrip mismatch: got %q, want %q", decrypted, plaintext)
	}
}

func TestDecryptWrongPassphrase(t *testing.T) {
	ciphertext, err := Encrypt("correct-password", []byte("secret data"))
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	_, err = Decrypt("wrong-password", ciphertext)
	if err == nil {
		t.Fatal("expected error with wrong passphrase, got nil")
	}
}
