package cli

import (
	"github.com/spf13/cobra"
)

// identityFile is set by the --identity persistent flag.
var identityFile string

// vaultName is set by the --vault persistent flag.
var vaultName string

func NewRootCmd(version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "vaulty",
		Short:   "Secrets proxy for AI coding agents",
		Long:    "A local-only CLI daemon that acts as a secrets proxy for AI coding agents, so agents can make authenticated API calls without ever seeing raw credentials.",
		Version: version,
	}

	cmd.PersistentFlags().StringVarP(&identityFile, "identity", "i", "", "age identity (private key) file for team vault decryption")
	cmd.PersistentFlags().StringVarP(&vaultName, "vault", "V", "", "named vault to use (stored in vaults/<name>.age)")

	cmd.AddCommand(
		newInitCmd(),
		newSetCmd(),
		newListCmd(),
		newRemoveCmd(),
		newRotateCmd(),
		newStartCmd(),
		newStopCmd(),
		newProxyCmd(),
		newExecCmd(),
		newMCPCmd(),
		newKeychainCmd(),
		newTeamCmd(),
		newExportCmd(),
		newImportCmd(),
		newBackendCmd(),
		newImportEnvCmd(),
		newExportEnvCmd(),
		newImportRailsCmd(),
		newExportRailsCmd(),
		newImportDockerCmd(),
		newExportDockerCmd(),
		newImportK8sCmd(),
		newExportK8sCmd(),
	)

	return cmd
}
