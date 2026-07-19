//go:build windows

package mediacreator

import (
	"fmt"
	"os/exec"
	"syscall"

	"golang.org/x/sys/windows"
)

func openBrowser(url string) error {
	// Prefer cmd start — more reliable than rundll32 FileProtocolHandler alone.
	cmd := exec.Command("cmd", "/c", "start", "", url)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if err := cmd.Start(); err == nil {
		return nil
	}
	cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	return cmd.Start()
}

func notifyLaunch(url, logPath string) {
	title, _ := windows.UTF16PtrFromString("Vyntrio Media Creator")
	text, _ := windows.UTF16PtrFromString(fmt.Sprintf(
		"Vyntrio Media Creator is running.\n\nYour browser should open:\n%s\n\nIf nothing opens, paste that URL into Edge or Chrome.\n\nLog file:\n%s\n\nClick OK to dismiss this message. The writer keeps running until you stop it from the wizard or Task Manager.",
		url, logPath,
	))
	windows.MessageBox(0, text, title, windows.MB_OK|windows.MB_ICONINFORMATION)
}

func userFacingStartupError(err error, logPath string) {
	title, _ := windows.UTF16PtrFromString("Vyntrio Media Creator")
	text, _ := windows.UTF16PtrFromString(fmt.Sprintf(
		"Vyntrio Media Creator failed to start.\n\n%v\n\nDetails were written to:\n%s",
		err, logPath,
	))
	windows.MessageBox(0, text, title, windows.MB_OK|windows.MB_ICONERROR)
}
