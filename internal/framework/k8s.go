package framework

import (
	"encoding/base64"
	"fmt"
	"io"
	"sort"

	"gopkg.in/yaml.v3"
)

// k8sSecret is a minimal representation of a Kubernetes Secret manifest.
type k8sSecret struct {
	APIVersion string            `yaml:"apiVersion"`
	Kind       string            `yaml:"kind"`
	Metadata   k8sMetadata       `yaml:"metadata"`
	Type       string            `yaml:"type,omitempty"`
	Data       map[string]string `yaml:"data"`
}

type k8sMetadata struct {
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace,omitempty"`
}

// ParseK8sSecret parses a Kubernetes Secret YAML manifest and returns the
// decoded secret data. Values in .data are base64-decoded.
func ParseK8sSecret(data []byte) (map[string]string, error) {
	var secret k8sSecret
	if err := yaml.Unmarshal(data, &secret); err != nil {
		return nil, fmt.Errorf("parsing K8s Secret manifest: %w", err)
	}

	if secret.Kind != "Secret" {
		return nil, fmt.Errorf("expected Kind: Secret, got %q", secret.Kind)
	}

	result := make(map[string]string, len(secret.Data))
	for key, encoded := range secret.Data {
		decoded, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			return nil, fmt.Errorf("decoding base64 for key %q: %w", key, err)
		}
		result[key] = string(decoded)
	}

	return result, nil
}

// WriteK8sSecret generates a Kubernetes Secret YAML manifest with base64-encoded values.
func WriteK8sSecret(w io.Writer, name, namespace string, secrets map[string]string) error {
	keys := make([]string, 0, len(secrets))
	for k := range secrets {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build output manually for deterministic key ordering
	if _, err := fmt.Fprintln(w, "apiVersion: v1"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "kind: Secret"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "metadata:"); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  name: %s\n", name); err != nil {
		return err
	}
	if namespace != "" {
		if _, err := fmt.Fprintf(w, "  namespace: %s\n", namespace); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(w, "type: Opaque"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "data:"); err != nil {
		return err
	}
	for _, k := range keys {
		encoded := base64.StdEncoding.EncodeToString([]byte(secrets[k]))
		if _, err := fmt.Fprintf(w, "  %s: %s\n", k, encoded); err != nil {
			return err
		}
	}
	return nil
}
