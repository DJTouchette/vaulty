//go:build !windows

package daemon

import (
	"os"
	"syscall"
)

// StopProcess sends SIGTERM to the process with the given PID.
func StopProcess(pid int) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return process.Signal(syscall.SIGTERM)
}
