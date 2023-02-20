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

package main

import (
	"fmt"

	"github.com/iDigitalFlame/xmt/device/winapi"
)

func exampleSignedDrivers() {
	err := winapi.EnumDrivers(func(h uintptr, s string) error {
		v, err := winapi.FileSigningIssuerName(s)
		if err != nil || len(v) == 0 {
			return nil
		}
		switch winapi.FnvHash(v) {
		case 0x1FB166BC: // Microsoft Windows
			fallthrough
		case 0x4C18C11F: // Microsoft Windows Hardware Abstraction Layer Publisher
			return nil
		}
		fmt.Printf("Unsigned/non-MS: 0x%X: %s [%s]\n", h, s, v)
		return nil
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(winapi.IsSecureBootEnabled())
	fmt.Println(winapi.GetCodeIntegrityState())
}
