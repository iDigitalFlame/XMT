//go:build windows && go1.17
// +build windows,go1.17

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

package winapi

import (
	"syscall"

	// Importing unsafe to use "linkname"
	_ "unsafe"
)

func removeCtrlHandler() {
	var (
		h = syscall.NewCallback(ctrlHandler)
		z = callbackasmAddr(0) // Base of runtime.callbackasm
		e = callbackasmAddr(1) - z
	)
	if e <= 0 {
		// Weird shit, shouldn't happen.
		return
	}
	// Instead of trying to use it directly (which never works), we grab a callback
	// index and determine the entrySize between 1 and 0.
	//
	// We then loop until we hit the handle or hit the callbackasm address instead.
	for ; h >= z; h -= e {
		if h == 0 {
			// DO NOT SET HANDLERS TO IGNORE.
			break
		}
		if r, _, _ := syscallN(setConsoleCtrlHandler, h, 0); r > 0 {
			break
		}
	}
}

//go:linkname ctrlHandler runtime.ctrlHandler
func ctrlHandler(uint32) uintptr

//go:linkname callbackasmAddr runtime.callbackasmAddr
func callbackasmAddr(int) uintptr
