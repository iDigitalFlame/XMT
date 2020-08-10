// +build windows

package cmd

import (
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync/atomic"
	"unicode/utf16"
	"unsafe"

	"github.com/iDigitalFlame/xmt/device/devtools"
	"github.com/iDigitalFlame/xmt/util"
	"github.com/iDigitalFlame/xmt/util/xerr"
	"golang.org/x/sys/windows"
)

const secStandard uint32 = windows.PROCESS_TERMINATE | windows.SYNCHRONIZE |
	windows.PROCESS_QUERY_INFORMATION | windows.PROCESS_CREATE_PROCESS |
	windows.PROCESS_SUSPEND_RESUME | windows.PROCESS_DUP_HANDLE

var (
	dllKernel32 = windows.NewLazySystemDLL("kernel32.dll")

	funcRtlCloneUserProcess = dllNtdll.NewProc("RtlCloneUserProcess")

	funcAllocConsole                      = dllKernel32.NewProc("AllocConsole")
	funcCreateProcess                     = dllKernel32.NewProc("CreateProcessW")
	funcCreateProcessAsUser               = dllKernel32.NewProc("CreateProcessAsUserW")
	funcUpdateProcThreadAttribute         = dllKernel32.NewProc("UpdateProcThreadAttribute")
	funcInitializeProcThreadAttributeList = dllKernel32.NewProc("InitializeProcThreadAttributeList")
)

type file interface {
	File() (*os.File, error)
}
type clientID struct {
	UniqueProcess uintptr
	UniqueThread  uintptr
}
type imageInfo struct {
	_       uintptr
	_       uint32
	_, _    uint64
	_       uint32
	_, _    uint16
	_       uint32
	_, _    uint16
	_       uint32
	_, _, _ uint16
	_, _, _ uint8
	_, _, _ uint32
}
type processInfo struct {
	Length   uint32
	Process  uintptr
	Thread   uintptr
	ClientID clientID
	_        imageInfo
}
type closer windows.Handle
type startupAttrs struct {
	_, _, _, _, _, _ uint64
}
type startupInfoEx struct {
	StartupInfo   windows.StartupInfo
	AttributeList *startupAttrs
}

func (p *Process) wait() {
	var (
		x   = make(chan error)
		err error
	)
	go waitFunc(p.opts.info.Process, windows.INFINITE, x)
	select {
	case err = <-x:
	case <-p.ctx.Done():
	}
	if err != nil {
		p.stopWith(err)
		return
	}
	if p.ctx.Err() != nil {
		p.stopWith(p.ctx.Err())
		return
	}
	err = windows.GetExitCodeProcess(p.opts.info.Process, &p.exit)
	if atomic.StoreUint32(&p.once, 2); err != nil {
		p.stopWith(err)
		return
	}
	if p.exit != 0 {
		p.stopWith(&ExitError{Exit: p.exit})
		return
	}
	p.stopWith(nil)
}
func (o *options) close() {
	for i := range o.closers {
		o.closers[i].Close()
	}
	if o.parent > 0 {
		windows.CloseHandle(o.parent)
	}
	o.parent, o.closers = 0, nil
}
func (c container) clear() {
	c.pid, c.name, c.choices, c.elevated = 0, "", nil, false
}
func (c closer) Close() error {
	return windows.CloseHandle(windows.Handle(c))
}
func createEnv(s []string) (*uint16, error) {
	if len(s) == 0 {
		return nil, nil
	}
	var t, i, l int
	for _, s := range s {
		if q := strings.IndexByte(s, 61); q <= 0 {
			return nil, xerr.New(`invalid environment string "` + s + `"`)
		}
		t += len(s) + 1
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
	return &utf16.Encode([]rune(string(b)))[0], nil
}
func readFunc(r io.ReadCloser, w io.Writer) {
	io.Copy(w, r)
	r.Close()
}
func wait(h windows.Handle, d uint32) error {
	switch r, err := windows.WaitForSingleObject(h, d); {
	case r == windows.WAIT_OBJECT_0:
		return nil
	default:
		return err
	}
}
func writeFunc(r io.Reader, w io.WriteCloser) {
	io.Copy(w, r)
	w.Close()
}
func waitFunc(h windows.Handle, d uint32, c chan<- error) {
	c <- wait(h, d)
}
func (o *options) startInfo() (*windows.StartupInfo, error) {
	s := &windows.StartupInfo{X: o.X, Y: o.Y, XSize: o.W, YSize: o.H, Flags: o.Flags, ShowWindow: o.Mode}
	if len(o.Title) > 0 {
		var err error
		if s.Title, err = windows.UTF16PtrFromString(o.Title); err != nil {
			return nil, xerr.Wrap(`cannot convert title "`+o.Title+`"`, err)
		}
	}
	o.parent, o.closers = 0, nil
	return s, nil
}
func (c container) getParent(a uint32) (windows.Handle, error) {
	if c.pid > 0 {
		h, err := windows.OpenProcess(a, true, uint32(c.pid))
		if h == 0 {
			return 0, xerr.Wrap("winapi OpenProcess PID "+strconv.Itoa(int(c.pid))+" error", err)
		}
		return h, nil
	}
	h, err := getProcessByName(a, c.name, c.choices, c.pid < 0, c.elevated)
	if err != nil {
		if c.pid < 0 {
			return 0, err
		}
		if len(c.name) > 0 {
			return 0, xerr.Wrap(c.name, err)
		}
		return 0, xerr.Wrap("["+strings.Join(c.choices, ", ")+"]", err)
	}
	return h, nil
}
func (o *options) readHandle(r io.Reader, m bool) (windows.Handle, error) {
	if !m && r == nil {
		return 0, nil
	}
	var h uintptr
	if r != nil {
		switch i := r.(type) {
		case *os.File:
			h = i.Fd()
		case file:
			f, err := i.File()
			if err != nil {
				return 0, xerr.Wrap("cannot obtain file handle for "+reflect.TypeOf(r).String(), err)
			}
			h = f.Fd()
		default:
			x, y, err := os.Pipe()
			if err != nil {
				return 0, xerr.Wrap("cannot open os pipe", err)
			}
			h = x.Fd()
			o.closers = append(o.closers, x)
			o.closers = append(o.closers, y)
			go writeFunc(r, y)
		}
		if h == 0 {
			return 0, nil
		}
	} else {
		f, err := os.Open(os.DevNull)
		if err != nil {
			return 0, xerr.Wrap("cannot open null device", err)
		}
		o.closers = append(o.closers, f)
		h = f.Fd()
	}
	var (
		v   = windows.CurrentProcess()
		n   windows.Handle
		err error
	)
	if o.parent > 0 {
		v = o.parent
	}
	if err = windows.DuplicateHandle(windows.CurrentProcess(), windows.Handle(h), v, &n, 0, true, windows.DUPLICATE_SAME_ACCESS); err != nil {
		return 0, xerr.Wrap("cannot duplicate handle 0x"+strconv.FormatUint(uint64(h), 16), err)
	}
	o.closers = append(o.closers, closer(n))
	return n, nil
}
func (o *options) writeHandle(w io.Writer, m bool) (windows.Handle, error) {
	if !m && w == nil {
		return 0, nil
	}
	var h uintptr
	if w != nil {
		switch i := w.(type) {
		case *os.File:
			h = i.Fd()
		case file:
			f, err := i.File()
			if err != nil {
				return 0, xerr.Wrap("cannot obtain file handle for "+reflect.TypeOf(w).String(), err)
			}
			h = f.Fd()
		default:
			x, y, err := os.Pipe()
			if err != nil {
				return 0, xerr.Wrap("cannot open os pipe", err)
			}
			h = y.Fd()
			o.closers = append(o.closers, x)
			o.closers = append(o.closers, y)
			go readFunc(x, w)
		}
		if h == 0 {
			return 0, nil
		}
	} else {
		f, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0644)
		if err != nil {
			return 0, xerr.Wrap("cannot open null device", err)
		}
		o.closers = append(o.closers, f)
		h = f.Fd()
	}
	var (
		v   = windows.CurrentProcess()
		n   windows.Handle
		err error
	)
	if o.parent > 0 {
		v = o.parent
	}
	if err = windows.DuplicateHandle(windows.CurrentProcess(), windows.Handle(h), v, &n, 0, true, windows.DUPLICATE_SAME_ACCESS); err != nil {
		return 0, xerr.Wrap("cannot duplicate handle 0x"+strconv.FormatUint(uint64(h), 16), err)
	}
	o.closers = append(o.closers, closer(n))
	return n, nil
}
func newParentEx(p windows.Handle, i *windows.StartupInfo) (*startupInfoEx, error) {
	var (
		s uint64
		x startupInfoEx
	)
	if _, _, err := funcInitializeProcThreadAttributeList.Call(0, 1, 0, uintptr(unsafe.Pointer(&s))); s < 48 {
		return nil, xerr.Wrap("winapi InitializeProcThreadAttributeList error", err)
	}
	x.AttributeList = new(startupAttrs)
	r, _, err := funcInitializeProcThreadAttributeList.Call(
		uintptr(unsafe.Pointer(x.AttributeList)), 1, 0, uintptr(unsafe.Pointer(&s)),
	)
	if r == 0 {
		return nil, xerr.Wrap("winapi InitializeProcThreadAttributeList error", err)
	}
	if i != nil {
		x.StartupInfo = *i
	}
	x.StartupInfo.Cb = uint32(unsafe.Sizeof(x))
	r, _, err = funcUpdateProcThreadAttribute.Call(
		uintptr(unsafe.Pointer(x.AttributeList)), 0, 0x00020000,
		uintptr(unsafe.Pointer(&p)), uintptr(unsafe.Sizeof(p)), 0, 0,
	)
	if r == 0 {
		return nil, xerr.Wrap("winapi UpdateProcThreadAttribute error", err)
	}
	return &x, nil
}
func getProcessByName(a uint32, n string, x []string, v, y bool) (windows.Handle, error) {
	h, err := windows.CreateToolhelp32Snapshot(0x00000002, 0)
	if err != nil {
		return 0, xerr.Wrap("winapi CreateToolhelp32Snapshot error", err)
	}
	if y {
		devtools.AdjustPrivileges("SeDebugPrivilege")
	}
	var (
		e    windows.ProcessEntry32
		q    []string
		c    []windows.Handle
		p    = uint32(os.Getpid())
		z    windows.Token
		o    windows.Handle
		f, j bool
		s, m = "", strings.ToLower(n)
	)
	if e.Size = uint32(unsafe.Sizeof(e)); len(x) > 0 {
		q := make([]string, len(x))
		for i := range x {
			q[i] = strings.ToLower(x[i])
		}
	}
	for err = windows.Process32First(h, &e); err == nil; err = windows.Process32Next(h, &e) {
		if e.ProcessID == p {
			continue
		}
		if s = strings.ToLower(windows.UTF16ToString(e.ExeFile[:])); len(s) == 0 {
			continue
		}
		for i := range q {
			if q[i] == s {
				f = true
				break
			}
		}
		if !v && ((len(q) > 0 && !f) || (len(m) > 0 && s != m)) {
			continue
		}
		if o, err = windows.OpenProcess(a, true, e.ProcessID); err != nil || o == 0 {
			continue
		}
		if y {
			if err := windows.OpenProcessToken(o, windows.TOKEN_QUERY, &z); err != nil {
				windows.CloseHandle(o)
				continue
			}
			j = z.IsElevated()
			if z.Close(); !j {
				windows.CloseHandle(o)
				continue
			}
		}
		if v {
			c = append(c, o)
			continue
		}
		break
	}
	if windows.CloseHandle(h); v && len(c) > 0 {
		b := c[int(util.FastRandN(len(c)))]
		for i := range c {
			if c[i] == b {
				continue
			}
			windows.CloseHandle(c[i])
		}
		return b, nil
	}
	if err != nil {
		return 0, ErrNoProcessFound
	}
	return windows.Handle(o), nil
}
func run(name, cmd, dir string, p, t *windows.SecurityAttributes, f uint32, e *uint16, s *windows.StartupInfo, x *startupInfoEx, u *windows.Token, i *windows.ProcessInformation) error {
	var (
		err     error
		r, z    uintptr
		n, c, d *uint16
	)
	if len(name) > 0 {
		if n, err = windows.UTF16PtrFromString(name); err != nil {
			return xerr.Wrap(`cannot convert "`+name+`"`, err)
		}
	}
	if len(cmd) > 0 {
		if c, err = windows.UTF16PtrFromString(cmd); err != nil {
			return xerr.Wrap(`cannot convert "`+cmd+`"`, err)
		}
	}
	if len(dir) > 0 {
		if d, err = windows.UTF16PtrFromString(dir); err != nil {
			return xerr.Wrap(`cannot convert "`+dir+`"`, err)
		}
	}
	if e != nil {
		f |= windows.CREATE_UNICODE_ENVIRONMENT
	}
	if x != nil {
		if x.StartupInfo.Cb == 0 {
			x.StartupInfo.Cb = uint32(unsafe.Sizeof(&x))
		}
		z = uintptr(unsafe.Pointer(x))
		f |= 0x00080000
	} else if s != nil {
		if s.Cb == 0 {
			s.Cb = uint32(unsafe.Sizeof(&s))
		}
		z = uintptr(unsafe.Pointer(s))
	}
	if u != nil {
		r, _, err = funcCreateProcessAsUser.Call(
			uintptr(*u),
			uintptr(unsafe.Pointer(n)), uintptr(unsafe.Pointer(c)), uintptr(unsafe.Pointer(p)),
			uintptr(unsafe.Pointer(t)), uintptr(1), uintptr(f), uintptr(unsafe.Pointer(e)),
			uintptr(unsafe.Pointer(d)), z, uintptr(unsafe.Pointer(i)),
		)
	} else {
		r, _, err = funcCreateProcess.Call(
			uintptr(unsafe.Pointer(n)), uintptr(unsafe.Pointer(c)), uintptr(unsafe.Pointer(p)),
			uintptr(unsafe.Pointer(t)), uintptr(1), uintptr(f), uintptr(unsafe.Pointer(e)),
			uintptr(unsafe.Pointer(d)), z, uintptr(unsafe.Pointer(i)),
		)
	}
	if r == 0 {
		return xerr.Wrap("winapi CreateProcess error", err)
	}
	return nil
}
