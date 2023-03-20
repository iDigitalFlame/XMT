//go:build windows && crypt
// +build windows,crypt

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
	dllExt    = crypt.Get(8)[1:] // *.dll
	sysRoot   = crypt.Get(99)   // \SystemRoot
	debugPriv = crypt.Get(100)   // SeDebugPrivilege

	dllAmsi       = &lazyDLL{name: crypt.Get(101)} // amsi.dll
	dllNtdll      = &lazyDLL{name: crypt.Get(102)} // ntdll.dll
	dllGdi32      = &lazyDLL{name: crypt.Get(103)} // gdi32.dll
	dllUser32     = &lazyDLL{name: crypt.Get(104)} // user32.dll
	dllWinhttp    = &lazyDLL{name: crypt.Get(105)} // winhttp.dll
	dllDbgHelp    = &lazyDLL{name: crypt.Get(106)} // DbgHelp.dll
	dllCrypt32    = &lazyDLL{name: crypt.Get(107)} // crypt32.dll
	dllKernel32   = &lazyDLL{name: crypt.Get(108)} // kernel32.dll
	dllAdvapi32   = &lazyDLL{name: crypt.Get(109)} // advapi32.dll
	dllWtsapi32   = &lazyDLL{name: crypt.Get(110)} // wtsapi32.dll
	dllKernelBase = &lazyDLL{name: crypt.Get(111)} // kernelbase.dll
)
