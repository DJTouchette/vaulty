package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

type mcpConfig struct {
	MCPServers map[string]mcpServerEntry `json:"mcpServers"`
}

type mcpServerEntry struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

func newMCPInitCmd() *cobra.Command {
	var dir string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Write .mcp.json for Claude Code / Cursor integration",
		Long:  "Creates a .mcp.json file in the target directory so Claude Code, Cursor, or other MCP clients can discover the Vaulty MCP server.",
		RunE: func(cmd *cobra.Command, args []string) error {
			target := dir
			if target == "" {
				wd, err := os.Getwd()
				if err != nil {
					return fmt.Errorf("getting working directory: %w", err)
				}
				target = wd
			}

			path := filepath.Join(target, ".mcp.json")

			if _, err := os.Stat(path); err == nil {
				// File exists — check if vaulty is already configured
				existing, err := os.ReadFile(path)
				if err != nil {
					return fmt.Errorf("reading existing %s: %w", path, err)
				}
				var cfg mcpConfig
				if err := json.Unmarshal(existing, &cfg); err == nil {
					if _, ok := cfg.MCPServers["vaulty"]; ok {
						fmt.Printf("Vaulty is already configured in %s\n", path)
						return nil
					}
				}
				// Merge vaulty into existing config
				var raw map[string]json.RawMessage
				if err := json.Unmarshal(existing, &raw); err != nil {
					return fmt.Errorf("parsing existing %s: %w", path, err)
				}
				servers := make(map[string]json.RawMessage)
				if s, ok := raw["mcpServers"]; ok {
					if err := json.Unmarshal(s, &servers); err != nil {
						return fmt.Errorf("parsing mcpServers in %s: %w", path, err)
					}
				}
				entry, _ := json.Marshal(mcpServerEntry{
					Command: "vaulty",
					Args:    []string{"mcp", "start"},
				})
				servers["vaulty"] = entry
				raw["mcpServers"], _ = json.Marshal(servers)
				return writeJSONFile(path, raw)
			}

			// Create new file
			cfg := mcpConfig{
				MCPServers: map[string]mcpServerEntry{
					"vaulty": {
						Command: "vaulty",
						Args:    []string{"mcp", "start"},
					},
				},
			}
			return writeJSONFile(path, cfg)
		},
	}

	cmd.Flags().StringVar(&dir, "dir", "", "target directory (defaults to current directory)")
	return cmd
}

func writeJSONFile(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling JSON: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}
	fmt.Printf("Wrote %s\n", path)
	return nil
}
