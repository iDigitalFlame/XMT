//go:build windows && crypt

package device

import (
	"os"

	"github.com/iDigitalFlame/xmt/util/crypt"
)

var (
	// ShellArgs is the default machine specific command shell arguments to run
	// commands.
	ShellArgs = crypt.Get(37) // /c
	// PowerShell is the path to the PowerShell binary, which is based on the
	// underlying OS type.
	PowerShell = crypt.Get(38) // powershell.exe
	home       = crypt.Get(19) // %USERPROFILE%
	debugDlls  = crypt.Get(24) // hal.dll\nwmi.dll\nwpx.dll\nwdc.dll\nzipfldr.dll

)

func shell() string {
	if s, ok := os.LookupEnv(crypt.Get(39)); ok { // ComSpec
		return s
	}
	if d, ok := os.LookupEnv(crypt.Get(40)); ok { // WinDir
		p := d + crypt.Get(41) // \system32\cmd.exe
		if s, err := os.Stat(p); err == nil && !s.IsDir() {
			return p
		}
	}
	return crypt.Get(42) // %WinDir%\system32\cmd.exe
}
