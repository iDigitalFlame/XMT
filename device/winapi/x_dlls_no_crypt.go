//go:build windows && !crypt
// +build windows,!crypt

package winapi

const debugPriv = "SeDebugPrivilege"

var (
	dllKernel32 = &lazyDLL{Name: "kernel32.dll"}

	funcLoadLibraryEx       = dllKernel32.proc("LoadLibraryExW")
	funcGetSystemDirectory  = dllKernel32.proc("GetSystemDirectoryW")
	funcOpenProcess         = dllKernel32.proc("OpenProcess")
	funcOpenThread          = dllKernel32.proc("OpenThread")
	funcCloseHandle         = dllKernel32.proc("CloseHandle")
	funcGetCurrentProcessID = dllKernel32.proc("GetCurrentProcessId")

	dllNtdll    = &lazyDLL{Name: "ntdll.dll"}
	dllWinhttp  = &lazyDLL{Name: "winhttp.dll"}
	dllAdvapi32 = &lazyDLL{Name: "advapi32.dll"}

	//funcRtlCloneUserProcess                                 = dllNtdll.proc("RtlCloneUserProcess")
	//funcAllocConsole                                        = dllKernel32.proc("AllocConsole")

	funcReadFile                                            = dllKernel32.proc("ReadFile")
	funcLsaClose                                            = dllAdvapi32.proc("LsaClose")
	funcWriteFile                                           = dllKernel32.proc("WriteFile")
	funcOpenMutex                                           = dllKernel32.proc("OpenMutexW")
	funcLocalFree                                           = dllKernel32.proc("LocalFree")
	funcOpenEvent                                           = dllKernel32.proc("OpenEventW")
	funcCreateFile                                          = dllKernel32.proc("CreateFileW")
	funcGetVersion                                          = dllKernel32.proc("GetVersion")
	funcCancelIoEx                                          = dllKernel32.proc("CancelIoEx")
	funcLoadLibrary                                         = dllKernel32.proc("LoadLibraryW")
	funcCreateMutex                                         = dllKernel32.proc("CreateMutexW")
	funcCreateEvent                                         = dllKernel32.proc("CreateEventW")
	funcNtTraceEvent                                        = dllNtdll.proc("NtTraceEvent")
	funcResumeThread                                        = dllKernel32.proc("ResumeThread")
	funcThread32Next                                        = dllKernel32.proc("Thread32Next")
	funcRegDeleteKey                                        = dllAdvapi32.proc("RegDeleteKeyW")
	funcGetProcessID                                        = dllKernel32.proc("GetProcessId")
	funcRevertToSelf                                        = dllAdvapi32.proc("RevertToSelf")
	funcRegEnumValue                                        = dllAdvapi32.proc("RegEnumValueW")
	funcWaitNamedPipe                                       = dllKernel32.proc("WaitNamedPipeW")
	funcCreateProcess                                       = dllKernel32.proc("CreateProcessW")
	funcSuspendThread                                       = dllKernel32.proc("SuspendThread")
	funcProcess32Next                                       = dllKernel32.proc("Process32NextW")
	funcRegSetValueEx                                       = dllAdvapi32.proc("RegSetValueExW")
	funcThread32First                                       = dllKernel32.proc("Thread32First")
	funcLsaOpenPolicy                                       = dllAdvapi32.proc("LsaOpenPolicy")
	funcOpenSemaphore                                       = dllKernel32.proc("OpenSemaphoreW")
	funcVirtualProtect                                      = dllKernel32.proc("VirtualProtect")
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
	funcConnectNamedPipe                                    = dllKernel32.proc("ConnectNamedPipe")
	funcTerminateProcess                                    = dllKernel32.proc("TerminateProcess")
	funcDuplicateTokenEx                                    = dllAdvapi32.proc("DuplicateTokenEx")
	funcNtSuspendProcess                                    = dllNtdll.proc("NtSuspendProcess")
	funcNtCreateThreadEx                                    = dllNtdll.proc("NtCreateThreadEx")
	funcGetLogicalDrives                                    = dllKernel32.proc("GetLogicalDrives")
	funcOpenProcessToken                                    = dllAdvapi32.proc("OpenProcessToken")
	funcIsDebuggerPresent                                   = dllKernel32.proc("IsDebuggerPresent")
	funcGetExitCodeThread                                   = dllKernel32.proc("GetExitCodeThread")
	funcGetExitCodeProcess                                  = dllKernel32.proc("GetExitCodeProcess")
	funcGetTokenInformation                                 = dllAdvapi32.proc("GetTokenInformation")
	funcGetOverlappedResult                                 = dllKernel32.proc("GetOverlappedResult")
	funcNtFreeVirtualMemory                                 = dllNtdll.proc("NtFreeVirtualMemory")
	funcWaitForSingleObject                                 = dllKernel32.proc("WaitForSingleObject")
	funcDisconnectNamedPipe                                 = dllKernel32.proc("DisconnectNamedPipe")
	funcNtWriteVirtualMemory                                = dllNtdll.proc("NtWriteVirtualMemory")
	funcLookupPrivilegeValue                                = dllAdvapi32.proc("LookupPrivilegeValueW")
	funcConvertSIDToStringSID                               = dllAdvapi32.proc("ConvertSidToStringSidW")
	funcAdjustTokenPrivileges                               = dllAdvapi32.proc("AdjustTokenPrivileges")
	funcNtProtectVirtualMemory                              = dllNtdll.proc("NtProtectVirtualMemory")
	funcCreateProcessWithToken                              = dllAdvapi32.proc("CreateProcessWithTokenW")
	funcNtAllocateVirtualMemory                             = dllNtdll.proc("NtAllocateVirtualMemory")
	funcRtlSetProcessIsCritical                             = dllNtdll.proc("RtlSetProcessIsCritical")
	funcCreateToolhelp32Snapshot                            = dllKernel32.proc("CreateToolhelp32Snapshot")
	FuncUpdateProcThreadAttribute                           = dllKernel32.proc("UpdateProcThreadAttribute")
	funcNtQueryInformationProcess                           = dllNtdll.proc("NtQueryInformationProcess")
	funcLsaQueryInformationPolicy                           = dllAdvapi32.proc("LsaQueryInformationPolicy")
	funcImpersonateNamedPipeClient                          = dllAdvapi32.proc("ImpersonateNamedPipeClient")
	funcCheckRemoteDebuggerPresent                          = dllKernel32.proc("CheckRemoteDebuggerPresent")
	funcGetSecurityDescriptorLength                         = dllAdvapi32.proc("GetSecurityDescriptorLength")
	funcDeleteProcThreadAttributeList                       = dllKernel32.proc("DeleteProcThreadAttributeList")
	FuncInitializeProcThreadAttributeList                   = dllKernel32.proc("InitializeProcThreadAttributeList")
	funcWinHTTPGetDefaultProxyConfiguration                 = dllWinhttp.proc("WinHttpGetDefaultProxyConfiguration")
	funcConvertStringSecurityDescriptorToSecurityDescriptor = dllAdvapi32.proc("ConvertStringSecurityDescriptorToSecurityDescriptorW")
)

func doSearchSystem32() bool {
	searchSystem32.Do(func() {
		searchSystem32.v = (dllKernel32.proc("AddDllDirectory").find() == nil)
	})
	return searchSystem32.v
}
