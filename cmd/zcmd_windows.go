// +build windows

package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"
	"syscall"
	"unicode/utf16"
	"unsafe"

	"github.com/iDigitalFlame/xmt/device"
	"github.com/iDigitalFlame/xmt/util"
	"golang.org/x/sys/windows"

	ps "github.com/shirou/gopsutil/process"
)

var (
	dllKernel32 = windows.NewLazySystemDLL("kernel32.dll")

	funcOpenProcess                       = dllKernel32.NewProc("OpenProcess")
	funcCreateProcessW                    = dllKernel32.NewProc("CreateProcessW")
	funcTerminateThread                   = dllKernel32.NewProc("TerminateThread")
	funcTerminateProcess                  = dllKernel32.NewProc("TerminateProcess")
	funcGetExitCodeProcess                = dllKernel32.NewProc("GetExitCodeProcess")
	funcWaitForSingleObject               = dllKernel32.NewProc("WaitForSingleObject")
	funcCreateProcessAsUser               = dllKernel32.NewProc("CreateProcessAsUserA")
	funcUpdateProcThreadAttribute         = dllKernel32.NewProc("UpdateProcThreadAttribute")
	funcInitializeProcThreadAttributeList = dllKernel32.NewProc("InitializeProcThreadAttributeList")
)

type file interface {
	File() (*os.File, error)
}
type startupAttrs struct {
	_, _, _, _, _, _ uint64
}
type startupInfoEx struct {
	StartupInfo   windows.StartupInfo
	AttributeList *startupAttrs
}

func (p *Process) wait() {
	if err := wait(p.opts.info.Process, 0xFFFFFFFF); err != nil {
		p.stopWith(err)
		return
	}
	if p.ctx.Err() != nil {
		return
	}
	if r, _, err := funcGetExitCodeProcess.Call(uintptr(p.opts.info.Process), uintptr(unsafe.Pointer(&p.exit))); r == 0 {
		p.stopWith(fmt.Errorf("winapi GetExitCodeProcess error: %w", err))
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
func (c container) pid() (int32, error) {
	if c.PID > 0 {
		return c.PID, nil
	}
	l, err := ps.Processes()
	if err != nil {
		return 0, err
	}
	if len(l) == 0 {
		return 0, ErrNoProcessFound
	}
	switch {
	case len(c.Name) > 0:
		s := strings.ToLower(string(c.Name))
		for i := range l {
			if uint32(l[i].Pid) == device.Local.PID {
				continue
			}
			if _, err := l[i].Exe(); err != nil {
				continue
			}
			n, err := l[i].Name()
			if err != nil || strings.ToLower(n) != s {
				continue
			}
			return l[i].Pid, nil
		}
		return 0, fmt.Errorf("%s: %w", c.Name, ErrNoProcessFound)
	case c.Choices != nil && len(c.Choices) > 0:
		s := make([]string, len(c.Choices))
		for i := range c.Choices {
			s[i] = strings.ToLower(string(c.Choices[i]))
		}
		for i := range l {
			if uint32(l[i].Pid) == device.Local.PID {
				continue
			}
			if _, err := l[i].Exe(); err != nil {
				continue
			}
			n, err := l[i].Name()
			if err != nil {
				continue
			}
			for x := range s {
				if strings.ToLower(n) == s[x] {
					return l[i].Pid, nil
				}
			}
		}
		return 0, fmt.Errorf("[%s]: %w", strings.Join(c.Choices, ", "), ErrNoProcessFound)
	}
	a := make([]int32, 0, len(l))
	for i := range l {
		if uint32(l[i].Pid) == device.Local.PID {
			continue
		}
		if _, err := l[i].Exe(); err != nil {
			continue
		}
		a = append(a, l[i].Pid)
	}
	if len(a) == 1 {
		return a[0], nil
	}
	return a[util.Rand.Intn(len(a))], nil
}
func createEnv(s []string) (*uint16, error) {
	if len(s) == 0 {
		return nil, nil
	}
	var t, i int
	for _, s := range s {
		if q := strings.IndexRune(s, '='); q <= 0 {
			return nil, fmt.Errorf("invalid environment string %q", s)
		}
		t += len(s) + 1
	}
	t += 1
	b := make([]byte, t)
	for _, v := range s {
		l := len(v)
		copy(b[i:i+l], []byte(v))
		copy(b[i+l:i+l+1], []byte{0})
		i = i + l + 1
	}
	copy(b[i:i+1], []byte{0})
	return &utf16.Encode([]rune(string(b)))[0], nil
}
func readFunc(r io.ReadCloser, w io.Writer) {
	io.Copy(w, r)
	r.Close()
}
func writeFunc(r io.Reader, w io.WriteCloser) {
	io.Copy(w, r)
	w.Close()
}
func wait(proc windows.Handle, d uint32) error {
	switch r, _, err := funcWaitForSingleObject.Call(uintptr(proc), uintptr(d)); {
	case r == 0x00000000:
		return nil
	default:
		return err
	}
}
func (o *options) startInfo() (*windows.StartupInfo, error) {
	s := &windows.StartupInfo{X: o.X, Y: o.Y, XSize: o.W, YSize: o.H, Flags: o.Flags, ShowWindow: o.Mode}
	if len(o.Title) > 0 {
		var err error
		if s.Title, err = syscall.UTF16PtrFromString(o.Title); err != nil {
			return nil, fmt.Errorf("could not convert window title: %w", err)
		}
	}
	o.parent, o.closers = 0, nil
	return s, nil
}
func (o *options) readHandle(r io.Reader) (windows.Handle, error) {
	if r == nil {
		return 0, nil
	}
	var h uintptr
	switch r.(type) {
	case *os.File:
		h = r.(*os.File).Fd()
	case file:
		f, err := r.(file).File()
		if err != nil {
			return 0, err
		}
		h = f.Fd()
	default:
		x, y, err := os.Pipe()
		if err != nil {
			return 0, err
		}
		h = x.Fd()
		o.closers = append(o.closers, x)
		o.closers = append(o.closers, y)
		go writeFunc(r, y)
	}
	if h == 0 {
		return 0, nil
	}
	if o.parent > 0 {
		n, err := dupHandle(o.parent, h)
		if err != nil {
			return 0, err
		}
		return n, nil
	}
	return windows.Handle(h), nil
}
func (o *options) writeHandle(w io.Writer) (windows.Handle, error) {
	if w == nil {
		return 0, nil
	}
	var h uintptr
	switch w.(type) {
	case *os.File:
		h = w.(*os.File).Fd()
	case file:
		f, err := w.(file).File()
		if err != nil {
			return 0, err
		}
		h = f.Fd()
	default:
		x, y, err := os.Pipe()
		if err != nil {
			return 0, err
		}
		h = y.Fd()
		o.closers = append(o.closers, x)
		o.closers = append(o.closers, y)
		go readFunc(x, w)
	}
	if h == 0 {
		return 0, nil
	}
	if o.parent > 0 {
		n, err := dupHandle(o.parent, h)
		if err != nil {
			return 0, err
		}
		return n, nil
	}
	return windows.Handle(h), nil
}
func openProcess(pid int32, access uintptr) (windows.Handle, error) {
	h, _, err := funcOpenProcess.Call(access, 0, uintptr(pid))
	if h == 0 {
		return 0, fmt.Errorf("winapi OpenProcess PID %d error: %w", pid, err)
	}
	return windows.Handle(h), nil
}
func dupHandle(proc windows.Handle, handle uintptr) (windows.Handle, error) {
	i, err := syscall.GetCurrentProcess()
	if err != nil {
		return 0, err
	}
	var n syscall.Handle
	if err = syscall.DuplicateHandle(i, syscall.Handle(handle), syscall.Handle(proc), &n, 0, true, syscall.DUPLICATE_SAME_ACCESS); err != nil {
		return 0, err
	}
	return windows.Handle(n), nil
}
func newParentEx(p windows.Handle, start *windows.StartupInfo) (*startupInfoEx, error) {
	var (
		s uint64
		x startupInfoEx
	)
	_, _, err := funcInitializeProcThreadAttributeList.Call(0, 1, 0, uintptr(unsafe.Pointer(&s)))
	if s < 48 {
		return nil, fmt.Errorf("winapi InitializeProcThreadAttributeList error: %w", err)
	}
	x.AttributeList = new(startupAttrs)
	r, _, err := funcInitializeProcThreadAttributeList.Call(
		uintptr(unsafe.Pointer(x.AttributeList)), 1, 0, uintptr(unsafe.Pointer(&s)),
	)
	if r == 0 {
		return nil, fmt.Errorf("winapi InitializeProcThreadAttributeList error: %w", err)
	}
	if start != nil {
		x.StartupInfo = *start
	}
	x.StartupInfo.Cb = uint32(unsafe.Sizeof(x))
	r, _, err = funcUpdateProcThreadAttribute.Call(
		uintptr(unsafe.Pointer(x.AttributeList)), 0, 0x00020000,
		uintptr(unsafe.Pointer(&p)), uintptr(unsafe.Sizeof(p)),
		uintptr(unsafe.Pointer(nil)), uintptr(unsafe.Pointer(nil)),
	)
	if r == 0 {
		return nil, fmt.Errorf("winapi UpdateProcThreadAttribute error: %w", err)
	}
	return &x, nil
}
func run(name, cmd, dir string, p, t *windows.SecurityAttributes, flags uint32, env *uint16, s *windows.StartupInfo, x *startupInfoEx, u *windows.Token, i *windows.ProcessInformation) error {
	var (
		err     error
		r, z    uintptr
		n, c, d *uint16
	)
	if len(name) > 0 {
		if n, err = windows.UTF16PtrFromString(name); err != nil {
			return err
		}
	}
	if len(cmd) > 0 {
		if c, err = windows.UTF16PtrFromString(cmd); err != nil {
			return err
		}
	}
	if len(dir) > 0 {
		if d, err = windows.UTF16PtrFromString(dir); err != nil {
			return err
		}
	}
	if x == nil && s == nil {
		return ErrNoStartupInfo
	}
	if env != nil {
		flags |= syscall.CREATE_UNICODE_ENVIRONMENT
	}
	if x != nil {
		if x.StartupInfo.Cb == 0 {
			x.StartupInfo.Cb = uint32(unsafe.Sizeof(&x))
		}
		z = uintptr(unsafe.Pointer(x))
		flags |= 0x00080000
	} else {
		if s.Cb == 0 {
			s.Cb = uint32(unsafe.Sizeof(&s))
		}
		z = uintptr(unsafe.Pointer(s))
	}
	if u != nil {
		r, _, err = funcCreateProcessAsUser.Call(
			uintptr(*u),
			uintptr(unsafe.Pointer(n)), uintptr(unsafe.Pointer(c)), uintptr(unsafe.Pointer(p)),
			uintptr(unsafe.Pointer(t)), uintptr(1), uintptr(flags), uintptr(unsafe.Pointer(env)),
			uintptr(unsafe.Pointer(d)), z, uintptr(unsafe.Pointer(i)),
		)
	} else {
		r, _, err = funcCreateProcessW.Call(
			uintptr(unsafe.Pointer(n)), uintptr(unsafe.Pointer(c)), uintptr(unsafe.Pointer(p)),
			uintptr(unsafe.Pointer(t)), uintptr(1), uintptr(flags), uintptr(unsafe.Pointer(env)),
			uintptr(unsafe.Pointer(d)), z, uintptr(unsafe.Pointer(i)),
		)
	}
	if r == 0 {
		return fmt.Errorf("winapi CreateProcessW error: %w", err)
	}
	return nil
}
