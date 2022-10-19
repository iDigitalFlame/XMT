//go:build windows && (altload || crypt)

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
	funcLoadLibraryEx = dllKernelBase.proc(0x68D28778)
	funcFormatMessage = dllKernelBase.proc(0x8233A148)

	funcNtClose                     = dllNtdll.proc(0x36291E41)
	funcRtlFreeHeap                 = dllNtdll.proc(0xBC880A2D) // Does not make a Syscall
	funcNtTraceEvent                = dllNtdll.proc(0x89F984CE)
	funcEtwEventWrite               = dllNtdll.proc(0xD32A6690)
	funcDbgBreakPoint               = dllNtdll.proc(0x6861210F)
	funcNtOpenProcess               = dllNtdll.proc(0x57367582)
	funcNtResumeThread              = dllNtdll.proc(0xA6F798EA)
	funcNtCreateSection             = dllNtdll.proc(0x40A2511C)
	funcNtSuspendThread             = dllNtdll.proc(0x9D419019)
	funcNtResumeProcess             = dllNtdll.proc(0xB5333DBD)
	funcRtlAllocateHeap             = dllNtdll.proc(0x50AA445E) // Does not make a Syscall
	funcNtDuplicateToken            = dllNtdll.proc(0x7A75D3A1)
	funcEtwEventRegister            = dllNtdll.proc(0xC0B4D94C) // Does not make a Syscall
	funcNtSuspendProcess            = dllNtdll.proc(0x8BD95BF8)
	funcNtCreateThreadEx            = dllNtdll.proc(0x8E6261C)
	funcNtDuplicateObject           = dllNtdll.proc(0xAD2BC047)
	funcNtTerminateThread           = dllNtdll.proc(0x18157A24)
	funcNtOpenThreadToken           = dllNtdll.proc(0x82EEAAFE)
	funcEtwEventWriteFull           = dllNtdll.proc(0xAC8A097)
	funcRtlReAllocateHeap           = dllNtdll.proc(0xA51D1975) // Does not make a Syscall
	funcNtMapViewOfSection          = dllNtdll.proc(0x704A2F2C)
	funcNtTerminateProcess          = dllNtdll.proc(0xB3AC5173)
	funcNtOpenProcessToken          = dllNtdll.proc(0xB2CA3641)
	funcRtlCopyMappedMemory         = dllNtdll.proc(0x381752E6) // Does not make a Syscall
	funcNtFreeVirtualMemory         = dllNtdll.proc(0x8C399853)
	funcNtImpersonateThread         = dllNtdll.proc(0x12724B12)
	funcNtUnmapViewOfSection        = dllNtdll.proc(0x19B022D)
	funcNtWriteVirtualMemory        = dllNtdll.proc(0x2012F428)
	funcNtProtectVirtualMemory      = dllNtdll.proc(0xD86AFCB8)
	funcNtSetInformationThread      = dllNtdll.proc(0x5F74B08D)
	funcRtlGetNtVersionNumbers      = dllNtdll.proc(0xD476F98B) // Does not make a Syscall
	funcEtwNotificationRegister     = dllNtdll.proc(0x7B7F821F) // Does not make a Syscall
	funcNtAllocateVirtualMemory     = dllNtdll.proc(0x46D22D36)
	funcRtlSetProcessIsCritical     = dllNtdll.proc(0xEE7639E9) // Does not make a Syscall
	funcNtFlushInstructionCache     = dllNtdll.proc(0xEFB80179)
	funcNtAdjustTokenPrivileges     = dllNtdll.proc(0x6CCF6931)
	funcNtQueryInformationThread    = dllNtdll.proc(0x115412D)
	funcNtQuerySystemInformation    = dllNtdll.proc(0x337C7C64)
	funcNtQueryInformationProcess   = dllNtdll.proc(0xC88AB8C)
	funcRtlLengthSecurityDescriptor = dllNtdll.proc(0xF5677F7C) // Does not make a Syscall

	funcSetEvent                      = dllKernelBase.proc(0xD99E4045)
	funcReadFile                      = dllKernelBase.proc(0xEBE8E9AF)
	funcWriteFile                     = dllKernelBase.proc(0x567775AC)
	funcOpenMutex                     = dllKernelBase.proc(0x56F8CC91)
	funcLocalFree                     = dllKernelBase.proc(0x3A5DD394)
	funcOpenEvent                     = dllKernelBase.proc(0x3D0B286)
	funcOpenThread                    = dllKernelBase.proc(0x9D818049)
	funcHeapCreate                    = dllKernelBase.proc(0x65E14E43)
	funcCreateFile                    = dllKernelBase.proc(0xBD1BFDAE)
	funcCancelIoEx                    = dllKernelBase.proc(0x7BCF40)
	funcDebugBreak                    = dllKernelBase.proc(0x7F7E4A57)
	funcFreeLibrary                   = dllKernelBase.proc(0x4176E4F8)
	funcHeapDestroy                   = dllKernelBase.proc(0xF1E923B)
	funcCreateMutex                   = dllKernelBase.proc(0x3FFF8555)
	funcCreateEvent                   = dllKernelBase.proc(0x9C12E8F2)
	funcWaitNamedPipe                 = dllKernelBase.proc(0x7851B108)
	funcOpenSemaphore                 = dllKernelBase.proc(0xEFE004)
	funcSetThreadToken                = dllKernelBase.proc(0x4C3BD602)
	funcIsWow64Process                = dllKernelBase.proc(0xC7D72D9F)
	funcIsWellKnownSID                = dllKernelBase.proc(0xF855936A)
	funcCreateNamedPipe               = dllKernelBase.proc(0xF05E3B8B)
	funcGetLogicalDrives              = dllKernelBase.proc(0xDD09448B)
	funcConnectNamedPipe              = dllKernelBase.proc(0xEE1FF6A8)
	funcGetModuleHandleEx             = dllKernelBase.proc(0x2FFDCF65)
	funcIsDebuggerPresent             = dllKernelBase.proc(0x88BFA355)
	funcCreateWellKnownSid            = dllKernelBase.proc(0x25F61A8E)
	funcGetCurrentThreadID            = dllKernelBase.proc(0x3C31D725)
	funcGetOverlappedResult           = dllKernelBase.proc(0x1C7ADC04)
	funcDisconnectNamedPipe           = dllKernelBase.proc(0xCC9E66D6)
	funcSetTokenInformation           = dllKernelBase.proc(0xCDECBE4C)
	funcGetTokenInformation           = dllKernelBase.proc(0xCB5ED050)
	funcGetCurrentProcessID           = dllKernelBase.proc(0x75FCB062)
	funcWaitForSingleObject           = dllKernelBase.proc(0x96B1F869)
	funcWaitForMultipleObjects        = dllKernelBase.proc(0xEC29F0D6)
	funcImpersonateLoggedOnUser       = dllKernelBase.proc(0xC389384)
	funcUpdateProcThreadAttribute     = dllKernelBase.proc(0xEB87DE36)
	funcImpersonateNamedPipeClient    = dllKernelBase.proc(0x2BA3D9CE)
	funcDeleteProcThreadAttributeList = dllKernelBase.proc(0xAE159724)

	funcLoadLibrary                = dllKernel32.proc(0x9322F2CD)
	funcThread32Next               = dllKernel32.proc(0x9B4B1895)
	funcCreateProcess              = dllKernel32.proc(0x19C69863)
	funcProcess32Next              = dllKernel32.proc(0x80132847)
	funcThread32First              = dllKernel32.proc(0xC5311BC8)
	funcProcess32First             = dllKernel32.proc(0xD4C414BE)
	funcCreateMailslot             = dllKernel32.proc(0xB10785BB)
	funcCreateSemaphore            = dllKernel32.proc(0xE540398)
	funcK32EnumDeviceDrivers       = dllKernel32.proc(0x779D5EFF)
	funcK32GetModuleInformation    = dllKernel32.proc(0xFD5B63D5)
	funcCreateToolhelp32Snapshot   = dllKernel32.proc(0xBAA64095)
	funcK32GetDeviceDriverBaseName = dllKernel32.proc(0xB376188E)
	funcSetProcessWorkingSetSizeEx = dllKernel32.proc(0xAB634AE1)

	funcLsaClose                                            = dllAdvapi32.proc(0xB9C1C829)
	funcLogonUser                                           = dllAdvapi32.proc(0x5BAC4A5A)
	funcRegFlushKey                                         = dllAdvapi32.proc(0x8177DB3A)
	funcRegEnumValue                                        = dllAdvapi32.proc(0x42EC9414)
	funcRegSetValueEx                                       = dllAdvapi32.proc(0xC0050EDC)
	funcLsaOpenPolicy                                       = dllAdvapi32.proc(0x34D221F9)
	funcRegDeleteTree                                       = dllAdvapi32.proc(0x35CED63F)
	funcRegDeleteKeyEx                                      = dllAdvapi32.proc(0xF888EF35)
	funcRegDeleteValue                                      = dllAdvapi32.proc(0x717D1086)
	funcRegCreateKeyEx                                      = dllAdvapi32.proc(0xA656F848)
	funcSetServiceStatus                                    = dllAdvapi32.proc(0xC09B613A)
	funcLookupAccountSid                                    = dllAdvapi32.proc(0x59E27333)
	funcLookupPrivilegeValue                                = dllAdvapi32.proc(0xEC6FF8D6)
	funcConvertSIDToStringSID                               = dllAdvapi32.proc(0x7AAB722D)
	funcCreateProcessWithToken                              = dllAdvapi32.proc(0xC20739FE)
	funcCreateProcessWithLogon                              = dllAdvapi32.proc(0x62F9BC50)
	funcInitiateSystemShutdownEx                            = dllAdvapi32.proc(0xDA8731DD)
	funcLsaQueryInformationPolicy                           = dllAdvapi32.proc(0xD67C4D8B)
	funcStartServiceCtrlDispatcher                          = dllAdvapi32.proc(0x99A279E7)
	funcRegisterServiceCtrlHandlerEx                        = dllAdvapi32.proc(0x5046FA66)
	funcQueryServiceDynamicInformation                      = dllAdvapi32.proc(0x2F5CB537)
	funcConvertStringSecurityDescriptorToSecurityDescriptor = dllAdvapi32.proc(0x9EF78621)

	funcGetDC                      = dllUser32.proc(0xC9AB9064)
	funcIsZoomed                   = dllUser32.proc(0xC087590F)
	funcIsIconic                   = dllUser32.proc(0x836DE3D2)
	funcSetFocus                   = dllUser32.proc(0x1AF3F781)
	funcReleaseDC                  = dllUser32.proc(0x934A6B3)
	funcSendInput                  = dllUser32.proc(0xB22A0065)
	funcBlockInput                 = dllUser32.proc(0x1359E3BC)
	funcShowWindow                 = dllUser32.proc(0xB408886A)
	funcMessageBox                 = dllUser32.proc(0x1C4E3F6C)
	funcEnumWindows                = dllUser32.proc(0x9A29AD49)
	funcEnableWindow               = dllUser32.proc(0x64DED01C)
	funcSetWindowPos               = dllUser32.proc(0x57C8D93B)
	funcGetWindowText              = dllUser32.proc(0x123362FD)
	funcGetWindowInfo              = dllUser32.proc(0x971B836B)
	funcGetMonitorInfo             = dllUser32.proc(0x9B68BE4A)
	funcGetWindowLongW             = dllUser32.proc(0x31A5F5B0)
	funcSetWindowLongW             = dllUser32.proc(0x8BD0F82C)
	funcIsWindowVisible            = dllUser32.proc(0x244822C5)
	funcGetDesktopWindow           = dllUser32.proc(0x1921BE95)
	funcSendNotifyMessage          = dllUser32.proc(0xDEBEDBC0)
	funcEnumDisplayMonitors        = dllUser32.proc(0x6FA69AB9)
	funcEnumDisplaySettings        = dllUser32.proc(0x83B28A2E)
	funcGetWindowTextLength        = dllUser32.proc(0x85381939)
	funcSetForegroundWindow        = dllUser32.proc(0x52EF9094)
	funcSystemParametersInfo       = dllUser32.proc(0xF1855EA9)
	funcSetLayeredWindowAttributes = dllUser32.proc(0x950A5A2E)

	funcBitBlt                 = dllGdi32.proc(0x4C7E7258)
	funcDeleteDC               = dllGdi32.proc(0x3C53364B)
	funcGetDIBits              = dllGdi32.proc(0x35F5C026)
	funcSelectObject           = dllGdi32.proc(0xFBC3B004)
	funcDeleteObject           = dllGdi32.proc(0x2AAC1D49)
	funcCreateCompatibleDC     = dllGdi32.proc(0xD5203D54)
	funcCreateCompatibleBitmap = dllGdi32.proc(0xC2BE1C3E)

	funcWTSFreeMemory              = dllWtsapi32.proc(0x8264A52C)
	funcWTSOpenServer              = dllWtsapi32.proc(0xFE2B3B89)
	funcWTSCloseServer             = dllWtsapi32.proc(0x1BCAB670)
	funcWTSSendMessage             = dllWtsapi32.proc(0xACD5E389)
	funcWTSLogoffSession           = dllWtsapi32.proc(0xE355D47E)
	funcWTSEnumerateSessions       = dllWtsapi32.proc(0x81A0698B)
	funcWTSDisconnectSession       = dllWtsapi32.proc(0x9A352247)
	funcWTSEnumerateProcesses      = dllWtsapi32.proc(0x9BC0257D)
	funcWTSQuerySessionInformation = dllWtsapi32.proc(0xCEFF39A)

	funcMiniDumpWriteDump = dllDbgHelp.proc(0x499916F9)

	funcWinHTTPGetDefaultProxyConfiguration = dllWinhttp.proc(0xFD091ACC)

	funcAmsiScanBuffer = dllAmsi.proc(0x7AB1BB42)
	funcAmsiInitialize = dllAmsi.proc(0xBFB2E53D)
	funcAmsiScanString = dllAmsi.proc(0x18AB3DF)
)

func doSearchSystem32() bool {
	searchSystem32.Do(func() {
		searchSystem32.v = (dllKernel32.proc(0xE849CE13).find() == nil)
	})
	return searchSystem32.v
}
