//go:build windows && crypt

package winapi

import "github.com/iDigitalFlame/xmt/util/crypt"

var (
	debugPriv = crypt.Get(123) // SeDebugPrivilege

	dllKernel32 = &lazyDLL{Name: crypt.Get(124)} // kernel32.dll

	funcLoadLibraryEx       = dllKernel32.proc(crypt.Get(125)) // LoadLibraryExW
	funcGetSystemDirectory  = dllKernel32.proc(crypt.Get(126)) // GetSystemDirectoryW
	funcOpenProcess         = dllKernel32.proc(crypt.Get(127)) // OpenProcess
	funcOpenThread          = dllKernel32.proc(crypt.Get(128)) // OpenThread
	funcCloseHandle         = dllKernel32.proc(crypt.Get(129)) // CloseHandle
	funcGetCurrentProcessID = dllKernel32.proc(crypt.Get(130)) // GetCurrentProcessId

	dllNtdll    = &lazyDLL{Name: crypt.Get(131)} // ntdll.dll
	dllGdi32    = &lazyDLL{Name: crypt.Get(217)} // gdi32.dll
	dllUser32   = &lazyDLL{Name: crypt.Get(216)} // user32.dll
	dllWinhttp  = &lazyDLL{Name: crypt.Get(132)} // winhttp.dll
	dllDbgHelp  = &lazyDLL{Name: crypt.Get(235)} // DbgHelp.dll
	dllAdvapi32 = &lazyDLL{Name: crypt.Get(133)} // advapi32.dll

	funcGetDC                                               = dllUser32.proc(crypt.Get(229))   // GetDC
	funcBitBlt                                              = dllGdi32.proc(crypt.Get(227))    // BitBlt
	funcHeapFree                                            = dllKernel32.proc(crypt.Get(220)) // HeapFree
	funcReadFile                                            = dllKernel32.proc(crypt.Get(134)) // ReadFile
	funcLsaClose                                            = dllAdvapi32.proc(crypt.Get(135)) // LsaClose
	funcDeleteDC                                            = dllGdi32.proc(crypt.Get(222))    // DeleteDC
	funcWriteFile                                           = dllKernel32.proc(crypt.Get(136)) // WriteFile
	funcOpenMutex                                           = dllKernel32.proc(crypt.Get(137)) // OpenMutexW
	funcLocalFree                                           = dllKernel32.proc(crypt.Get(138)) // LocalFree
	funcOpenEvent                                           = dllKernel32.proc(crypt.Get(139)) // OpenEventW
	funcGetDIBits                                           = dllGdi32.proc(crypt.Get(228))    // GetDIBits
	funcReleaseDC                                           = dllUser32.proc(crypt.Get(230))   // ReleaseDC
	funcHeapAlloc                                           = dllKernel32.proc(crypt.Get(219)) // HeapAlloc
	funcCreateFile                                          = dllKernel32.proc(crypt.Get(140)) // CreateFileW
	funcGetVersion                                          = dllKernel32.proc(crypt.Get(141)) // GetVersion
	funcCancelIoEx                                          = dllKernel32.proc(crypt.Get(142)) // CancelIoEx
	funcLoadLibrary                                         = dllKernel32.proc(crypt.Get(143)) // LoadLibraryW
	funcCreateMutex                                         = dllKernel32.proc(crypt.Get(144)) // CreateMutexW
	funcCreateEvent                                         = dllKernel32.proc(crypt.Get(145)) // CreateEventW
	funcHeapReAlloc                                         = dllKernel32.proc(crypt.Get(218)) // HeapReAlloc
	funcSelectObject                                        = dllGdi32.proc(crypt.Get(226))    // SelectObject
	funcDeleteObject                                        = dllGdi32.proc(crypt.Get(223))    // DeleteObject
	funcNtTraceEvent                                        = dllNtdll.proc(crypt.Get(146))    // NtTraceEvent
	funcResumeThread                                        = dllKernel32.proc(crypt.Get(147)) // ResumeThread
	funcThread32Next                                        = dllKernel32.proc(crypt.Get(148)) // Thread32Next
	funcGetProcessID                                        = dllKernel32.proc(crypt.Get(150)) // GetProcessId
	funcRevertToSelf                                        = dllAdvapi32.proc(crypt.Get(151)) // RevertToSelf
	funcRegEnumValue                                        = dllAdvapi32.proc(crypt.Get(152)) // RegEnumValueW
	funcModule32Next                                        = dllKernel32.proc(crypt.Get(247)) // Module32NextW
	funcModule32First                                       = dllKernel32.proc(crypt.Get(248)) // Module32FirstW
	funcWaitNamedPipe                                       = dllKernel32.proc(crypt.Get(153)) // WaitNamedPipeW
	funcCreateProcess                                       = dllKernel32.proc(crypt.Get(154)) // CreateProcessW
	funcSuspendThread                                       = dllKernel32.proc(crypt.Get(155)) // SuspendThread
	funcProcess32Next                                       = dllKernel32.proc(crypt.Get(156)) // Process32NextW
	funcRegSetValueEx                                       = dllAdvapi32.proc(crypt.Get(157)) // RegSetValueExW
	funcThread32First                                       = dllKernel32.proc(crypt.Get(158)) // Thread32First
	funcLsaOpenPolicy                                       = dllAdvapi32.proc(crypt.Get(159)) // LsaOpenPolicy
	funcOpenSemaphore                                       = dllKernel32.proc(crypt.Get(160)) // OpenSemaphoreW
	funcRegDeleteTree                                       = dllAdvapi32.proc(crypt.Get(242)) // RegDeleteTreeW
	funcRtlCopyMemory                                       = dllNtdll.proc(crypt.Get(221))    // RtlCopyMemory
	funcGetProcessHeap                                      = dllKernel32.proc(crypt.Get(249)) // GetProcessHeap
	funcRegDeleteKeyEx                                      = dllAdvapi32.proc(crypt.Get(149)) // RegDeleteKeyExW
	funcGetMonitorInfo                                      = dllUser32.proc(crypt.Get(231))   // GetMonitorInfoW
	funcVirtualProtect                                      = dllKernel32.proc(crypt.Get(161)) // VirtualProtect
	funcIsWellKnownSID                                      = dllAdvapi32.proc(crypt.Get(162)) // IsWellKnownSid
	funcProcess32First                                      = dllKernel32.proc(crypt.Get(163)) // Process32FirstW
	funcCreateMailslot                                      = dllKernel32.proc(crypt.Get(164)) // CreateMailslotW
	funcRegCreateKeyEx                                      = dllAdvapi32.proc(crypt.Get(165)) // RegCreateKeyExW
	funcSetThreadToken                                      = dllAdvapi32.proc(crypt.Get(166)) // SetThreadToken
	funcRegDeleteValue                                      = dllAdvapi32.proc(crypt.Get(167)) // RegDeleteValueW
	funcCreateNamedPipe                                     = dllKernel32.proc(crypt.Get(168)) // CreateNamedPipeW
	funcDuplicateHandle                                     = dllKernel32.proc(crypt.Get(169)) // DuplicateHandle
	funcCreateSemaphore                                     = dllKernel32.proc(crypt.Get(170)) // CreateSemaphoreW
	funcTerminateThread                                     = dllKernel32.proc(crypt.Get(171)) // TerminateThread
	funcOpenThreadToken                                     = dllAdvapi32.proc(crypt.Get(172)) // OpenThreadToken
	funcNtResumeProcess                                     = dllNtdll.proc(crypt.Get(173))    // NtResumeProcess
	funcSetServiceStatus                                    = dllAdvapi32.proc(crypt.Get(212)) // SetServiceStatus
	funcConnectNamedPipe                                    = dllKernel32.proc(crypt.Get(174)) // ConnectNamedPipe
	funcTerminateProcess                                    = dllKernel32.proc(crypt.Get(175)) // TerminateProcess
	funcDuplicateTokenEx                                    = dllAdvapi32.proc(crypt.Get(176)) // DuplicateTokenEx
	funcNtSuspendProcess                                    = dllNtdll.proc(crypt.Get(177))    // NtSuspendProcess
	funcNtCreateThreadEx                                    = dllNtdll.proc(crypt.Get(178))    // NtCreateThreadEx
	funcGetLogicalDrives                                    = dllKernel32.proc(crypt.Get(209)) // GetLogicalDrives
	funcOpenProcessToken                                    = dllAdvapi32.proc(crypt.Get(179)) // OpenProcessToken
	funcGetDesktopWindow                                    = dllUser32.proc(crypt.Get(232))   // GetDesktopWindow
	funcGetModuleHandleEx                                   = dllKernel32.proc(crypt.Get(246)) // GetModuleHandleExW
	funcIsDebuggerPresent                                   = dllKernel32.proc(crypt.Get(180)) // IsDebuggerPresent
	funcMiniDumpWriteDump                                   = dllDbgHelp.proc(crypt.Get(236))  // MiniDumpWriteDump
	funcGetExitCodeThread                                   = dllKernel32.proc(crypt.Get(181)) // GetExitCodeThread
	funcGetExitCodeProcess                                  = dllKernel32.proc(crypt.Get(182)) // GetExitCodeProcess
	funcCreateCompatibleDC                                  = dllGdi32.proc(crypt.Get(224))    // CreateCompatibleDC
	funcGetCurrentThreadID                                  = dllKernel32.proc(crypt.Get(244)) // GetCurrentThreadId
	funcEnumDisplayMonitors                                 = dllUser32.proc(crypt.Get(233))   // EnumDisplayMonitors
	funcEnumDisplaySettings                                 = dllUser32.proc(crypt.Get(234))   // EnumDisplaySettingsW
	funcGetTokenInformation                                 = dllAdvapi32.proc(crypt.Get(183)) // GetTokenInformation
	funcGetOverlappedResult                                 = dllKernel32.proc(crypt.Get(184)) // GetOverlappedResult
	funcNtFreeVirtualMemory                                 = dllNtdll.proc(crypt.Get(185))    // NtFreeVirtualMemory
	funcWaitForSingleObject                                 = dllKernel32.proc(crypt.Get(186)) // WaitForSingleObject
	funcDisconnectNamedPipe                                 = dllKernel32.proc(crypt.Get(187)) // DisconnectNamedPipe
	funcNtWriteVirtualMemory                                = dllNtdll.proc(crypt.Get(188))    // NtWriteVirtualMemory
	funcLookupPrivilegeValue                                = dllAdvapi32.proc(crypt.Get(189)) // LookupPrivilegeValueW
	funcConvertSIDToStringSID                               = dllAdvapi32.proc(crypt.Get(190)) // ConvertSidToStringSidW
	funcAdjustTokenPrivileges                               = dllAdvapi32.proc(crypt.Get(191)) // AdjustTokenPrivileges
	funcCreateCompatibleBitmap                              = dllGdi32.proc(crypt.Get(225))    // CreateCompatibleBitmap
	funcNtProtectVirtualMemory                              = dllNtdll.proc(crypt.Get(192))    // NtProtectVirtualMemory
	funcCreateProcessWithToken                              = dllAdvapi32.proc(crypt.Get(193)) // CreateProcessWithTokenW
	funcNtAllocateVirtualMemory                             = dllNtdll.proc(crypt.Get(194))    // NtAllocateVirtualMemory
	funcRtlSetProcessIsCritical                             = dllNtdll.proc(crypt.Get(195))    // RtlSetProcessIsCritical
	funcNtQueryInformationThread                            = dllNtdll.proc(crypt.Get(245))    // NtQueryInformationThread
	funcCreateToolhelp32Snapshot                            = dllKernel32.proc(crypt.Get(196)) // CreateToolhelp32Snapshot
	funcUpdateProcThreadAttribute                           = dllKernel32.proc(crypt.Get(197)) // UpdateProcThreadAttribute
	funcNtQueryInformationProcess                           = dllNtdll.proc(crypt.Get(198))    // NtQueryInformationProcess
	funcLsaQueryInformationPolicy                           = dllAdvapi32.proc(crypt.Get(199)) // LsaQueryInformationPolicy
	funcStartServiceCtrlDispatcher                          = dllAdvapi32.proc(crypt.Get(213)) // StartServiceCtrlDispatcherW
	funcImpersonateNamedPipeClient                          = dllAdvapi32.proc(crypt.Get(200)) // ImpersonateNamedPipeClient
	funcCheckRemoteDebuggerPresent                          = dllKernel32.proc(crypt.Get(201)) // CheckRemoteDebuggerPresent
	funcGetSecurityDescriptorLength                         = dllAdvapi32.proc(crypt.Get(202)) // GetSecurityDescriptorLength
	funcRegisterServiceCtrlHandlerEx                        = dllAdvapi32.proc(crypt.Get(215)) // RegisterServiceCtrlHandlerExW
	funcDeleteProcThreadAttributeList                       = dllKernel32.proc(crypt.Get(203)) // DeleteProcThreadAttributeList
	funcQueryServiceDynamicInformation                      = dllAdvapi32.proc(crypt.Get(214)) // QueryServiceDynamicInformation
	funcInitializeProcThreadAttributeList                   = dllKernel32.proc(crypt.Get(204)) // InitializeProcThreadAttributeList
	funcWinHTTPGetDefaultProxyConfiguration                 = dllWinhttp.proc(crypt.Get(205))  // WinHttpGetDefaultProxyConfiguration
	funcConvertStringSecurityDescriptorToSecurityDescriptor = dllAdvapi32.proc(crypt.Get(206)) // ConvertStringSecurityDescriptorToSecurityDescriptorW

)

func doSearchSystem32() bool {
	searchSystem32.Do(func() {
		searchSystem32.v = (dllKernel32.proc(crypt.Get(207)).find() == nil) // AddDllDirectory
	})
	return searchSystem32.v
}
