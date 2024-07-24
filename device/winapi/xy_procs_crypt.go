//go:build windows && (altload || crypt)
// +build windows
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
	funcLoadLibraryEx = dllKernelBase.Proc(0x68D28778)
	funcFormatMessage = dllKernelBase.Proc(0x8233A148)

	funcNtClose                       = dllNtdll.sysProc(0x36291E41)
	funcNtSetEvent                    = dllNtdll.sysProc(0x5E5D5E5B)
	funcRtlFreeHeap                   = dllNtdll.Proc(0xBC880A2D)
	funcNtTraceEvent                  = dllNtdll.sysProc(0x89F984CE)
	funcNtOpenThread                  = dllNtdll.sysProc(0x7319665F)
	funcRtlCreateHeap                 = dllNtdll.Proc(0xA1846AB)
	funcEtwEventWrite                 = dllNtdll.Proc(0xD32A6690) // >= WinVista
	funcDbgBreakPoint                 = dllNtdll.Proc(0x6861210F)
	funcNtOpenProcess                 = dllNtdll.sysProc(0x57367582)
	funcRtlDestroyHeap                = dllNtdll.Proc(0x167E8613)
	funcNtResumeThread                = dllNtdll.sysProc(0xA6F798EA)
	funcNtCreateSection               = dllNtdll.sysProc(0x40A2511C)
	funcNtSuspendThread               = dllNtdll.sysProc(0x9D419019)
	funcNtResumeProcess               = dllNtdll.sysProc(0xB5333DBD)
	funcRtlAllocateHeap               = dllNtdll.Proc(0x50AA445E)
	funcNtDuplicateToken              = dllNtdll.sysProc(0x7A75D3A1)
	funcEtwEventRegister              = dllNtdll.Proc(0xC0B4D94C) // >= WinVista
	funcNtSuspendProcess              = dllNtdll.sysProc(0x8BD95BF8)
	funcNtCreateThreadEx              = dllNtdll.sysProc(0x8E6261C)  // >= WinVista (Xp sub = RtlCreateUserThread)
	funcNtCancelIoFileEx              = dllNtdll.sysProc(0xD4909C18) // >= WinVista (Xp sub = NtCancelIoFile)
	funcNtDuplicateObject             = dllNtdll.sysProc(0xAD2BC047)
	funcNtTerminateThread             = dllNtdll.sysProc(0x18157A24)
	funcNtOpenThreadToken             = dllNtdll.sysProc(0x82EEAAFE)
	funcEtwEventWriteFull             = dllNtdll.Proc(0xAC8A097) // >= WinVista
	funcRtlReAllocateHeap             = dllNtdll.Proc(0xA51D1975)
	funcNtMapViewOfSection            = dllNtdll.sysProc(0x704A2F2C)
	funcNtTerminateProcess            = dllNtdll.sysProc(0xB3AC5173)
	funcNtOpenProcessToken            = dllNtdll.sysProc(0xB2CA3641)
	funcRtlCopyMappedMemory           = dllNtdll.Proc(0x381752E6) // >= WinS2003 (Not in XP sub = RtlMoveMemory)
	funcNtFreeVirtualMemory           = dllNtdll.sysProc(0x8C399853)
	funcNtImpersonateThread           = dllNtdll.sysProc(0x12724B12)
	funcNtUnmapViewOfSection          = dllNtdll.sysProc(0x19B022D)
	funcNtWriteVirtualMemory          = dllNtdll.sysProc(0x2012F428)
	funcNtDeviceIoControlFile         = dllNtdll.sysProc(0x5D0C9026)
	funcNtWaitForSingleObject         = dllNtdll.sysProc(0x46D9033C)
	funcNtSetInformationToken         = dllNtdll.sysProc(0x43623A4)
	funcNtProtectVirtualMemory        = dllNtdll.sysProc(0xD86AFCB8)
	funcNtSetInformationThread        = dllNtdll.sysProc(0x5F74B08D)
	funcRtlGetNtVersionNumbers        = dllNtdll.Proc(0xD476F98B)
	funcEtwNotificationRegister       = dllNtdll.Proc(0x7B7F821F) // >= WinVista
	funcNtAllocateVirtualMemory       = dllNtdll.sysProc(0x46D22D36)
	funcRtlSetProcessIsCritical       = dllNtdll.Proc(0xEE7639E9)
	funcNtFlushInstructionCache       = dllNtdll.sysProc(0xEFB80179)
	funcNtAdjustTokenPrivileges       = dllNtdll.sysProc(0x6CCF6931)
	funcNtQueryInformationToken       = dllNtdll.sysProc(0x63C176C4)
	funcNtQueryInformationThread      = dllNtdll.sysProc(0x115412D)
	funcNtQuerySystemInformation      = dllNtdll.sysProc(0x337C7C64)
	funcNtWaitForMultipleObjects      = dllNtdll.sysProc(0x5DF74043)
	funcNtQueryInformationProcess     = dllNtdll.sysProc(0xC88AB8C)
	funcRtlWow64GetProcessMachines    = dllNtdll.Proc(0x982D219D) // == 64bit/ARM64
	funcRtlLengthSecurityDescriptor   = dllNtdll.Proc(0xF5677F7C)
	funcRtlGetDaclSecurityDescriptor  = dllNtdll.Proc(0x13464D36)
	funcRtlGetSaclSecurityDescriptor  = dllNtdll.Proc(0xE72F0F6F)
	funcRtlGetGroupSecurityDescriptor = dllNtdll.Proc(0xD1F4CD59)
	funcRtlGetOwnerSecurityDescriptor = dllNtdll.Proc(0xB5D71CF9)

	funcReadFile                  = dllKernelBase.Proc(0xEBE8E9AF)
	funcWriteFile                 = dllKernelBase.Proc(0x567775AC)
	funcOpenMutex                 = dllKernelBase.Proc(0x56F8CC91)
	funcLocalFree                 = dllKernelBase.Proc(0x3A5DD394)
	funcOpenEvent                 = dllKernelBase.Proc(0x3D0B286)
	funcCreateFile                = dllKernelBase.Proc(0xBD1BFDAE)
	funcDebugBreak                = dllKernelBase.Proc(0x7F7E4A57)
	funcCreateMutex               = dllKernelBase.Proc(0x3FFF8555)
	funcCreateEvent               = dllKernelBase.Proc(0x9C12E8F2)
	funcWaitNamedPipe             = dllKernelBase.Proc(0x7851B108)
	funcOpenSemaphore             = dllKernelBase.Proc(0xEFE004)
	funcCreateNamedPipe           = dllKernelBase.Proc(0xF05E3B8B)
	funcConnectNamedPipe          = dllKernelBase.Proc(0xEE1FF6A8)
	funcGetModuleHandleEx         = dllKernelBase.Proc(0x2FFDCF65)
	funcOutputDebugString         = dllKernelBase.Proc(0x58448029)
	funcGetCurrentThreadID        = dllKernelBase.Proc(0x3C31D725)
	funcGetOverlappedResult       = dllKernelBase.Proc(0x1C7ADC04)
	funcDisconnectNamedPipe       = dllKernelBase.Proc(0xCC9E66D6)
	funcGetCurrentProcessID       = dllKernelBase.Proc(0x75FCB062)
	funcUpdateProcThreadAttribute = dllKernelBase.Proc(0xEB87DE36) // >= WinVista

	funcIsWellKnownSID             = dllKernelOrAdvapi.Proc(0xF855936A) // >= Win7 kernelbase.dll else advapi32.dll
	funcCreateWellKnownSid         = dllKernelOrAdvapi.Proc(0x25F61A8E) // >= Win7 kernelbase.dll else advapi32.dll
	funcImpersonateNamedPipeClient = dllKernelOrAdvapi.Proc(0x2BA3D9CE) // >= Win7 kernelbase.dll else advapi32.dll

	funcCreateProcess              = dllKernel32.Proc(0x19C69863)
	funcCreateMailslot             = dllKernel32.Proc(0xB10785BB)
	funcCreateSemaphore            = dllKernel32.Proc(0xE540398)
	funcK32EnumDeviceDrivers       = dllKernel32.Proc(0x779D5EFF) // >= Win7 (Xp sub = psapi.EnumDeviceDrivers)
	funcK32GetModuleInformation    = dllKernel32.Proc(0xFD5B63D5) // >= Win7 (Xp sub = psapi.GetModuleInformation)
	funcSetProcessWorkingSetSizeEx = dllKernel32.Proc(0xAB634AE1) // >= WinS2003 (Not in XP sub = SetProcessWorkingSetSize)
	funcK32GetDeviceDriverFileName = dllKernel32.Proc(0x9EF6FF6D) // >= Win7 (Xp sub = psapi.GetDeviceDriverFileNameW)

	funcLsaClose                                            = dllAdvapi32.Proc(0xB9C1C829)
	funcLogonUser                                           = dllAdvapi32.Proc(0x5BAC4A5A)
	funcRegFlushKey                                         = dllAdvapi32.Proc(0x8177DB3A)
	funcRegEnumValue                                        = dllAdvapi32.Proc(0x42EC9414)
	funcRegSetValueEx                                       = dllAdvapi32.Proc(0xC0050EDC)
	funcLsaOpenPolicy                                       = dllAdvapi32.Proc(0x34D221F9)
	funcRegDeleteTree                                       = dllAdvapi32.Proc(0x35CED63F) // >= WinVista
	funcRegDeleteKeyEx                                      = dllAdvapi32.Proc(0xF888EF35) // >= WinVista (Xp sub = RegDeleteKey)
	funcRegDeleteValue                                      = dllAdvapi32.Proc(0x717D1086)
	funcRegCreateKeyEx                                      = dllAdvapi32.Proc(0xA656F848)
	funcSetServiceStatus                                    = dllAdvapi32.Proc(0xC09B613A)
	funcLookupAccountSid                                    = dllAdvapi32.Proc(0x59E27333)
	funcGetNamedSecurityInfo                                = dllAdvapi32.Proc(0x411B68C7)
	funcSetNamedSecurityInfo                                = dllAdvapi32.Proc(0xFA5B67F3)
	funcLookupPrivilegeValue                                = dllAdvapi32.Proc(0xEC6FF8D6)
	funcConvertSIDToStringSID                               = dllAdvapi32.Proc(0x7AAB722D)
	funcCreateProcessWithToken                              = dllAdvapi32.Proc(0xC20739FE) // >= WinS2003 (Not in XP)
	funcCreateProcessWithLogon                              = dllAdvapi32.Proc(0x62F9BC50)
	funcInitiateSystemShutdownEx                            = dllAdvapi32.Proc(0xDA8731DD)
	funcLsaQueryInformationPolicy                           = dllAdvapi32.Proc(0xD67C4D8B)
	funcStartServiceCtrlDispatcher                          = dllAdvapi32.Proc(0x99A279E7)
	funcRegisterServiceCtrlHandlerEx                        = dllAdvapi32.Proc(0x5046FA66)
	funcQueryServiceDynamicInformation                      = dllAdvapi32.Proc(0x2F5CB537) // >= Win8
	funcConvertStringSecurityDescriptorToSecurityDescriptor = dllAdvapi32.Proc(0x9EF78621)

	funcGetDC                      = dllUser32.Proc(0xC9AB9064)
	funcSetFocus                   = dllUser32.Proc(0x1AF3F781)
	funcReleaseDC                  = dllUser32.Proc(0x934A6B3)
	funcSendInput                  = dllUser32.Proc(0xB22A0065)
	funcBlockInput                 = dllUser32.Proc(0x1359E3BC)
	funcShowWindow                 = dllUser32.Proc(0xB408886A)
	funcMessageBox                 = dllUser32.Proc(0x1C4E3F6C)
	funcEnumWindows                = dllUser32.Proc(0x9A29AD49)
	funcEnableWindow               = dllUser32.Proc(0x64DED01C)
	funcSetWindowPos               = dllUser32.Proc(0x57C8D93B)
	funcGetWindowText              = dllUser32.Proc(0x123362FD)
	funcGetWindowInfo              = dllUser32.Proc(0x971B836B)
	funcGetMonitorInfo             = dllUser32.Proc(0x9B68BE4A)
	funcGetWindowLongW             = dllUser32.Proc(0x31A5F5B0)
	funcSetWindowLongW             = dllUser32.Proc(0x8BD0F82C)
	funcGetDesktopWindow           = dllUser32.Proc(0x1921BE95)
	funcSendNotifyMessage          = dllUser32.Proc(0xDEBEDBC0)
	funcEnumDisplayMonitors        = dllUser32.Proc(0x6FA69AB9)
	funcEnumDisplaySettings        = dllUser32.Proc(0x83B28A2E)
	funcGetWindowTextLength        = dllUser32.Proc(0x85381939)
	funcSetForegroundWindow        = dllUser32.Proc(0x52EF9094)
	funcSystemParametersInfo       = dllUser32.Proc(0xF1855EA9)
	funcSetLayeredWindowAttributes = dllUser32.Proc(0x950A5A2E)

	funcCryptMsgClose              = dllCrypt32.Proc(0x9B5720EA)
	funcCertCloseStore             = dllCrypt32.Proc(0xF614DAC4)
	funcCryptQueryObject           = dllCrypt32.Proc(0xEAEDD248)
	funcCryptMsgGetParam           = dllCrypt32.Proc(0xEE8C1C55)
	funcCertGetNameString          = dllCrypt32.Proc(0x3F6B7692)
	funcCertFindCertificateInStore = dllCrypt32.Proc(0x38707435)
	funcCertFreeCertificateContext = dllCrypt32.Proc(0x6F27DE27)

	funcBitBlt                 = dllGdi32.Proc(0x4C7E7258)
	funcDeleteDC               = dllGdi32.Proc(0x3C53364B)
	funcGetDIBits              = dllGdi32.Proc(0x35F5C026)
	funcSelectObject           = dllGdi32.Proc(0xFBC3B004)
	funcDeleteObject           = dllGdi32.Proc(0x2AAC1D49)
	funcCreateCompatibleDC     = dllGdi32.Proc(0xD5203D54)
	funcCreateCompatibleBitmap = dllGdi32.Proc(0xC2BE1C3E)

	funcWTSOpenServer              = dllWtsapi32.Proc(0xFE2B3B89)
	funcWTSCloseServer             = dllWtsapi32.Proc(0x1BCAB670)
	funcWTSSendMessage             = dllWtsapi32.Proc(0xACD5E389)
	funcWTSLogoffSession           = dllWtsapi32.Proc(0xE355D47E)
	funcWTSEnumerateSessions       = dllWtsapi32.Proc(0x81A0698B)
	funcWTSDisconnectSession       = dllWtsapi32.Proc(0x9A352247)
	funcWTSEnumerateProcesses      = dllWtsapi32.Proc(0x9BC0257D)
	funcWTSQuerySessionInformation = dllWtsapi32.Proc(0xCEFF39A)

	funcMiniDumpWriteDump = dllDbgHelp.Proc(0x499916F9)

	funcWinHTTPGetDefaultProxyConfiguration = dllWinhttp.Proc(0xFD091ACC) // >= WinXP_SP3

	funcAmsiScanBuffer = dllAmsi.Proc(0x7AB1BB42) // >= Win10
	funcAmsiInitialize = dllAmsi.Proc(0xBFB2E53D) // >= Win10
	funcAmsiScanString = dllAmsi.Proc(0x18AB3DF)  // >= Win10
)

func doSearchSystem32() bool {
	searchSystem32.Do(func() {
		searchSystem32.v = funcAddDllDirectory > 0 // >= Win8 / ~Win7
	})
	return searchSystem32.v
}
