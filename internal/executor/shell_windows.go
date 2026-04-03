//go:build windows

package executor

func shellCommand(command string) (string, []string) {
	return "cmd", []string{"/C", command}
}
