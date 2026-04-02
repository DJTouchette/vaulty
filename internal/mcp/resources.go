package mcp

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/djtouchette/vaulty/internal/audit"
	"github.com/djtouchette/vaulty/internal/policy"
	"github.com/djtouchette/vaulty/internal/vault"
	"github.com/pelletier/go-toml/v2"
)

// ResourceInfo describes an MCP resource.
type ResourceInfo struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MimeType    string `json:"mimeType"`
}

// ResourceHandler handles MCP resource requests.
type ResourceHandler struct {
	vault  *vault.Vault
	config *policy.Config
	logger *audit.Logger
}

// NewResourceHandler creates a new resource handler.
func NewResourceHandler(v *vault.Vault, cfg *policy.Config, logger *audit.Logger) *ResourceHandler {
	return &ResourceHandler{
		vault:  v,
		config: cfg,
		logger: logger,
	}
}

// ListResources returns the available MCP resources.
func (r *ResourceHandler) ListResources() []ResourceInfo {
	return []ResourceInfo{
		{
			URI:         "vaulty://secrets",
			Name:        "Secret Names",
			Description: "List of stored secret names (no values)",
			MimeType:    "text/plain",
		},
		{
			URI:         "vaulty://policy",
			Name:        "Policy Config",
			Description: "Active policy configuration (TOML)",
			MimeType:    "application/toml",
		},
		{
			URI:         "vaulty://audit",
			Name:        "Audit Log",
			Description: "Last 50 lines of the audit log",
			MimeType:    "application/jsonl",
		},
	}
}

// ReadResource reads the content of a resource by URI.
func (r *ResourceHandler) ReadResource(uri string) (string, error) {
	switch uri {
	case "vaulty://secrets":
		return r.readSecrets()
	case "vaulty://policy":
		return r.readPolicy()
	case "vaulty://audit":
		return r.readAudit()
	default:
		return "", fmt.Errorf("unknown resource: %s", uri)
	}
}

func (r *ResourceHandler) readSecrets() (string, error) {
	names := r.vault.List()
	if len(names) == 0 {
		return "No secrets stored.", nil
	}
	return strings.Join(names, "\n"), nil
}

func (r *ResourceHandler) readPolicy() (string, error) {
	data, err := toml.Marshal(r.config)
	if err != nil {
		return "", fmt.Errorf("marshaling policy: %w", err)
	}
	return string(data), nil
}

func (r *ResourceHandler) readAudit() (string, error) {
	path := r.logger.Path()
	if path == "" {
		return "No audit log configured.", nil
	}

	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "Audit log is empty.", nil
		}
		return "", fmt.Errorf("opening audit log: %w", err)
	}
	defer f.Close()

	// Read all lines, keep last 50
	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("reading audit log: %w", err)
	}

	if len(lines) == 0 {
		return "Audit log is empty.", nil
	}

	if len(lines) > 50 {
		lines = lines[len(lines)-50:]
	}
	return strings.Join(lines, "\n"), nil
}
