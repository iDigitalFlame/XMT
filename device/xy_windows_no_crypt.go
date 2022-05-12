//go:build windows && !crypt

package device

import "os"

const (
	// ShellArgs is the default machine specific command shell arguments to run
	// commands.
	ShellArgs = "/c"
	// PowerShell is the path to the PowerShell binary, which is based on the
	// underlying OS type.
	PowerShell = "powershell.exe"
	home       = "%USERPROFILE%"
	debugDlls  = "hal.dll\nwmi.dll\nwpx.dll\nwdc.dll\nzipfldr.dll"
)

func shell() string {
	if s, ok := os.LookupEnv("ComSpec"); ok {
		return s
	}
	if d, ok := os.LookupEnv("WinDir"); ok {
		p := d + `\system32\cmd.exe`
		if s, err := os.Stat(p); err == nil && !s.IsDir() {
			return p
		}
	}
	return `%WinDir%\system32\cmd.exe`
}
