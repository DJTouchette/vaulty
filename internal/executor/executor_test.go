package executor

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/djtouchette/vaulty/internal/proxy"
)

// echoEnvCmd returns a shell command that echoes an environment variable,
// using the correct syntax for the current platform.
func echoEnvCmd(varName string) string {
	if runtime.GOOS == "windows" {
		return "echo %" + varName + "%"
	}
	return "echo $" + varName
}

func TestRunSimpleCommand(t *testing.T) {
	redactor := proxy.NewRedactor(map[string]string{})

	result, err := Run("echo hello", nil, "", redactor)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("exit code = %d, want 0", result.ExitCode)
	}

	if strings.TrimSpace(result.Stdout) != "hello" {
		t.Errorf("stdout = %q, want hello", result.Stdout)
	}
}

func TestRunWithSecretInjection(t *testing.T) {
	secrets := map[string]string{
		"MY_SECRET": "super_secret_value",
	}
	redactor := proxy.NewRedactor(secrets)

	result, err := Run(echoEnvCmd("MY_SECRET"), secrets, "", redactor)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if strings.Contains(result.Stdout, "super_secret_value") {
		t.Error("stdout should not contain raw secret value")
	}

	if !strings.Contains(result.Stdout, "[VAULTY:MY_SECRET]") {
		t.Errorf("stdout should contain redacted placeholder, got %q", result.Stdout)
	}
}

func TestRunWithNonZeroExit(t *testing.T) {
	redactor := proxy.NewRedactor(map[string]string{})

	cmd := "exit 42"
	if runtime.GOOS == "windows" {
		cmd = "exit /b 42"
	}

	result, err := Run(cmd, nil, "", redactor)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if result.ExitCode != 42 {
		t.Errorf("exit code = %d, want 42", result.ExitCode)
	}
}

func TestRunStderrRedaction(t *testing.T) {
	secrets := map[string]string{
		"DB_PASS": "p@ssw0rd",
	}
	redactor := proxy.NewRedactor(secrets)

	cmd := "echo p@ssw0rd >&2"
	if runtime.GOOS == "windows" {
		cmd = "echo p@ssw0rd 1>&2"
	}

	result, err := Run(cmd, secrets, "", redactor)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if strings.Contains(result.Stderr, "p@ssw0rd") {
		t.Error("stderr should not contain raw secret value")
	}

	if !strings.Contains(result.Stderr, "[VAULTY:DB_PASS]") {
		t.Errorf("stderr should contain redacted placeholder, got %q", result.Stderr)
	}
}

func TestRunWithWorkDir(t *testing.T) {
	redactor := proxy.NewRedactor(map[string]string{})

	tmpDir := t.TempDir()

	cmd := "pwd"
	if runtime.GOOS == "windows" {
		cmd = "cd"
	}

	result, err := Run(cmd, nil, tmpDir, redactor)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Resolve symlinks for macOS where /var -> /private/var
	got, _ := filepath.EvalSymlinks(strings.TrimSpace(result.Stdout))
	want, _ := filepath.EvalSymlinks(tmpDir)

	if got != want {
		t.Errorf("stdout = %q, want %s", got, want)
	}
}
