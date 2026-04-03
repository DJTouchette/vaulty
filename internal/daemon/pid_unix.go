//go:build !windows

package daemon

import "os"

// IsProcessAlive checks whether the process with the given PID is still running.
// On Unix, sending signal 0 checks for existence without affecting the process.
func IsProcessAlive(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// Signal(nil) returns nil if the process exists and we have permission.
	return process.Signal(nil) == nil
}
