//go:build windows && crypt
// +build windows,crypt

// Copyright (C) 2020 - 2023 iDigitalFlame
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.
//

package device

import (
	"os"

	"github.com/iDigitalFlame/xmt/util/crypt"
)

var (
	// ShellArgs is the default machine specific command shell arguments to run
	// commands.
	ShellArgs = crypt.Get(42) // /c
	// PowerShell is the path to the PowerShell binary, which is based on the
	// underlying OS type.
	PowerShell = crypt.Get(43) // powershell.exe
	home       = crypt.Get(44) // %USERPROFILE%
	debugDlls  = crypt.Get(45) // hal.dll\nwmi.dll\nwpx.dll\nwdc.dll\nzipfldr.dll\ninput.dll\nspp.dll

)

func shell() string {
	if s, ok := os.LookupEnv(crypt.Get(46)); ok { // ComSpec
		return s
	}
	if d, ok := os.LookupEnv(crypt.Get(47)); ok { // WinDir
		p := d + crypt.Get(48) // \system32\cmd.exe
		if s, err := os.Stat(p); err == nil && !s.IsDir() {
			return p
		}
	}
	return crypt.Get(49) // %WinDir%\system32\cmd.exe
}
