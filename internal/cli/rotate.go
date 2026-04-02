package cli

import (
	"fmt"

	"github.com/djtouchette/vaulty/internal/policy"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func newRotateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rotate <name>",
		Short: "Rotate a secret's value",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			cfg, err := policy.LoadOrDefault("")
			if err != nil {
				return err
			}

			vaultPath := resolveVaultPath(cfg.Vault.Path)
			h, err := openVault(vaultPath)
			if err != nil {
				return err
			}
			defer h.Vault.Zero()

			if !h.Vault.Has(name) {
				return fmt.Errorf("secret %s not found", name)
			}

			// Prompt for new secret value
			fmt.Print("New value: ")
			raw, err := term.ReadPassword(0)
			fmt.Println()
			if err != nil {
				return fmt.Errorf("reading value: %w", err)
			}

			newValue := string(raw)
			if newValue == "" {
				return fmt.Errorf("secret value cannot be empty")
			}

			h.Vault.Set(name, newValue)

			if err := h.Vault.Save(vaultPath, h.Passphrase); err != nil {
				return err
			}

			fmt.Printf("Secret %s rotated.\n", name)
			return nil
		},
	}

	return cmd
}
