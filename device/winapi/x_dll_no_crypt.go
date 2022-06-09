//go:build windows && !crypt

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

const debugPriv = "SeDebugPrivilege"

var (
	dllKernel32 = &lazyDLL{name: "kernel32.dll"}
	dllNtdll    = &lazyDLL{name: "ntdll.dll"}
	dllGdi32    = &lazyDLL{name: "gdi32.dll"}
	dllUser32   = &lazyDLL{name: "user32.dll"}
	dllWinhttp  = &lazyDLL{name: "winhttp.dll"}
	dllDbgHelp  = &lazyDLL{name: "DbgHelp.dll"}
	dllAdvapi32 = &lazyDLL{name: "advapi32.dll"}
)
