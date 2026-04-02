package backend

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// HashiCorpBackend retrieves secrets from HashiCorp Vault via the vault CLI.
type HashiCorpBackend struct {
	addr  string
	mount string
}

// NewHashiCorpBackend creates a HashiCorpBackend. mount defaults to "secret" if empty.
// Returns an error if the vault CLI is not installed.
func NewHashiCorpBackend(addr, mount string) (*HashiCorpBackend, error) {
	if _, err := exec.LookPath("vault"); err != nil {
		return nil, fmt.Errorf("vault CLI not found — install it from https://developer.hashicorp.com/vault/install")
	}
	if mount == "" {
		mount = "secret"
	}
	return &HashiCorpBackend{addr: addr, mount: mount}, nil
}

func (h *HashiCorpBackend) Name() string {
	return "hashicorp-vault"
}

// List returns secret names from HashiCorp Vault.
func (h *HashiCorpBackend) List() ([]string, error) {
	args := h.CmdArgsList()

	cmd := exec.Command("vault", args...)
	h.setEnv(cmd)

	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("vault kv list: %w", err)
	}

	var names []string
	if err := json.Unmarshal(out, &names); err != nil {
		return nil, fmt.Errorf("parsing vault kv list output: %w", err)
	}
	return names, nil
}

// Get retrieves a secret value from HashiCorp Vault.
// Tries to read a single "value" field first. If that fails, returns the full JSON of the data object.
func (h *HashiCorpBackend) Get(name string) (string, error) {
	// Try single-field first
	argsField := h.CmdArgsGetField(name)
	cmdField := exec.Command("vault", argsField...)
	h.setEnv(cmdField)

	out, err := cmdField.Output()
	if err == nil {
		return strings.TrimSpace(string(out)), nil
	}

	// Fall back to full JSON
	argsJSON := h.CmdArgsGetJSON(name)
	cmdJSON := exec.Command("vault", argsJSON...)
	h.setEnv(cmdJSON)

	out, err = cmdJSON.Output()
	if err != nil {
		return "", fmt.Errorf("vault kv get: %w", err)
	}

	// Parse the KV v2 response and extract .data.data
	var resp struct {
		Data struct {
			Data json.RawMessage `json:"data"`
		} `json:"data"`
	}
	if err := json.Unmarshal(out, &resp); err != nil {
		return "", fmt.Errorf("parsing vault kv get output: %w", err)
	}
	return string(resp.Data.Data), nil
}

// CmdArgsList returns the command arguments that would be used for List.
// Exported for testing without running the CLI.
func (h *HashiCorpBackend) CmdArgsList() []string {
	return []string{"kv", "list", "-format=json", h.mount}
}

// CmdArgsGetField returns the command arguments for Get with -field=value.
// Exported for testing without running the CLI.
func (h *HashiCorpBackend) CmdArgsGetField(name string) []string {
	return []string{"kv", "get", "-field=value", "-mount=" + h.mount, name}
}

// CmdArgsGetJSON returns the command arguments for Get with -format=json.
// Exported for testing without running the CLI.
func (h *HashiCorpBackend) CmdArgsGetJSON(name string) []string {
	return []string{"kv", "get", "-format=json", "-mount=" + h.mount, name}
}

func (h *HashiCorpBackend) setEnv(cmd *exec.Cmd) {
	cmd.Env = os.Environ()
	if h.addr != "" {
		cmd.Env = append(cmd.Env, "VAULT_ADDR="+h.addr)
	}
}
