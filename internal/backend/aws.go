package backend

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// AWSBackend retrieves secrets from AWS Secrets Manager via the aws CLI.
type AWSBackend struct {
	region   string
	profile  string
	endpoint string
}

// NewAWSBackend creates an AWSBackend. Returns an error if the aws CLI is not installed.
func NewAWSBackend(region, profile string) (*AWSBackend, error) {
	if _, err := exec.LookPath("aws"); err != nil {
		return nil, fmt.Errorf("aws CLI not found — install it from https://aws.amazon.com/cli/")
	}
	return &AWSBackend{region: region, profile: profile}, nil
}

// NewAWSBackendWithEndpoint creates an AWSBackend with a custom endpoint URL (e.g., for LocalStack).
func NewAWSBackendWithEndpoint(region, profile, endpoint string) (*AWSBackend, error) {
	b, err := NewAWSBackend(region, profile)
	if err != nil {
		return nil, err
	}
	b.endpoint = endpoint
	return b, nil
}

func (a *AWSBackend) Name() string {
	return "aws-secrets-manager"
}

// List returns secret names from AWS Secrets Manager.
func (a *AWSBackend) List() ([]string, error) {
	args := []string{"secretsmanager", "list-secrets",
		"--query", "SecretList[].Name", "--output", "json"}
	args = a.appendFlags(args)

	out, err := exec.Command("aws", args...).Output()
	if err != nil {
		return nil, fmt.Errorf("aws secretsmanager list-secrets: %w", err)
	}

	var names []string
	if err := json.Unmarshal(out, &names); err != nil {
		return nil, fmt.Errorf("parsing aws list-secrets output: %w", err)
	}
	return names, nil
}

// Get retrieves a secret value by name from AWS Secrets Manager.
func (a *AWSBackend) Get(name string) (string, error) {
	args := []string{"secretsmanager", "get-secret-value",
		"--secret-id", name,
		"--query", "SecretString", "--output", "text"}
	args = a.appendFlags(args)

	out, err := exec.Command("aws", args...).Output()
	if err != nil {
		return "", fmt.Errorf("aws secretsmanager get-secret-value: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// CmdArgs returns the command arguments that would be used for Get.
// Exported for testing without running the CLI.
func (a *AWSBackend) CmdArgsGet(name string) []string {
	args := []string{"secretsmanager", "get-secret-value",
		"--secret-id", name,
		"--query", "SecretString", "--output", "text"}
	return a.appendFlags(args)
}

// CmdArgsList returns the command arguments that would be used for List.
// Exported for testing without running the CLI.
func (a *AWSBackend) CmdArgsList() []string {
	args := []string{"secretsmanager", "list-secrets",
		"--query", "SecretList[].Name", "--output", "json"}
	return a.appendFlags(args)
}

func (a *AWSBackend) appendFlags(args []string) []string {
	if a.endpoint != "" {
		args = append(args, "--endpoint-url", a.endpoint)
	}
	if a.region != "" {
		args = append(args, "--region", a.region)
	}
	if a.profile != "" {
		args = append(args, "--profile", a.profile)
	}
	return args
}
