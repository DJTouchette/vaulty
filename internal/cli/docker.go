package cli

import (
	"fmt"
	"os"

	"github.com/djtouchette/vaulty/internal/framework"
	"github.com/djtouchette/vaulty/internal/policy"
	"github.com/spf13/cobra"
)

func newExportDockerCmd() *cobra.Command {
	var (
		outFile    string
		service    string
		secretsDir string
	)

	cmd := &cobra.Command{
		Use:   "export-docker",
		Short: "Export vault secrets for Docker/Compose",
		Long: `Exports vault secrets as a docker-compose.override.yml or as Docker secret files.

By default generates docker-compose.override.yml with environment variables.
Use --secrets-dir to write each secret as a separate file (Docker secrets format).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := policy.LoadOrDefault("")
			if err != nil {
				return err
			}

			h, err := openVault(cfg.Vault.Path)
			if err != nil {
				return err
			}
			defer h.Vault.Zero()

			names := h.Vault.List()
			secrets := make(map[string]string, len(names))
			for _, name := range names {
				val, _ := h.Vault.Get(name)
				secrets[name] = val
			}

			// Docker secret files mode
			if secretsDir != "" {
				if err := framework.WriteSecretFiles(secretsDir, secrets); err != nil {
					return err
				}
				fmt.Printf("Wrote %d secret files to %s/\n", len(secrets), secretsDir)
				return nil
			}

			// Compose override mode
			svcName := service
			if svcName == "" {
				svcName = "app"
			}

			if outFile != "" {
				f, err := os.Create(outFile)
				if err != nil {
					return fmt.Errorf("creating %s: %w", outFile, err)
				}
				defer f.Close()
				if err := framework.WriteComposeOverride(f, secrets, svcName); err != nil {
					return err
				}
				fmt.Printf("Exported %d secrets to %s.\n", len(secrets), outFile)
				return nil
			}

			return framework.WriteComposeOverride(os.Stdout, secrets, svcName)
		},
	}

	cmd.Flags().StringVar(&outFile, "out", "", "output file (default: stdout)")
	cmd.Flags().StringVar(&service, "service", "", "Docker Compose service name (default: app)")
	cmd.Flags().StringVar(&secretsDir, "secrets-dir", "", "write each secret as a file in this directory")
	return cmd
}

func newImportDockerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import-docker <compose-file>",
		Short: "Import environment variables from a docker-compose.yml",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			composePath := args[0]

			data, err := os.ReadFile(composePath)
			if err != nil {
				return fmt.Errorf("reading %s: %w", composePath, err)
			}

			secrets, err := framework.ParseComposeEnv(data)
			if err != nil {
				return err
			}

			if len(secrets) == 0 {
				fmt.Println("No environment variables found in compose file.")
				return nil
			}

			cfg, err := policy.LoadOrDefault("")
			if err != nil {
				return err
			}

			h, err := openVault(cfg.Vault.Path)
			if err != nil {
				return err
			}
			defer h.Vault.Zero()

			for key, val := range secrets {
				h.Vault.Set(key, val)
			}

			if err := h.Vault.Save(cfg.Vault.Path, h.Passphrase); err != nil {
				return err
			}

			fmt.Printf("Imported %d environment variables from %s.\n", len(secrets), composePath)
			return nil
		},
	}

	return cmd
}
