package backend

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// GCPBackend retrieves secrets from GCP Secret Manager via the gcloud CLI.
type GCPBackend struct {
	project string
}

// NewGCPBackend creates a GCPBackend. Returns an error if the gcloud CLI is not installed.
func NewGCPBackend(project string) (*GCPBackend, error) {
	if _, err := exec.LookPath("gcloud"); err != nil {
		return nil, fmt.Errorf("gcloud CLI not found — install it from https://cloud.google.com/sdk/docs/install")
	}
	return &GCPBackend{project: project}, nil
}

func (g *GCPBackend) Name() string {
	return "gcp-secret-manager"
}

// List returns secret names from GCP Secret Manager.
func (g *GCPBackend) List() ([]string, error) {
	args := g.CmdArgsList()

	out, err := exec.Command("gcloud", args...).Output()
	if err != nil {
		return nil, fmt.Errorf("gcloud secrets list: %w", err)
	}

	// gcloud returns JSON array of objects with "name" field containing the full resource path
	// e.g. "projects/my-project/secrets/my-secret"
	var entries []struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(out, &entries); err != nil {
		return nil, fmt.Errorf("parsing gcloud secrets list output: %w", err)
	}

	names := make([]string, len(entries))
	for i, e := range entries {
		// Extract just the secret name from the full resource path
		parts := strings.Split(e.Name, "/")
		names[i] = parts[len(parts)-1]
	}
	return names, nil
}

// Get retrieves the latest version of a secret from GCP Secret Manager.
func (g *GCPBackend) Get(name string) (string, error) {
	args := g.CmdArgsGet(name)

	out, err := exec.Command("gcloud", args...).Output()
	if err != nil {
		return "", fmt.Errorf("gcloud secrets versions access: %w", err)
	}
	return string(out), nil
}

// CmdArgsList returns the command arguments that would be used for List.
// Exported for testing without running the CLI.
func (g *GCPBackend) CmdArgsList() []string {
	args := []string{"secrets", "list", "--format", "json"}
	if g.project != "" {
		args = append(args, "--project", g.project)
	}
	return args
}

// CmdArgsGet returns the command arguments that would be used for Get.
// Exported for testing without running the CLI.
func (g *GCPBackend) CmdArgsGet(name string) []string {
	args := []string{"secrets", "versions", "access", "latest", "--secret=" + name}
	if g.project != "" {
		args = append(args, "--project=" + g.project)
	}
	return args
}
