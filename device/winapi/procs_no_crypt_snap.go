//go:build windows && snap && !crypt && !altload
// +build windows,snap,!crypt,!altload

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

var (
	funcThread32Next             = dllKernel32.proc("Thread32Next")
	funcThread32First            = dllKernel32.proc("Thread32First")
	funcProcess32Next            = dllKernel32.proc("Process32NextW")
	funcProcess32First           = dllKernel32.proc("Process32FirstW")
	funcCreateToolhelp32Snapshot = dllKernel32.proc("CreateToolhelp32Snapshot")
)
