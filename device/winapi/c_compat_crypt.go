//go:build windows && (altload || crypt) && !go1.11
// +build windows
// +build altload crypt
// +build !go1.11

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
	funcRtlMoveMemory  = dllNtdll.proc(0xA0CE107B)
	funcNtCancelIoFile = dllNtdll.sysProc(0xF402EB27)

	funcCreateRemoteThread       = dllKernel32.proc(0xEE34539B)
	funcSetProcessWorkingSetSize = dllKernel32.proc(0x28085D42)

	funcRegDeleteKey         = dllAdvapi32.proc(0xE93C7BC8)
	funcCheckTokenMembership = dllAdvapi32.proc(0xE42E234E)

	funcEnumDeviceDrivers       = dllPsapi.proc(0x36EBB2F5)
	funcGetModuleInformation    = dllPsapi.proc(0xC94AC5BB)
	funcGetDeviceDriverFileName = dllPsapi.proc(0x1F449D97)
)
