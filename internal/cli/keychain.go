package cli

import (
	"fmt"

	"github.com/djtouchette/vaulty/internal/policy"
	"github.com/djtouchette/vaulty/internal/vault"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func newKeychainCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "keychain",
		Short: "Manage vault passphrase in the OS keychain",
	}

	cmd.AddCommand(
		newKeychainSaveCmd(),
		newKeychainDeleteCmd(),
		newKeychainStatusCmd(),
	)

	return cmd
}

func newKeychainSaveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "save",
		Short: "Save vault passphrase to the OS keychain",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := policy.LoadOrDefault("")
			if err != nil {
				return err
			}

			fmt.Print("Passphrase: ")
			pass, err := term.ReadPassword(0)
			fmt.Println()
			if err != nil {
				return fmt.Errorf("reading passphrase: %w", err)
			}

			// Verify the passphrase is correct by opening the vault
			v, err := vault.Open(cfg.Vault.Path, string(pass))
			if err != nil {
				return err
			}
			v.Zero()

			service := vault.DefaultService()
			account := vault.KeyringAccount(cfg.Vault.Path)

			if err := vault.SavePassphrase(service, account, string(pass)); err != nil {
				return fmt.Errorf("saving to keychain: %w", err)
			}

			fmt.Println("Passphrase saved to OS keychain.")
			return nil
		},
	}
}

func newKeychainDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete",
		Short: "Remove vault passphrase from the OS keychain",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := policy.LoadOrDefault("")
			if err != nil {
				return err
			}

			service := vault.DefaultService()
			account := vault.KeyringAccount(cfg.Vault.Path)

			if err := vault.DeletePassphrase(service, account); err != nil {
				return fmt.Errorf("removing from keychain: %w", err)
			}

			fmt.Println("Passphrase removed from OS keychain.")
			return nil
		},
	}
}

func newKeychainStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Check if vault passphrase is stored in the OS keychain",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := policy.LoadOrDefault("")
			if err != nil {
				return err
			}

			service := vault.DefaultService()
			account := vault.KeyringAccount(cfg.Vault.Path)

			if vault.HasPassphrase(service, account) {
				fmt.Println("Passphrase is stored in the OS keychain.")
			} else {
				fmt.Println("No passphrase found in the OS keychain.")
			}
			return nil
		},
	}
}
