//go:build !windows

package executor

func shellCommand(command string) (string, []string) {
	return "sh", []string{"-c", command}
}
