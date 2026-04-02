package policy

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

// BackendConfig defines an external secret backend.
type BackendConfig struct {
	Type     string `toml:"type" yaml:"type"`
	Region   string `toml:"region,omitempty" yaml:"region,omitempty"`
	Profile  string `toml:"profile,omitempty" yaml:"profile,omitempty"`
	Endpoint string `toml:"endpoint,omitempty" yaml:"endpoint,omitempty"`
	Project  string `toml:"project,omitempty" yaml:"project,omitempty"`
	Addr     string `toml:"addr,omitempty" yaml:"addr,omitempty"`
	Mount    string `toml:"mount,omitempty" yaml:"mount,omitempty"`
	OpVault  string `toml:"op_vault,omitempty" yaml:"op_vault,omitempty"`
	TTL      string `toml:"ttl,omitempty" yaml:"ttl,omitempty"`
}

// Config represents the parsed vaulty.toml or vaulty.yaml file.
type Config struct {
	Vault    VaultConfig                `toml:"vault" yaml:"vault"`
	Backends map[string]BackendConfig   `toml:"backends,omitempty" yaml:"backends,omitempty"`
	Secrets  map[string]SecretPolicy    `toml:"secrets" yaml:"secrets"`

	// path tracks where this config was loaded from / should be written to
	path string
}

// VaultConfig holds vault-level settings.
type VaultConfig struct {
	Path          string `toml:"path" yaml:"path"`
	IdleTimeout   string `toml:"idle_timeout" yaml:"idle_timeout"`
	Socket        string `toml:"socket" yaml:"socket"`
	HTTPPort      int    `toml:"http_port" yaml:"http_port"`
	Notifications bool   `toml:"notifications" yaml:"notifications"`
}

// SecretPolicy defines access constraints for a secret.
type SecretPolicy struct {
	Description     string   `toml:"description,omitempty" yaml:"description,omitempty"`
	AllowedDomains  []string `toml:"allowed_domains,omitempty" yaml:"allowed_domains,omitempty"`
	AllowedCommands []string `toml:"allowed_commands,omitempty" yaml:"allowed_commands,omitempty"`
	InjectAs        string   `toml:"inject_as,omitempty" yaml:"inject_as,omitempty"`
	HeaderName      string   `toml:"header_name,omitempty" yaml:"header_name,omitempty"`
	AlsoInject      []string `toml:"also_inject,omitempty" yaml:"also_inject,omitempty"`
	Vault           string   `toml:"vault,omitempty" yaml:"vault,omitempty"`
	AutoApprove     bool     `toml:"auto_approve,omitempty" yaml:"auto_approve,omitempty"`
	Backend         string   `toml:"backend,omitempty" yaml:"backend,omitempty"`
}

// NewLocalConfig creates a config for a project-local .vaulty/ directory.
func NewLocalConfig(configPath, vaultPath string) *Config {
	cfg := defaultConfig()
	cfg.Vault.Path = vaultPath
	cfg.path = configPath
	return cfg
}

func defaultConfig() *Config {
	return &Config{
		Vault: VaultConfig{
			Path:        "~/.config/vaulty/vault.age",
			IdleTimeout: "8h",
			Socket:      "/tmp/vaulty.sock",
			HTTPPort:    19876,
		},
		Secrets: map[string]SecretPolicy{},
	}
}

// LoadOrDefault loads config from the given path, or searches common locations.
// Supports both TOML (.toml) and YAML (.yaml, .yml) formats.
// If no config file is found, returns a default config.
func LoadOrDefault(path string) (*Config, error) {
	if path != "" {
		return loadFrom(path)
	}

	// Search order: for each directory, try .toml then .yaml then .yml
	dirs := []string{
		".",
		".vaulty",
	}
	if home, err := os.UserHomeDir(); err == nil {
		dirs = append(dirs, filepath.Join(home, ".config", "vaulty"))
	}

	for _, dir := range dirs {
		for _, name := range []string{"vaulty.toml", "vaulty.yaml", "vaulty.yml"} {
			p := filepath.Join(dir, name)
			if dir == "." {
				p = name
			}
			if _, err := os.Stat(p); err == nil {
				return loadFrom(p)
			}
		}
	}

	cfg := defaultConfig()
	// Default to writing in ~/.config/vaulty/
	if home, err := os.UserHomeDir(); err == nil {
		cfg.path = filepath.Join(home, ".config", "vaulty", "vaulty.toml")
	} else {
		cfg.path = "vaulty.toml"
	}
	return cfg, nil
}

func isYAMLPath(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".yaml" || ext == ".yml"
}

func loadFrom(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	cfg := defaultConfig()
	if isYAMLPath(path) {
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("parsing YAML config: %w", err)
		}
	} else {
		if err := toml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("parsing TOML config: %w", err)
		}
	}
	cfg.path = path
	return cfg, nil
}

// Write saves the config to its original path.
func (c *Config) Write() error {
	return c.writeTo(c.path)
}

// WriteDefault writes the config file only if it doesn't already exist.
func (c *Config) WriteDefault() error {
	if _, err := os.Stat(c.path); err == nil {
		return nil // already exists
	}
	return c.writeTo(c.path)
}

func (c *Config) writeTo(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	var data []byte
	var err error
	if isYAMLPath(path) {
		data, err = yaml.Marshal(c)
		if err != nil {
			return fmt.Errorf("marshaling YAML config: %w", err)
		}
	} else {
		data, err = toml.Marshal(c)
		if err != nil {
			return fmt.Errorf("marshaling TOML config: %w", err)
		}
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	return nil
}

// SetSecretPolicy adds or updates the policy for a secret.
func (c *Config) SetSecretPolicy(name string, p SecretPolicy) {
	if c.Secrets == nil {
		c.Secrets = map[string]SecretPolicy{}
	}
	c.Secrets[name] = p
}

// GetSecretPolicy returns the policy for a secret, or an empty policy if not found.
func (c *Config) GetSecretPolicy(name string) SecretPolicy {
	if c.Secrets == nil {
		return SecretPolicy{}
	}
	return c.Secrets[name]
}

// RemoveSecretPolicy removes the policy for a secret.
func (c *Config) RemoveSecretPolicy(name string) {
	delete(c.Secrets, name)
}

// Path returns the config file path.
func (c *Config) Path() string {
	return c.path
}
