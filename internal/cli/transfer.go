package cli

import (
	"fmt"
	"os"

	"github.com/djtouchette/vaulty/internal/policy"
	"github.com/djtouchette/vaulty/internal/vault"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func newExportCmd() *cobra.Command {
	var outFile string

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export vault secrets as an encrypted snapshot",
		RunE: func(cmd *cobra.Command, args []string) error {
			if outFile == "" {
				return fmt.Errorf("--out flag is required")
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

			// Load recipients for the source vault (if any)
			recipients, err := vault.LoadRecipients(vaultPath)
			if err != nil {
				return fmt.Errorf("loading recipients: %w", err)
			}

			data, err := vault.Export(h.Vault, h.Passphrase, recipients)
			if err != nil {
				return err
			}

			if err := os.WriteFile(outFile, data, 0600); err != nil {
				return fmt.Errorf("writing export file: %w", err)
			}

			names := h.Vault.List()
			fmt.Printf("Exported %d secret(s) to %s\n", len(names), outFile)
			return nil
		},
	}

	cmd.Flags().StringVar(&outFile, "out", "", "output file path (required)")
	return cmd
}

func newImportCmd() *cobra.Command {
	var (
		fromFile  string
		overwrite bool
	)

	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import secrets from an encrypted snapshot into the vault",
		RunE: func(cmd *cobra.Command, args []string) error {
			if fromFile == "" {
				return fmt.Errorf("--from flag is required")
			}

			importData, err := os.ReadFile(fromFile)
			if err != nil {
				return fmt.Errorf("reading import file: %w", err)
			}

			// Open import file — use same credentials as current vault
			importIdentity := identityFile
			if importIdentity == "" {
				importIdentity = os.Getenv("VAULTY_IDENTITY")
			}

			var importPass string
			if importIdentity == "" {
				// We need a passphrase for the import file. Prompt specifically.
				fmt.Print("Import file passphrase: ")
				importPass, err = readPassword()
				if err != nil {
					return err
				}
			}

			srcVault, err := vault.Import(importData, importPass, importIdentity)
			if err != nil {
				return err
			}
			defer srcVault.Zero()

			cfg, err := policy.LoadOrDefault("")
			if err != nil {
				return err
			}

			vaultPath := resolveVaultPath(cfg.Vault.Path)

			// Create the vault if it doesn't exist
			if !vault.Exists(vaultPath) {
				dstPass, err := getPassphrase(vaultPath)
				if err != nil {
					return fmt.Errorf("getting passphrase for new vault: %w", err)
				}
				if err := vault.Create(vaultPath, dstPass); err != nil {
					return err
				}
			}

			h, err := openVault(vaultPath)
			if err != nil {
				return err
			}
			defer h.Vault.Zero()

			count := vault.MergeVaults(h.Vault, srcVault, overwrite)

			if err := h.Vault.Save(vaultPath, h.Passphrase); err != nil {
				return err
			}

			fmt.Printf("Imported %d secret(s) into vault.\n", count)
			return nil
		},
	}

	cmd.Flags().StringVar(&fromFile, "from", "", "import file path (required)")
	cmd.Flags().BoolVar(&overwrite, "overwrite", false, "overwrite existing secrets")
	return cmd
}

func readPassword() (string, error) {
	raw, err := term.ReadPassword(0)
	fmt.Println()
	if err != nil {
		return "", fmt.Errorf("reading passphrase: %w", err)
	}
	return string(raw), nil
}
