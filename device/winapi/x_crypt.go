//go:build windows && crypt

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

import "github.com/iDigitalFlame/xmt/util/crypt"

var (
	dllExt    = crypt.Get(0)[1:] // *.dll
	debugPriv = crypt.Get(81)    // SeDebugPrivilege

	dllAmsi       = &lazyDLL{name: crypt.Get(82)}  // amsi.dll
	dllNtdll      = &lazyDLL{name: crypt.Get(83)}  // ntdll.dll
	dllGdi32      = &lazyDLL{name: crypt.Get(84)}  // gdi32.dll
	dllUser32     = &lazyDLL{name: crypt.Get(85)}  // user32.dll
	dllWinhttp    = &lazyDLL{name: crypt.Get(86)}  // winhttp.dll
	dllDbgHelp    = &lazyDLL{name: crypt.Get(87)}  // DbgHelp.dll
	dllCrypt32    = &lazyDLL{name: crypt.Get(100)} // crypt32.dll
	dllAdvapi32   = &lazyDLL{name: crypt.Get(88)}  // advapi32.dll
	dllWtsapi32   = &lazyDLL{name: crypt.Get(89)}  // wtsapi32.dll
	dllKernel32   = &lazyDLL{name: crypt.Get(90)}  // kernel32.dll
	dllKernelBase = &lazyDLL{name: crypt.Get(91)}  // kernelbase.dll
)
