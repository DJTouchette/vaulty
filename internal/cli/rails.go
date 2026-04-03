package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/djtouchette/vaulty/internal/framework"
	"github.com/djtouchette/vaulty/internal/policy"
	"github.com/spf13/cobra"
)

func newImportRailsCmd() *cobra.Command {
	var env string

	cmd := &cobra.Command{
		Use:   "import-rails",
		Short: "Import secrets from Rails encrypted credentials",
		Long: `Decrypts and imports Rails credentials into the vault.

By default reads config/credentials.yml.enc with config/master.key.
Use --env to read config/credentials/<env>.yml.enc instead.
The master key is read from RAILS_MASTER_KEY env var or config/master.key.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Determine encrypted credentials path
			var encPath string
			if env != "" {
				encPath = filepath.Join("config", "credentials", env+".yml.enc")
			} else {
				encPath = filepath.Join("config", "credentials.yml.enc")
			}

			// Determine master key path
			var keyPath string
			masterKeyEnv := os.Getenv("RAILS_MASTER_KEY")
			if masterKeyEnv != "" {
				// Write env var to temp file for the decrypt function
				tmpDir, err := os.MkdirTemp("", "vaulty-rails-*")
				if err != nil {
					return fmt.Errorf("creating temp dir: %w", err)
				}
				defer os.RemoveAll(tmpDir)
				keyPath = filepath.Join(tmpDir, "master.key")
				if err := os.WriteFile(keyPath, []byte(masterKeyEnv), 0600); err != nil {
					return fmt.Errorf("writing temp key: %w", err)
				}
			} else {
				keyPath = filepath.Join("config", "master.key")
			}

			plaintext, err := framework.DecryptRailsCredentials(encPath, keyPath)
			if err != nil {
				return fmt.Errorf("decrypting Rails credentials: %w\n\nAlternatively, pipe decrypted YAML via:\n  rails credentials:show | vaulty import-env /dev/stdin", err)
			}

			secrets, err := framework.ParseRailsCredentials(plaintext)
			if err != nil {
				return err
			}

			if len(secrets) == 0 {
				fmt.Println("No secrets found in credentials.")
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

			fmt.Printf("Imported %d secrets from Rails credentials.\n", len(secrets))
			return nil
		},
	}

	cmd.Flags().StringVar(&env, "env", "", "Rails environment (e.g. production) — reads config/credentials/<env>.yml.enc")
	return cmd
}

func newExportRailsCmd() *cobra.Command {
	var outFile string

	cmd := &cobra.Command{
		Use:   "export-rails",
		Short: "Export vault secrets as Rails credentials YAML",
		Long: `Exports vault secrets as nested YAML suitable for Rails credentials.

Keys are unflattened from SECTION_KEY format to nested YAML.
For example: AWS_ACCESS_KEY_ID becomes aws.access_key_id.`,
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

			data, err := framework.WriteRailsCredentials(secrets)
			if err != nil {
				return err
			}

			if outFile != "" {
				if err := os.WriteFile(outFile, data, 0600); err != nil {
					return fmt.Errorf("writing %s: %w", outFile, err)
				}
				fmt.Printf("Exported %d secrets to %s.\n", len(secrets), outFile)
				return nil
			}

			os.Stdout.Write(data)
			return nil
		},
	}

	cmd.Flags().StringVar(&outFile, "out", "", "output file (default: stdout)")
	return cmd
}
