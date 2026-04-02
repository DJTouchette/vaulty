package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/djtouchette/vaulty/internal/daemon"
	"github.com/djtouchette/vaulty/internal/policy"
	"github.com/djtouchette/vaulty/internal/vault"
	"github.com/spf13/cobra"
)

func newStartCmd() *cobra.Command {
	var foreground bool
	var extraVaults string

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the Vaulty daemon",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check if already running
			pidPath := daemon.PIDFilePath()
			if data, err := os.ReadFile(pidPath); err == nil {
				pid, _ := strconv.Atoi(string(data))
				if pid > 0 && daemon.IsProcessAlive(pid) {
					return fmt.Errorf("daemon already running (pid %d)", pid)
				}
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

			if foreground {
				vaults := map[string]*vault.Vault{"": h.Vault}

				// Load additional named vaults
				if extraVaults != "" {
					for _, name := range splitTrimVaults(extraVaults) {
						namedPath := vault.ResolveVaultPath(name, cfg.Vault.Path)
						nh, err := openVault(namedPath)
						if err != nil {
							return fmt.Errorf("opening vault %q: %w", name, err)
						}
						vaults[name] = nh.Vault
					}
				}

				d, err := daemon.New(vaults, cfg)
				if err != nil {
					return err
				}

				fmt.Printf("Vaulty daemon started (pid %d)\n", os.Getpid())
				if cfg.Vault.Socket != "" {
					fmt.Printf("Listening on %s\n", cfg.Vault.Socket)
				}
				if cfg.Vault.HTTPPort > 0 {
					fmt.Printf("Listening on http://127.0.0.1:%d\n", cfg.Vault.HTTPPort)
				}

				return d.Run(context.Background())
			}

			// Background: re-exec ourselves with --foreground
			h.Vault.Zero() // Don't keep secrets in this process

			exe, err := os.Executable()
			if err != nil {
				return fmt.Errorf("finding executable: %w", err)
			}

			childArgs := []string{"start", "--foreground"}
			childEnv := os.Environ()

			// Pass credentials to the child process
			if identityFile != "" {
				childArgs = append(childArgs, "--identity", identityFile)
			} else if h.Passphrase != "" {
				childEnv = append(childEnv, fmt.Sprintf("VAULTY_PASSPHRASE=%s", h.Passphrase))
			}

			// Pass vault flags to the child process
			if vaultName != "" {
				childArgs = append(childArgs, "--vault", vaultName)
			}
			if extraVaults != "" {
				childArgs = append(childArgs, "--vaults", extraVaults)
			}

			child := exec.Command(exe, childArgs...)
			child.Stdin = nil
			child.Stdout = nil
			child.Stderr = nil
			child.Env = childEnv
			child.SysProcAttr = nil

			if err := child.Start(); err != nil {
				return fmt.Errorf("starting daemon: %w", err)
			}

			fmt.Printf("Vaulty daemon started (pid %d)\n", child.Process.Pid)
			if cfg.Vault.Socket != "" {
				fmt.Printf("Listening on %s\n", cfg.Vault.Socket)
			}
			if cfg.Vault.HTTPPort > 0 {
				fmt.Printf("Listening on http://127.0.0.1:%d\n", cfg.Vault.HTTPPort)
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&foreground, "foreground", false, "run in foreground (don't daemonize)")
	cmd.Flags().StringVar(&extraVaults, "vaults", "", "additional named vaults to load (comma-separated)")
	return cmd
}

func splitTrimVaults(s string) []string {
	parts := strings.Split(s, ",")
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
