package cli

import (
	"fmt"
	"os"

	"github.com/djtouchette/vaulty/internal/audit"
	"github.com/djtouchette/vaulty/internal/mcp"
	"github.com/djtouchette/vaulty/internal/policy"
	"github.com/spf13/cobra"
)

func newMCPCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "mcp",
		Short: "Start Vaulty as an MCP server (stdio transport)",
		Long:  "Starts Vaulty as an MCP server for direct integration with Claude Code, Cursor, etc. Communicates via JSON-RPC over stdin/stdout.",
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

			logger, err := audit.NewLogger("~/.config/vaulty/audit.log")
			if err != nil {
				return err
			}
			defer logger.Close()

			handler := mcp.NewHandler(h.Vault, cfg, logger)
			resources := mcp.NewResourceHandler(h.Vault, cfg, logger)
			server := mcp.NewServer(handler, resources, os.Stdin, os.Stdout)

			fmt.Fprintln(os.Stderr, "Vaulty MCP server started")
			return server.Run()
		},
	}
}
