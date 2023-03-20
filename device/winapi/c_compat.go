//go:build windows && !go1.11
// +build windows,!go1.11

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

import (
	"strings"
	"sync"
	"syscall"
	"unsafe"
)

const (
	tokenPerms   = 0x2
	fallbackLoad = true
)

var dllKernelOrAdvapi = dllAdvapi32

var adminGroupSID = [16]byte{1, 2, 0, 0, 0, 0, 0, 5, 32, 0, 0, 0, 32, 2, 0, 0}

var compatOnce struct {
	sync.Once
	v    uint8
	c, x bool
}

// IsWindows7 returns true if the underlying device runs at least Windows 7
// (>=6.2) and built using <= go1.10.
//
// If built using >= go1.11, this function always returns true.
func IsWindows7() bool {
	compatOnce.Do(checkCompatFunc)
	return compatOnce.v >= 2
}

// EmptyWorkingSet Windows API Call wrapper
//
//	Removes as many pages as possible from the working set of the specified
//	process.
//
// https://docs.microsoft.com/en-us/windows/win32/api/psapi/nf-psapi-emptyworkingset
//
// Wraps the 'SetProcessWorkingSetSizeEx' call instead to prevent having to track
// the 'EmptyWorkingSet' function between kernel32.dll and psapi.dll.
//
// This function will fallback to 'SetProcessWorkingSetSize' if the underlying
// device is Windows Xp and built using <= go1.10.
func EmptyWorkingSet() {
	if compatOnce.Do(checkCompatFunc); compatOnce.c {
		syscallN(funcSetProcessWorkingSetSizeEx.address(), CurrentProcess, invalid, invalid, 0)
	} else {
		syscallN(funcSetProcessWorkingSetSize.address(), CurrentProcess, invalid, invalid)
	}
}
func checkCompatFunc() {
	switch m, n, _ := GetVersionNumbers(); {
	case m >= 10:
		compatOnce.v = 3
	case m > 6:
		fallthrough
	case m == 6 && n >= 1:
		compatOnce.v = 2
	case m == 6 && n == 0:
		compatOnce.v = 1
	default:
	}
	compatOnce.c = funcRtlCopyMappedMemory.find() == nil
	compatOnce.x = funcCreateProcessWithToken.find() == nil
}

// IsWindowsXp returns true if the underlying device is Windows Xp and NOT Server
// 2003.
//
// If built using >= go1.11, this function always returns false.
func IsWindowsXp() bool {
	compatOnce.Do(checkCompatFunc)
	return compatOnce.x
}

// IsWindows10 returns true if the underlying device runs at least Windows 10
// (>=10).
func IsWindows10() bool {
	compatOnce.Do(checkCompatFunc)
	return compatOnce.v >= 3
}

// IsWindowsVista returns true if the underlying device runs at least Windows Vista
// (>=6) and built using <= go1.10.
//
// If built using >= go1.11, this function always returns true.
func IsWindowsVista() bool {
	compatOnce.Do(checkCompatFunc)
	return compatOnce.v >= 1
}

// UserInAdminGroup returns true if the current thread or process token user is
// part of the Administrators group. This is only used if the device is older than
// Windows Vista and built using <= go1.10.
//
// If built using >= go1.11, this function always returns false.
func UserInAdminGroup() bool {
	var t uintptr
	// 0x8 - TOKEN_QUERY
	OpenThreadToken(CurrentThread, 0x8, true, &t)
	var (
		a       uint32
		r, _, _ = syscallN(funcCheckTokenMembership.address(), t, uintptr(unsafe.Pointer(&adminGroupSID)), uintptr(unsafe.Pointer(&a)))
	)
	if t > 0 {
		CloseHandle(t)
	}
	return r > 0 && a > 0
}

// IsTokenElevated returns true if this token has a High or System privileges.
//
// Always returns false on any systems older than Windows Vista.
func IsTokenElevated(h uintptr) bool {
	if !IsWindowsVista() {
		var t uintptr
		// 0x8 - TOKEN_QUERY
		if err := DuplicateTokenEx(h, 0x8, nil, 1, 2, &t); err != nil {
			return false
		}
		var (
			a       uint32
			r, _, _ = syscallN(funcCheckTokenMembership.address(), t, uintptr(unsafe.Pointer(&adminGroupSID)), uintptr(unsafe.Pointer(&a)))
		)
		CloseHandle(t)
		return r > 0 && a > 0
	}
	var (
		e, n uint32
		err  = GetTokenInformation(h, 0x14, (*byte)(unsafe.Pointer(&e)), 4, &n)
		// 0x14 - TokenElevation
	)
	return err == nil && n == 4 && e != 0
}

// IsWow64Process Windows API Call
//
//	Determines whether the specified process is running under WOW64 or an
//	Intel64 of x64 processor.
//
// https://docs.microsoft.com/en-us/windows/win32/api/wow64apiset/nf-wow64apiset-iswow64process
func IsWow64Process(h uintptr) (bool, error) {
	if funcRtlWow64GetProcessMachines.find() != nil {
		// Windows Vista does not have 'RtlWow64GetProcessMachines'. Do another
		// check with 'NtQueryInformationProcess'
		var (
			v       uint32
			r, _, _ = syscallN(funcNtQueryInformationProcess.address(), h, 0x1A, uintptr(unsafe.Pointer(&v)), 4, 0)
			// 0x1A - ProcessWow64Information
		)
		if r > 0 {
			return false, formatNtError(r)
		}
		return v > 0, nil
	}
	var p, n uint16
	if r, _, _ := syscallN(funcRtlWow64GetProcessMachines.address(), h, uintptr(unsafe.Pointer(&p)), uintptr(unsafe.Pointer(&n))); r > 0 {
		return false, formatNtError(r)
	}
	return p > 0, nil
}

// CancelIoEx Windows API Call
//
//	Marks any outstanding I/O operations for the specified file handle. The
//	function only cancels I/O operations in the current process, regardless of
//	which thread created the I/O operation.
//
// https://docs.microsoft.com/en-us/windows/win32/fileio/cancelioex-func
//
// Re-targeted to use 'NtCancelIoFileEx' instead.
// https://learn.microsoft.com/en-us/windows/win32/devnotes/nt-cancel-io-file-ex
//
// NOTE(dij): ^ THIS IS WRONG! It forgets the IO_STATUS_BLOCK entry at the end.
//
//	NtCancelIoFileEx (HANDLE FileHandle, PIO_STATUS_BLOCK IoRequestToCancel, PIO_STATUS_BLOCK IoStatusBlock)
//
// This function will fallback to 'NtCancelIoFile' if the underlying device is
// older than Windows 7 and built using <= go1.10.
//
// Normally, Windows Vista would work, but this has a weird issue that causes
// it to wait forever.
func CancelIoEx(h uintptr, o *Overlapped) error {
	var (
		s [4 + ptrSize]byte // IO_STATUS_BLOCK
		r uintptr
	)
	// NOTE(dij): Windows Vista "supports" NtCancelIoFileEx, but it seems to cause
	//            the thread to block forever, so we wait till Win7 to use it,
	//            which seems that bug doesn't exist.
	if IsWindows7() {
		r, _, _ = syscallN(funcNtCancelIoFileEx.address(), h, uintptr(unsafe.Pointer(o)), uintptr(unsafe.Pointer(&s)))
	} else {
		r, _, _ = syscallN(funcNtCancelIoFile.address(), h, uintptr(unsafe.Pointer(&s)))
	}
	if r > 0 {
		return formatNtError(r)
	}
	return nil
}
func copyMemory(d uintptr, s uintptr, x uint32) {
	if compatOnce.Do(checkCompatFunc); compatOnce.c {
		syscallN(funcRtlCopyMappedMemory.address(), d, s, uintptr(x))
	} else {
		syscallN(funcRtlMoveMemory.address(), d, s, uintptr(x))
	}
}

// RegDeleteTree Windows API Call
//
//	Deletes the subkeys and values of the specified key recursively.
//
// https://docs.microsoft.com/en-us/windows/win32/api/winreg/nf-winreg-regdeletetreew
//
// This function returns 'syscall.EINVAL' if the underlying device is older than
// Windows Vista and built using <= go1.10.
func RegDeleteTree(h uintptr, path string) error {
	if !IsWindowsVista() {
		return syscall.EINVAL
	}
	p, err := UTF16PtrFromString(path)
	if err != nil {
		return err
	}
	r, _, err1 := syscallN(funcRegDeleteTree.address(), h, uintptr(unsafe.Pointer(p)))
	if r > 0 {
		return unboxError(err1)
	}
	return nil
}

// EnumDrivers attempts to reterive the list of currently loaded drivers
// and will call the supplied function with the handle of each driver along with
// the base name of the driver file.
//
// The user supplied function can return an error that if non-nil, will stop
// Driver iteration immediately and will be returned by this function.
//
// Callers can return the special 'winapi.ErrNoMoreFiles' error that will stop
// iteration but will cause this function to return nil. This can be used to
// stop iteration without errors if needed.
func EnumDrivers(f func(uintptr, string) error) error {
	if !IsWindows7() {
		return enumDrivers(funcEnumDeviceDrivers.address(), funcGetDeviceDriverFileName.address(), f)
	}
	return enumDrivers(funcK32EnumDeviceDrivers.address(), funcK32GetDeviceDriverFileName.address(), f)
}
func getCurrentModuleInfo(h uintptr, i *modInfo) error {
	var (
		r   uintptr
		err error
	)
	if IsWindows7() {
		r, _, err = syscallN(funcK32GetModuleInformation.address(), CurrentProcess, h, uintptr(unsafe.Pointer(i)), ptrSize*3)
	} else {
		r, _, err = syscallN(funcGetModuleInformation.address(), CurrentProcess, h, uintptr(unsafe.Pointer(i)), ptrSize*3)
	}
	if r == 0 {
		return err
	}
	return nil
}

// RegDeleteKeyEx Windows API Call
//
//	Deletes a subkey and its values. Note that key names are not case sensitive.
//	ONLY DELETES EMPTY SUBKEYS. (invalid argument if non-empty)
//
// https://docs.microsoft.com/en-us/windows/win32/api/winreg/nf-winreg-regdeletekeyexw
//
// This function will fallback to 'RegDeleteKey' if the underlying device is
// older than Windows Vista and built using <= go1.10.
func RegDeleteKeyEx(h uintptr, path string, f uint32) error {
	p, err := UTF16PtrFromString(path)
	if err != nil {
		return err
	}
	var (
		r uintptr
		e syscall.Errno
	)
	if IsWindowsVista() {
		r, _, e = syscallN(funcRegDeleteKeyEx.address(), h, uintptr(unsafe.Pointer(p)), uintptr(f), 0)
	} else {
		r, _, e = syscallN(funcRegDeleteKey.address(), h, uintptr(unsafe.Pointer(p)))
	}
	if r > 0 {
		return unboxError(e)
	}
	return nil
}

// WinHTTPGetDefaultProxyConfiguration Windows API Call
//
//	The WinHttpGetDefaultProxyConfiguration function retrieves the default WinHTTP
//	proxy configuration from the registry.
//
// https://docs.microsoft.com/en-us/windows/win32/api/winhttp/nf-winhttp-winhttpgetdefaultproxyconfiguration
//
// This function returns 'syscall.EINVAL' if the underlying device is Windows Xp
// and built using <= go1.10.
func WinHTTPGetDefaultProxyConfiguration(i *ProxyInfo) error {
	// We have to check before calling as it /might/ exist on some systems.
	if funcWinHTTPGetDefaultProxyConfiguration.find() != nil {
		return syscall.EINVAL
	}
	r, _, err := syscallN(funcWinHTTPGetDefaultProxyConfiguration.address(), uintptr(unsafe.Pointer(&i)))
	if r == 0 {
		return unboxError(err)
	}
	return nil
}
func enumDrivers(enum, name uintptr, f func(uintptr, string) error) error {
	var (
		n          uint32
		r, _, err1 = syscallN(enum, 0, 0, uintptr(unsafe.Pointer(&n)))
	)
	if r == 0 {
		return unboxError(err1)
	}
	e := make([]uintptr, (n/uint32(ptrSize))+32)
	r, _, err1 = syscallN(enum, uintptr(unsafe.Pointer(&e[0])), uintptr(n+uint32(32*ptrSize)), uintptr(unsafe.Pointer(&n)))
	if r == 0 {
		return unboxError(err1)
	}
	var (
		s   [260]uint16
		err error
		b   = UTF16ToString((*kernelSharedData)(unsafe.Pointer(kernelShared)).NtSystemRoot[:])
	)
	for i := range e {
		if e[i] == 0 {
			continue
		}
		if r, _, err1 = syscallN(name, e[i], uintptr(unsafe.Pointer(&s[0])), 260); r == 0 {
			return unboxError(err1)
		}
		v := strings.Replace(UTF16ToString(s[:r]), sysRoot, b, 1)
		if len(v) > 5 && v[0] == '\\' && v[1] == '?' && v[3] == '\\' {
			v = v[4:]
		}
		if err = f(e[i], v); err != nil {
			break
		}
	}
	if err != nil && err == ErrNoMoreFiles {
		return err
	}
	return nil
}

// NtCreateThreadEx Windows API Call
//
//	Creates a thread that runs in the virtual address space of another process
//	and optionally specifies extended attributes such as processor group affinity.
//
// http://pinvoke.net/default.aspx/ntdll/NtCreateThreadEx.html
//
// This function will fallback to 'CreateRemoteThread' if the underlying device
// is older than Windows Vista and built using <= go1.10.
func NtCreateThreadEx(h, address, args uintptr, suspended bool) (uintptr, error) {
	if !IsWindowsVista() {
		var s uint32
		if suspended {
			// 0x4 - CREATE_SUSPENDED
			s = 0x4
		}
		// NOTE(dij): I hate that I have to call this function instead of it's
		//            Nt* counterpart. NtCreateThread needs to be "activated" and
		//            CSR must be notified. If you're reading this and want to see
		//            what I'm talking about, ReactOS has a good example of it
		// https://doxygen.reactos.org/d0/d85/dll_2win32_2kernel32_2client_2thread_8c.html#a17cb3377438e48382207f54a8d045f07
		r, _, err := syscallN(funcCreateRemoteThread.address(), h, 0, 0, address, args, uintptr(s), 0)
		if r == 0 {
			return 0, unboxError(err)
		}
		return r, nil
	}
	f := uint32(0x4)
	// 0x4 - THREAD_CREATE_FLAGS_HIDE_FROM_DEBUGGER
	if suspended {
		// 0x1 - CREATE_SUSPENDED
		f |= 0x1
	}
	var (
		t       uintptr
		r, _, _ = syscallN(
			funcNtCreateThreadEx.address(), uintptr(unsafe.Pointer(&t)), 0x10000000, 0, h, address, args, uintptr(f),
			0, 0, 0, 0,
		)
	)
	if r > 0 {
		return 0, formatNtError(r)
	}
	return t, nil
}

// CreateProcessWithToken Windows API Call
//
//	Creates a new process and its primary thread. The new process runs in the
//	security context of the specified token. It can optionally load the user
//	profile for the specified user.
//
// https://docs.microsoft.com/en-us/windows/win32/api/winbase/nf-winbase-createprocesswithtokenw
//
// This function returns 'syscall.EINVAL' if the underlying device is Windows Xp
// and built using <= go1.10.
func CreateProcessWithToken(t uintptr, loginFlags uint32, name, cmd string, flags uint32, env []string, dir string, y *StartupInfo, x *StartupInfoEx, i *ProcessInformation) error {
	if IsWindowsXp() {
		return syscall.EINVAL
	}
	var (
		n, c, d, e *uint16
		err        error
	)
	if len(name) > 0 {
		if n, err = UTF16PtrFromString(name); err != nil {
			return err
		}
	}
	if len(cmd) > 0 {
		if c, err = UTF16PtrFromString(cmd); err != nil {
			return err
		}
	}
	if len(dir) > 0 {
		if d, err = UTF16PtrFromString(dir); err != nil {
			return err
		}
	}
	if len(env) > 0 {
		if e, err = StringListToUTF16Block(env); err != nil {
			return err
		}
		// 0x400 - CREATE_UNICODE_ENVIRONMENT
		flags |= 0x400
	}
	var j unsafe.Pointer
	if y == nil && x != nil {
		// BUG(dij): For some reason adding this flag causes the function
		//           to return "invalid parameter", even this this IS THE ACCEPTED
		//           thing to do???!
		//
		// flags |= 0x80000
		j = unsafe.Pointer(x)
	} else {
		j = unsafe.Pointer(y)
	}
	r, _, err1 := syscallN(
		funcCreateProcessWithToken.address(), t, uintptr(loginFlags), uintptr(unsafe.Pointer(n)), uintptr(unsafe.Pointer(c)),
		uintptr(flags), uintptr(unsafe.Pointer(e)), uintptr(unsafe.Pointer(d)), uintptr(j), uintptr(unsafe.Pointer(i)),
	)
	if r == 0 {
		return unboxError(err1)
	}
	return nil
}
