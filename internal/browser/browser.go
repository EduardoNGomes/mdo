package browser

import (
	"fmt"
	"os/exec"
	"runtime"
)

func Open(url string) error {
	var command string
	var args []string

	switch runtime.GOOS {
	case "linux":
		command, args = "xdg-open", []string{url}
	case "darwin":
		command, args = "open", []string{url}
	case "windows":
		command, args = "rundll32", []string{"url.dll,FileProtocolHandler", url}
	default:
		return fmt.Errorf("sistema operacional %q não suportado", runtime.GOOS)
	}

	cmd := exec.Command(command, args...)
	if err := cmd.Start(); err != nil {
		return err
	}
	return cmd.Process.Release()
}
