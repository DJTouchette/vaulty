package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Entry represents a single audit log entry.
type Entry struct {
	Timestamp string `json:"ts"`
	Action    string `json:"action"`            // proxy, exec, denied
	Secret    string `json:"secret"`
	Target    string `json:"target,omitempty"`   // URL for proxy
	Method    string `json:"method,omitempty"`   // HTTP method
	Status    int    `json:"status,omitempty"`   // HTTP status code
	Command   string `json:"command,omitempty"`  // for exec
	ExitCode  *int   `json:"exit_code,omitempty"`
	Reason    string `json:"reason,omitempty"`   // for denied
}

// Logger writes append-only audit log entries.
type Logger struct {
	mu   sync.Mutex
	path string
	file *os.File
}

// NewLogger creates a new audit logger writing to the given path.
func NewLogger(path string) (*Logger, error) {
	path = expandPath(path)

	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, fmt.Errorf("creating audit log directory: %w", err)
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return nil, fmt.Errorf("opening audit log: %w", err)
	}

	return &Logger{path: path, file: f}, nil
}

// Log writes an audit entry.
func (l *Logger) Log(entry Entry) error {
	entry.Timestamp = time.Now().UTC().Format(time.RFC3339)

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshaling audit entry: %w", err)
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if _, err := l.file.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("writing audit entry: %w", err)
	}
	return nil
}

// LogProxy logs a proxy request.
func (l *Logger) LogProxy(secret, method, target string, status int) error {
	return l.Log(Entry{
		Action: "proxy",
		Secret: secret,
		Method: method,
		Target: target,
		Status: status,
	})
}

// LogExec logs a command execution.
func (l *Logger) LogExec(secret, command string, exitCode int) error {
	return l.Log(Entry{
		Action:   "exec",
		Secret:   secret,
		Command:  command,
		ExitCode: &exitCode,
	})
}

// LogDenied logs a denied request.
func (l *Logger) LogDenied(secret, target, reason string) error {
	return l.Log(Entry{
		Action: "denied",
		Secret: secret,
		Target: target,
		Reason: reason,
	})
}

// LogApproval logs an approval decision.
func (l *Logger) LogApproval(secret, target, decision string) error {
	return l.Log(Entry{
		Action: "approval",
		Secret: secret,
		Target: target,
		Reason: decision,
	})
}

// Path returns the audit log file path.
func (l *Logger) Path() string {
	return l.path
}

// Close closes the log file.
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.file.Close()
}

func expandPath(path string) string {
	if len(path) > 1 && path[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}
