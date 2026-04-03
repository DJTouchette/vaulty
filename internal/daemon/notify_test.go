package daemon

import (
	"strings"
	"testing"
)

func TestNotifierDisabledNoop(t *testing.T) {
	n := NewNotifier(false)
	// Should not panic or error when disabled.
	n.NotifyDenied("API_KEY", "https://evil.com", "domain not in allowlist")
}

func TestNotifierEnabledDoesNotPanic(t *testing.T) {
	// Even when enabled, if notify-send is not installed this must not panic.
	n := NewNotifier(true)
	n.NotifyDenied("DB_URL", "curl http://evil.com", "command not in allowlist")
}

func TestFormatBody(t *testing.T) {
	body := FormatBody("API_KEY", "https://evil.com/steal", "domain not in allowlist")

	if !strings.Contains(body, "Secret: API_KEY") {
		t.Errorf("body missing secret name: %q", body)
	}
	if !strings.Contains(body, "Target: https://evil.com/steal") {
		t.Errorf("body missing target: %q", body)
	}
	if !strings.Contains(body, "Reason: domain not in allowlist") {
		t.Errorf("body missing reason: %q", body)
	}
}

func TestFormatBodyStructure(t *testing.T) {
	body := FormatBody("SECRET", "target", "reason")
	lines := strings.Split(body, "\n")
	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d: %q", len(lines), body)
	}
}
