//go:build !windows

package mediacreator

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}

func notifyLaunch(url, logPath string) {
	fmt.Fprintf(os.Stderr, "vyntrio-media-creator: open %s\n", url)
	fmt.Fprintf(os.Stderr, "vyntrio-media-creator: log %s\n", logPath)
}

func userFacingStartupError(err error, logPath string) {
	fmt.Fprintf(os.Stderr, "vyntrio-media-creator: failed to start: %v\n", err)
	fmt.Fprintf(os.Stderr, "vyntrio-media-creator: log %s\n", logPath)
}
