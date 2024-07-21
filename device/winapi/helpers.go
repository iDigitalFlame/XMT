//go:build windows
// +build windows

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
	"errors"
	"io"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/iDigitalFlame/xmt/util/bugtrack"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

const (
	ptrSize      = unsafe.Sizeof(uintptr(0))
	kernelShared = uintptr(0x7FFE0000)
)

var caught struct{}

//go:linkname allm runtime.allm
var allm unsafe.Pointer

// We have this to be used to prevent crashing the stack of the program
// when we call minidump as we need to track extra parameters.
// The lock will stay enabled until it's done, so it's "thread safe".
var dumpStack dumpParam

//go:linkname setConsoleCtrlHandler runtime._SetConsoleCtrlHandler
var setConsoleCtrlHandler uintptr

var dumpCallbackOnce struct {
	_ [0]func()
	sync.Once
	f uintptr
}

type dumpCallback struct {
	// DO NOT REORDER
	Func uintptr
	Args uintptr
}
type kernelSharedData struct {
	// DO NOT REORDER
	_          [20]byte
	SystemTime struct {
		LowPart  uint32
		HighPart int32
		_        int32
	}
	_                     [16]byte
	NtSystemRoot          [260]uint16
	MaxStackTraceDepth    uint32
	_                     uint32
	_                     [32]byte
	NtBuildNumber         uint32
	_                     [8]byte
	NtMajorVersion        uint32
	NtMinorVersion        uint32
	ProcessorFeatures     [64]byte
	_                     [20]byte
	SystemExpirationDate  uint64
	_                     [4]byte
	KdDebuggerEnabled     uint8
	MitigationPolicies    uint8
	_                     [2]byte
	ActiveConsoleID       uint32
	_                     [12]byte
	NumberOfPhysicalPages uint32
	SafeBootMode          uint8
	VirtualizationFlags   uint8
	_                     [2]byte
	SharedDataFlags       uint32
}

// KillRuntime attempts to walk through the process threads and will forcefully
// kill all Golang based OS-Threads based on their starting address (which
// should be the same when starting from CGo).
//
// This will attempt to determine the base thread and any children that may be
// running and take action on what type of host we're in to best end the
// runtime without crashing.
//
// This function can be used on binaries, shared libraries or Zombified processes.
//
// DO NOT EXPECT ANYTHING (INCLUDING DEFERS) TO HAPPEN AFTER THIS FUNCTION.
func KillRuntime() {
	runtime.GC()
	debug.FreeOSMemory()
	runtime.LockOSThread()
	killRuntime()
	// Below shouldn't run.
	runtime.UnlockOSThread()
}
func killRuntime() {
	//// Workflow for killRuntime()
	//
	// 1 - Find the module that's us (this thread)
	// 2 - Find the base of the module we are in
	// 3 - Enumerate the runtime's M to find all open threads (finally)
	// 4 - Look through process threads to see if any other threads exist
	// 5 - Collect threads that exist in the base address space
	// > 6 - If we are in the base, it's a binary - syscall.Exit(0)
	// 7 - Check suspend cout of each thread in base address to see if we're a Zombie
	// > 8 - If only one thread in base address is suspended, we're a Zombie - syscall.Exit(0)
	// 9 - Iterate through all our threads and terminate them
	// > 0 - Terminate self thread
	//
	var q uintptr
	// 0x2 - GET_MODULE_HANDLE_EX_FLAG_UNCHANGED_REFCOUNT
	if r, _, err := syscallN(funcGetModuleHandleEx.address(), 0x2, 0, uintptr(unsafe.Pointer(&q))); r == 0 {
		if bugtrack.Enabled {
			bugtrack.Track("winapi.killRuntime(): GetModuleHandleEx failed err=%s", err.Error())
		}
		return
	}
	var k modInfo
	if err := getCurrentModuleInfo(q, &k); err != nil {
		if bugtrack.Enabled {
			bugtrack.Track("winapi.killRuntime(): GetModuleInformation failed err=%s", err.Error())
		}
		return
	}
	a, e := k.Base, k.Base+uintptr(k.Size)
	if bugtrack.Enabled {
		bugtrack.Track("winapi.killRuntime(): Module range a=%d, e=%d", a, e)
	}
	runtime.GC()
	debug.FreeOSMemory()
	var (
		x = make(map[uint32]struct{}, 8)
		g = make([]uintptr, 0, 8)
	)
	for i := uintptr(allm); ; {
		if h := *(*uintptr)(unsafe.Pointer(i + ptrThread)); h > 0 {
			if z, err := getThreadID(h); err == nil {
				if x[z] = caught; bugtrack.Enabled {
					bugtrack.Track("winapi.killRuntime(): Found runtime thread ID z=%d, h=%d", z, h)
				}
			}
			g = append(g, h)
		}
		n := (*uintptr)(unsafe.Pointer(i + ptrNext))
		if n == nil || *n == 0 {
			break // Reached bottom of linked list
		}
		i = *n
	}
	var (
		y = getCurrentThreadID()
		m = make([]uintptr, 0, len(x))
		b bool
		z uint8
	)
	err := EnumThreads(GetCurrentProcessID(), func(t ThreadEntry) error {
		// 0x43 - THREAD_QUERY_INFORMATION  | THREAD_SUSPEND_RESUME | THREAD_TERMINATE
		h, err1 := t.Handle(0x43)
		if err1 != nil {
			if bugtrack.Enabled {
				bugtrack.Track("winapi.killRuntime(): Thread failed to have it's handle opened t.TID=%d, err1=%s!", t.TID, err1)
			}
			// NOTE(dij): Workaround on attribute weirdness where we can see our
			//            threads, but we can't get a handle to them. If we are
			//            using QSI, we can see if it's suspended and we can add
			//            it to the Zombie flag as 9/10 it's a Zombie.
			//
			//            We also check to see if it's part of the runtime's threads
			//            as it shouldn't. This makes false-positives less likely.
			var (
				q, _  = t.IsSuspended()
				_, ok = x[t.TID]
			)
			if q && !ok {
				if z++; bugtrack.Enabled {
					bugtrack.Track("winapi.killRuntime(): Failed thread seems to be a Zombie thread t.TID=%d!", t.TID)
				}
			}
			return nil // Continue on handle errors instead of bailing.
		}
		s, err1 := getThreadStartAddress(h)
		if err1 != nil {
			return err1
		}
		if t.TID == y { // Skip the current thread
			if b = s >= a && s < e; b {
				return ErrNoMoreFiles
			}
			if bugtrack.Enabled {
				bugtrack.Track("winapi.killRuntime(): Found our thread t.TID=%d y=%d, s=%d, b=%t", t.TID, y, s, b)
			}
			return nil
		}
		k, err1 := t.suspended(h)
		if err1 != nil {
			return err1
		}
		if (s > a && s < e) && k && z < 0xFF { // Prevent overflow here
			z++
		}
		if _, ok := x[t.TID]; !ok {
			CloseHandle(h)
			return nil
		}
		m = append(m, h)
		return nil
	})
	if err != nil {
		if bugtrack.Enabled {
			bugtrack.Track("winapi.killRuntime(): EnumThreads failed err=%s", err.Error())
		}
		return
	}
	// Unmap all function mappings (if any)
	if FuncUnmapAll(); b || len(m) == 0 {
		for i := range m {
			CloseHandle(m[i])
		}
		if g, m = nil, nil; b {
			if bugtrack.Enabled {
				bugtrack.Track("winapi.killRuntime(): We're in the base thread, we can exit normally.")
			}
			// Base thread (us), is in the base module address
			// This is a binary, it's safe to exit cleanly.
			syscall.Exit(0)
			return
		}
		if bugtrack.Enabled {
			bugtrack.Track("winapi.killRuntime(): Failed to find base threads!")
		}
		return
	}
	if z == 1 {
		if bugtrack.Enabled {
			bugtrack.Track("winapi.killRuntime(): Zombie check passed z=%d", z)
		}
		for i := range m {
			CloseHandle(m[i])
		}
		if g, m = nil, nil; bugtrack.Enabled {
			bugtrack.Track("winapi.killRuntime(): We're a Zombie, we can exit normally.")
		}
		// Out of all the base threads, only one exists and is suspended,
		// 99% chance this is a Zombified process, it's ok to exit cleanly.
		syscall.Exit(0)
		return
	}
	if bugtrack.Enabled {
		bugtrack.Track("winapi.killRuntime(): Zombie check failed z=%d", z)
	}
	freeChunkHeap()
	// NOTE(dij): Potential footgun? Free all loaded libaries since we're leaving
	//            but not /exiting/. FreeLibrary shouldn't cause an issue as it
	//            /should/ only clean unused libraries after we are done.
	//            ntdll.dll will NOT be unloaded.
	freeLoadedLibaries()
	// Stop all running Goroutines
	stopTheWorld("exit")
	// Disable the CTRL console handler that Go sets.
	removeCtrlHandler()
	// What's left is that we're probally injected into memory somewhere, and
	// we just need to nuke the runtime without affecting the host.
	for i := range g {
		CloseHandle(g[i])
	}
	g = nil
	for i := range m {
		if err = TerminateThread(m[i], 0); err != nil {
			break
		}
	}
	// Close all timers and open handles
	// Even if the world is stopped, we still run into this occasionally. So it's
	// down here instead.
	destoryAllM()
	for i := range m {
		CloseHandle(m[i])
	}
	if m = nil; err != nil {
		if bugtrack.Enabled {
			bugtrack.Track("winapi.killRuntime(): Terminate error err=%s", err.Error())
		}
		return
	}
	if bugtrack.Enabled {
		bugtrack.Track("winapi.killRuntime(): Bye bye!")
	}
	EmptyWorkingSet()
	freeRuntimeMemory() // Buck Stops here.
}

// Getppid returns the Parent Process ID of this Process by reading the PEB.
// If this fails, this returns zero.
func Getppid() uint32 {
	var (
		p       processBasicInfo
		r, _, _ = syscallN(
			funcNtQueryInformationProcess.address(), CurrentProcess, 0, uintptr(unsafe.Pointer(&p)),
			unsafe.Sizeof(p), 0,
		)
	)
	if r > 0 {
		return 0
	}
	return uint32(p.InheritedFromUniqueProcessID)
}
func createDumpFunc() {
	dumpCallbackOnce.f = syscall.NewCallback(dumpCallbackFunc)
}

// InSafeMode returns true if the current device was booted into Safe Mode, false
// otherwise.
func InSafeMode() bool {
	return (*kernelSharedData)(unsafe.Pointer(kernelShared)).SafeBootMode > 0
}

// IsDebugged attempts to check multiple system calls in order to determine
// REAL debugging status.
//
// NOTE: Systems that are "Debug" / "Checked" versions of Windows will always
// return false!
//
// This function checks in this order:
//
//   - KSHARED.KdDebuggerEnabled
//   - KSHARED.SharedDataFlags.DbgErrorPortPresent
//   - NtQuerySystemInformation/SystemKernelDebuggerInformation
//   - IsDebuggerPresent (from PEB)
//   - NtGlobalFlag (from PEB)
//   - OutputDebugStringA
//   - CheckRemoteDebuggerPresent
//
// Errors make the function return false only if they are the last call.
func IsDebugged() bool {
	switch s := (*kernelSharedData)(unsafe.Pointer(kernelShared)); {
	case s.KdDebuggerEnabled > 1:
		return true
	case s.SharedDataFlags&0x1 != 0: // 0x1 - DbgErrorPortPresent
		// NOTE(dij): This returns true when on a Debug/Checked version on Windows.
		//            Not sure if we want to ignore this or not, but I doubt that
		//            actual systems are using "Multiprocessor Debug/Checked" unless
		//            the system is a driver test or builder.
		return true
	}
	var (
		d uint16
		x uint32
	)
	// 0x23 - SystemKernelDebuggerInformation
	syscallN(funcNtQuerySystemInformation.address(), 0x23, uintptr(unsafe.Pointer(&d)), 2, uintptr(unsafe.Pointer(&x)))
	// The SYSTEM_KERNEL_DEBUGGER_INFORMATION short offset 1 (last 8 bits) is not
	// filled out by systems older than Vista, so we ignore them.
	if x == 2 && ((d&0xFF) > 1 || ((d>>8) == 0 && IsWindowsVista())) {
		return true
	}
	switch p, err := getProcessPeb(); {
	case err != nil:
	case p.BeingDebugged > 0:
		return true
	case p.NtGlobalFlag&(0x70) != 0: // 0x70 - FLG_HEAP_ENABLE_TAIL_CHECK | FLG_HEAP_ENABLE_FREE_CHECK | FLG_HEAP_VALIDATE_PARAMETERS
		return true
	}
	o := [2]byte{'_', 0}
	// Take advantage of a "bug" in OutputDebugStringA where the "r2" return value
	// will NOT be zero when a debugger is present to receive the debug string.
	if _, r, _ := syscallN(funcOutputDebugString.address(), uintptr(unsafe.Pointer(&o[0]))); r > 0 {
		return true
	}
	// 0x400 - PROCESS_QUERY_INFORMATION
	h, err := OpenProcess(0x400, false, GetCurrentProcessID())
	if err != nil {
		return false
	}
	var v bool
	err = CheckRemoteDebuggerPresent(h, &v)
	CloseHandle(h)
	return err == nil && v
}

//go:linkname stopTheWorld runtime.stopTheWorld
func stopTheWorld(string)

// IsSystemEval returns true if the KSHARED_USER_DATA.SystemExpirationDate value
// is greater than zero.
//
// SystemExpirationDate is the time that remains in any evaluation copies of
// Windows. This can be used to find systems that may be used for testing and
// are not production machines.
func IsSystemEval() bool {
	return (*kernelSharedData)(unsafe.Pointer(kernelShared)).SystemExpirationDate > 0
}

// IsUACEnabled returns true if UAC (User Account Control) is enabled, false
// otherwise.
func IsUACEnabled() bool {
	// 0x2 - DbgElevationEnabled
	return (*kernelSharedData)(unsafe.Pointer(kernelShared)).SharedDataFlags&0x2 != 0
}
func freeLoadedLibaries() {
	dllAmsi.Free()
	dllGdi32.Free()
	dllUser32.Free()
	dllWinhttp.Free()
	dllDbgHelp.Free()
	dllAdvapi32.Free()
	dllWtsapi32.Free()
	dllKernel32.Free()
	dllKernelBase.Free()
}

// ErasePEHeader erases the first page of the mapped PE memory data. This is
// recommended to ONLY use when using a shipped binary.
//
// Any errors found during zeroing will returned.
//
// Retrieved from: https://github.com/LordNoteworthy/al-khaser/blob/master/al-khaser/AntiDump/ErasePEHeaderFromMemory.cpp
func ErasePEHeader() error {
	var (
		h          uintptr
		r, _, err1 = syscallN(funcGetModuleHandleEx.address(), 0x2, 0, uintptr(unsafe.Pointer(&h)))
		// 0x2 - GET_MODULE_HANDLE_EX_FLAG_UNCHANGED_REFCOUNT
	)
	if r == 0 {
		return unboxError(err1)
	}
	var (
		n      = uint32(syscall.Getpagesize())
		o, err = NtProtectVirtualMemory(CurrentProcess, h, n, 0x40)
		// 0x40 - PAGE_EXECUTE_READWRITE
	)
	if err != nil {
		return err
	}
	for i := uint32(0); i < n; i++ {
		(*(*[1]byte)(unsafe.Pointer(h + uintptr(i))))[0] = 0
	}
	_, err = NtProtectVirtualMemory(CurrentProcess, h, n, o)
	return err
}
func (p *dumpParam) close() {
	heapFree(p.b, p.h)
	heapDestroy(p.b)
	p.Unlock()
}

// Untrust will attempt to revoke all Token permissions and change the Token
// integrity level to "Untrusted".
//
// This effectively revokes all permissions for the application with the supplied
// PID to run.
//
// Ensure a call to 'GetDebugPrivilege' is made first before starting.
//
// Thanks for the find by @zha0gongz1 in their article:
//
//	https://golangexample.com/without-closing-windows-defender-to-make-defender-useless-by-removing-its-token-privileges-and-lowering-the-token-integrity/
func Untrust(p uint32) error {
	// 0x400 - PROCESS_QUERY_INFORMATION
	h, err := OpenProcess(0x400, false, p)
	if err != nil {
		return err
	}
	var t uintptr
	// 0x200A8 - TOKEN_READ | TOKEN_ADJUST_PRIVILEGES | TOKEN_ADJUST_DEFAULT | TOKEN_QUERY
	if err = OpenProcessToken(h, 0x200A8, &t); err != nil {
		CloseHandle(h)
		return err
	}
	var n uint32
	// 0x3 - TokenPrivileges
	if err = GetTokenInformation(t, 0x3, nil, 0, &n); n == 0 {
		CloseHandle(h)
		CloseHandle(t)
		return err
	}
	b := make([]byte, n)
	// 0x3 - TokenPrivileges
	if err = GetTokenInformation(t, 0x3, &b[0], n, &n); err != nil {
		CloseHandle(h)
		CloseHandle(t)
		return err
	}
	_ = b[n-1]
	// NOTE(dij): Loop over all the privileges and disable them. Yes we
	//            call "disableAll", but this is a failsafe.
	for c, i, a := uint32(b[3])<<24|uint32(b[2])<<16|uint32(b[1])<<8|uint32(b[0]), uint32(12), uint32(0); a < c && i < n; a, i = a+1, i+12 {
		b[i], b[i+1], b[i+2], b[i+3] = 0x4, 0, 0, 0
	}
	if err = AdjustTokenPrivileges(t, false, unsafe.Pointer(&b[0]), n, nil, nil); err != nil {
		CloseHandle(h)
		CloseHandle(t)
		return err
	}
	// We don't care if this errors.
	if AdjustTokenPrivileges(t, true, nil, 0, nil, nil); !IsWindowsVista() {
		CloseHandle(h)
		CloseHandle(t)
		return nil
	}
	var (
		c = uint32(32)
		s [32]byte
	)
	// 0x41 - WinUntrustedLabelSid
	r, _, err1 := syscallN(funcCreateWellKnownSid.address(), 0x41, 0, uintptr(unsafe.Pointer(&s[0])), uintptr(unsafe.Pointer(&c)))
	if r == 0 {
		CloseHandle(h)
		CloseHandle(t)
		return unboxError(err1)
	}
	var x SIDAndAttributes
	// 0x20 - SE_GROUP_INTEGRITY
	x.Sid, x.Attributes = (*SID)(unsafe.Pointer(&s[0])), 0x20
	// 0x19 - TokenIntegrityLevel
	r, _, _ = syscallN(funcNtSetInformationToken.address(), t, 0x19, uintptr(unsafe.Pointer(&x)), uintptr(c+4))
	CloseHandle(h)
	if CloseHandle(t); r == 0 {
		return nil
	}
	return formatNtError(r)
}

// SystemDirectory Windows API Call
//
//	Retrieves the path of the system directory. The system directory contains
//	system files such as dynamic-link libraries and drivers.
//
// https://docs.microsoft.com/en-us/windows/win32/api/sysinfoapi/nf-sysinfoapi-getsystemdirectoryw
//
// Technically a link to the runtime "GetSystemDirectory" cached API call.
func SystemDirectory() string {
	return systemDirectoryPrefix
}

// GetDebugPrivilege is a quick helper function that will attempt to grant the
// caller the "SeDebugPrivilege" privilege.
func GetDebugPrivilege() error {
	var (
		t   uintptr
		err = OpenProcessToken(CurrentProcess, 0x200E8, &t)
		// 0x200E8 - TOKEN_READ (STANDARD_RIGHTS_READ | TOKEN_QUERY) | TOKEN_WRITE
		//            (TOKEN_ADJUST_PRIVILEGES | TOKEN_ADJUST_GROUPS | TOKEN_ADJUST_DEFAULT)
	)
	if err != nil {
		return err
	}
	var p privileges
	if err = LookupPrivilegeValue("", debugPriv, &p.Privileges[0].Luid); err != nil {
		CloseHandle(t)
		return err
	}
	p.Privileges[0].Attributes, p.PrivilegeCount = 0x2, 1 // SE_PRIVILEGE_ENABLED
	err = AdjustTokenPrivileges(t, false, unsafe.Pointer(&p), uint32(unsafe.Sizeof(p)), nil, nil)
	CloseHandle(t)
	return err
}
func fullPath(n string) string {
	if !isBaseName(n) {
		return n
	}
	return systemDirectoryPrefix + n
}

// IsUTCTime checks the current system TimeZone information to see if the device
// is set to the UTC time zone. Most systems in debugging/logging environments will
// have this set.
//
// This function detects UTC as it's biases are always zero and is the only time
// zone that has this feature.
func IsUTCTime() (bool, error) {
	var (
		t       timeZoneInfo
		n       int
		r, _, _ = syscallN(funcNtQuerySystemInformation.address(), 0x2C, uintptr(unsafe.Pointer(&t)), 172, uintptr(unsafe.Pointer(&n)))
		// 0x2C - SystemCurrentTimeZoneInformation
	)
	if r > 0 {
		return false, formatNtError(r)
	}
	return t.Bias == 0 && t.DaylightBias == 0 && t.StdBias == 0, nil
}

// GetKernelTime returns the system time based on the KSHARED_USER_DATA struct in
// memory that is converted to a time.Time struct.
//
// This can be used to get the system time without relying on any API calls.
//
// NOTE(dij): Supposedly Go already reads this for 'time.Now()'?
func GetKernelTime() time.Time {
	var (
		s = (*kernelSharedData)(unsafe.Pointer(kernelShared))
		t = time.Unix(0, ((int64(s.SystemTime.HighPart)<<32|int64(s.SystemTime.LowPart))-epoch)*100)
	)
	return t
}
func getCurrentThreadID() uint32 {
	r, _, _ := syscallN(funcGetCurrentThreadID.address())
	return uint32(r)
}
func (p *dumpParam) init() error {
	p.Lock()
	var err error
	// 2 << 20 = ~20MB
	if p.b, err = heapCreate(2 << 20); err != nil {
		return err
	}
	if p.h, err = heapAlloc(p.b, 2<<20, true); err != nil {
		heapDestroy(p.b)
		return err
	}
	p.s, p.w = 2<<20, 0
	dumpCallbackOnce.Do(createDumpFunc)
	return nil
}

// LoadLibraryAddress is a simple function that returns the raw address of the
// 'LoadLibraryW' function in 'kernel32.dll' that's currently loaded.
func LoadLibraryAddress() uintptr {
	return funcLoadLibrary
}

// IsStackTracingEnabled returns true if the KSHARED_USER_DATA.MaxStackTraceDepth
// value is greater than zero.
//
// MaxStackTraceDepth is a value that represents the stack trace depth if tracing
// is enabled. If this flag is greater than zero, it is likely that some form of
// debug tracing is enabled.
func IsStackTracingEnabled() bool {
	return (*kernelSharedData)(unsafe.Pointer(kernelShared)).MaxStackTraceDepth > 0
}

// GetSystemSID will attempt to determine the System SID value and return it.
func GetSystemSID() (*SID, error) {
	var (
		o lsaAttributes
		h uintptr
	)
	o.Length = uint32(unsafe.Sizeof(o))
	r, _, err := syscallN(funcLsaOpenPolicy.address(), 0, uintptr(unsafe.Pointer(&o)), 1, uintptr(unsafe.Pointer(&h)))
	if r > 0 {
		return nil, unboxError(err)
	}
	i := new(lsaAccountDomainInfo)
	r, _, err = syscallN(funcLsaQueryInformationPolicy.address(), h, 5, uintptr(unsafe.Pointer(&i)))
	if syscallN(funcLsaClose.address(), h); r > 0 {
		return nil, unboxError(err)
	}
	// TODO(dij): There is a memory leak here!
	//            Need to call 'localFree' with the ptr to 'i'.
	return i.SID, nil
}
func heapFree(h, m uintptr) error {
	r, _, err := syscallN(funcRtlFreeHeap.address(), h, 0, m)
	if r == 0 {
		return unboxError(err)
	}
	return nil
}
func heapDestroy(h uintptr) error {
	r, _, err := syscallN(funcRtlDestroyHeap.address(), h)
	if r == 0 {
		return unboxError(err)
	}
	return nil
}

// SetWallpaper uses the 'SystemParametersInfo' API call to set the user's
// wallpaper. Changes take effect immediately.
//
// https://docs.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-systemparametersinfoa
func SetWallpaper(s string) error {
	v, err := UTF16PtrFromString(s)
	if err != nil {
		return err
	}
	// 0x14 - SPI_SETDESKWALLPAPER
	r, _, err1 := syscallN(funcSystemParametersInfo.address(), 0x14, 1, uintptr(unsafe.Pointer(v)), 0x3)
	if r == 0 {
		return unboxError(err1)
	}
	return nil
}

// SetHighContrast uses the 'SystemParametersInfo' API call to trigger the
// HighContrast theme setting. Set to 'True' to enable it and 'False' to disbale
// it.
//
// https://docs.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-systemparametersinfoa
func SetHighContrast(e bool) error {
	var c highContrast
	if c.Size = uint32(unsafe.Sizeof(c)); e {
		c.Flags = 1
	}
	// 0x43 - SPI_SETHIGHCONTRAST
	r, _, err := syscallN(funcSystemParametersInfo.address(), 0x43, 0, uintptr(unsafe.Pointer(&c)), 0x3)
	if r == 0 {
		return unboxError(err)
	}
	return nil
}

// InWow64Process is a helper function that just calls'IsWow64Process' with the
// 'CurrentProcess' handle to determine if the current process is a WOW64 process.
func InWow64Process() (bool, error) {
	return IsWow64Process(CurrentProcess)
}

/*
// IsVirtualizationEnabled return true if the current device processor has the
// PF_VIRT_FIRMWARE_ENABLED flag. This will just indicate if the device has the
// capability to run Virtual Machines. This is commonly not the case of many VMs
// themselves.
//
// Be cautious, as many hypervisors have the ability to still expose this CPU
// flag to guests.
//
// This flag is grabbed from KSHARED_USER_DATA.
func IsVirtualizationEnabled() bool {
	// 0x15 - PF_VIRT_FIRMWARE_ENABLED
	return (*kernelSharedData)(unsafe.Pointer(kernelShared)).ProcessorFeatures[0x15] > 0
}*/

// SetCommandLine will attempt to read the Process PEB and overrite the
// 'ProcessParameters.CommandLine' property with the supplied string value.
//
// This will NOT change the ImagePath or Binary Name.
//
// This will return any errors that occur during reading the PEB.
//
// DOES NOT WORK ON WOW6432 PEBs!
//   - These are in a separate memory space and seem to only be read once? or the
//     data is copied somewhere else. Even if I call 'NtWow64QueryInformationProcess64'
//     and change it, it does NOT seem to care. *shrug* who TF uses x86 anyway in 2022!?
//
// TODO(dij): Since we have backwards compatibility now. The 32bit PEB can be read
// using NtQueryInformationProcess/ProcessWow64Information which returns
// 32bit pointer to the PEB in 32bit mode.
func SetCommandLine(s string) error {
	c, err := UTF16FromString(s)
	if err != nil {
		return err
	}
	p, err := getProcessPeb()
	if err != nil {
		return err
	}
	p.ProcessParameters.CommandLine.Buffer = &c[0]
	p.ProcessParameters.CommandLine.Length = uint16(len(c)*2) - 1
	p.ProcessParameters.CommandLine.MaximumLength = p.ProcessParameters.CommandLine.Length
	return nil
}

// SwapMouseButtons uses the 'SystemParametersInfo' API call to trigger the
// swapping of the left and right mouse buttons. Set to 'True' to swap and
// 'False' to disable it.
//
// https://docs.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-systemparametersinfoa
func SwapMouseButtons(e bool) error {
	var v uint32
	if e {
		v = 1
	}
	// 0x21 - SPI_SETMOUSEBUTTONSWAP
	r, _, err := syscallN(funcSystemParametersInfo.address(), 0x21, uintptr(v), 0, 0x3)
	if r == 0 {
		return unboxError(err)
	}
	return nil
}
func formatNtError(e uintptr) error {
	// NOTE(dij): Not loading NTDLL here as we /should/ already have loaded it
	//            as we're calling this function due to an Nt* function error
	//            status. If not, this just acts like a standard 'FormatMessage'
	//            call.
	var (
		o       [300]uint16
		r, _, _ = syscallN(funcFormatMessage.address(), 0x3A00, dllNtdll.addr, e, 0x409, uintptr(unsafe.Pointer(&o)), 0x12C, 0)
		// 0x3A00 - FORMAT_MESSAGE_ARGUMENT_ARRAY | FORMAT_MESSAGE_FROM_HMODULE |
		//          FORMAT_MESSAGE_FROM_SYSTEM | FORMAT_MESSAGE_IGNORE_INSERTS
		// 0x409  - English LANG and English SUB
	)
	if r == 0 {
		return syscall.Errno(e)
	}
	v := r
	// Remove newline at the end
	for ; r > 0; r-- {
		if o[r] == '\n' || o[r] == '\r' {
			if r > 1 && (o[r-1] == '\n' || o[r-1] == '\r') {
				r--
			}
			break
		}
	}
	// CAan't find it? Just return what we have.
	if r == 0 {
		return errors.New(UTF16ToString(o[:v]))
	}
	// Remove prepended "{TYPE}" string
	if o[0] == '{' {
		for i := uintptr(1); i < r; i++ {
			if o[i] == '\n' || o[i] == '\r' {
				if i+1 < r && (o[i+1] == '\n' || o[i+1] == '\r') {
					i++
				}
				return errors.New(UTF16ToString(o[i+1 : r]))
			}
		}
	}
	return errors.New(UTF16ToString(o[:r]))
}

// GetLocalUser attempts to return the username associated with the current Thread
// or Process.
//
// This function will first check if the Thread is using a Token (Impersonation)
// and if not it will then pull the Token for the Process instead.
//
// This function will concationate the domain (or local workstation) name if the
// Token provides one.
//
// If any errors occur, an empty string with the error will be returned.
func GetLocalUser() (string, error) {
	var t uintptr
	// 0x20008 - TOKEN_READ | TOKEN_QUERY
	if err := OpenThreadToken(CurrentThread, 0x20008, true, &t); err != nil {
		if err = OpenProcessToken(CurrentProcess, 0x20008, &t); err != nil {
			return "", err
		}
	}
	u, err := UserFromToken(t)
	if CloseHandle(t); err != nil {
		return "", err
	}
	return u, nil
}

// CheckDebugWithLoad will attempt to check for a debugger by loading a non-loaded
// DLL specified and will check for exclusive access (which is false for debuggers).
//
// If the file can be opened, the library is freed and the file is closed. This
// will return true ONLY if opening for exclusive access fails.
//
// Any errors opening or loading DLLs will silently return false.
func CheckDebugWithLoad(d string) bool {
	var (
		p      = fullPath(d)
		n, err = UTF16PtrFromString(p)
	)
	if err != nil {
		return false
	}
	var (
		h       uintptr
		r, _, _ = syscallN(funcGetModuleHandleEx.address(), 0x2, uintptr(unsafe.Pointer(n)), uintptr(unsafe.Pointer(&h)))
		// 0x2 - GET_MODULE_HANDLE_EX_FLAG_UNCHANGED_REFCOUNT
	)
	if r > 0 {
		return false
	}
	if h, err = loadLibraryEx(p); err != nil || h == 0 {
		return false
	}
	// 0x80000000 - FILE_FLAG_WRITE_THROUGH
	// 0x0        - EXCLUSIVE
	// 0x3        - OPEN_EXISTING
	f, err := CreateFile(p, 0x80000000, 0, nil, 0x3, 0, 0)
	if syscall.FreeLibrary(syscall.Handle(h)); err != nil {
		return err.(syscall.Errno) != 0x2
	}
	CloseHandle(f)
	return false
}

// IsUserNetworkToken will return true if the origin of the Token was a LoginUser
// network impersonation API call and NOT a duplicated Token via Token or Thread
// impersonation.
func IsUserNetworkToken(t uintptr) bool {
	if t == 0 {
		return false
	}
	var (
		n   uint32
		b   [16]byte
		err = GetTokenInformation(t, 0x7, &b[0], 16, &n)
		// 0x7 - TokenSource
	)
	if err != nil {
		return false
	}
	// Match [65 100 118 97 112 105 32 32] == "Advapi"
	return b[0] == 65 && b[1] == 100 && b[6] == 32 && b[7] == 32
}

// EnablePrivileges will attempt to enable the supplied Windows privilege values
// on the current process's Token.
//
// Errors during encoding, lookup or assignment will be returned and not all
// privileges will be assigned, if they occur.
func EnablePrivileges(s ...string) error {
	if len(s) == 0 {
		return nil
	}
	var (
		t   uintptr
		err = OpenProcessToken(CurrentProcess, 0x200E8, &t)
		// 0x200E8 - TOKEN_READ (STANDARD_RIGHTS_READ | TOKEN_QUERY) | TOKEN_WRITE
		//            (TOKEN_ADJUST_PRIVILEGES | TOKEN_ADJUST_GROUPS | TOKEN_ADJUST_DEFAULT)
	)
	if err != nil {
		return xerr.Wrap("OpenProcessToken", err)
	}
	err = EnableTokenPrivileges(t, s...)
	CloseHandle(t)
	return err
}

// IsSecureBootEnabled returns true if Secure Boot is enabled in the current device.
//
// This function returns true or false and any errors that may occur during checking
// for secure boot.
func IsSecureBootEnabled() (bool, error) {
	var (
		i       uint16
		n       uint32
		r, _, _ = syscallN(funcNtQuerySystemInformation.address(), 0x91, uintptr(unsafe.Pointer(&i)), 2, uintptr(unsafe.Pointer(&n)))
		// 0x91 - SystemSecureBootInformation
	)
	if r > 0 {
		return false, formatNtError(r)
	}
	return (i & 0xFF) == 1, nil
}

// SetAllThreadsToken sets the Token for all current Golang threads. This is an
// easy way to do thread impersonation across the entire runtime.
//
// Calls 'ForEachThread' -> 'SetThreadToken' under the hood.
func SetAllThreadsToken(h uintptr) error {
	return ForEachThread(func(t uintptr) error { return SetThreadToken(t, h) })
}
func getProcessPeb() (*processPeb, error) {
	/*
		PVOID64 GetPeb64()
		{
			PVOID64 peb64 = NULL;

			if (API::IsAvailable(API_IDENTIFIER::API_NtWow64QueryInformationProcess64))
			{
				PROCESS_BASIC_INFORMATION_WOW64 pbi64 = {};

				auto NtWow64QueryInformationProcess64 = static_cast<pNtWow64QueryInformationProcess64>(API::GetAPI(API_IDENTIFIER::API_NtWow64QueryInformationProcess64));
				NTSTATUS status = NtWow64QueryInformationProcess64(GetCurrentProcess(), ProcessBasicInformation, &pbi64, sizeof(pbi64), nullptr);
				if ( NT_SUCCESS ( status ) )
					peb64 = pbi64.PebBaseAddress;
			}

			return peb64;
		}
	*/
	var (
		p       processBasicInfo
		r, _, _ = syscallN(
			funcNtQueryInformationProcess.address(), CurrentProcess, 0, uintptr(unsafe.Pointer(&p)),
			unsafe.Sizeof(p), 0,
		)
	)
	if r > 0 {
		return nil, formatNtError(r)
	}
	return (*processPeb)(unsafe.Pointer(p.PebBaseAddress)), nil
}

// ImpersonatePipeToken will attempt to impersonate the Token used by the Named
// Pipe client.
//
// This function is only usable on Windows with a Server Pipe handle.
//
// BUG(dij): I'm not sure if this is broken or this is how it's handled. I'm
//
//	getting error 5.
//
// Pipe insights:
//
//	https://papers.vx-underground.org/papers/Windows/System%20Components%20and%20Abuse/Offensive%20Windows%20IPC%20Internals%201%20Named%20Pipes.pdf
func ImpersonatePipeToken(h uintptr) error {
	// NOTE(dij): For best results, we FIRST impersonate the token, THEN
	//            we try to set the token to each user thread with a duplicated
	//            token set to impersonate. (Similar to an Impersonate call).
	runtime.LockOSThread()
	if err := ImpersonateNamedPipeClient(h); err != nil {
		runtime.UnlockOSThread()
		return err
	}
	var y uintptr
	// 0xF01FF - TOKEN_ALL_ACCESS
	if err := OpenThreadToken(CurrentThread, 0xF01FF, true, &y); err != nil {
		runtime.UnlockOSThread()
		return err
	}
	err := SetAllThreadsToken(y)
	CloseHandle(y)
	runtime.UnlockOSThread()
	return err
}
func heapCreate(n uint64) (uintptr, error) {
	// 0x1002 - MEM_COMMIT? | HEAP_GROWABLE
	r, _, err := syscallN(funcRtlCreateHeap.address(), 0x1002, 0, 0, uintptr(n), 0, 0)
	if r == 0 {
		return 0, unboxError(err)
	}
	return r, nil
}
func (p *dumpParam) resize(n uint64) error {
	if n < p.s {
		return nil
	}
	var (
		v      = (p.s + n) * 2
		h, err = heapReAlloc(p.b, p.h, v, false)
	)
	if err != nil {
		return err
	}
	p.h, p.s = h, v
	return nil
}

// PhysicalInfo will query the system using NtQuerySystemInformation to grab the
// number of CPUs installed and the current memory (in MB) that is avaliable to
// the system (installed physically).
func PhysicalInfo() (uint8, uint32, error) {
	var (
		n       uint64
		i       systemBasicInfo
		r, _, _ = syscallN(funcNtQuerySystemInformation.address(), 0x0, uintptr(unsafe.Pointer(&i)), unsafe.Sizeof(i), uintptr(unsafe.Pointer(&n)))
		// 0x0 - SystemBasicInformation
	)
	if r > 0 {
		return 0, 0, formatNtError(r)
	}
	return i.NumProc, uint32((uint64(i.PageSize)*uint64(i.PhysicalPages))/0x100000) + 1, nil
}
func getThreadID(h uintptr) (uint32, error) {
	var (
		t       threadBasicInfo
		r, _, _ = syscallN(funcNtQueryInformationThread.address(), h, 0, uintptr(unsafe.Pointer(&t)), unsafe.Sizeof(t), 0)
	)
	if r > 0 {
		return 0, formatNtError(r)
	}
	return uint32(t.ClientID.Thread), nil
}

// GetCodeIntegrityState returns a bitvalue that returns the Code Integrity status
// of the current device. If the return value is zero without an error, this means
// that code integrity is disabled.
func GetCodeIntegrityState() (uint32, error) {
	var (
		n       uint32
		s       = [2]uint32{8, 0}
		r, _, _ = syscallN(funcNtQuerySystemInformation.address(), 0x67, uintptr(unsafe.Pointer(&s)), 8, uintptr(unsafe.Pointer(&n)))
		// 0x67 - SystemCodeIntegrityInformation
	)
	if r > 0 {
		return 0, formatNtError(r)
	}
	return s[1], nil
}
func (p *dumpParam) write(w io.Writer) error {
	var (
		b      = (*[]byte)(unsafe.Pointer(&SliceHeader{Data: unsafe.Pointer(p.h), Len: int(p.w), Cap: int(p.w)}))
		n, err = w.Write(*b)
	)
	if b, *b = nil, nil; err != nil {
		return err
	}
	if n != int(p.w) {
		return io.ErrShortWrite
	}
	return nil
}

// GetDiskSize returns the size in bytes of the disk by it's NT path or the path
// to a partition or volume on the disk.
//
// Any errors encountered during reading will be returned.
//
// The name can be in the format of an NT path such as:
//
//   - \\.\C:
//   - \\.\PhysicalDrive0
//
// Both are equal on /most/ systems.
func GetDiskSize(name string) (uint64, error) {
	// 0x1 - FILE_SHARE_READ
	// 0x3 - OPEN_EXISTING
	h, err := CreateFile(name, 0, 0x1, nil, 0x3, 0, 0)
	if err != nil {
		return 0, err
	}
	var (
		g diskGeometryEx
		s [4 + ptrSize]byte // IO_STATUS_BLOCK
	)
	// 0x700A0 - IOCTL_DISK_GET_DRIVE_GEOMETRY_EX
	r, _, err := syscallN(funcNtDeviceIoControlFile.address(), h, 0, 0, 0, uintptr(unsafe.Pointer(&s)), 0x700A0, 0, 0, uintptr(unsafe.Pointer(&g)), 0x20+ptrSize)
	if CloseHandle(h); r > 0 {
		return 0, formatNtError(r)
	}
	return g.Size, nil
}

// UserFromToken will attempt to get the User SID from the supplied Token and
// return the associated Username and Domain string from the SID.
func UserFromToken(h uintptr) (string, error) {
	u, err := GetTokenUser(h)
	if err != nil {
		return "", err
	}
	return u.User.Sid.UserName()
}

// ForEachThread is a helper function that allows a function to be executed with
// the handle of the Thread.
//
// This function only returns an error if enumerating the Threads generates an
// error or the supplied function returns an error.
//
// This function ONLY targets Golang threads. To target all Process threads,
// use 'ForEachProcThread'.
func ForEachThread(f func(uintptr) error) error {
	var err error
	for i := uintptr(allm); ; {
		if h := *(*uintptr)(unsafe.Pointer(i + ptrThread)); h > 0 {
			if err = f(h); err != nil {
				break
			}
		}
		n := (*uintptr)(unsafe.Pointer(i + ptrNext))
		if n == nil || *n == 0 {
			break // Reached bottom of linked list
		}
		i = *n
	}
	return err
}

// GetTokenUser retrieves access token user account information and SID.
func GetTokenUser(h uintptr) (*TokenUser, error) {
	u, err := getTokenInfo(h, 1, 50)
	if err != nil {
		return nil, err
	}
	return (*TokenUser)(u), nil
}

// GetVersionNumbers returns the NTDLL internal version numbers as Major, Minor
// and Build.
//
// This function should return the correct values regardless of manifest version.
func GetVersionNumbers() (uint32, uint32, uint16) {
	var m, n, b uint32
	syscallN(funcRtlGetNtVersionNumbers.address(), uintptr(unsafe.Pointer(&m)), uintptr(unsafe.Pointer(&n)), uintptr(unsafe.Pointer(&b)))
	return m, n, uint16(b)
}
func enablePrivileges(h uintptr, s []string) error {
	var (
		p   privileges
		err error
	)
	for i := range s {
		if i > 5 {
			break
		}
		if err = LookupPrivilegeValue("", s[i], &p.Privileges[i].Luid); err != nil {
			if xerr.ExtendedInfo {
				return xerr.Wrap(`cannot lookup "`+s[i]+`"`, err)
			}
			return xerr.Wrap("cannot lookup Privilege", err)
		}
		p.Privileges[i].Attributes = 0x2 // SE_PRIVILEGE_ENABLED
	}
	p.PrivilegeCount = uint32(len(s))
	if err = AdjustTokenPrivileges(h, false, unsafe.Pointer(&p), uint32(unsafe.Sizeof(p)), nil, nil); err != nil {
		return xerr.Wrap("cannot assign all Privileges", err)
	}
	return nil
}

// GetProcessFileName will attempt to retrieve the basename of the process
// related to the open Process handle supplied.
func GetProcessFileName(h uintptr) (string, error) {
	var (
		u ntUnicodeString
		n uint32
	)
	r, _, _ := syscallN(
		funcNtQueryInformationProcess.address(), h, 0x1B, uintptr(unsafe.Pointer(&u)),
		unsafe.Sizeof(u)+260, uintptr(unsafe.Pointer(&n)),
	)
	// 0x1B - ProcessImageFileName
	if r > 0 {
		return "", formatNtError(r)
	}
	v := UTF16ToString(u.Buffer[4:n])
	for i := len(v) - 1; i > 0; i-- {
		if v[i] == '\\' {
			return v[i+1:], nil
		}
	}
	return v, nil
}

// ForEachProcThread is a helper function that allows a function to be executed
// with the handle of the Thread.
//
// This function only returns an error if enumerating the Threads generates an
// error or the supplied function returns an error.
//
// This function targets ALL threads (including non-Golang threads). To target
// all only Golang threads, use 'ForEachThread'.
func ForEachProcThread(f func(uintptr) error) error {
	return EnumThreads(GetCurrentProcessID(), func(t ThreadEntry) error {
		// old (0xE0 - THREAD_QUERY_INFORMATION | THREAD_SET_INFORMATION | THREAD_SET_THREAD_TOKEN)
		// 0x1FFFFF - THREAD_ALL_ACCESS
		v, err := t.Handle(0x1FFFFF)
		if err != nil {
			return err
		}
		err = f(v)
		CloseHandle(v)
		return err
	})
}
func getThreadStartAddress(h uintptr) (uintptr, error) {
	var (
		i       uintptr
		r, _, _ = syscallN(funcNtQueryInformationThread.address(), h, 0x9, uintptr(unsafe.Pointer(&i)), ptrSize, 0)
		// 0x9 - ThreadQuerySetWin32StartAddress
	)
	if r > 0 {
		return 0, formatNtError(r)
	}
	return i, nil
}

// FileSigningIssuerName attempts to read the Authenticate signing certificate
// issuer name for the specified file path.
//
// If the file does not exist or a certificate cannot be found, this returns the
// error 'syscall.EINVAL'.
//
// If the function success, the return result will be the string name of the
// certificate issuer.
func FileSigningIssuerName(path string) (string, error) {
	f, err1 := UTF16PtrFromString(path)
	if err1 != nil {
		return "", err1
	}
	var (
		s, h      uintptr
		r, _, err = syscallN(
			funcCryptQueryObject.address(), 0x1, uintptr(unsafe.Pointer(f)), 0x400, 0x2,
			0, 0, 0, 0, uintptr(unsafe.Pointer(&s)), uintptr(unsafe.Pointer(&h)), 0,
		)
		// 0x1   - CERT_QUERY_OBJECT_FILE
		// 0x400 - CERT_QUERY_CONTENT_FLAG_PKCS7_SIGNED_EMBED
		// 0x2   - CERT_QUERY_FORMAT_FLAG_BINARY
	)
	if r == 0 {
		if err == 0x80092009 { // 0x80092009 - Object not found, file isn't signed.
			return "", syscall.EINVAL
		}
		return "", unboxError(err)
	}
	var x uint32
	// 0x6 - CMSG_SIGNER_INFO_PARAM
	if r, _, err = syscallN(funcCryptMsgGetParam.address(), h, 0x6, 0, 0, uintptr(unsafe.Pointer(&x))); r == 0 {
		syscallN(funcCryptMsgClose.address(), h)
		syscallN(funcCertCloseStore.address(), s)
		return "", unboxError(err)
	}
	b := make([]byte, x)
	// 0x6 - CMSG_SIGNER_INFO_PARAM
	r, _, err = syscallN(funcCryptMsgGetParam.address(), h, 0x6, 0, uintptr(unsafe.Pointer(&b[0])), uintptr(unsafe.Pointer(&x)))
	if syscallN(funcCryptMsgClose.address(), h); r == 0 {
		syscallN(funcCertCloseStore.address(), s)
		return "", unboxError(err)
	}
	var (
		v = (*certSigner)(unsafe.Pointer(&b[0]))
		i = certInfo{Issuer: v.Issuer, Serial: v.Serial}
	)
	// 0x10001 - X509_ASN_ENCODING | PKCS_7_ASN_ENCODING
	// 0xB0000 - CERT_FIND_SUBJECT_CERT
	r, _, err = syscallN(funcCertFindCertificateInStore.address(), s, 0x10001, 0, 0xB0000, uintptr(unsafe.Pointer(&i)), 0)
	if syscallN(funcCertCloseStore.address(), s); r == 0 {
		return "", unboxError(err)
	}
	var (
		n string
		k uintptr
	)
	// 0x4 - CERT_NAME_SIMPLE_DISPLAY_TYPE
	// 0x0 - CERT_NAME_ISSUER_FLAG
	if k, _, err = syscallN(funcCertGetNameString.address(), r, 0x4, 0x0, 0, 0, 0); k > 0 {
		c := make([]uint16, k)
		if k, _, err = syscallN(funcCertGetNameString.address(), r, 0x4, 0x0, 0, uintptr(unsafe.Pointer(&c[0])), k); k > 0 {
			n = UTF16ToString(c[:k])
		}
	}
	if syscallN(funcCertFreeCertificateContext.address(), r); k == 0 {
		return "", unboxError(err)
	}
	return n, nil
}

// StringListToUTF16Block creates a UTF16 encoded block for usage as a Process
// environment block.
//
// This function returns an error if any of the environment strings are not in
// the 'KEY=VALUE' format or contain a NUL byte.
func StringListToUTF16Block(s []string) (*uint16, error) {
	if len(s) == 0 {
		return nil, nil
	}
	var t, i, l int
	for _, x := range s {
		for v := range x {
			if x[v] == 0 {
				return nil, syscall.EINVAL
			}
		}
		if q := strings.IndexByte(x, '='); q <= 0 {
			if xerr.ExtendedInfo {
				return nil, xerr.Sub(`invalid env value "`+x+`"`, 0x17)
			}
			return nil, xerr.Sub("invalid env value", 0x17)
		}
		t += len(x) + 1
	}
	t++
	b := make([]byte, t)
	for _, v := range s {
		l = len(v)
		copy(b[i:i+l], v)
		b[i+l] = 0
		i = i + l + 1
	}
	b[i] = 0
	return &UTF16EncodeStd([]rune(string(b)))[0], nil
}

// EnableTokenPrivileges will attempt to enable the supplied Windows privilege
// values on the supplied process Token.
//
// Errors during encoding, lookup or assignment will be returned and not all
// privileges will be assigned, if they occur.
func EnableTokenPrivileges(h uintptr, s ...string) error {
	if len(s) == 0 {
		return nil
	}
	if len(s) <= 5 {
		return enablePrivileges(h, s)
	}
	for x, w := 0, 0; x < len(s); {
		if w = 5; x+w > len(s) {
			w = len(s) - x
		}
		if err := enablePrivileges(h, s[x:x+w]); err != nil {
			return err
		}
		x += w
	}
	return nil
}
func heapAlloc(h uintptr, s uint64, z bool) (uintptr, error) {
	var f uint32
	if z {
		f |= 0x08
	}
	r, _, err := syscallN(funcRtlAllocateHeap.address(), h, uintptr(f), uintptr(s))
	if r == 0 {
		return 0, unboxError(err)
	}
	return r, nil
}
func (p *dumpParam) copy(o uint64, b uintptr, s uint32) error {
	if err := p.resize(o + uint64(s)); err != nil {
		return err
	}
	copyMemory(p.h+uintptr(o), b, s)
	p.w += uint64(s)
	return nil
}
func heapReAlloc(h, m uintptr, s uint64, z bool) (uintptr, error) {
	var f uint32
	if z {
		// 0x8 - HEAP_ZERO_MEMORY
		f |= 0x8
	}
	r, _, err := syscallN(funcRtlReAllocateHeap.address(), h, uintptr(f), m, uintptr(s))
	if r == 0 {
		return 0, unboxError(err)
	}
	return r, nil
}
func dumpCallbackFunc(_ uintptr, i uintptr, r *dumpOutput) uintptr {
	switch *(*uint32)(unsafe.Pointer(i + 4 + ptrSize)) {
	case 11:
		r.Status = 1
	case 12:
		var (
			o = *(*uint64)(unsafe.Pointer(i + 8 + (ptrSize * 2)))   // Offset
			b = *(*uintptr)(unsafe.Pointer(i + 16 + (ptrSize * 2))) // Buffer
			s = *(*uint32)(unsafe.Pointer(i + 16 + (ptrSize * 3)))  // Size
		)
		if err := dumpStack.copy(o, b, s); err != nil {
			r.Status = 1
			return 0
		}
		r.Status = 0
	case 13:
		r.Status = 0
	case 16, 17:
		r.Status = 0
	}
	return 1
}
func getTokenInfo(t uintptr, c uint32, i int) (unsafe.Pointer, error) {
	for n := uint32(i); ; {
		var (
			b   = make([]byte, n)
			err = GetTokenInformation(t, c, &b[0], uint32(len(b)), &n)
		)
		if err == nil {
			return unsafe.Pointer(&b[0]), nil
		}
		if err != syscall.ERROR_INSUFFICIENT_BUFFER {
			return nil, err
		}
		if n <= uint32(len(b)) {
			return nil, err
		}
	}
}

// MiniDumpWriteDump Windows API Call
//
//	Writes user-mode minidump information to the specified file handle.
//
// https://docs.microsoft.com/en-us/windows/win32/api/minidumpapiset/nf-minidumpapiset-minidumpwritedump
//
// Updated version that will take and use the supplied Writer instead of the file
// handle is zero.
//
// This function may fail if attempting to dump a process that is a different CPU
// architecture than the host process.
//
// Dumping to a Writer instead of a file is not avaliable on systems older than
// Windows Vista and will return 'syscall.EINVAL' instead.
func MiniDumpWriteDump(h uintptr, pid uint32, o uintptr, f uint32, w io.Writer) error {
	if o > 0 {
		r, _, err := syscallN(funcMiniDumpWriteDump.address(), h, uintptr(pid), o, uintptr(f), 0, 0, 0)
		if r == 0 {
			return unboxError(err)
		}
		return nil
	}
	if !IsWindowsVista() {
		return syscall.EINVAL
	}
	if err := dumpStack.init(); err != nil {
		return err
	}
	var (
		a          = dumpCallback{Func: dumpCallbackOnce.f}
		r, _, err1 = syscallN(funcMiniDumpWriteDump.address(), h, uintptr(pid), 0, uintptr(f), 0, 0, uintptr(unsafe.Pointer(&a)))
	)
	if r == 0 {
		dumpStack.close()
		return unboxError(err1)
	}
	err := dumpStack.write(w)
	dumpStack.close()
	return err
}
