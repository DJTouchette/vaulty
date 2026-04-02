package framework

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteComposeOverride(t *testing.T) {
	secrets := map[string]string{
		"DB_URL":  "postgres://localhost/mydb",
		"API_KEY": "sk-123",
	}

	var buf bytes.Buffer
	if err := WriteComposeOverride(&buf, secrets, "web"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()

	if !strings.Contains(out, "services:") {
		t.Error("output should contain 'services:'")
	}
	if !strings.Contains(out, "  web:") {
		t.Error("output should contain '  web:'")
	}
	if !strings.Contains(out, "    environment:") {
		t.Error("output should contain '    environment:'")
	}
	if !strings.Contains(out, "API_KEY:") {
		t.Error("output should contain API_KEY")
	}
	if !strings.Contains(out, "DB_URL:") {
		t.Error("output should contain DB_URL")
	}

	// API_KEY should come before DB_URL (alphabetical)
	apiIdx := strings.Index(out, "API_KEY")
	dbIdx := strings.Index(out, "DB_URL")
	if apiIdx > dbIdx {
		t.Error("keys should be sorted alphabetically")
	}
}

func TestWriteSecretFiles(t *testing.T) {
	dir := t.TempDir()
	secrets := map[string]string{
		"db_password": "s3cret",
		"api_key":     "key-123",
	}

	if err := WriteSecretFiles(dir, secrets); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for name, want := range secrets {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Errorf("reading %s: %v", name, err)
			continue
		}
		if string(data) != want {
			t.Errorf("%s = %q, want %q", name, string(data), want)
		}
	}
}

func TestParseComposeEnvMapping(t *testing.T) {
	yaml := `services:
  web:
    environment:
      API_KEY: abc123
      DB_URL: postgres://localhost/mydb
  worker:
    environment:
      REDIS_URL: redis://localhost:6379
`
	got, err := ParseComposeEnv([]byte(yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := map[string]string{
		"API_KEY":   "abc123",
		"DB_URL":    "postgres://localhost/mydb",
		"REDIS_URL": "redis://localhost:6379",
	}

	for key, want := range tests {
		if got[key] != want {
			t.Errorf("%s = %q, want %q", key, got[key], want)
		}
	}
}

func TestParseComposeEnvList(t *testing.T) {
	yaml := `services:
  web:
    environment:
      - API_KEY=abc123
      - DB_URL=postgres://localhost/mydb
`
	got, err := ParseComposeEnv([]byte(yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got["API_KEY"] != "abc123" {
		t.Errorf("API_KEY = %q, want %q", got["API_KEY"], "abc123")
	}
	if got["DB_URL"] != "postgres://localhost/mydb" {
		t.Errorf("DB_URL = %q, want %q", got["DB_URL"], "postgres://localhost/mydb")
	}
}

func TestParseComposeEnvEmpty(t *testing.T) {
	yaml := `services:
  web:
    image: nginx
`
	got, err := ParseComposeEnv([]byte(yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 0 {
		t.Errorf("expected 0 keys, got %d", len(got))
	}
}
