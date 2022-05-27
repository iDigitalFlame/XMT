//go:build windows && crypt

package winapi

import "github.com/iDigitalFlame/xmt/util/crypt"

var (
	debugPriv = crypt.Get(96) // SeDebugPrivilege

	dllKernel32 = &lazyDLL{name: crypt.Get(97)}  // kernel32.dll
	dllNtdll    = &lazyDLL{name: crypt.Get(98)}  // ntdll.dll
	dllGdi32    = &lazyDLL{name: crypt.Get(99)} // gdi32.dll
	dllUser32   = &lazyDLL{name: crypt.Get(100)} // user32.dll
	dllWinhttp  = &lazyDLL{name: crypt.Get(101)} // winhttp.dll
	dllDbgHelp  = &lazyDLL{name: crypt.Get(102)} // DbgHelp.dll
	dllAdvapi32 = &lazyDLL{name: crypt.Get(103)} // advapi32.dll
)
