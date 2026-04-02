package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/djtouchette/vaulty/internal/policy"
	"github.com/spf13/cobra"
)

func newRemoveCmd() *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove a secret from the vault",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			if !yes {
				fmt.Printf("Remove %s? (y/N): ", name)
				reader := bufio.NewReader(os.Stdin)
				answer, _ := reader.ReadString('\n')
				if strings.TrimSpace(strings.ToLower(answer)) != "y" {
					fmt.Println("Aborted.")
					return nil
				}
			}

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

			h.Vault.Remove(name)

			if err := h.Vault.Save(vaultPath, h.Passphrase); err != nil {
				return err
			}

			cfg.RemoveSecretPolicy(name)
			if err := cfg.Write(); err != nil {
				return fmt.Errorf("writing config: %w", err)
			}

			fmt.Println("Removed.")
			return nil
		},
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "skip confirmation")
	return cmd
}
