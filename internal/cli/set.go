package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/djtouchette/vaulty/internal/policy"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func newSetCmd() *cobra.Command {
	var (
		value       string
		domains     string
		commands    string
		description string
	)

	cmd := &cobra.Command{
		Use:   "set <name>",
		Short: "Add or update a secret",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			cfg, err := policy.LoadOrDefault("")
			if err != nil {
				return err
			}

			// Get secret value
			secretValue := value
			if secretValue == "" {
				stat, _ := os.Stdin.Stat()
				if (stat.Mode() & os.ModeCharDevice) == 0 {
					// Piped input
					scanner := bufio.NewScanner(os.Stdin)
					if scanner.Scan() {
						secretValue = scanner.Text()
					}
				} else {
					// Interactive
					fmt.Print("Enter value: ")
					raw, err := term.ReadPassword(0)
					fmt.Println()
					if err != nil {
						return fmt.Errorf("reading value: %w", err)
					}
					secretValue = string(raw)
				}
			}

			if secretValue == "" {
				return fmt.Errorf("secret value cannot be empty")
			}

			// Open vault, set secret, save
			vaultPath := resolveVaultPath(cfg.Vault.Path)
			h, err := openVault(vaultPath)
			if err != nil {
				return err
			}
			defer h.Vault.Zero()

			h.Vault.Set(name, secretValue)

			if err := h.Vault.Save(vaultPath, h.Passphrase); err != nil {
				return err
			}

			// Update policy in config
			sp := policy.SecretPolicy{}
			if description != "" {
				sp.Description = description
			}
			if domains != "" {
				sp.AllowedDomains = splitTrim(domains)
			}
			if commands != "" {
				sp.AllowedCommands = splitTrim(commands)
			}
			cfg.SetSecretPolicy(name, sp)
			if err := cfg.Write(); err != nil {
				return fmt.Errorf("writing config: %w", err)
			}

			fmt.Printf("Secret %s stored.\n", name)
			return nil
		},
	}

	cmd.Flags().StringVar(&value, "value", "", "secret value (or pipe via stdin)")
	cmd.Flags().StringVar(&description, "description", "", "description of what this secret is for (shown to agents)")
	cmd.Flags().StringVar(&domains, "domains", "", "allowed domains (comma-separated)")
	cmd.Flags().StringVar(&commands, "commands", "", "allowed commands (comma-separated)")
	return cmd
}

func splitTrim(s string) []string {
	parts := strings.Split(s, ",")
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
