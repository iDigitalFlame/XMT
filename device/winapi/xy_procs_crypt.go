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
	funcLoadLibraryEx       = dllKernel32.proc(0x68D28778)
	funcGetSystemDirectory  = dllKernel32.proc(0xE8DD936C)
	funcOpenProcess         = dllKernel32.proc(0xEAC1E408)
	funcOpenThread          = dllKernel32.proc(0x9D818049)
	funcCloseHandle         = dllKernel32.proc(0x6F767D81)
	funcGetCurrentProcessID = dllKernel32.proc(0x75FCB062)

	funcGetDC                                               = dllUser32.proc(0xC9AB9064)
	funcBitBlt                                              = dllGdi32.proc(0x4C7E7258)
	funcSetEvent                                            = dllKernel32.proc(0xD99E4045)
	funcIsZoomed                                            = dllUser32.proc(0xC087590F)
	funcIsIconic                                            = dllUser32.proc(0x836DE3D2)
	funcHeapFree                                            = dllKernel32.proc(0xA931332D)
	funcReadFile                                            = dllKernel32.proc(0xEBE8E9AF)
	funcLsaClose                                            = dllAdvapi32.proc(0xB9C1C829)
	funcDeleteDC                                            = dllGdi32.proc(0x3C53364B)
	funcSetFocus                                            = dllUser32.proc(0x1AF3F781)
	funcLogonUser                                           = dllAdvapi32.proc(0x5BAC4A5A)
	funcWriteFile                                           = dllKernel32.proc(0x567775AC)
	funcOpenMutex                                           = dllKernel32.proc(0x56F8CC91)
	funcLocalFree                                           = dllKernel32.proc(0x3A5DD394)
	funcOpenEvent                                           = dllKernel32.proc(0x3D0B286)
	funcGetDIBits                                           = dllGdi32.proc(0x35F5C026)
	funcReleaseDC                                           = dllUser32.proc(0x934A6B3)
	funcHeapAlloc                                           = dllKernel32.proc(0xC9CEF18E)
	funcSendInput                                           = dllUser32.proc(0xB22A0065)
	funcHeapCreate                                          = dllKernel32.proc(0x65E14E43)
	funcCreateFile                                          = dllKernel32.proc(0xBD1BFDAE)
	funcGetVersion                                          = dllKernel32.proc(0x87C40B51)
	funcCancelIoEx                                          = dllKernel32.proc(0x7BCF40)
	funcBlockInput                                          = dllUser32.proc(0x1359E3BC)
	funcShowWindow                                          = dllUser32.proc(0xB408886A)
	funcMessageBox                                          = dllUser32.proc(0x1C4E3F6C)
	funcFreeLibrary                                         = dllKernel32.proc(0x4176E4F8)
	funcHeapDestroy                                         = dllKernel32.proc(0xF1E923B)
	funcLoadLibrary                                         = dllKernel32.proc(0x9322F2CD)
	funcCreateMutex                                         = dllKernel32.proc(0x3FFF8555)
	funcCreateEvent                                         = dllKernel32.proc(0x9C12E8F2)
	funcHeapReAlloc                                         = dllKernel32.proc(0x3854EE6B)
	funcEnumWindows                                         = dllUser32.proc(0x9A29AD49)
	funcEnableWindow                                        = dllUser32.proc(0x64DED01C)
	funcSelectObject                                        = dllGdi32.proc(0xFBC3B004)
	funcDeleteObject                                        = dllGdi32.proc(0x2AAC1D49)
	funcNtTraceEvent                                        = dllNtdll.proc(0x89F984CE)
	funcResumeThread                                        = dllKernel32.proc(0x7AE85CB4)
	funcThread32Next                                        = dllKernel32.proc(0x9B4B1895)
	funcGetProcessID                                        = dllKernel32.proc(0xE58B35D9)
	funcRevertToSelf                                        = dllAdvapi32.proc(0x244DD3E6)
	funcRegEnumValue                                        = dllAdvapi32.proc(0x42EC9414)
	funcModule32Next                                        = dllKernel32.proc(0xFE05ABD4)
	funcSetWindowPos                                        = dllUser32.proc(0x57C8D93B)
	funcGetWindowText                                       = dllUser32.proc(0x123362FD)
	funcModule32First                                       = dllKernel32.proc(0x68184B5B)
	funcWaitNamedPipe                                       = dllKernel32.proc(0x7851B108)
	funcCreateProcess                                       = dllKernel32.proc(0x19C69863)
	funcSuspendThread                                       = dllKernel32.proc(0xE57214B)
	funcProcess32Next                                       = dllKernel32.proc(0x80132847)
	funcRegSetValueEx                                       = dllAdvapi32.proc(0xC0050EDC)
	funcThread32First                                       = dllKernel32.proc(0xC5311BC8)
	funcLsaOpenPolicy                                       = dllAdvapi32.proc(0x34D221F9)
	funcOpenSemaphore                                       = dllKernel32.proc(0xEFE004)
	funcRegDeleteTree                                       = dllAdvapi32.proc(0x35CED63F)
	funcRtlCopyMemory                                       = dllNtdll.proc(0xC5A43FC3)
	funcGetWindowInfo                                       = dllUser32.proc(0x971B836B)
	funcDbgBreakPoint                                       = dllNtdll.proc(0x6861210F)
	funcRegDeleteKeyEx                                      = dllAdvapi32.proc(0xF888EF35)
	funcGetMonitorInfo                                      = dllUser32.proc(0x9B68BE4A)
	funcIsWellKnownSID                                      = dllAdvapi32.proc(0xF855936A)
	funcProcess32First                                      = dllKernel32.proc(0xD4C414BE)
	funcCreateMailslot                                      = dllKernel32.proc(0xB10785BB)
	funcRegCreateKeyEx                                      = dllAdvapi32.proc(0xA656F848)
	funcSetThreadToken                                      = dllAdvapi32.proc(0x4C3BD602)
	funcRegDeleteValue                                      = dllAdvapi32.proc(0x717D1086)
	funcCreateNamedPipe                                     = dllKernel32.proc(0xF05E3B8B)
	funcDuplicateHandle                                     = dllKernel32.proc(0x3627A3C2)
	funcCreateSemaphore                                     = dllKernel32.proc(0xE540398)
	funcTerminateThread                                     = dllKernel32.proc(0x4CBF9B0A)
	funcOpenThreadToken                                     = dllAdvapi32.proc(0x4F4CA738)
	funcNtResumeProcess                                     = dllNtdll.proc(0xB5333DBD)
	funcIsWindowVisible                                     = dllUser32.proc(0x244822C5)
	funcLookupAccountSid                                    = dllAdvapi32.proc(0x59E27333)
	funcSetServiceStatus                                    = dllAdvapi32.proc(0xC09B613A)
	funcConnectNamedPipe                                    = dllKernel32.proc(0xEE1FF6A8)
	funcTerminateProcess                                    = dllKernel32.proc(0xFE03BE5D)
	funcDuplicateTokenEx                                    = dllAdvapi32.proc(0x575AEC28)
	funcNtSuspendProcess                                    = dllNtdll.proc(0x8BD95BF8)
	funcNtCreateThreadEx                                    = dllNtdll.proc(0x8E6261C)
	funcGetLogicalDrives                                    = dllKernel32.proc(0xDD09448B)
	funcOpenProcessToken                                    = dllAdvapi32.proc(0x5B4DD11F)
	funcGetDesktopWindow                                    = dllUser32.proc(0x1921BE95)
	funcGetWindowLongPtr                                    = dllUser32.proc(0x1945E84)
	funcSetWindowLongPtr                                    = dllUser32.proc(0xBA92D0C0)
	funcSendNotifyMessage                                   = dllUser32.proc(0xDEBEDBC0)
	funcGetModuleHandleEx                                   = dllKernel32.proc(0x2FFDCF65)
	funcIsDebuggerPresent                                   = dllKernel32.proc(0x88BFA355)
	funcMiniDumpWriteDump                                   = dllDbgHelp.proc(0x499916F9)
	funcGetExitCodeThread                                   = dllKernel32.proc(0xAD1DB0E0)
	funcGetExitCodeProcess                                  = dllKernel32.proc(0x6EFD1BF)
	funcCreateCompatibleDC                                  = dllGdi32.proc(0xD5203D54)
	funcGetCurrentThreadID                                  = dllKernel32.proc(0x3C31D725)
	funcCreateWellKnownSid                                  = dllAdvapi32.proc(0x25F61A8E)
	funcEnumDisplayMonitors                                 = dllUser32.proc(0x6FA69AB9)
	funcEnumDisplaySettings                                 = dllUser32.proc(0x83B28A2E)
	funcSetTokenInformation                                 = dllAdvapi32.proc(0xCDECBE4C)
	funcGetTokenInformation                                 = dllAdvapi32.proc(0xCB5ED050)
	funcGetOverlappedResult                                 = dllKernel32.proc(0x1C7ADC04)
	funcNtFreeVirtualMemory                                 = dllNtdll.proc(0x8C399853)
	funcWaitForSingleObject                                 = dllKernel32.proc(0x2CECF27A)
	funcDisconnectNamedPipe                                 = dllKernel32.proc(0xCC9E66D6)
	funcGetWindowTextLength                                 = dllUser32.proc(0x85381939)
	funcSetForegroundWindow                                 = dllUser32.proc(0x52EF9094)
	funcNtWriteVirtualMemory                                = dllNtdll.proc(0x2012F428)
	funcLookupPrivilegeValue                                = dllAdvapi32.proc(0xEC6FF8D6)
	funcSystemParametersInfo                                = dllUser32.proc(0xF1855EA9)
	funcConvertSIDToStringSID                               = dllAdvapi32.proc(0x7AAB722D)
	funcAdjustTokenPrivileges                               = dllAdvapi32.proc(0xC6B20D5F)
	funcCreateCompatibleBitmap                              = dllGdi32.proc(0xC2BE1C3E)
	funcWaitForMultipleObjects                              = dllKernel32.proc(0x440039C5)
	funcNtProtectVirtualMemory                              = dllNtdll.proc(0xD86AFCB8)
	funcCreateProcessWithToken                              = dllAdvapi32.proc(0xC20739FE)
	funcCreateProcessWithLogon                              = dllAdvapi32.proc(0x62F9BC50)
	funcImpersonateLoggedOnUser                             = dllAdvapi32.proc(0xC389384)
	funcNtAllocateVirtualMemory                             = dllNtdll.proc(0x46D22D36)
	funcRtlSetProcessIsCritical                             = dllNtdll.proc(0xEE7639E9)
	funcNtQueryInformationThread                            = dllNtdll.proc(0x115412D)
	funcCreateToolhelp32Snapshot                            = dllKernel32.proc(0xBAA64095)
	funcUpdateProcThreadAttribute                           = dllKernel32.proc(0xEB87DE36)
	funcNtQueryInformationProcess                           = dllNtdll.proc(0xC88AB8C)
	funcLsaQueryInformationPolicy                           = dllAdvapi32.proc(0xD67C4D8B)
	funcSetLayeredWindowAttributes                          = dllUser32.proc(0x950A5A2E)
	funcSetProcessWorkingSetSizeEx                          = dllKernel32.proc(0xAB634AE1)
	funcStartServiceCtrlDispatcher                          = dllAdvapi32.proc(0x99A279E7)
	funcImpersonateNamedPipeClient                          = dllAdvapi32.proc(0x2BA3D9CE)
	funcCheckRemoteDebuggerPresent                          = dllKernel32.proc(0x68D617E9)
	funcGetSecurityDescriptorLength                         = dllAdvapi32.proc(0xB8E54A56)
	funcRegisterServiceCtrlHandlerEx                        = dllAdvapi32.proc(0x5046FA66)
	funcDeleteProcThreadAttributeList                       = dllKernel32.proc(0xAE159724)
	funcQueryServiceDynamicInformation                      = dllAdvapi32.proc(0x2F5CB537)
	funcInitializeProcThreadAttributeList                   = dllKernel32.proc(0xC1C9947D)
	funcWinHTTPGetDefaultProxyConfiguration                 = dllWinhttp.proc(0xFD091ACC)
	funcConvertStringSecurityDescriptorToSecurityDescriptor = dllAdvapi32.proc(0x9EF78621)
)

func doSearchSystem32() bool {
	searchSystem32.Do(func() {
		searchSystem32.v = (dllKernel32.proc(0xE849CE13).find() == nil)
	})
	return searchSystem32.v
}
