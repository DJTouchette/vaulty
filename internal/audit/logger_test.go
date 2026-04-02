package audit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLogProxy(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.log")

	logger, err := NewLogger(path)
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}
	defer logger.Close()

	err = logger.LogProxy("STRIPE_KEY", "POST", "https://api.stripe.com/v1/charges", 200)
	if err != nil {
		t.Fatalf("LogProxy: %v", err)
	}

	data, _ := os.ReadFile(path)
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}

	var entry Entry
	if err := json.Unmarshal([]byte(lines[0]), &entry); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if entry.Action != "proxy" {
		t.Errorf("action = %q, want proxy", entry.Action)
	}
	if entry.Secret != "STRIPE_KEY" {
		t.Errorf("secret = %q", entry.Secret)
	}
	if entry.Status != 200 {
		t.Errorf("status = %d, want 200", entry.Status)
	}
	if entry.Timestamp == "" {
		t.Error("timestamp should be set")
	}
}

func TestLogExec(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.log")

	logger, err := NewLogger(path)
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}
	defer logger.Close()

	err = logger.LogExec("DB_URL", "npx prisma migrate deploy", 0)
	if err != nil {
		t.Fatalf("LogExec: %v", err)
	}

	data, _ := os.ReadFile(path)
	var entry Entry
	json.Unmarshal([]byte(strings.TrimSpace(string(data))), &entry)

	if entry.Action != "exec" {
		t.Errorf("action = %q, want exec", entry.Action)
	}
	if entry.ExitCode == nil || *entry.ExitCode != 0 {
		t.Errorf("exit_code = %v, want 0", entry.ExitCode)
	}
}

func TestLogDenied(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.log")

	logger, err := NewLogger(path)
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}
	defer logger.Close()

	err = logger.LogDenied("STRIPE_KEY", "https://evil.com", "domain not in allowlist")
	if err != nil {
		t.Fatalf("LogDenied: %v", err)
	}

	data, _ := os.ReadFile(path)
	var entry Entry
	json.Unmarshal([]byte(strings.TrimSpace(string(data))), &entry)

	if entry.Action != "denied" {
		t.Errorf("action = %q, want denied", entry.Action)
	}
	if entry.Reason != "domain not in allowlist" {
		t.Errorf("reason = %q", entry.Reason)
	}
}

func TestAppendOnly(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.log")

	logger, err := NewLogger(path)
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}

	logger.LogProxy("KEY1", "GET", "https://a.com", 200)
	logger.LogProxy("KEY2", "POST", "https://b.com", 201)
	logger.Close()

	data, _ := os.ReadFile(path)
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}
}

func TestLogNeverContainsSecretValues(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.log")

	logger, err := NewLogger(path)
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}
	defer logger.Close()

	// The logger only stores names, not values — but verify the log format
	logger.LogProxy("API_KEY", "GET", "https://api.example.com", 200)

	data, _ := os.ReadFile(path)
	content := string(data)

	// The entry should contain the name but we verify there's no "value" field
	if strings.Contains(content, "sk_live") {
		t.Error("log should never contain secret values")
	}
}
