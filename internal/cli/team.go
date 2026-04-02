package cli

import (
	"fmt"

	"github.com/djtouchette/vaulty/internal/policy"
	"github.com/djtouchette/vaulty/internal/vault"
	"github.com/spf13/cobra"
)

func newTeamCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "team",
		Short: "Manage team sharing (recipients for the vault)",
		Long:  "Add, list, or remove age public key recipients so team members can decrypt the vault.",
	}

	cmd.AddCommand(
		newTeamAddCmd(),
		newTeamListCmd(),
		newTeamRemoveCmd(),
	)

	return cmd
}

func newTeamAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add <public-key-or-file>",
		Short: "Add a team member's age public key",
		Long:  "Add an age X25519 public key as a vault recipient. Accepts a raw public key (age1...) or a path to a file containing one.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := policy.LoadOrDefault("")
			if err != nil {
				return err
			}

			if err := vault.AddRecipient(cfg.Vault.Path, args[0]); err != nil {
				return err
			}

			fmt.Println("Recipient added.")
			return nil
		},
	}
}

func newTeamListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List current vault recipients",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := policy.LoadOrDefault("")
			if err != nil {
				return err
			}

			recipients, err := vault.ListRecipients(cfg.Vault.Path)
			if err != nil {
				return err
			}

			if len(recipients) == 0 {
				fmt.Println("No team recipients configured.")
				return nil
			}

			for _, r := range recipients {
				fmt.Println(r)
			}
			return nil
		},
	}
}

func newTeamRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <public-key>",
		Short: "Remove a recipient",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := policy.LoadOrDefault("")
			if err != nil {
				return err
			}

			if err := vault.RemoveRecipient(cfg.Vault.Path, args[0]); err != nil {
				return err
			}

			fmt.Println("Recipient removed.")
			return nil
		},
	}
}
