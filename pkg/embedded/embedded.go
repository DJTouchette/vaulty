// Package embedded exports vaulty's CLI command tree for embedding in other tools.
package embedded

import (
	"github.com/djtouchette/vaulty/internal/cli"
	"github.com/spf13/cobra"
)

// NewCommand returns vaulty's root cobra command.
// Callers can execute it directly or attach it as a subcommand.
func NewCommand(version string) *cobra.Command {
	return cli.NewRootCmd(version)
}
