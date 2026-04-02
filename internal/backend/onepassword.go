package backend

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// OnePasswordBackend retrieves secrets from 1Password via the op CLI.
type OnePasswordBackend struct {
	vault string // 1Password vault name
}

// NewOnePasswordBackend creates a OnePasswordBackend. Returns an error if the op CLI is not installed.
func NewOnePasswordBackend(opVault string) (*OnePasswordBackend, error) {
	if _, err := exec.LookPath("op"); err != nil {
		return nil, fmt.Errorf("op CLI not found — install it from https://developer.1password.com/docs/cli/get-started/")
	}
	return &OnePasswordBackend{vault: opVault}, nil
}

func (o *OnePasswordBackend) Name() string {
	return "1password"
}

// List returns item titles from the 1Password vault.
func (o *OnePasswordBackend) List() ([]string, error) {
	args := o.CmdArgsList()

	out, err := exec.Command("op", args...).Output()
	if err != nil {
		return nil, fmt.Errorf("op item list: %w", err)
	}

	var items []struct {
		Title string `json:"title"`
	}
	if err := json.Unmarshal(out, &items); err != nil {
		return nil, fmt.Errorf("parsing op item list output: %w", err)
	}

	names := make([]string, len(items))
	for i, item := range items {
		names[i] = item.Title
	}
	return names, nil
}

// Get retrieves a secret from 1Password using the op:// URI scheme.
func (o *OnePasswordBackend) Get(name string) (string, error) {
	args := o.CmdArgsGet(name)

	out, err := exec.Command("op", args...).Output()
	if err != nil {
		// Fallback: try op item get with --fields
		argsFallback := o.CmdArgsGetFallback(name)
		out, err = exec.Command("op", argsFallback...).Output()
		if err != nil {
			return "", fmt.Errorf("op read: %w", err)
		}

		var field struct {
			Value string `json:"value"`
		}
		if err := json.Unmarshal(out, &field); err != nil {
			return "", fmt.Errorf("parsing op item get output: %w", err)
		}
		return field.Value, nil
	}
	return strings.TrimSpace(string(out)), nil
}

// CmdArgsList returns the command arguments that would be used for List.
// Exported for testing without running the CLI.
func (o *OnePasswordBackend) CmdArgsList() []string {
	args := []string{"item", "list", "--format", "json"}
	if o.vault != "" {
		args = append(args, "--vault", o.vault)
	}
	return args
}

// CmdArgsGet returns the command arguments that would be used for Get (primary).
// Exported for testing without running the CLI.
func (o *OnePasswordBackend) CmdArgsGet(name string) []string {
	return []string{"read", fmt.Sprintf("op://%s/%s/password", o.vault, name)}
}

// CmdArgsGetFallback returns the fallback command arguments for Get.
// Exported for testing without running the CLI.
func (o *OnePasswordBackend) CmdArgsGetFallback(name string) []string {
	args := []string{"item", "get", name, "--fields", "password", "--format", "json"}
	if o.vault != "" {
		args = append(args, "--vault", o.vault)
	}
	return args
}
