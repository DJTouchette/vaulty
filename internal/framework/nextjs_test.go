package framework

import "testing"

func TestIsPublicEnvVar(t *testing.T) {
	tests := []struct {
		key  string
		want bool
	}{
		{"NEXT_PUBLIC_API_URL", true},
		{"NEXT_PUBLIC_", true},
		{"API_KEY", false},
		{"NEXT_PUBLICX", false},
		{"next_public_key", false},
		{"", false},
	}

	for _, tt := range tests {
		got := IsPublicEnvVar(tt.key)
		if got != tt.want {
			t.Errorf("IsPublicEnvVar(%q) = %v, want %v", tt.key, got, tt.want)
		}
	}
}

func TestClassifyNextJSEnv(t *testing.T) {
	secrets := map[string]string{
		"NEXT_PUBLIC_API_URL":    "https://api.example.com",
		"NEXT_PUBLIC_SITE_NAME": "My Site",
		"DATABASE_URL":           "postgres://localhost/db",
		"SECRET_KEY":             "abc123",
	}

	public, private := ClassifyNextJSEnv(secrets)

	if len(public) != 2 {
		t.Errorf("expected 2 public vars, got %d", len(public))
	}
	if len(private) != 2 {
		t.Errorf("expected 2 private vars, got %d", len(private))
	}

	if public["NEXT_PUBLIC_API_URL"] != "https://api.example.com" {
		t.Errorf("missing NEXT_PUBLIC_API_URL from public")
	}
	if public["NEXT_PUBLIC_SITE_NAME"] != "My Site" {
		t.Errorf("missing NEXT_PUBLIC_SITE_NAME from public")
	}
	if private["DATABASE_URL"] != "postgres://localhost/db" {
		t.Errorf("missing DATABASE_URL from private")
	}
	if private["SECRET_KEY"] != "abc123" {
		t.Errorf("missing SECRET_KEY from private")
	}
}

func TestClassifyNextJSEnvEmpty(t *testing.T) {
	public, private := ClassifyNextJSEnv(map[string]string{})
	if len(public) != 0 || len(private) != 0 {
		t.Error("expected empty maps for empty input")
	}
}

func TestClassifyNextJSEnvAllPublic(t *testing.T) {
	secrets := map[string]string{
		"NEXT_PUBLIC_A": "a",
		"NEXT_PUBLIC_B": "b",
	}

	public, private := ClassifyNextJSEnv(secrets)
	if len(public) != 2 {
		t.Errorf("expected 2 public, got %d", len(public))
	}
	if len(private) != 0 {
		t.Errorf("expected 0 private, got %d", len(private))
	}
}
