package cli

import (
	"fmt"
	"os"
	"strconv"

	"github.com/djtouchette/vaulty/internal/daemon"
	"github.com/spf13/cobra"
)

func newStopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop the Vaulty daemon",
		RunE: func(cmd *cobra.Command, args []string) error {
			pidPath := daemon.PIDFilePath()

			data, err := os.ReadFile(pidPath)
			if err != nil {
				return fmt.Errorf("daemon not running (no PID file at %s)", pidPath)
			}

			pid, err := strconv.Atoi(string(data))
			if err != nil {
				return fmt.Errorf("invalid PID file: %w", err)
			}

			if err := daemon.StopProcess(pid); err != nil {
				// Process might already be dead
				os.Remove(pidPath)
				return fmt.Errorf("stopping pid %d: %w", pid, err)
			}

			fmt.Println("Daemon stopped. Secrets cleared from memory.")
			return nil
		},
	}
}
