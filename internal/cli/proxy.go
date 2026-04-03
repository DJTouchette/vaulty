package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/djtouchette/vaulty/internal/daemon"
	"github.com/djtouchette/vaulty/internal/policy"
	"github.com/spf13/cobra"
)

func newProxyCmd() *cobra.Command {
	var (
		secret  string
		headers []string
		body    string
	)

	cmd := &cobra.Command{
		Use:   "proxy <METHOD> <URL>",
		Short: "Make an authenticated HTTP request through Vaulty",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			method := strings.ToUpper(args[0])
			url := args[1]

			if secret == "" {
				return fmt.Errorf("--secret is required")
			}

			cfg, err := policy.LoadOrDefault("")
			if err != nil {
				return err
			}

			client := daemon.NewClient(cfg.Vault.Socket, cfg.Vault.HTTPPort)

			headerMap := make(map[string]string)
			for _, h := range headers {
				parts := strings.SplitN(h, ":", 2)
				if len(parts) == 2 {
					headerMap[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
				}
			}

			resp, err := client.Send(daemon.Request{
				Action:  "proxy",
				Method:  method,
				URL:     url,
				Secret:  secret,
				Headers: headerMap,
				Body:    body,
			})
			if err != nil {
				return err
			}

			if resp.Error != "" {
				return fmt.Errorf("%s", resp.Error)
			}

			fmt.Fprintf(os.Stderr, "HTTP %d\n", resp.Status)
			fmt.Print(resp.Body)
			return nil
		},
	}

	cmd.Flags().StringVar(&secret, "secret", "", "secret name for authentication")
	cmd.Flags().StringArrayVar(&headers, "header", nil, "additional headers (Key: Value)")
	cmd.Flags().StringVar(&body, "body", "", "request body")
	return cmd
}
