//go:build windows && !crypt && !altload
// +build windows,!crypt,!altload

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
	funcLoadLibraryEx = dllKernelBase.proc("LoadLibraryExW")
	funcFormatMessage = dllKernelBase.proc("FormatMessageW")

	funcNtClose                     = dllNtdll.sysProc("NtClose")
	funcNtSetEvent                  = dllNtdll.sysProc("NtSetEvent")
	funcRtlFreeHeap                 = dllNtdll.proc("RtlFreeHeap")
	funcNtTraceEvent                = dllNtdll.sysProc("NtTraceEvent")
	funcNtOpenThread                = dllNtdll.sysProc("NtOpenThread")
	funcRtlCreateHeap               = dllNtdll.proc("RtlCreateHeap")
	funcEtwEventWrite               = dllNtdll.proc("EtwEventWrite") // >= WinVista
	funcDbgBreakPoint               = dllNtdll.proc("DbgBreakPoint")
	funcNtOpenProcess               = dllNtdll.sysProc("NtOpenProcess")
	funcRtlDestroyHeap              = dllNtdll.proc("RtlDestroyHeap")
	funcNtResumeThread              = dllNtdll.sysProc("NtResumeThread")
	funcNtCreateSection             = dllNtdll.sysProc("NtCreateSection")
	funcNtSuspendThread             = dllNtdll.sysProc("NtSuspendThread")
	funcNtResumeProcess             = dllNtdll.sysProc("NtResumeProcess")
	funcRtlAllocateHeap             = dllNtdll.proc("RtlAllocateHeap")
	funcNtDuplicateToken            = dllNtdll.sysProc("NtDuplicateToken")
	funcEtwEventRegister            = dllNtdll.proc("EtwEventRegister") // >= WinVista
	funcNtSuspendProcess            = dllNtdll.sysProc("NtSuspendProcess")
	funcNtCreateThreadEx            = dllNtdll.sysProc("NtCreateThreadEx") // >= WinVista (Xp sub = RtlCreateUserThread)
	funcNtCancelIoFileEx            = dllNtdll.sysProc("NtCancelIoFileEx") // >= WinVista (Xp sub = NtCancelIoFile)
	funcNtDuplicateObject           = dllNtdll.sysProc("NtDuplicateObject")
	funcNtTerminateThread           = dllNtdll.sysProc("NtTerminateThread")
	funcNtOpenThreadToken           = dllNtdll.sysProc("NtOpenThreadToken")
	funcEtwEventWriteFull           = dllNtdll.proc("EtwEventWriteFull") // >= WinVista
	funcRtlReAllocateHeap           = dllNtdll.proc("RtlReAllocateHeap")
	funcNtMapViewOfSection          = dllNtdll.sysProc("NtMapViewOfSection")
	funcNtTerminateProcess          = dllNtdll.sysProc("NtTerminateProcess")
	funcNtOpenProcessToken          = dllNtdll.sysProc("NtOpenProcessToken")
	funcRtlCopyMappedMemory         = dllNtdll.proc("RtlCopyMappedMemory") // >= WinS2003 (Not in XP sub = RtlMoveMemory)
	funcNtFreeVirtualMemory         = dllNtdll.sysProc("NtFreeVirtualMemory")
	funcNtImpersonateThread         = dllNtdll.sysProc("NtImpersonateThread")
	funcNtUnmapViewOfSection        = dllNtdll.sysProc("NtUnmapViewOfSection")
	funcNtWriteVirtualMemory        = dllNtdll.sysProc("NtWriteVirtualMemory")
	funcNtDeviceIoControlFile       = dllNtdll.sysProc("NtDeviceIoControlFile")
	funcNtWaitForSingleObject       = dllNtdll.sysProc("NtWaitForSingleObject")
	funcNtSetInformationToken       = dllNtdll.sysProc("NtSetInformationToken")
	funcNtProtectVirtualMemory      = dllNtdll.sysProc("NtProtectVirtualMemory")
	funcNtSetInformationThread      = dllNtdll.sysProc("NtSetInformationThread")
	funcRtlGetNtVersionNumbers      = dllNtdll.proc("RtlGetNtVersionNumbers")
	funcEtwNotificationRegister     = dllNtdll.proc("EtwNotificationRegister") // >= WinVista
	funcNtAllocateVirtualMemory     = dllNtdll.sysProc("NtAllocateVirtualMemory")
	funcRtlSetProcessIsCritical     = dllNtdll.proc("RtlSetProcessIsCritical")
	funcNtFlushInstructionCache     = dllNtdll.sysProc("NtFlushInstructionCache")
	funcNtAdjustTokenPrivileges     = dllNtdll.sysProc("NtAdjustPrivilegesToken")
	funcNtQueryInformationToken     = dllNtdll.sysProc("NtQueryInformationToken")
	funcNtQueryInformationThread    = dllNtdll.sysProc("NtQueryInformationThread")
	funcNtQuerySystemInformation    = dllNtdll.sysProc("NtQuerySystemInformation")
	funcNtWaitForMultipleObjects    = dllNtdll.sysProc("NtWaitForMultipleObjects")
	funcNtQueryInformationProcess   = dllNtdll.sysProc("NtQueryInformationProcess")
	funcRtlWow64GetProcessMachines  = dllNtdll.proc("RtlWow64GetProcessMachines") // == 64bit/ARM64
	funcRtlLengthSecurityDescriptor = dllNtdll.proc("RtlLengthSecurityDescriptor")

	funcReadFile                  = dllKernelBase.proc("ReadFile")
	funcWriteFile                 = dllKernelBase.proc("WriteFile")
	funcOpenMutex                 = dllKernelBase.proc("OpenMutexW")
	funcLocalFree                 = dllKernelBase.proc("LocalFree")
	funcOpenEvent                 = dllKernelBase.proc("OpenEventW")
	funcCreateFile                = dllKernelBase.proc("CreateFileW")
	funcDebugBreak                = dllKernelBase.proc("DebugBreak")
	funcCreateMutex               = dllKernelBase.proc("CreateMutexW")
	funcCreateEvent               = dllKernelBase.proc("CreateEventW")
	funcWaitNamedPipe             = dllKernelBase.proc("WaitNamedPipeW")
	funcOpenSemaphore             = dllKernelBase.proc("OpenSemaphoreW")
	funcCreateNamedPipe           = dllKernelBase.proc("CreateNamedPipeW")
	funcConnectNamedPipe          = dllKernelBase.proc("ConnectNamedPipe")
	funcGetModuleHandleEx         = dllKernelBase.proc("GetModuleHandleExW")
	funcOutputDebugString         = dllKernelBase.proc("OutputDebugStringA")
	funcGetCurrentThreadID        = dllKernelBase.proc("GetCurrentThreadId")
	funcGetOverlappedResult       = dllKernelBase.proc("GetOverlappedResult")
	funcDisconnectNamedPipe       = dllKernelBase.proc("DisconnectNamedPipe")
	funcGetCurrentProcessID       = dllKernelBase.proc("GetCurrentProcessId")
	funcUpdateProcThreadAttribute = dllKernelBase.proc("UpdateProcThreadAttribute") // >= WinVista

	funcIsWellKnownSID             = dllKernelOrAdvapi.proc("IsWellKnownSid")             // >= Win7 kernelbase.dll else advapi32.dll
	funcCreateWellKnownSid         = dllKernelOrAdvapi.proc("CreateWellKnownSid")         // >= Win7 kernelbase.dll else advapi32.dll
	funcImpersonateNamedPipeClient = dllKernelOrAdvapi.proc("ImpersonateNamedPipeClient") // >= Win7 kernelbase.dll else advapi32.dll

	funcCreateProcess              = dllKernel32.proc("CreateProcessW")
	funcCreateMailslot             = dllKernel32.proc("CreateMailslotW")
	funcCreateSemaphore            = dllKernel32.proc("CreateSemaphoreW")
	funcK32EnumDeviceDrivers       = dllKernel32.proc("K32EnumDeviceDrivers")        // >= Win7 (Xp sub = psapi.EnumDeviceDrivers)
	funcK32GetModuleInformation    = dllKernel32.proc("K32GetModuleInformation")     // >= Win7 (Xp sub = psapi.GetModuleInformation)
	funcSetProcessWorkingSetSizeEx = dllKernel32.proc("SetProcessWorkingSetSizeEx")  // >= WinS2003 (Not in XP sub = SetProcessWorkingSetSize)
	funcK32GetDeviceDriverFileName = dllKernel32.proc("K32GetDeviceDriverFileNameW") // >= Win7 (Xp sub = psapi.GetDeviceDriverFileNameW)

	funcLsaClose                                            = dllAdvapi32.proc("LsaClose")
	funcLogonUser                                           = dllAdvapi32.proc("LogonUserW")
	funcRegFlushKey                                         = dllAdvapi32.proc("RegFlushKey")
	funcRegEnumValue                                        = dllAdvapi32.proc("RegEnumValueW")
	funcRegSetValueEx                                       = dllAdvapi32.proc("RegSetValueExW")
	funcLsaOpenPolicy                                       = dllAdvapi32.proc("LsaOpenPolicy")
	funcRegDeleteTree                                       = dllAdvapi32.proc("RegDeleteTreeW")  // >= WinVista
	funcRegDeleteKeyEx                                      = dllAdvapi32.proc("RegDeleteKeyExW") // >= WinVista (Xp sub = RegDeleteKey)
	funcRegDeleteValue                                      = dllAdvapi32.proc("RegDeleteValueW")
	funcRegCreateKeyEx                                      = dllAdvapi32.proc("RegCreateKeyExW")
	funcSetServiceStatus                                    = dllAdvapi32.proc("SetServiceStatus")
	funcLookupAccountSid                                    = dllAdvapi32.proc("LookupAccountSidW")
	funcLookupPrivilegeValue                                = dllAdvapi32.proc("LookupPrivilegeValueW")
	funcConvertSIDToStringSID                               = dllAdvapi32.proc("ConvertSidToStringSidW")
	funcCreateProcessWithToken                              = dllAdvapi32.proc("CreateProcessWithTokenW") // >= WinS2003 (Not in XP)
	funcCreateProcessWithLogon                              = dllAdvapi32.proc("CreateProcessWithLogonW")
	funcInitiateSystemShutdownEx                            = dllAdvapi32.proc("InitiateSystemShutdownExW")
	funcLsaQueryInformationPolicy                           = dllAdvapi32.proc("LsaQueryInformationPolicy")
	funcStartServiceCtrlDispatcher                          = dllAdvapi32.proc("StartServiceCtrlDispatcherW")
	funcRegisterServiceCtrlHandlerEx                        = dllAdvapi32.proc("RegisterServiceCtrlHandlerExW")
	funcQueryServiceDynamicInformation                      = dllAdvapi32.proc("QueryServiceDynamicInformation") // >= Win8
	funcConvertStringSecurityDescriptorToSecurityDescriptor = dllAdvapi32.proc("ConvertStringSecurityDescriptorToSecurityDescriptorW")

	funcGetDC                      = dllUser32.proc("GetDC")
	funcSetFocus                   = dllUser32.proc("SetFocus")
	funcReleaseDC                  = dllUser32.proc("ReleaseDC")
	funcSendInput                  = dllUser32.proc("SendInput")
	funcBlockInput                 = dllUser32.proc("BlockInput")
	funcShowWindow                 = dllUser32.proc("ShowWindow")
	funcMessageBox                 = dllUser32.proc("MessageBoxW")
	funcEnumWindows                = dllUser32.proc("EnumWindows")
	funcEnableWindow               = dllUser32.proc("EnableWindow")
	funcSetWindowPos               = dllUser32.proc("SetWindowPos")
	funcGetWindowText              = dllUser32.proc("GetWindowTextW")
	funcGetWindowInfo              = dllUser32.proc("GetWindowInfo")
	funcGetMonitorInfo             = dllUser32.proc("GetMonitorInfoW")
	funcGetWindowLongW             = dllUser32.proc("GetWindowLongW")
	funcSetWindowLongW             = dllUser32.proc("SetWindowLongW")
	funcGetDesktopWindow           = dllUser32.proc("GetDesktopWindow")
	funcSendNotifyMessage          = dllUser32.proc("SendNotifyMessageW")
	funcEnumDisplayMonitors        = dllUser32.proc("EnumDisplayMonitors")
	funcEnumDisplaySettings        = dllUser32.proc("EnumDisplaySettingsW")
	funcGetWindowTextLength        = dllUser32.proc("GetWindowTextLengthW")
	funcSetForegroundWindow        = dllUser32.proc("SetForegroundWindow")
	funcSystemParametersInfo       = dllUser32.proc("SystemParametersInfoW")
	funcSetLayeredWindowAttributes = dllUser32.proc("SetLayeredWindowAttributes")

	funcCryptMsgClose              = dllCrypt32.proc("CryptMsgClose")
	funcCertCloseStore             = dllCrypt32.proc("CertCloseStore")
	funcCryptQueryObject           = dllCrypt32.proc("CryptQueryObject")
	funcCryptMsgGetParam           = dllCrypt32.proc("CryptMsgGetParam")
	funcCertGetNameString          = dllCrypt32.proc("CertGetNameStringW")
	funcCertFindCertificateInStore = dllCrypt32.proc("CertFindCertificateInStore")
	funcCertFreeCertificateContext = dllCrypt32.proc("CertFreeCertificateContext")

	funcBitBlt                 = dllGdi32.proc("BitBlt")
	funcDeleteDC               = dllGdi32.proc("DeleteDC")
	funcGetDIBits              = dllGdi32.proc("GetDIBits")
	funcSelectObject           = dllGdi32.proc("SelectObject")
	funcDeleteObject           = dllGdi32.proc("DeleteObject")
	funcCreateCompatibleDC     = dllGdi32.proc("CreateCompatibleDC")
	funcCreateCompatibleBitmap = dllGdi32.proc("CreateCompatibleBitmap")

	funcWTSFreeMemory              = dllWtsapi32.proc("WTSFreeMemory")
	funcWTSOpenServer              = dllWtsapi32.proc("WTSOpenServerW")
	funcWTSCloseServer             = dllWtsapi32.proc("WTSCloseServer")
	funcWTSSendMessage             = dllWtsapi32.proc("WTSSendMessageW")
	funcWTSLogoffSession           = dllWtsapi32.proc("WTSLogoffSession")
	funcWTSEnumerateSessions       = dllWtsapi32.proc("WTSEnumerateSessionsW")
	funcWTSDisconnectSession       = dllWtsapi32.proc("WTSDisconnectSession")
	funcWTSEnumerateProcesses      = dllWtsapi32.proc("WTSEnumerateProcessesW")
	funcWTSQuerySessionInformation = dllWtsapi32.proc("WTSQuerySessionInformationW")

	funcMiniDumpWriteDump = dllDbgHelp.proc("MiniDumpWriteDump")

	funcWinHTTPGetDefaultProxyConfiguration = dllWinhttp.proc("WinHttpGetDefaultProxyConfiguration") // >= WinXP_SP3

	funcAmsiScanBuffer = dllAmsi.proc("AmsiScanBuffer")
	funcAmsiInitialize = dllAmsi.proc("AmsiInitialize")
	funcAmsiScanString = dllAmsi.proc("AmsiScanString")
)

func doSearchSystem32() bool {
	searchSystem32.Do(func() {
		searchSystem32.v = funcAddDllDirectory > 0 // >= Win8 / ~Win7
	})
	return searchSystem32.v
}
