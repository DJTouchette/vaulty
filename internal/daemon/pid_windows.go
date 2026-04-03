//go:build windows

package daemon

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// IsProcessAlive checks whether the process with the given PID is still running.
// On Windows, os.FindProcess always succeeds, so we use tasklist to verify.
func IsProcessAlive(pid int) bool {
	cmd := exec.Command("tasklist", "/FI", "PID eq "+strconv.Itoa(pid), "/NH")
	out, err := cmd.Output()
	if err != nil {
		// Fallback: try FindProcess + Signal. On Windows this is unreliable
		// but better than assuming dead.
		process, err := os.FindProcess(pid)
		if err != nil {
			return false
		}
		return process.Signal(os.Interrupt) != nil
	}
	// tasklist output contains the PID number if the process exists.
	return strings.Contains(string(out), strconv.Itoa(pid))
}
