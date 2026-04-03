package executor

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	"github.com/djtouchette/vaulty/internal/proxy"
)

// Result holds the result of an executed command.
type Result struct {
	ExitCode int
	Stdout   string
	Stderr   string
}

// Run executes a shell command with secrets injected as environment variables.
// Output is redacted to prevent secret leakage.
func Run(command string, secrets map[string]string, workDir string, redactor *proxy.Redactor) (*Result, error) {
	shell, args := shellCommand(command)
	cmd := exec.Command(shell, args...)

	// Build environment: inherit current env, then overlay secrets
	env := os.Environ()
	for name, value := range secrets {
		env = append(env, fmt.Sprintf("%s=%s", name, value))
	}
	cmd.Env = env

	if workDir != "" {
		cmd.Dir = workDir
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return nil, fmt.Errorf("executing command: %w", err)
		}
	}

	return &Result{
		ExitCode: exitCode,
		Stdout:   redactor.RedactString(stdout.String()),
		Stderr:   redactor.RedactString(stderr.String()),
	}, nil
}
