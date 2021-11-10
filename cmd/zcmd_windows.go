//go:build windows
// +build windows

package cmd

import (
	"io"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"unicode/utf16"
	"unsafe"

	"github.com/iDigitalFlame/xmt/util/xerr"
	"golang.org/x/sys/windows"
)

const secStandard uint32 = windows.PROCESS_TERMINATE | windows.SYNCHRONIZE |
	windows.PROCESS_QUERY_INFORMATION | windows.PROCESS_CREATE_PROCESS |
	windows.PROCESS_SUSPEND_RESUME | windows.PROCESS_DUP_HANDLE

var (
	secProtect uint64 = 0x100100000000

	dllKernel32 = windows.NewLazySystemDLL("kernel32.dll")

	funcNtResumeProcess     = dllNtdll.NewProc("NtResumeProcess")
	funcNtSuspendProcess    = dllNtdll.NewProc("NtSuspendProcess")
	funcRtlCloneUserProcess = dllNtdll.NewProc("RtlCloneUserProcess")

	funcLoadLibrary                       = dllKernel32.NewProc(loadLibFunc)
	funcAllocConsole                      = dllKernel32.NewProc("AllocConsole")
	funcCreateProcess                     = dllKernel32.NewProc("CreateProcessW")
	funcCreateProcessAsUser               = dllKernel32.NewProc("CreateProcessAsUserW")
	funcUpdateProcThreadAttribute         = dllKernel32.NewProc("UpdateProcThreadAttribute")
	funcDeleteProcThreadAttributeList     = dllKernel32.NewProc("DeleteProcThreadAttributeList")
	funcInitializeProcThreadAttributeList = dllKernel32.NewProc("InitializeProcThreadAttributeList")

	verOnce   sync.Once
	verPast61 bool
)

type file interface {
	File() (*os.File, error)
}
type clientID struct {
	Process, Thread uintptr
}

// DO NOT REORDER
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

// DO NOT REORDER
type processInfo struct {
	Length          uint32
	Process, Thread uintptr
	ClientID        clientID
	_               imageInfo
}
type closer windows.Handle
type startupAttrs struct {
	_, _, _, _, _, _, _, _, _ uint64
}

// DO NOT REORDER
type startupInfoEx struct {
	StartupInfo   windows.StartupInfo
	AttributeList *startupAttrs
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
func (c closer) Close() error {
	return windows.CloseHandle(windows.Handle(c))
}
func wait(h windows.Handle) error {
	if r, err := windows.WaitForSingleObject(h, windows.INFINITE); r != windows.WAIT_OBJECT_0 {
		return err
	}
	return nil
}
func (p *Process) wait(e *startupInfoEx) {
	var (
		x   = make(chan error)
		err error
	)
	go func() {
		x <- wait(p.opts.info.Process)
		close(x)
	}()
	select {
	case err = <-x:
	case <-p.ctx.Done():
	}
	if e != nil {
		funcDeleteProcThreadAttributeList.Call(uintptr(unsafe.Pointer(e.AttributeList)))
		e.AttributeList = nil
		e = nil
	}
	if err != nil {
		p.stopWith(exitStopped, err)
		return
	}
	if err2 := p.ctx.Err(); err2 != nil {
		p.stopWith(exitStopped, err2)
		return
	}
	err = windows.GetExitCodeProcess(p.opts.info.Process, &p.exit)
	if atomic.StoreUint32(&p.cookie, 2); err != nil {
		p.stopWith(exitStopped, err)
		return
	}
	if p.exit != 0 {
		p.stopWith(p.exit, &ExitError{Exit: p.exit})
		return
	}
	p.stopWith(p.exit, nil)
}
func createEnv(s []string) (*uint16, error) {
	if len(s) == 0 {
		return nil, nil
	}
	var t, i, l int
	for _, s := range s {
		if q := strings.IndexByte(s, 61); q <= 0 {
			return nil, xerr.New(`invalid env string "` + s + `"`)
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
func (o *options) startInfo() (*windows.StartupInfo, error) {
	s := &windows.StartupInfo{X: o.X, Y: o.Y, XSize: o.W, YSize: o.H, Flags: o.Flags, ShowWindow: o.Mode}
	if len(o.Title) > 0 {
		var err error
		if s.Title, err = windows.UTF16PtrFromString(o.Title); err != nil {
			return nil, xerr.Wrap("cannot convert title", err)
		}
	}
	o.parent, o.closers = 0, nil
	return s, nil
}
func (o *options) reader(r io.Reader, m bool) (windows.Handle, error) {
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
				return 0, xerr.Wrap("cannot obtain file handle", err)
			}
			h = f.Fd()
		default:
			x, y, err := os.Pipe()
			if err != nil {
				return 0, xerr.Wrap("cannot create pipe", err)
			}
			h = x.Fd()
			o.closers = append(o.closers, x)
			o.closers = append(o.closers, y)
			go func() {
				io.Copy(y, r)
				y.Close()
			}()
		}
		if h == 0 {
			return 0, nil
		}
	} else {
		f, err := os.OpenFile(os.DevNull, os.O_RDONLY, 0)
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
		return 0, xerr.Wrap("DuplicateHandle", err)
	}
	o.closers = append(o.closers, closer(n))
	return n, nil
}
func (o *options) writer(w io.Writer, m bool) (windows.Handle, error) {
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
				return 0, xerr.Wrap("cannot obtain file handle", err)
			}
			h = f.Fd()
		default:
			x, y, err := os.Pipe()
			if err != nil {
				return 0, xerr.Wrap("cannot create pipe", err)
			}
			h = y.Fd()
			o.closers = append(o.closers, x)
			o.closers = append(o.closers, y)
			go func() {
				io.Copy(w, x)
				x.Close()
			}()
		}
		if h == 0 {
			return 0, nil
		}
	} else {
		f, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
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
		return 0, xerr.Wrap("DuplicateHandle", err)
	}
	o.closers = append(o.closers, closer(n))
	return n, nil
}
func parentEx(p windows.Handle, i *windows.StartupInfo) (*startupInfoEx, error) {
	var (
		s, w uint64
		c    uintptr
		x    startupInfoEx
	)
	if protectEnable {
		verOnce.Do(func() {
			if v, err := windows.GetVersion(); err == nil && v > 0 {
				if m := byte(v & 0xFF); m >= 6 {
					verPast61 = byte((v&0xFFFF)>>8) >= 1
				}
			}
		})
	}
	if protectEnable && verPast61 {
		w, c = 72, 2
	} else {
		w, c = 48, 1
	}
	if _, _, err := funcInitializeProcThreadAttributeList.Call(0, c, 0, uintptr(unsafe.Pointer(&s))); s < w {
		return nil, xerr.Wrap("InitializeProcThreadAttributeList", err)
	}
	x.AttributeList = new(startupAttrs)
	r, _, err := funcInitializeProcThreadAttributeList.Call(
		uintptr(unsafe.Pointer(x.AttributeList)), c, 0, uintptr(unsafe.Pointer(&s)),
	)
	if r == 0 {
		return nil, xerr.Wrap("InitializeProcThreadAttributeList", err)
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
		funcDeleteProcThreadAttributeList.Call(uintptr(unsafe.Pointer(x.AttributeList)))
		return nil, xerr.Wrap("UpdateProcThreadAttribute", err)
	}
	if protectEnable && verPast61 {
		r, _, err = funcUpdateProcThreadAttribute.Call(
			uintptr(unsafe.Pointer(x.AttributeList)), 0, 0x00020007,
			uintptr(unsafe.Pointer(&secProtect)), uintptr(unsafe.Sizeof(secProtect)), 0, 0,
		)
		if r == 0 {
			funcDeleteProcThreadAttributeList.Call(uintptr(unsafe.Pointer(x.AttributeList)))
			return nil, xerr.Wrap("UpdateProcThreadAttribute", err)
		}
	}
	return &x, nil
}
func run(name, cmd, dir string, p, t *windows.SecurityAttributes, f uint32, e *uint16, s *windows.StartupInfo, x *startupInfoEx, u *windows.Token, i *windows.ProcessInformation) error {
	var (
		err     error
		r, z    uintptr
		n, c, d *uint16
	)
	if len(name) > 0 {
		if n, err = windows.UTF16PtrFromString(name); err != nil {
			return xerr.Wrap("cannot convert name", err)
		}
	}
	if len(cmd) > 0 {
		if c, err = windows.UTF16PtrFromString(cmd); err != nil {
			return xerr.Wrap("cannot convert cmd", err)
		}
	}
	if len(dir) > 0 {
		if d, err = windows.UTF16PtrFromString(dir); err != nil {
			return xerr.Wrap("cannot convert directory", err)
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
		return xerr.Wrap("CreateProcess", err)
	}
	return nil
}
