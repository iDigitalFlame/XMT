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

package evade

import "github.com/iDigitalFlame/xmt/device"

// ZeroTraceEvent will attempt to zero out the NtTraceEvent function call with
// a NOP.
//
// This will return an error if it fails.
//
// This is just a wrapper for the winapi function call as we want to keep the
// function body in winapi for easy crypt wrapping.
//
// Always returns 'ErrNoWindows' on non-Windows devices.
func ZeroTraceEvent() error {
	return device.ErrNoWindows
}

// ReloadDLL is a function shamelessly stolen from the sliver project. This
// function will read a DLL file from on-disk and rewrite over it's current
// in-memory contents to erase any hooks placed on function calls.
//
// Re-mastered and refactored to be less memory hungry and easier to read :P
//
// Orig src here:
//   https://github.com/BishopFox/sliver/blob/f94f0fc938ca3871390c5adfa71bf4edc288022d/implant/sliver/evasion/evasion_windows.go#L116
//
// Always returns 'ErrNoWindows' on non-Windows devices.
func ReloadDLL(_ string) error {
	return device.ErrNoWindows
}

// CheckDLL is a similar function to ReloadDLL.
// This function will return 'true' and 'nil' if the contents in memory match the
// contents of the file on disk. Otherwise it returns false.
//
// Always returns 'ErrNoWindows' on non-Windows devices.
func CheckDLL(_ string) (bool, error) {
	return false, device.ErrNoWindows
}
