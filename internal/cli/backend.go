package cli

import (
	"fmt"
	"sort"

	"github.com/djtouchette/vaulty/internal/backend"
	"github.com/djtouchette/vaulty/internal/policy"
	"github.com/spf13/cobra"
)

func newBackendCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backend",
		Short: "Manage cloud provider secret backends",
		Long:  "List configured backends, browse their secrets, and pull secrets into the local vault.",
	}

	cmd.AddCommand(
		newBackendListCmd(),
		newBackendSecretsCmd(),
		newBackendPullCmd(),
	)

	return cmd
}

func newBackendListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List configured backends",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := policy.LoadOrDefault("")
			if err != nil {
				return err
			}

			if len(cfg.Backends) == 0 {
				fmt.Println("No backends configured. Add a [backends.<name>] section to vaulty.toml.")
				return nil
			}

			names := make([]string, 0, len(cfg.Backends))
			for name := range cfg.Backends {
				names = append(names, name)
			}
			sort.Strings(names)

			for _, name := range names {
				bc := cfg.Backends[name]
				fmt.Printf("%-24s type=%s", name, bc.Type)
				if bc.Region != "" {
					fmt.Printf("  region=%s", bc.Region)
				}
				if bc.Profile != "" {
					fmt.Printf("  profile=%s", bc.Profile)
				}
				if bc.Project != "" {
					fmt.Printf("  project=%s", bc.Project)
				}
				if bc.Addr != "" {
					fmt.Printf("  addr=%s", bc.Addr)
				}
				if bc.Mount != "" {
					fmt.Printf("  mount=%s", bc.Mount)
				}
				if bc.OpVault != "" {
					fmt.Printf("  op_vault=%s", bc.OpVault)
				}
				fmt.Println()
			}
			return nil
		},
	}
}

func newBackendSecretsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "secrets <backend-name>",
		Short: "List secrets available from a backend",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			backendName := args[0]

			cfg, err := policy.LoadOrDefault("")
			if err != nil {
				return err
			}

			bc, ok := cfg.Backends[backendName]
			if !ok {
				return fmt.Errorf("backend %q not found in config — run 'vaulty backend list' to see configured backends", backendName)
			}

			b, err := backend.NewBackend(toBackendConfig(bc))
			if err != nil {
				return err
			}

			secrets, err := b.List()
			if err != nil {
				return err
			}

			if len(secrets) == 0 {
				fmt.Printf("No secrets found in backend %q.\n", backendName)
				return nil
			}

			for _, s := range secrets {
				fmt.Println(s)
			}
			return nil
		},
	}
}

func newBackendPullCmd() *cobra.Command {
	var asName string

	cmd := &cobra.Command{
		Use:   "pull <backend-name> <secret-name>",
		Short: "Pull a secret from a backend into the local vault",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			backendName := args[0]
			secretName := args[1]

			cfg, err := policy.LoadOrDefault("")
			if err != nil {
				return err
			}

			bc, ok := cfg.Backends[backendName]
			if !ok {
				return fmt.Errorf("backend %q not found in config — run 'vaulty backend list' to see configured backends", backendName)
			}

			b, err := backend.NewBackend(toBackendConfig(bc))
			if err != nil {
				return err
			}

			value, err := b.Get(secretName)
			if err != nil {
				return err
			}

			vaultName := secretName
			if asName != "" {
				vaultName = asName
			}

			h, err := openVault(cfg.Vault.Path)
			if err != nil {
				return err
			}
			defer h.Vault.Zero()

			h.Vault.Set(vaultName, value)

			if err := h.Vault.Save(cfg.Vault.Path, h.Passphrase); err != nil {
				return err
			}

			// Update policy to record which backend this secret came from
			sp := cfg.GetSecretPolicy(vaultName)
			sp.Backend = backendName
			cfg.SetSecretPolicy(vaultName, sp)
			if err := cfg.Write(); err != nil {
				return fmt.Errorf("writing config: %w", err)
			}

			fmt.Printf("Secret %q pulled from %s and stored as %q.\n", secretName, backendName, vaultName)
			return nil
		},
	}

	cmd.Flags().StringVar(&asName, "as", "", "store the secret under a different name in the vault")
	return cmd
}

// toBackendConfig converts a policy.BackendConfig to a backend.BackendConfig.
func toBackendConfig(pc policy.BackendConfig) backend.BackendConfig {
	return backend.BackendConfig{
		Type:     pc.Type,
		Region:   pc.Region,
		Profile:  pc.Profile,
		Endpoint: pc.Endpoint,
		Project:  pc.Project,
		Addr:     pc.Addr,
		Mount:    pc.Mount,
		OpVault:  pc.OpVault,
		TTL:      pc.TTL,
	}
}
