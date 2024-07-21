//go:build windows && !crypt
// +build windows,!crypt

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

const (
	dllExt    = ".dll"
	sysRoot   = `\SystemRoot`
	debugPriv = "SeDebugPrivilege"
)

var (
	dllAmsi       = &LazyDLL{name: "amsi.dll"}
	dllNtdll      = &LazyDLL{name: "ntdll.dll"}
	dllGdi32      = &LazyDLL{name: "gdi32.dll"}
	dllUser32     = &LazyDLL{name: "user32.dll"}
	dllWinhttp    = &LazyDLL{name: "winhttp.dll"}
	dllDbgHelp    = &LazyDLL{name: "DbgHelp.dll"}
	dllCrypt32    = &LazyDLL{name: "crypt32.dll"}
	dllKernel32   = &LazyDLL{name: "kernel32.dll"}
	dllAdvapi32   = &LazyDLL{name: "advapi32.dll"}
	dllWtsapi32   = &LazyDLL{name: "wtsapi32.dll"}
	dllKernelBase = &LazyDLL{name: "kernelbase.dll"}
)
