//go:build windows && snap && (altload || crypt)
// +build windows
// +build snap
// +build altload crypt

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

// All hashes are in the FNV format.
/*
def fnv(v):
    h = 2166136261
    for n in v:
        h *= 16777619
        h ^= ord(n)
        h = h&0xFFFFFFFF
    return "0x" + hex(h).upper()[2:]
*/

var (
	funcThread32Next             = dllKernel32.proc(0x9B4B1895)
	funcThread32First            = dllKernel32.proc(0xC5311BC8)
	funcProcess32Next            = dllKernel32.proc(0x80132847)
	funcProcess32First           = dllKernel32.proc(0xD4C414BE)
	funcCreateToolhelp32Snapshot = dllKernel32.proc(0xBAA64095)
)
