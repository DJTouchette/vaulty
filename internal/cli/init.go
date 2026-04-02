package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/djtouchette/vaulty/internal/policy"
	"github.com/djtouchette/vaulty/internal/vault"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// promptNewPassphrase gets a passphrase from VAULTY_PASSPHRASE env var or interactive prompt.
func promptNewPassphrase() (string, error) {
	if pass := os.Getenv("VAULTY_PASSPHRASE"); pass != "" {
		return pass, nil
	}
	fmt.Print("Set your passphrase: ")
	raw, err := term.ReadPassword(0)
	fmt.Println()
	if err != nil {
		return "", fmt.Errorf("reading passphrase: %w", err)
	}
	if len(raw) == 0 {
		return "", fmt.Errorf("passphrase cannot be empty")
	}
	return string(raw), nil
}

func newInitCmd() *cobra.Command {
	var force bool
	var local bool

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create a new vault",
		RunE: func(cmd *cobra.Command, args []string) error {
			if local {
				return initLocal(force)
			}

			cfg, err := policy.LoadOrDefault("")
			if err != nil {
				return err
			}

			path := cfg.Vault.Path

			if vault.Exists(path) && !force {
				return fmt.Errorf("vault already exists at %s (use --force to overwrite)", path)
			}

			pass, err := promptNewPassphrase()
			if err != nil {
				return err
			}

			if err := vault.Create(path, pass); err != nil {
				return err
			}

			if err := cfg.WriteDefault(); err != nil {
				return fmt.Errorf("writing default config: %w", err)
			}

			fmt.Printf("Vault initialized at %s\n", path)
			offerKeychainSave(path, pass)
			fmt.Println("Run `vaulty set <name>` to add secrets.")
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "overwrite existing vault")
	cmd.Flags().BoolVar(&local, "local", false, "create a project-local .vaulty/ directory")
	return cmd
}

func initLocal(force bool) error {
	vaultPath := filepath.Join(".vaulty", "vault.age")
	configPath := filepath.Join(".vaulty", "vaulty.toml")

	if vault.Exists(vaultPath) && !force {
		return fmt.Errorf("vault already exists at %s (use --force to overwrite)", vaultPath)
	}

	pass, err := promptNewPassphrase()
	if err != nil {
		return err
	}

	if err := vault.Create(vaultPath, pass); err != nil {
		return err
	}

	cfg := policy.NewLocalConfig(configPath, vaultPath)
	if err := cfg.WriteDefault(); err != nil {
		return fmt.Errorf("writing local config: %w", err)
	}

	fmt.Printf("Vault initialized at %s\n", vaultPath)
	fmt.Printf("Config written to %s\n", configPath)
	fmt.Println("Reminder: add .vaulty/vault.age to your .gitignore")
	offerKeychainSave(vaultPath, pass)
	fmt.Println("Run `vaulty set <name>` to add secrets.")
	return nil
}

// offerKeychainSave prompts the user to save the passphrase to the OS keychain.
func offerKeychainSave(vaultPath, passphrase string) {
	fmt.Print("Save passphrase to OS keychain? (y/N): ")
	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	if strings.TrimSpace(strings.ToLower(answer)) != "y" {
		return
	}

	service := vault.DefaultService()
	account := vault.KeyringAccount(vaultPath)
	if err := vault.SavePassphrase(service, account, passphrase); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not save to keychain: %v\n", err)
		return
	}
	fmt.Println("Passphrase saved to OS keychain.")
}
