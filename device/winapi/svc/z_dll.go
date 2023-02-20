//go:build windows && svcdll
// +build windows,svcdll

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

package svc

import "syscall"

const serviceType = 0x20

func serviceWireThread(_ string) error {
	// NOTE(dij): If we are running as a shared DLL inside svchost.exe, there's
	//            no need to init the dispatcher and wire to it. We're already
	//            IN the wire thread (ie: we are the DLL 'ServiceMain' function).
	if r := serviceMain(0, nil); r > 0 {
		return syscall.Errno(r)
	}
	return nil
}
