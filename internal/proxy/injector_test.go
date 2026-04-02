package proxy

import (
	"encoding/base64"
	"net/http"
	"strings"
	"testing"
)

func TestInjectBearer(t *testing.T) {
	req, _ := http.NewRequest("GET", "https://api.example.com", nil)
	err := InjectSecret(req, "sk_test_123", InjectBearer, "")
	if err != nil {
		t.Fatalf("InjectSecret: %v", err)
	}

	got := req.Header.Get("Authorization")
	want := "Bearer sk_test_123"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestInjectBasic(t *testing.T) {
	req, _ := http.NewRequest("GET", "https://api.example.com", nil)
	secret := "user:pass"
	err := InjectSecret(req, secret, InjectBasic, "")
	if err != nil {
		t.Fatalf("InjectSecret: %v", err)
	}

	got := req.Header.Get("Authorization")
	wantPrefix := "Basic "
	if !strings.HasPrefix(got, wantPrefix) {
		t.Fatalf("got %q, want prefix %q", got, wantPrefix)
	}

	encoded := strings.TrimPrefix(got, wantPrefix)
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if string(decoded) != secret {
		t.Errorf("decoded %q, want %q", decoded, secret)
	}
}

func TestInjectHeader(t *testing.T) {
	req, _ := http.NewRequest("GET", "https://api.example.com", nil)
	err := InjectSecret(req, "my-api-key", InjectHeader, "X-API-Key")
	if err != nil {
		t.Fatalf("InjectSecret: %v", err)
	}

	got := req.Header.Get("X-API-Key")
	if got != "my-api-key" {
		t.Errorf("got %q, want %q", got, "my-api-key")
	}
}

func TestInjectHeaderMissingName(t *testing.T) {
	req, _ := http.NewRequest("GET", "https://api.example.com", nil)
	err := InjectSecret(req, "my-api-key", InjectHeader, "")
	if err == nil {
		t.Fatal("expected error for missing header_name")
	}
}

func TestInjectQuery(t *testing.T) {
	req, _ := http.NewRequest("GET", "https://api.example.com/path?existing=1", nil)
	err := InjectSecret(req, "my-key", InjectQuery, "")
	if err != nil {
		t.Fatalf("InjectSecret: %v", err)
	}

	got := req.URL.Query().Get("key")
	if got != "my-key" {
		t.Errorf("got %q, want %q", got, "my-key")
	}

	// Existing params preserved
	if req.URL.Query().Get("existing") != "1" {
		t.Error("existing query param lost")
	}
}

func TestInjectDefaultIsBearer(t *testing.T) {
	req, _ := http.NewRequest("GET", "https://api.example.com", nil)
	err := InjectSecret(req, "sk_test_123", "", "")
	if err != nil {
		t.Fatalf("InjectSecret: %v", err)
	}

	got := req.Header.Get("Authorization")
	if got != "Bearer sk_test_123" {
		t.Errorf("default mode: got %q", got)
	}
}
