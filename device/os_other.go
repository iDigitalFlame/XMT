//go:build js

// Copyright (C) 2020 - 2022 iDigitalFlame
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
	"io"
	"syscall"

	"github.com/iDigitalFlame/xmt/cmd/filter"
)

const (
	// OS is the local machine's Operating System type.
	OS = Unsupported

	// Shell is the default machine specific command shell.
	Shell = ""
	// ShellArgs is the default machine specific command shell arguments to run
	// commands.
	ShellArgs = ""
	// PowerShell is the path to the PowerShell binary, which is based on the
	// underlying OS type.
	PowerShell = ""
	home       = ""
)

// IsDebugged returns true if the current process is attached by a debugger.
func IsDebugged() bool {
	return false
}

// Mounts attempts to get the mount points on the local device.
//
// On Windows devices, this is the drive letters avaliable, otherwise on nix*
// systems, this will be the mount points on the system.
//
// The return result (if no errors occurred) will be a string list of all the
// mount points (or Windows drive letters).
func Mounts() ([]string, error) {
	return nil, syscall.EINVAL
}

// DumpProcess will attempt to copy the memory of the targeted Filter to the
// supplied Writer. This fill select the first process that matches the Filter.
//
// If the Filter is nil or empty or if an error occurs during reading/writing
// an error will be returned.
func DumpProcess(_ *filter.Filter, _ io.Writer) error {
	return syscall.EINVAL
}
