package clipboard

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
)

// CopyToClipboard copies text to clipboard based on platform
func CopyToClipboard(text string) error {
	var cmd *exec.Cmd

	// Check if running on Android via Termux
	if _, err := os.Stat("/data/data/com.termux"); err == nil {
		cmd = exec.Command("termux-clipboard-set")
		stdin, err := cmd.StdinPipe()
		if err != nil {
			return err
		}
		go func() {
			defer stdin.Close()
			io.WriteString(stdin, text)
		}()
	} else if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		cmd = exec.Command("xclip", "-selection", "clipboard")
		stdin, err := cmd.StdinPipe()
		if err != nil {
			return err
		}
		go func() {
			defer stdin.Close()
			io.WriteString(stdin, text)
		}()
	} else if runtime.GOOS == "windows" {
		cmd = exec.Command("powershell", "-command", "Set-Clipboard", "-Value", text)
	} else {
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return cmd.Run()
}
