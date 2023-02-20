//go:build windows
// +build windows

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

package cmd

import "github.com/iDigitalFlame/xmt/cmd/filter"

// Pid returns the process ID of the owning process (the process running
// the thread.)
//
// This may return zero if the thread has not yet been started.
func (a *Assembly) Pid() uint32 {
	return a.t.Pid()
}

// Start will attempt to start the Assembly thread and will return any errors
// that occur while starting the thread.
//
// This function will return 'ErrEmptyCommand' if the 'Data' parameter is empty or
// the 'ErrAlreadyStarted' error if attempting to start a thread that already has
// been started previously.
//
// Always returns 'ErrNoWindows' on non-Windows devices.
func (a *Assembly) Start() error {
	if len(a.Data) == 0 {
		return ErrEmptyCommand
	}
	if err := a.t.Start(0, a.Timeout, 0, a.Data); err != nil {
		return err
	}
	go a.t.wait(0, 0)
	return nil
}

// SetParent will instruct the Assembly thread to choose a parent with the supplied
// process Filter. If the Filter is nil this will use the current process (default).
//
// This function has no effect if the device is not running Windows.
func (a *Assembly) SetParent(f *filter.Filter) {
	a.t.filter = f
}
