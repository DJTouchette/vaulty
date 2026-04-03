package cli

import (
	"fmt"
	"os"

	"github.com/djtouchette/vaulty/internal/framework"
	"github.com/djtouchette/vaulty/internal/policy"
	"github.com/spf13/cobra"
)

func newImportK8sCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import-k8s <manifest.yaml>",
		Short: "Import secrets from a Kubernetes Secret manifest",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			manifestPath := args[0]

			data, err := os.ReadFile(manifestPath)
			if err != nil {
				return fmt.Errorf("reading %s: %w", manifestPath, err)
			}

			secrets, err := framework.ParseK8sSecret(data)
			if err != nil {
				return err
			}

			if len(secrets) == 0 {
				fmt.Println("No secrets found in manifest.")
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

			fmt.Printf("Imported %d secrets from %s.\n", len(secrets), manifestPath)
			return nil
		},
	}

	return cmd
}

func newExportK8sCmd() *cobra.Command {
	var (
		name      string
		namespace string
		outFile   string
	)

	cmd := &cobra.Command{
		Use:   "export-k8s",
		Short: "Export vault secrets as a Kubernetes Secret manifest",
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				return fmt.Errorf("--name is required")
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

			names := h.Vault.List()
			secrets := make(map[string]string, len(names))
			for _, n := range names {
				val, _ := h.Vault.Get(n)
				secrets[n] = val
			}

			if outFile != "" {
				f, err := os.Create(outFile)
				if err != nil {
					return fmt.Errorf("creating %s: %w", outFile, err)
				}
				defer f.Close()
				if err := framework.WriteK8sSecret(f, name, namespace, secrets); err != nil {
					return err
				}
				fmt.Printf("Exported %d secrets to %s.\n", len(secrets), outFile)
				return nil
			}

			return framework.WriteK8sSecret(os.Stdout, name, namespace, secrets)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Kubernetes Secret name (required)")
	cmd.Flags().StringVar(&namespace, "namespace", "", "Kubernetes namespace")
	cmd.Flags().StringVar(&outFile, "out", "", "output file (default: stdout)")
	return cmd
}
