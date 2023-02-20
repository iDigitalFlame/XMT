//go:build windows && !crypt && !altload && !go1.11
// +build windows,!crypt,!altload,!go1.11

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
	funcRtlMoveMemory  = dllNtdll.proc("RtlMoveMemory")
	funcNtCancelIoFile = dllNtdll.sysProc("NtCancelIoFile")

	funcCreateRemoteThread       = dllKernel32.proc("CreateRemoteThread")
	funcSetProcessWorkingSetSize = dllKernel32.proc("SetProcessWorkingSetSize")

	funcRegDeleteKey         = dllAdvapi32.proc("RegDeleteKeyW")
	funcCheckTokenMembership = dllAdvapi32.proc("CheckTokenMembership")

	funcEnumDeviceDrivers       = dllPsapi.proc("EnumDeviceDrivers")
	funcGetModuleInformation    = dllPsapi.proc("GetModuleInformation")
	funcGetDeviceDriverFileName = dllPsapi.proc("GetDeviceDriverFileNameW")
)
