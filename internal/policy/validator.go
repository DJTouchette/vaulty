package policy

import (
	"fmt"
	"net/url"
	"strings"
)

// ValidateDomain checks if the target URL's host is in the secret's allowed domains.
// Returns nil if allowed, error if denied. An empty allowlist permits all domains.
func (c *Config) ValidateDomain(secretName, targetURL string) error {
	sp := c.GetSecretPolicy(secretName)

	if len(sp.AllowedDomains) == 0 {
		return nil // wildcard: no restrictions
	}

	parsed, err := url.Parse(targetURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	host := parsed.Hostname()
	for _, allowed := range sp.AllowedDomains {
		if strings.EqualFold(host, allowed) {
			return nil
		}
	}

	return fmt.Errorf("domain %q not in allowlist for secret %s (allowed: %s)",
		host, secretName, strings.Join(sp.AllowedDomains, ", "))
}

// ValidateCommand checks if the command contains one of the allowed command names.
// Returns nil if allowed, error if denied. An empty allowlist permits all commands.
func (c *Config) ValidateCommand(secretName, command string) error {
	sp := c.GetSecretPolicy(secretName)

	if len(sp.AllowedCommands) == 0 {
		return nil // wildcard: no restrictions
	}

	for _, allowed := range sp.AllowedCommands {
		if strings.Contains(command, allowed) {
			return nil
		}
	}

	return fmt.Errorf("command not in allowlist for secret %s (allowed: %s)",
		secretName, strings.Join(sp.AllowedCommands, ", "))
}
