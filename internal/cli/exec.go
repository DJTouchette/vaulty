package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/djtouchette/vaulty/internal/daemon"
	"github.com/djtouchette/vaulty/internal/policy"
	"github.com/spf13/cobra"
)

func newExecCmd() *cobra.Command {
	var secrets []string

	cmd := &cobra.Command{
		Use:   "exec [--secret NAME]... -- <command> [args...]",
		Short: "Run a command with secrets injected as environment variables",
		DisableFlagParsing: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("command required after --")
			}
			if len(secrets) == 0 {
				return fmt.Errorf("at least one --secret is required")
			}

			command := strings.Join(args, " ")

			cfg, err := policy.LoadOrDefault("")
			if err != nil {
				return err
			}

			client := daemon.NewClient(cfg.Vault.Socket, cfg.Vault.HTTPPort)

			resp, err := client.Send(daemon.Request{
				Action:  "exec",
				Command: command,
				Secrets: secrets,
			})
			if err != nil {
				return err
			}

			if resp.Error != "" {
				return fmt.Errorf("%s", resp.Error)
			}

			if resp.Stdout != "" {
				fmt.Print(resp.Stdout)
			}
			if resp.Stderr != "" {
				fmt.Fprint(os.Stderr, resp.Stderr)
			}

			if resp.ExitCode != nil && *resp.ExitCode != 0 {
				os.Exit(*resp.ExitCode)
			}
			return nil
		},
	}

	cmd.Flags().StringArrayVar(&secrets, "secret", nil, "secret names to inject")
	return cmd
}
