//go:build windows && !crypt

package winapi

const debugPriv = "SeDebugPrivilege"

var (
	dllKernel32 = &lazyDLL{name: "kernel32.dll"}
	dllNtdll    = &lazyDLL{name: "ntdll.dll"}
	dllGdi32    = &lazyDLL{name: "gdi32.dll"}
	dllUser32   = &lazyDLL{name: "user32.dll"}
	dllWinhttp  = &lazyDLL{name: "winhttp.dll"}
	dllDbgHelp  = &lazyDLL{name: "DbgHelp.dll"}
	dllAdvapi32 = &lazyDLL{name: "advapi32.dll"}
)
