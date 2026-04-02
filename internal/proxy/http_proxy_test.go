package proxy

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDoRequest(t *testing.T) {
	secret := "sk_live_test123"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify secret was injected
		auth := r.Header.Get("Authorization")
		if auth != "Bearer "+secret {
			t.Errorf("Authorization = %q, want Bearer %s", auth, secret)
		}

		// Verify custom header
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Content-Type = %q", r.Header.Get("Content-Type"))
		}

		// Echo back a response that contains the secret (to test redaction)
		w.Header().Set("X-Debug", secret)
		w.WriteHeader(200)
		w.Write([]byte(`{"key":"` + secret + `"}`))
	}))
	defer server.Close()

	redactor := NewRedactor(map[string]string{"API_KEY": secret})

	result, err := DoRequest(
		"POST",
		server.URL+"/v1/charges",
		map[string]string{"Content-Type": "application/json"},
		`{"amount":2000}`,
		secret,
		InjectBearer,
		"",
		redactor,
	)
	if err != nil {
		t.Fatalf("DoRequest: %v", err)
	}

	if result.StatusCode != 200 {
		t.Errorf("status = %d, want 200", result.StatusCode)
	}

	// Response body should be redacted
	if strings.Contains(result.Body, secret) {
		t.Error("response body should not contain raw secret")
	}
	if !strings.Contains(result.Body, "[VAULTY:API_KEY]") {
		t.Errorf("response body should contain redacted placeholder, got %q", result.Body)
	}

	// Response headers should be redacted
	if xDebug, ok := result.Headers["X-Debug"]; ok {
		if strings.Contains(xDebug, secret) {
			t.Error("response header should not contain raw secret")
		}
	}
}

func TestDoRequestBasicAuth(t *testing.T) {
	secret := "user:password"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Basic ") {
			t.Errorf("expected Basic auth, got %q", auth)
		}
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	}))
	defer server.Close()

	redactor := NewRedactor(map[string]string{"CRED": secret})

	result, err := DoRequest("GET", server.URL, nil, "", secret, InjectBasic, "", redactor)
	if err != nil {
		t.Fatalf("DoRequest: %v", err)
	}
	if result.StatusCode != 200 {
		t.Errorf("status = %d", result.StatusCode)
	}
}

func TestDoRequestCustomHeader(t *testing.T) {
	secret := "my-api-key-value"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-Key") != secret {
			t.Errorf("X-API-Key = %q", r.Header.Get("X-API-Key"))
		}
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	}))
	defer server.Close()

	redactor := NewRedactor(map[string]string{"KEY": secret})

	result, err := DoRequest("GET", server.URL, nil, "", secret, InjectHeader, "X-API-Key", redactor)
	if err != nil {
		t.Fatalf("DoRequest: %v", err)
	}
	if result.StatusCode != 200 {
		t.Errorf("status = %d", result.StatusCode)
	}
}
