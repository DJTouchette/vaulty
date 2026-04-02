package cli

import (
	"fmt"
	"sort"
	"strings"

	"github.com/djtouchette/vaulty/internal/policy"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List stored secrets (names and policies, never values)",
		RunE: func(cmd *cobra.Command, args []string) error {
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

			names := h.Vault.List()
			if len(names) == 0 {
				fmt.Println("No secrets stored.")
				return nil
			}

			sort.Strings(names)

			for _, name := range names {
				sp := cfg.GetSecretPolicy(name)
				var info []string
				if sp.Description != "" {
					info = append(info, sp.Description)
				}
				if len(sp.AllowedDomains) > 0 {
					info = append(info, "domains: "+strings.Join(sp.AllowedDomains, ", "))
				}
				if len(sp.AllowedCommands) > 0 {
					info = append(info, "commands: "+strings.Join(sp.AllowedCommands, ", "))
				}
				if len(info) > 0 {
					fmt.Printf("%-24s %s\n", name, strings.Join(info, "  "))
				} else {
					fmt.Println(name)
				}
			}
			return nil
		},
	}
}
