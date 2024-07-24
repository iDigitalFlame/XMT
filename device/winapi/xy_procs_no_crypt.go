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
	funcLoadLibraryEx = dllKernelBase.Proc("LoadLibraryExW")
	funcFormatMessage = dllKernelBase.Proc("FormatMessageW")

	funcNtClose                       = dllNtdll.sysProc("NtClose")
	funcNtSetEvent                    = dllNtdll.sysProc("NtSetEvent")
	funcRtlFreeHeap                   = dllNtdll.Proc("RtlFreeHeap")
	funcNtTraceEvent                  = dllNtdll.sysProc("NtTraceEvent")
	funcNtOpenThread                  = dllNtdll.sysProc("NtOpenThread")
	funcRtlCreateHeap                 = dllNtdll.Proc("RtlCreateHeap")
	funcEtwEventWrite                 = dllNtdll.Proc("EtwEventWrite") // >= WinVista
	funcDbgBreakPoint                 = dllNtdll.Proc("DbgBreakPoint")
	funcNtOpenProcess                 = dllNtdll.sysProc("NtOpenProcess")
	funcRtlDestroyHeap                = dllNtdll.Proc("RtlDestroyHeap")
	funcNtResumeThread                = dllNtdll.sysProc("NtResumeThread")
	funcNtCreateSection               = dllNtdll.sysProc("NtCreateSection")
	funcNtSuspendThread               = dllNtdll.sysProc("NtSuspendThread")
	funcNtResumeProcess               = dllNtdll.sysProc("NtResumeProcess")
	funcRtlAllocateHeap               = dllNtdll.Proc("RtlAllocateHeap")
	funcNtDuplicateToken              = dllNtdll.sysProc("NtDuplicateToken")
	funcEtwEventRegister              = dllNtdll.Proc("EtwEventRegister") // >= WinVista
	funcNtSuspendProcess              = dllNtdll.sysProc("NtSuspendProcess")
	funcNtCreateThreadEx              = dllNtdll.sysProc("NtCreateThreadEx") // >= WinVista (Xp sub = RtlCreateUserThread)
	funcNtCancelIoFileEx              = dllNtdll.sysProc("NtCancelIoFileEx") // >= WinVista (Xp sub = NtCancelIoFile)
	funcNtDuplicateObject             = dllNtdll.sysProc("NtDuplicateObject")
	funcNtTerminateThread             = dllNtdll.sysProc("NtTerminateThread")
	funcNtOpenThreadToken             = dllNtdll.sysProc("NtOpenThreadToken")
	funcEtwEventWriteFull             = dllNtdll.Proc("EtwEventWriteFull") // >= WinVista
	funcRtlReAllocateHeap             = dllNtdll.Proc("RtlReAllocateHeap")
	funcNtMapViewOfSection            = dllNtdll.sysProc("NtMapViewOfSection")
	funcNtTerminateProcess            = dllNtdll.sysProc("NtTerminateProcess")
	funcNtOpenProcessToken            = dllNtdll.sysProc("NtOpenProcessToken")
	funcRtlCopyMappedMemory           = dllNtdll.Proc("RtlCopyMappedMemory") // >= WinS2003 (Not in XP sub = RtlMoveMemory)
	funcNtFreeVirtualMemory           = dllNtdll.sysProc("NtFreeVirtualMemory")
	funcNtImpersonateThread           = dllNtdll.sysProc("NtImpersonateThread")
	funcNtUnmapViewOfSection          = dllNtdll.sysProc("NtUnmapViewOfSection")
	funcNtWriteVirtualMemory          = dllNtdll.sysProc("NtWriteVirtualMemory")
	funcNtDeviceIoControlFile         = dllNtdll.sysProc("NtDeviceIoControlFile")
	funcNtWaitForSingleObject         = dllNtdll.sysProc("NtWaitForSingleObject")
	funcNtSetInformationToken         = dllNtdll.sysProc("NtSetInformationToken")
	funcNtProtectVirtualMemory        = dllNtdll.sysProc("NtProtectVirtualMemory")
	funcNtSetInformationThread        = dllNtdll.sysProc("NtSetInformationThread")
	funcRtlGetNtVersionNumbers        = dllNtdll.Proc("RtlGetNtVersionNumbers")
	funcEtwNotificationRegister       = dllNtdll.Proc("EtwNotificationRegister") // >= WinVista
	funcNtAllocateVirtualMemory       = dllNtdll.sysProc("NtAllocateVirtualMemory")
	funcRtlSetProcessIsCritical       = dllNtdll.Proc("RtlSetProcessIsCritical")
	funcNtFlushInstructionCache       = dllNtdll.sysProc("NtFlushInstructionCache")
	funcNtAdjustTokenPrivileges       = dllNtdll.sysProc("NtAdjustPrivilegesToken")
	funcNtQueryInformationToken       = dllNtdll.sysProc("NtQueryInformationToken")
	funcNtQueryInformationThread      = dllNtdll.sysProc("NtQueryInformationThread")
	funcNtQuerySystemInformation      = dllNtdll.sysProc("NtQuerySystemInformation")
	funcNtWaitForMultipleObjects      = dllNtdll.sysProc("NtWaitForMultipleObjects")
	funcNtQueryInformationProcess     = dllNtdll.sysProc("NtQueryInformationProcess")
	funcRtlWow64GetProcessMachines    = dllNtdll.Proc("RtlWow64GetProcessMachines") // == 64bit/ARM64
	funcRtlLengthSecurityDescriptor   = dllNtdll.Proc("RtlLengthSecurityDescriptor")
	funcRtlGetDaclSecurityDescriptor  = dllNtdll.Proc("RtlGetDaclSecurityDescriptor")
	funcRtlGetSaclSecurityDescriptor  = dllNtdll.Proc("RtlGetSaclSecurityDescriptor")
	funcRtlGetGroupSecurityDescriptor = dllNtdll.Proc("RtlGetGroupSecurityDescriptor")
	funcRtlGetOwnerSecurityDescriptor = dllNtdll.Proc("RtlGetOwnerSecurityDescriptor")

	funcReadFile                  = dllKernelBase.Proc("ReadFile")
	funcWriteFile                 = dllKernelBase.Proc("WriteFile")
	funcOpenMutex                 = dllKernelBase.Proc("OpenMutexW")
	funcLocalFree                 = dllKernelBase.Proc("LocalFree")
	funcOpenEvent                 = dllKernelBase.Proc("OpenEventW")
	funcCreateFile                = dllKernelBase.Proc("CreateFileW")
	funcDebugBreak                = dllKernelBase.Proc("DebugBreak")
	funcCreateMutex               = dllKernelBase.Proc("CreateMutexW")
	funcCreateEvent               = dllKernelBase.Proc("CreateEventW")
	funcWaitNamedPipe             = dllKernelBase.Proc("WaitNamedPipeW")
	funcOpenSemaphore             = dllKernelBase.Proc("OpenSemaphoreW")
	funcCreateNamedPipe           = dllKernelBase.Proc("CreateNamedPipeW")
	funcConnectNamedPipe          = dllKernelBase.Proc("ConnectNamedPipe")
	funcGetModuleHandleEx         = dllKernelBase.Proc("GetModuleHandleExW")
	funcOutputDebugString         = dllKernelBase.Proc("OutputDebugStringA")
	funcGetCurrentThreadID        = dllKernelBase.Proc("GetCurrentThreadId")
	funcGetOverlappedResult       = dllKernelBase.Proc("GetOverlappedResult")
	funcDisconnectNamedPipe       = dllKernelBase.Proc("DisconnectNamedPipe")
	funcGetCurrentProcessID       = dllKernelBase.Proc("GetCurrentProcessId")
	funcUpdateProcThreadAttribute = dllKernelBase.Proc("UpdateProcThreadAttribute") // >= WinVista

	funcIsWellKnownSID             = dllKernelOrAdvapi.Proc("IsWellKnownSid")             // >= Win7 kernelbase.dll else advapi32.dll
	funcCreateWellKnownSid         = dllKernelOrAdvapi.Proc("CreateWellKnownSid")         // >= Win7 kernelbase.dll else advapi32.dll
	funcImpersonateNamedPipeClient = dllKernelOrAdvapi.Proc("ImpersonateNamedPipeClient") // >= Win7 kernelbase.dll else advapi32.dll

	funcCreateProcess              = dllKernel32.Proc("CreateProcessW")
	funcCreateMailslot             = dllKernel32.Proc("CreateMailslotW")
	funcCreateSemaphore            = dllKernel32.Proc("CreateSemaphoreW")
	funcK32EnumDeviceDrivers       = dllKernel32.Proc("K32EnumDeviceDrivers")        // >= Win7 (Xp sub = psapi.EnumDeviceDrivers)
	funcK32GetModuleInformation    = dllKernel32.Proc("K32GetModuleInformation")     // >= Win7 (Xp sub = psapi.GetModuleInformation)
	funcSetProcessWorkingSetSizeEx = dllKernel32.Proc("SetProcessWorkingSetSizeEx")  // >= WinS2003 (Not in XP sub = SetProcessWorkingSetSize)
	funcK32GetDeviceDriverFileName = dllKernel32.Proc("K32GetDeviceDriverFileNameW") // >= Win7 (Xp sub = psapi.GetDeviceDriverFileNameW)

	funcLsaClose                                            = dllAdvapi32.Proc("LsaClose")
	funcLogonUser                                           = dllAdvapi32.Proc("LogonUserW")
	funcRegFlushKey                                         = dllAdvapi32.Proc("RegFlushKey")
	funcRegEnumValue                                        = dllAdvapi32.Proc("RegEnumValueW")
	funcRegSetValueEx                                       = dllAdvapi32.Proc("RegSetValueExW")
	funcLsaOpenPolicy                                       = dllAdvapi32.Proc("LsaOpenPolicy")
	funcRegDeleteTree                                       = dllAdvapi32.Proc("RegDeleteTreeW")  // >= WinVista
	funcRegDeleteKeyEx                                      = dllAdvapi32.Proc("RegDeleteKeyExW") // >= WinVista (Xp sub = RegDeleteKey)
	funcRegDeleteValue                                      = dllAdvapi32.Proc("RegDeleteValueW")
	funcRegCreateKeyEx                                      = dllAdvapi32.Proc("RegCreateKeyExW")
	funcSetServiceStatus                                    = dllAdvapi32.Proc("SetServiceStatus")
	funcLookupAccountSid                                    = dllAdvapi32.Proc("LookupAccountSidW")
	funcGetNamedSecurityInfo                                = dllAdvapi32.Proc("GetNamedSecurityInfoW")
	funcSetNamedSecurityInfo                                = dllAdvapi32.Proc("SetNamedSecurityInfoW")
	funcLookupPrivilegeValue                                = dllAdvapi32.Proc("LookupPrivilegeValueW")
	funcConvertSIDToStringSID                               = dllAdvapi32.Proc("ConvertSidToStringSidW")
	funcCreateProcessWithToken                              = dllAdvapi32.Proc("CreateProcessWithTokenW") // >= WinS2003 (Not in XP)
	funcCreateProcessWithLogon                              = dllAdvapi32.Proc("CreateProcessWithLogonW")
	funcInitiateSystemShutdownEx                            = dllAdvapi32.Proc("InitiateSystemShutdownExW")
	funcLsaQueryInformationPolicy                           = dllAdvapi32.Proc("LsaQueryInformationPolicy")
	funcStartServiceCtrlDispatcher                          = dllAdvapi32.Proc("StartServiceCtrlDispatcherW")
	funcRegisterServiceCtrlHandlerEx                        = dllAdvapi32.Proc("RegisterServiceCtrlHandlerExW")
	funcQueryServiceDynamicInformation                      = dllAdvapi32.Proc("QueryServiceDynamicInformation") // >= Win8
	funcConvertStringSecurityDescriptorToSecurityDescriptor = dllAdvapi32.Proc("ConvertStringSecurityDescriptorToSecurityDescriptorW")

	funcGetDC                      = dllUser32.Proc("GetDC")
	funcSetFocus                   = dllUser32.Proc("SetFocus")
	funcReleaseDC                  = dllUser32.Proc("ReleaseDC")
	funcSendInput                  = dllUser32.Proc("SendInput")
	funcBlockInput                 = dllUser32.Proc("BlockInput")
	funcShowWindow                 = dllUser32.Proc("ShowWindow")
	funcMessageBox                 = dllUser32.Proc("MessageBoxW")
	funcEnumWindows                = dllUser32.Proc("EnumWindows")
	funcEnableWindow               = dllUser32.Proc("EnableWindow")
	funcSetWindowPos               = dllUser32.Proc("SetWindowPos")
	funcGetWindowText              = dllUser32.Proc("GetWindowTextW")
	funcGetWindowInfo              = dllUser32.Proc("GetWindowInfo")
	funcGetMonitorInfo             = dllUser32.Proc("GetMonitorInfoW")
	funcGetWindowLongW             = dllUser32.Proc("GetWindowLongW")
	funcSetWindowLongW             = dllUser32.Proc("SetWindowLongW")
	funcGetDesktopWindow           = dllUser32.Proc("GetDesktopWindow")
	funcSendNotifyMessage          = dllUser32.Proc("SendNotifyMessageW")
	funcEnumDisplayMonitors        = dllUser32.Proc("EnumDisplayMonitors")
	funcEnumDisplaySettings        = dllUser32.Proc("EnumDisplaySettingsW")
	funcGetWindowTextLength        = dllUser32.Proc("GetWindowTextLengthW")
	funcSetForegroundWindow        = dllUser32.Proc("SetForegroundWindow")
	funcSystemParametersInfo       = dllUser32.Proc("SystemParametersInfoW")
	funcSetLayeredWindowAttributes = dllUser32.Proc("SetLayeredWindowAttributes")

	funcCryptMsgClose              = dllCrypt32.Proc("CryptMsgClose")
	funcCertCloseStore             = dllCrypt32.Proc("CertCloseStore")
	funcCryptQueryObject           = dllCrypt32.Proc("CryptQueryObject")
	funcCryptMsgGetParam           = dllCrypt32.Proc("CryptMsgGetParam")
	funcCertGetNameString          = dllCrypt32.Proc("CertGetNameStringW")
	funcCertFindCertificateInStore = dllCrypt32.Proc("CertFindCertificateInStore")
	funcCertFreeCertificateContext = dllCrypt32.Proc("CertFreeCertificateContext")

	funcBitBlt                 = dllGdi32.Proc("BitBlt")
	funcDeleteDC               = dllGdi32.Proc("DeleteDC")
	funcGetDIBits              = dllGdi32.Proc("GetDIBits")
	funcSelectObject           = dllGdi32.Proc("SelectObject")
	funcDeleteObject           = dllGdi32.Proc("DeleteObject")
	funcCreateCompatibleDC     = dllGdi32.Proc("CreateCompatibleDC")
	funcCreateCompatibleBitmap = dllGdi32.Proc("CreateCompatibleBitmap")

	funcWTSOpenServer              = dllWtsapi32.Proc("WTSOpenServerW")
	funcWTSCloseServer             = dllWtsapi32.Proc("WTSCloseServer")
	funcWTSSendMessage             = dllWtsapi32.Proc("WTSSendMessageW")
	funcWTSLogoffSession           = dllWtsapi32.Proc("WTSLogoffSession")
	funcWTSEnumerateSessions       = dllWtsapi32.Proc("WTSEnumerateSessionsW")
	funcWTSDisconnectSession       = dllWtsapi32.Proc("WTSDisconnectSession")
	funcWTSEnumerateProcesses      = dllWtsapi32.Proc("WTSEnumerateProcessesW")
	funcWTSQuerySessionInformation = dllWtsapi32.Proc("WTSQuerySessionInformationW")

	funcMiniDumpWriteDump = dllDbgHelp.Proc("MiniDumpWriteDump")

	funcWinHTTPGetDefaultProxyConfiguration = dllWinhttp.Proc("WinHttpGetDefaultProxyConfiguration") // >= WinXP_SP3

	funcAmsiScanBuffer = dllAmsi.Proc("AmsiScanBuffer")
	funcAmsiInitialize = dllAmsi.Proc("AmsiInitialize")
	funcAmsiScanString = dllAmsi.Proc("AmsiScanString")
)

func doSearchSystem32() bool {
	searchSystem32.Do(func() {
		searchSystem32.v = funcAddDllDirectory > 0 // >= Win8 / ~Win7
	})
	return searchSystem32.v
}
