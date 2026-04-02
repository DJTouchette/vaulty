package backend

import (
	"fmt"
	"time"
)

// BackendConfig holds the configuration for a backend provider.
type BackendConfig struct {
	Type     string `toml:"type"`               // "aws-secrets-manager", "gcp-secret-manager", "hashicorp-vault", "1password"
	Region   string `toml:"region,omitempty"`   // AWS
	Profile  string `toml:"profile,omitempty"`  // AWS
	Endpoint string `toml:"endpoint,omitempty"` // AWS endpoint URL override (e.g., for LocalStack)
	Project  string `toml:"project,omitempty"`  // GCP
	Addr     string `toml:"addr,omitempty"`     // HashiCorp Vault address
	Mount    string `toml:"mount,omitempty"`    // HashiCorp Vault mount
	OpVault  string `toml:"op_vault,omitempty"` // 1Password vault
	TTL      string `toml:"ttl,omitempty"`      // Cache TTL (default "5m")
}

// NewBackend creates a SecretBackend from the given config, wrapped in a CachedBackend.
func NewBackend(cfg BackendConfig) (SecretBackend, error) {
	var inner SecretBackend
	var err error

	switch cfg.Type {
	case "aws-secrets-manager":
		if cfg.Endpoint != "" {
			inner, err = NewAWSBackendWithEndpoint(cfg.Region, cfg.Profile, cfg.Endpoint)
		} else {
			inner, err = NewAWSBackend(cfg.Region, cfg.Profile)
		}
	case "gcp-secret-manager":
		inner, err = NewGCPBackend(cfg.Project)
	case "hashicorp-vault":
		inner, err = NewHashiCorpBackend(cfg.Addr, cfg.Mount)
	case "1password":
		inner, err = NewOnePasswordBackend(cfg.OpVault)
	default:
		return nil, fmt.Errorf("unknown backend type %q — supported: aws-secrets-manager, gcp-secret-manager, hashicorp-vault, 1password", cfg.Type)
	}
	if err != nil {
		return nil, err
	}

	ttl := 5 * time.Minute
	if cfg.TTL != "" {
		ttl, err = time.ParseDuration(cfg.TTL)
		if err != nil {
			return nil, fmt.Errorf("invalid TTL %q: %w", cfg.TTL, err)
		}
	}

	return NewCachedBackend(inner, ttl), nil
}
