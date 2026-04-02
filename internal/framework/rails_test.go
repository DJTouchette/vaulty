package framework

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseRailsCredentials(t *testing.T) {
	yaml := `aws:
  access_key_id: AKIAIOSFODNN7EXAMPLE
  secret_access_key: wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
secret_key_base: abc123def456
`
	got, err := ParseRailsCredentials([]byte(yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := map[string]string{
		"AWS_ACCESS_KEY_ID":     "AKIAIOSFODNN7EXAMPLE",
		"AWS_SECRET_ACCESS_KEY": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		"SECRET_KEY_BASE":       "abc123def456",
	}

	for key, want := range tests {
		if got[key] != want {
			t.Errorf("%s = %q, want %q", key, got[key], want)
		}
	}
}

func TestFlattenYAML(t *testing.T) {
	data := map[string]interface{}{
		"database": map[string]interface{}{
			"host":     "localhost",
			"port":     5432,
			"username": "admin",
		},
		"api_key": "secret",
	}

	got := FlattenYAML("", data)

	tests := map[string]string{
		"DATABASE_HOST":     "localhost",
		"DATABASE_PORT":     "5432",
		"DATABASE_USERNAME": "admin",
		"API_KEY":           "secret",
	}

	for key, want := range tests {
		if got[key] != want {
			t.Errorf("%s = %q, want %q", key, got[key], want)
		}
	}
}

func TestFlattenYAMLNested(t *testing.T) {
	data := map[string]interface{}{
		"a": map[string]interface{}{
			"b": map[string]interface{}{
				"c": "deep",
			},
		},
	}

	got := FlattenYAML("", data)
	if got["A_B_C"] != "deep" {
		t.Errorf("A_B_C = %q, want %q", got["A_B_C"], "deep")
	}
}

func TestWriteRailsCredentials(t *testing.T) {
	secrets := map[string]string{
		"AWS_ACCESS_KEY_ID":     "AKID",
		"AWS_SECRET_ACCESS_KEY": "SECRET",
		"SECRET_KEY_BASE":       "abc123",
	}

	data, err := WriteRailsCredentials(secrets)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Parse it back to verify structure
	got, err := ParseRailsCredentials(data)
	if err != nil {
		t.Fatalf("re-parse error: %v", err)
	}

	for key, want := range secrets {
		if got[key] != want {
			t.Errorf("%s = %q, want %q", key, got[key], want)
		}
	}
}

func TestDecryptRailsCredentials(t *testing.T) {
	// Create a test encrypted payload using AES-256-GCM
	plaintext := []byte("test: value\n")

	// Generate a random 32-byte key
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatal(err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatal(err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		t.Fatal(err)
	}

	iv := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(iv); err != nil {
		t.Fatal(err)
	}

	sealed := gcm.Seal(nil, iv, plaintext, nil)
	// sealed = ciphertext + auth_tag
	ciphertext := sealed[:len(sealed)-gcm.Overhead()]
	authTag := sealed[len(sealed)-gcm.Overhead():]

	// Rails format: base64(encrypted)--base64(iv)--base64(tag)
	payload := base64.StdEncoding.EncodeToString(ciphertext) + "--" +
		base64.StdEncoding.EncodeToString(iv) + "--" +
		base64.StdEncoding.EncodeToString(authTag)

	dir := t.TempDir()
	encPath := filepath.Join(dir, "credentials.yml.enc")
	keyPath := filepath.Join(dir, "master.key")

	if err := os.WriteFile(encPath, []byte(payload), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(keyPath, []byte(hex.EncodeToString(key)), 0600); err != nil {
		t.Fatal(err)
	}

	got, err := DecryptRailsCredentials(encPath, keyPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if string(got) != string(plaintext) {
		t.Errorf("got %q, want %q", string(got), string(plaintext))
	}
}

func TestDecryptRailsCredentialsBadKey(t *testing.T) {
	dir := t.TempDir()
	encPath := filepath.Join(dir, "credentials.yml.enc")
	keyPath := filepath.Join(dir, "master.key")

	// Create a valid-looking payload with wrong key
	payload := base64.StdEncoding.EncodeToString([]byte("data")) + "--" +
		base64.StdEncoding.EncodeToString(make([]byte, 12)) + "--" +
		base64.StdEncoding.EncodeToString(make([]byte, 16))

	os.WriteFile(encPath, []byte(payload), 0600)
	os.WriteFile(keyPath, []byte(strings.Repeat("ab", 32)), 0600)

	_, err := DecryptRailsCredentials(encPath, keyPath)
	if err == nil {
		t.Fatal("expected error with wrong key")
	}
	if !strings.Contains(err.Error(), "wrong master key") {
		t.Errorf("error should mention wrong master key, got: %v", err)
	}
}
