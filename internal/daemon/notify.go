package daemon

import (
	"fmt"
	"os/exec"
)

// Notifier sends desktop notifications for denied secret access requests.
type Notifier struct {
	enabled bool
}

// NewNotifier creates a new Notifier. If enabled is false, all notification
// calls are no-ops.
func NewNotifier(enabled bool) *Notifier {
	return &Notifier{enabled: enabled}
}

// NotifyDenied sends a critical desktop notification about a denied request.
// It silently does nothing if notifications are disabled or notify-send is
// not installed.
func (n *Notifier) NotifyDenied(secretName, target, reason string) {
	if !n.enabled {
		return
	}

	summary := "Vaulty: access denied"
	body := fmt.Sprintf("Secret: %s\nTarget: %s\nReason: %s", secretName, target, reason)

	cmd := exec.Command("notify-send", "--urgency=critical", summary, body)
	// Silently ignore errors (e.g. notify-send not found).
	_ = cmd.Run()
}

// FormatBody returns the notification body text for a denied request.
// Exported for testing.
func FormatBody(secretName, target, reason string) string {
	return fmt.Sprintf("Secret: %s\nTarget: %s\nReason: %s", secretName, target, reason)
}
