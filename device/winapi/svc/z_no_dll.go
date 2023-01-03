//go:build windows && !svcdll

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

import "github.com/iDigitalFlame/xmt/device/winapi"

const serviceType = 0x10

func serviceWireThread(n string) error {
	v, err := winapi.UTF16PtrFromString(n)
	if err != nil {
		return err
	}
	t := [2]winapi.ServiceTableEntry{
		{Name: v, Proc: callBack.m}, {Name: nil, Proc: 0},
	}
	return winapi.StartServiceCtrlDispatcher(&t[0])
}
