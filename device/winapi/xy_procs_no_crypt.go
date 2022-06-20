//go:build windows && !crypt && !altload

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

var (
	funcLoadLibraryEx       = dllKernel32.proc("LoadLibraryExW")
	funcGetSystemDirectory  = dllKernel32.proc("GetSystemDirectoryW")
	funcOpenProcess         = dllKernel32.proc("OpenProcess")
	funcOpenThread          = dllKernel32.proc("OpenThread")
	funcCloseHandle         = dllKernel32.proc("CloseHandle")
	funcGetCurrentProcessID = dllKernel32.proc("GetCurrentProcessId")

	funcGetDC                                               = dllUser32.proc("GetDC")
	funcBitBlt                                              = dllGdi32.proc("BitBlt")
	funcSetEvent                                            = dllKernel32.proc("SetEvent")
	funcIsZoomed                                            = dllUser32.proc("IsZoomed")
	funcIsIconic                                            = dllUser32.proc("IsIconic")
	funcHeapFree                                            = dllKernel32.proc("HeapFree")
	funcReadFile                                            = dllKernel32.proc("ReadFile")
	funcLsaClose                                            = dllAdvapi32.proc("LsaClose")
	funcDeleteDC                                            = dllGdi32.proc("DeleteDC")
	funcSetFocus                                            = dllUser32.proc("SetFocus")
	funcLogonUser                                           = dllAdvapi32.proc("LogonUserW")
	funcWriteFile                                           = dllKernel32.proc("WriteFile")
	funcOpenMutex                                           = dllKernel32.proc("OpenMutexW")
	funcLocalFree                                           = dllKernel32.proc("LocalFree")
	funcOpenEvent                                           = dllKernel32.proc("OpenEventW")
	funcGetDIBits                                           = dllGdi32.proc("GetDIBits")
	funcReleaseDC                                           = dllUser32.proc("ReleaseDC")
	funcHeapAlloc                                           = dllKernel32.proc("HeapAlloc")
	funcSendInput                                           = dllUser32.proc("SendInput")
	funcHeapCreate                                          = dllKernel32.proc("HeapCreate")
	funcCreateFile                                          = dllKernel32.proc("CreateFileW")
	funcGetVersion                                          = dllKernel32.proc("GetVersion")
	funcCancelIoEx                                          = dllKernel32.proc("CancelIoEx")
	funcBlockInput                                          = dllUser32.proc("BlockInput")
	funcShowWindow                                          = dllUser32.proc("ShowWindow")
	funcMessageBox                                          = dllUser32.proc("MessageBoxW")
	funcFreeLibrary                                         = dllKernel32.proc("FreeLibrary")
	funcHeapDestroy                                         = dllKernel32.proc("HeapDestroy")
	funcLoadLibrary                                         = dllKernel32.proc("LoadLibraryW")
	funcCreateMutex                                         = dllKernel32.proc("CreateMutexW")
	funcCreateEvent                                         = dllKernel32.proc("CreateEventW")
	funcHeapReAlloc                                         = dllKernel32.proc("HeapReAlloc")
	funcEnumWindows                                         = dllUser32.proc("EnumWindows")
	funcEnableWindow                                        = dllUser32.proc("EnableWindow")
	funcSelectObject                                        = dllGdi32.proc("SelectObject")
	funcDeleteObject                                        = dllGdi32.proc("DeleteObject")
	funcNtTraceEvent                                        = dllNtdll.proc("NtTraceEvent")
	funcResumeThread                                        = dllKernel32.proc("ResumeThread")
	funcThread32Next                                        = dllKernel32.proc("Thread32Next")
	funcGetProcessID                                        = dllKernel32.proc("GetProcessId")
	funcRevertToSelf                                        = dllAdvapi32.proc("RevertToSelf")
	funcRegEnumValue                                        = dllAdvapi32.proc("RegEnumValueW")
	funcModule32Next                                        = dllKernel32.proc("Module32NextW")
	funcSetWindowPos                                        = dllUser32.proc("SetWindowPos")
	funcGetWindowText                                       = dllUser32.proc("GetWindowTextW")
	funcModule32First                                       = dllKernel32.proc("Module32FirstW")
	funcWaitNamedPipe                                       = dllKernel32.proc("WaitNamedPipeW")
	funcCreateProcess                                       = dllKernel32.proc("CreateProcessW")
	funcSuspendThread                                       = dllKernel32.proc("SuspendThread")
	funcProcess32Next                                       = dllKernel32.proc("Process32NextW")
	funcRegSetValueEx                                       = dllAdvapi32.proc("RegSetValueExW")
	funcThread32First                                       = dllKernel32.proc("Thread32First")
	funcLsaOpenPolicy                                       = dllAdvapi32.proc("LsaOpenPolicy")
	funcOpenSemaphore                                       = dllKernel32.proc("OpenSemaphoreW")
	funcRegDeleteTree                                       = dllAdvapi32.proc("RegDeleteTreeW")
	funcRtlCopyMemory                                       = dllNtdll.proc("RtlCopyMemory")
	funcGetWindowInfo                                       = dllUser32.proc("GetWindowInfo")
	funcDbgBreakPoint                                       = dllNtdll.proc("DbgBreakPoint")
	funcRegDeleteKeyEx                                      = dllAdvapi32.proc("RegDeleteKeyExW")
	funcGetMonitorInfo                                      = dllUser32.proc("GetMonitorInfoW")
	funcIsWellKnownSID                                      = dllAdvapi32.proc("IsWellKnownSid")
	funcProcess32First                                      = dllKernel32.proc("Process32FirstW")
	funcCreateMailslot                                      = dllKernel32.proc("CreateMailslotW")
	funcRegCreateKeyEx                                      = dllAdvapi32.proc("RegCreateKeyExW")
	funcSetThreadToken                                      = dllAdvapi32.proc("SetThreadToken")
	funcRegDeleteValue                                      = dllAdvapi32.proc("RegDeleteValueW")
	funcCreateNamedPipe                                     = dllKernel32.proc("CreateNamedPipeW")
	funcDuplicateHandle                                     = dllKernel32.proc("DuplicateHandle")
	funcCreateSemaphore                                     = dllKernel32.proc("CreateSemaphoreW")
	funcTerminateThread                                     = dllKernel32.proc("TerminateThread")
	funcOpenThreadToken                                     = dllAdvapi32.proc("OpenThreadToken")
	funcNtResumeProcess                                     = dllNtdll.proc("NtResumeProcess")
	funcIsWindowVisible                                     = dllUser32.proc("IsWindowVisible")
	funcLookupAccountSid                                    = dllAdvapi32.proc("LookupAccountSidW")
	funcSetServiceStatus                                    = dllAdvapi32.proc("SetServiceStatus")
	funcConnectNamedPipe                                    = dllKernel32.proc("ConnectNamedPipe")
	funcTerminateProcess                                    = dllKernel32.proc("TerminateProcess")
	funcDuplicateTokenEx                                    = dllAdvapi32.proc("DuplicateTokenEx")
	funcNtSuspendProcess                                    = dllNtdll.proc("NtSuspendProcess")
	funcNtCreateThreadEx                                    = dllNtdll.proc("NtCreateThreadEx")
	funcGetLogicalDrives                                    = dllKernel32.proc("GetLogicalDrives")
	funcOpenProcessToken                                    = dllAdvapi32.proc("OpenProcessToken")
	funcGetDesktopWindow                                    = dllUser32.proc("GetDesktopWindow")
	funcGetWindowLongPtr                                    = dllUser32.proc("GetWindowLongPtrW")
	funcSetWindowLongPtr                                    = dllUser32.proc("SetWindowLongPtrW")
	funcSendNotifyMessage                                   = dllUser32.proc("SendNotifyMessageW")
	funcGetModuleHandleEx                                   = dllKernel32.proc("GetModuleHandleExW")
	funcIsDebuggerPresent                                   = dllKernel32.proc("IsDebuggerPresent")
	funcMiniDumpWriteDump                                   = dllDbgHelp.proc("MiniDumpWriteDump")
	funcGetExitCodeThread                                   = dllKernel32.proc("GetExitCodeThread")
	funcGetExitCodeProcess                                  = dllKernel32.proc("GetExitCodeProcess")
	funcCreateCompatibleDC                                  = dllGdi32.proc("CreateCompatibleDC")
	funcGetCurrentThreadID                                  = dllKernel32.proc("GetCurrentThreadId")
	funcCreateWellKnownSid                                  = dllAdvapi32.proc("CreateWellKnownSid")
	funcEnumDisplayMonitors                                 = dllUser32.proc("EnumDisplayMonitors")
	funcEnumDisplaySettings                                 = dllUser32.proc("EnumDisplaySettingsW")
	funcSetTokenInformation                                 = dllAdvapi32.proc("SetTokenInformation")
	funcGetTokenInformation                                 = dllAdvapi32.proc("GetTokenInformation")
	funcGetOverlappedResult                                 = dllKernel32.proc("GetOverlappedResult")
	funcNtFreeVirtualMemory                                 = dllNtdll.proc("NtFreeVirtualMemory")
	funcWaitForSingleObject                                 = dllKernel32.proc("WaitForSingleObject")
	funcDisconnectNamedPipe                                 = dllKernel32.proc("DisconnectNamedPipe")
	funcGetWindowTextLength                                 = dllUser32.proc("GetWindowTextLengthW")
	funcSetForegroundWindow                                 = dllUser32.proc("SetForegroundWindow")
	funcNtWriteVirtualMemory                                = dllNtdll.proc("NtWriteVirtualMemory")
	funcLookupPrivilegeValue                                = dllAdvapi32.proc("LookupPrivilegeValueW")
	funcSystemParametersInfo                                = dllUser32.proc("SystemParametersInfoW")
	funcConvertSIDToStringSID                               = dllAdvapi32.proc("ConvertSidToStringSidW")
	funcAdjustTokenPrivileges                               = dllAdvapi32.proc("AdjustTokenPrivileges")
	funcCreateCompatibleBitmap                              = dllGdi32.proc("CreateCompatibleBitmap")
	funcWaitForMultipleObjects                              = dllKernel32.proc("WaitForMultipleObjects")
	funcNtProtectVirtualMemory                              = dllNtdll.proc("NtProtectVirtualMemory")
	funcCreateProcessWithToken                              = dllAdvapi32.proc("CreateProcessWithTokenW")
	funcCreateProcessWithLogon                              = dllAdvapi32.proc("CreateProcessWithLogonW")
	funcImpersonateLoggedOnUser                             = dllAdvapi32.proc("ImpersonateLoggedOnUser")
	funcNtAllocateVirtualMemory                             = dllNtdll.proc("NtAllocateVirtualMemory")
	funcRtlSetProcessIsCritical                             = dllNtdll.proc("RtlSetProcessIsCritical")
	funcNtQueryInformationThread                            = dllNtdll.proc("NtQueryInformationThread")
	funcCreateToolhelp32Snapshot                            = dllKernel32.proc("CreateToolhelp32Snapshot")
	funcUpdateProcThreadAttribute                           = dllKernel32.proc("UpdateProcThreadAttribute")
	funcNtQueryInformationProcess                           = dllNtdll.proc("NtQueryInformationProcess")
	funcLsaQueryInformationPolicy                           = dllAdvapi32.proc("LsaQueryInformationPolicy")
	funcSetLayeredWindowAttributes                          = dllUser32.proc("SetLayeredWindowAttributes")
	funcSetProcessWorkingSetSizeEx                          = dllKernel32.proc("SetProcessWorkingSetSizeEx")
	funcStartServiceCtrlDispatcher                          = dllAdvapi32.proc("StartServiceCtrlDispatcherW")
	funcImpersonateNamedPipeClient                          = dllAdvapi32.proc("ImpersonateNamedPipeClient")
	funcCheckRemoteDebuggerPresent                          = dllKernel32.proc("CheckRemoteDebuggerPresent")
	funcGetSecurityDescriptorLength                         = dllAdvapi32.proc("GetSecurityDescriptorLength")
	funcRegisterServiceCtrlHandlerEx                        = dllAdvapi32.proc("RegisterServiceCtrlHandlerExW")
	funcDeleteProcThreadAttributeList                       = dllKernel32.proc("DeleteProcThreadAttributeList")
	funcQueryServiceDynamicInformation                      = dllAdvapi32.proc("QueryServiceDynamicInformation")
	funcInitializeProcThreadAttributeList                   = dllKernel32.proc("InitializeProcThreadAttributeList")
	funcWinHTTPGetDefaultProxyConfiguration                 = dllWinhttp.proc("WinHttpGetDefaultProxyConfiguration")
	funcConvertStringSecurityDescriptorToSecurityDescriptor = dllAdvapi32.proc("ConvertStringSecurityDescriptorToSecurityDescriptorW")
)

func doSearchSystem32() bool {
	searchSystem32.Do(func() {
		searchSystem32.v = (dllKernel32.proc("AddDllDirectory").find() == nil)
	})
	return searchSystem32.v
}
