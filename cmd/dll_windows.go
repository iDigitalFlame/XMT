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

import (
	"github.com/iDigitalFlame/xmt/cmd/filter"
	"github.com/iDigitalFlame/xmt/device/winapi"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

// Pid returns the process ID of the owning process (the process running
// the thread.)
//
// This may return zero if the thread has not yet been started.
func (d *DLL) Pid() uint32 {
	return d.t.Pid()
}

// Start will attempt to start the DLL and will return an errors that occur while
// starting the DLL.
//
// This function will return 'ErrEmptyCommand' if the 'Path' parameter is empty
// and 'ErrAlreadyStarted' if attempting to start a DLL that already has been
// started previously.
//
// Always returns 'ErrNoWindows' on non-Windows devices.
func (d *DLL) Start() error {
	if len(d.Path) == 0 {
		return ErrEmptyCommand
	}
	if d.Running() {
		return ErrAlreadyStarted
	}
	p, err := winapi.UTF16FromString(d.Path)
	if err != nil {
		return xerr.Wrap("could not convert path", err)
	}
	b := make([]byte, (len(p)*2)+1)
	for i := 0; i < len(b)-1; i += 2 {
		b[i], b[i+1] = byte(p[i/2]), byte(p[i/2]>>8)
	}
	if err := d.t.Start(0, d.Timeout, winapi.LoadLibraryAddress(), b, d.SameProcess); err != nil {
		return err
	}
	b = nil
	go d.t.wait(0, 0)
	return nil
}

// SetParent will instruct the DLL to choose a parent with the supplied process
// Filter. If the Filter is nil this will use the current process (default).
//
// This function has no effect if the device is not running Windows.
func (d *DLL) SetParent(f *filter.Filter) {
	d.t.filter = f
}
