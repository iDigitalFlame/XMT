//go:build !windows

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

package cmd

import (
	"github.com/iDigitalFlame/xmt/cmd/filter"
	"github.com/iDigitalFlame/xmt/device"
)

// Pid returns the process ID of the owning process (the process running
// the thread.)
//
// This may return zero if the thread has not yet been started.
func (Assembly) Pid() uint32 {
	return 0
}

// Start will attempt to start the Assembly thread and will return any errors
// that occur while starting the thread.
//
// This function will return 'ErrEmptyCommand' if the 'Data' parameter is empty or
// the 'ErrAlreadyStarted' error if attempting to start a thread that already has
// been started previously.
//
// Always returns 'ErrNoWindows' on non-Windows devices.
func (Assembly) Start() error {
	return device.ErrNoWindows
}

// SetParent will instruct the Assembly thread to choose a parent with the supplied
// process Filter. If the Filter is nil this will use the current process (default).
//
// This function has no effect if the device is not running Windows.
func (Assembly) SetParent(_ *filter.Filter) {}
