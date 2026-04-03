//go:build windows

package daemon

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
)

// StopProcess terminates the process with the given PID.
// On Windows, SIGTERM is not available, so we use taskkill.
func StopProcess(pid int) error {
	// Try taskkill first for a clean shutdown.
	cmd := exec.Command("taskkill", "/PID", strconv.Itoa(pid))
	if err := cmd.Run(); err != nil {
		// Fallback: use os.Process.Kill (forceful).
		process, findErr := os.FindProcess(pid)
		if findErr != nil {
			return fmt.Errorf("finding process %d: %w", pid, findErr)
		}
		return process.Kill()
	}
	return nil
}
