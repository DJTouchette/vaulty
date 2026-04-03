package framework

import (
	"bytes"
	"encoding/base64"
	"strings"
	"testing"
)

func TestParseK8sSecret(t *testing.T) {
	manifest := `apiVersion: v1
kind: Secret
metadata:
  name: my-secret
  namespace: default
type: Opaque
data:
  API_KEY: ` + base64.StdEncoding.EncodeToString([]byte("sk-123")) + `
  DB_PASSWORD: ` + base64.StdEncoding.EncodeToString([]byte("s3cret")) + `
`

	got, err := ParseK8sSecret([]byte(manifest))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got["API_KEY"] != "sk-123" {
		t.Errorf("API_KEY = %q, want %q", got["API_KEY"], "sk-123")
	}
	if got["DB_PASSWORD"] != "s3cret" {
		t.Errorf("DB_PASSWORD = %q, want %q", got["DB_PASSWORD"], "s3cret")
	}
}

func TestParseK8sSecretWrongKind(t *testing.T) {
	manifest := `apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
data:
  KEY: value
`
	_, err := ParseK8sSecret([]byte(manifest))
	if err == nil {
		t.Fatal("expected error for wrong Kind")
	}
	if !strings.Contains(err.Error(), "ConfigMap") {
		t.Errorf("error should mention actual kind, got: %v", err)
	}
}

func TestParseK8sSecretBadBase64(t *testing.T) {
	manifest := `apiVersion: v1
kind: Secret
metadata:
  name: my-secret
data:
  KEY: not-valid-base64!!!
`
	_, err := ParseK8sSecret([]byte(manifest))
	if err == nil {
		t.Fatal("expected error for bad base64")
	}
}

func TestWriteK8sSecret(t *testing.T) {
	secrets := map[string]string{
		"API_KEY":     "sk-123",
		"DB_PASSWORD": "s3cret",
	}

	var buf bytes.Buffer
	if err := WriteK8sSecret(&buf, "my-secret", "production", secrets); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()

	if !strings.Contains(out, "apiVersion: v1") {
		t.Error("should contain apiVersion")
	}
	if !strings.Contains(out, "kind: Secret") {
		t.Error("should contain kind: Secret")
	}
	if !strings.Contains(out, "name: my-secret") {
		t.Error("should contain name")
	}
	if !strings.Contains(out, "namespace: production") {
		t.Error("should contain namespace")
	}
	if !strings.Contains(out, "type: Opaque") {
		t.Error("should contain type: Opaque")
	}

	// Verify base64 encoding
	wantAPIKey := base64.StdEncoding.EncodeToString([]byte("sk-123"))
	if !strings.Contains(out, "API_KEY: "+wantAPIKey) {
		t.Errorf("should contain base64-encoded API_KEY")
	}
}

func TestWriteK8sSecretNoNamespace(t *testing.T) {
	secrets := map[string]string{"KEY": "val"}

	var buf bytes.Buffer
	if err := WriteK8sSecret(&buf, "test", "", secrets); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(buf.String(), "namespace") {
		t.Error("should not contain namespace when empty")
	}
}

func TestK8sRoundTrip(t *testing.T) {
	original := map[string]string{
		"API_KEY":     "sk-test-123",
		"DB_PASSWORD": "p@ssw0rd!",
		"EMPTY":       "",
	}

	var buf bytes.Buffer
	if err := WriteK8sSecret(&buf, "test", "default", original); err != nil {
		t.Fatalf("write error: %v", err)
	}

	got, err := ParseK8sSecret(buf.Bytes())
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	for key, want := range original {
		if got[key] != want {
			t.Errorf("%s = %q, want %q", key, got[key], want)
		}
	}
}
