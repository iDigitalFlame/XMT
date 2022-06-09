//go:build windows && crypt

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

package winapi

import "github.com/iDigitalFlame/xmt/util/crypt"

var (
	debugPriv = crypt.Get(96) // SeDebugPrivilege

	dllKernel32 = &lazyDLL{name: crypt.Get(97)}  // kernel32.dll
	dllNtdll    = &lazyDLL{name: crypt.Get(98)}  // ntdll.dll
	dllGdi32    = &lazyDLL{name: crypt.Get(99)}  // gdi32.dll
	dllUser32   = &lazyDLL{name: crypt.Get(100)} // user32.dll
	dllWinhttp  = &lazyDLL{name: crypt.Get(101)} // winhttp.dll
	dllDbgHelp  = &lazyDLL{name: crypt.Get(102)} // DbgHelp.dll
	dllAdvapi32 = &lazyDLL{name: crypt.Get(103)} // advapi32.dll
)
