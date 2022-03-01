//go:build windows
// +build windows

package cmd

import (
	"context"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"unsafe"

	"github.com/iDigitalFlame/xmt/cmd/filter"
	"github.com/iDigitalFlame/xmt/device/winapi"
	"github.com/iDigitalFlame/xmt/util/bugtrack"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

var secProtect uint64 = 0x100100000000

var versionOnce struct {
	sync.Once
	v bool
}

type closer uintptr
type file interface {
	File() (*os.File, error)
}
type fileFd interface {
	Fd() uintptr
}
type executable struct {
	r              *os.File
	filter         *filter.Filter
	title          string
	closers        []io.Closer
	i              winapi.ProcessInformation
	token, parent  uintptr
	sf, x, y, w, h uint32
	mode           uint16
}
type closeFunc func() error

func checkVersion() bool {
	versionOnce.Do(func() {
		if v, err := winapi.GetVersion(); err == nil && v > 0 {
			if m := byte(v & 0xFF); m >= 6 {
				versionOnce.v = byte((v&0xFFFF)>>8) >= 1
			}
		}
	})
	return versionOnce.v
}
func wait(h uintptr) error {
	if r, err := winapi.WaitForSingleObject(h, -1); r != 0 {
		return err
	}
	return nil
}
func (e *executable) close() {
	for i := range e.closers {
		e.closers[i].Close()
	}
	e.parent, e.closers = 0, nil
}
func (c closer) Close() error {
	return winapi.CloseHandle(uintptr(c))
}
func (c closeFunc) Close() error {
	return c()
}
func (e *executable) Pid() uint32 {
	return e.i.ProcessID
}

// ResumeProcess will attempt to resume the process via it's PID. This will
// attempt to resume the process using an OS-dependent syscall.
//
// This will not affect already running processes.
func ResumeProcess(p uint32) error {
	h, err := winapi.OpenProcess(0x800, false, p)
	if err != nil {
		return err
	}
	err = winapi.ResumeProcess(h)
	winapi.CloseHandle(h)
	return err
}

// SuspendProcess will attempt to suspend the process via it's PID. This will
// attempt to suspend the process using an OS-dependent syscall.
//
// This will not affect already suspended processes.
func SuspendProcess(p uint32) error {
	h, err := winapi.OpenProcess(0x800, false, p)
	if err != nil {
		return err
	}
	err = winapi.SuspendProcess(h)
	winapi.CloseHandle(h)
	return err
}
func (e *executable) Resume() error {
	return winapi.ResumeProcess(e.i.Process)
}
func (e *executable) Suspend() error {
	return winapi.SuspendProcess(e.i.Process)
}
func (e *executable) isStarted() bool {
	return e.i.ProcessID > 0 && e.i.Process > 0
}
func (e *executable) Handle() uintptr {
	return e.i.Process
}
func (e *executable) SetToken(t uintptr) {
	e.token = t
}
func (e *executable) SetFullscreen(f bool) {
	if f {
		e.sf |= 0x20
	} else {
		e.sf = e.sf &^ 0x20
	}
}
func (e *executable) SetWindowDisplay(m int) {
	if m < 0 {
		e.sf = e.sf &^ 0x1
	} else {
		e.sf |= 0x1
		e.mode = uint16(m)
	}
}
func (e *executable) SetWindowTitle(s string) {
	if len(s) > 0 {
		e.sf |= 0x1000
		e.title = s
	} else {
		e.sf, e.title = e.sf&^0x1000, ""
	}
}
func (executable) SetUID(_ int32, _ *Process) {}
func (executable) SetGID(_ int32, _ *Process) {}
func (e *executable) SetWindowSize(w, h uint32) {
	e.sf |= 0x2
	e.w, e.h = w, h
}
func (executable) SetNoWindow(h bool, p *Process) {
	if h {
		p.flags |= 0x8000000
	} else {
		p.flags = p.flags &^ 0x8000000
	}
}
func (executable) SetDetached(d bool, p *Process) {
	if d {
		p.flags = (p.flags | 0x8) &^ 0x10
	} else {
		p.flags = p.flags &^ 0x8
	}
}
func (executable) SetChroot(_ string, _ *Process) {}
func (executable) SetSuspended(s bool, p *Process) {
	if s {
		p.flags |= 0x4
	} else {
		p.flags = p.flags &^ 0x4
	}
}
func (e *executable) SetWindowPosition(x, y uint32) {
	e.sf |= 0x4
	e.x, e.y = x, y
}
func (*executable) SetNewConsole(c bool, p *Process) {
	if c {
		p.flags |= 0x10
	} else {
		p.flags = p.flags &^ 0x10
	}
}
func (e *executable) kill(x uint32, p *Process) error {
	p.exit = x
	return winapi.TerminateProcess(e.i.Process, x)
}
func createEnvBlock(env []string, split bool) []string {
	if len(env) == 0 && !split {
		return syscall.Environ()[4:]
	}
	var (
		e = syscall.Environ()
		r = make([]string, len(env), len(env)+len(e))
	)
	if copy(r, env); !split {
		// NOTE(dij): If split == true, do NOT add any env vars, but DO
		//            check and add %SYSTEMROOT% if it doesn't exist in the
		//            supplied block.
		r = append(r, e...)
	}
	for i := range r {
		if len(r) > 11 {
			if x := strings.IndexByte(r[i], '='); x > 9 {
				if strings.EqualFold(r[i][:x], sysRoot) {
					return r
				}
			}
		}
	}
	return append(r, sysRoot+"="+os.Getenv(sysRoot))
}
func (e *executable) wait(x context.Context, p *Process) {
	if bugtrack.Enabled {
		defer bugtrack.Recover("cmd.executable.wait()")
	}
	var (
		w   = make(chan error)
		err error
	)
	go func() {
		if bugtrack.Enabled {
			defer bugtrack.Recover("cmd.executable.wait.func1()")
		}
		w <- wait(e.i.Process)
		close(w)
	}()
	select {
	case err = <-w:
	case <-x.Done():
	}
	if err != nil {
		p.stopWith(exitStopped, err)
		return
	}
	if err2 := x.Err(); err2 != nil {
		p.stopWith(exitStopped, err2)
		return
	}
	err = winapi.GetExitCodeProcess(e.i.Process, &p.exit)
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
func (e *executable) writer(w io.Writer) (uintptr, error) {
	var h uintptr
	if w == nil {
		f, err := os.OpenFile(os.DevNull, 1, 0)
		if err != nil {
			return 0, xerr.Wrap("cannot open null device", err)
		}
		e.closers, h = append(e.closers, f), f.Fd()
	} else {
		switch i := w.(type) {
		case file:
			f, err := i.File()
			if err != nil {
				return 0, xerr.Wrap("cannot obtain file handle", err)
			}
			h = f.Fd()
		case fileFd:
			h = i.Fd()
		default:
			x, y, err := os.Pipe()
			if err != nil {
				return 0, xerr.Wrap("cannot create Pipe", err)
			}
			h = y.Fd()
			e.closers = append(e.closers, x)
			e.closers = append(e.closers, y)
			go func() {
				if bugtrack.Enabled {
					defer bugtrack.Recover("cmd.options.writer.func1()")
				}
				io.Copy(w, x)
				x.Close()
			}()
		}
	}
	var (
		v, n uintptr = winapi.CurrentProcess, 0
		err  error
	)
	if e.parent > 0 {
		v = e.parent
	}
	if err = winapi.DuplicateHandle(winapi.CurrentProcess, h, v, &n, 0, true, 0x2); err != nil {
		return 0, xerr.Wrap("DuplicateHandle", err)
	}
	e.closers = append(e.closers, closer(n))
	return n, nil
}
func (e *executable) reader(r io.Reader) (uintptr, error) {
	var h uintptr
	if r == nil {
		f, err := os.OpenFile(os.DevNull, 0, 0)
		if err != nil {
			return 0, xerr.Wrap("cannot open null device", err)
		}
		e.closers, h = append(e.closers, f), f.Fd()
	} else {
		switch i := r.(type) {
		case file:
			f, err := i.File()
			if err != nil {
				return 0, xerr.Wrap("cannot obtain file handle", err)
			}
			h = f.Fd()
		case fileFd:
			h = i.Fd()
		default:
			x, y, err := os.Pipe()
			if err != nil {
				return 0, xerr.Wrap("cannot create Pipe", err)
			}
			h = x.Fd()
			e.closers = append(e.closers, x)
			e.closers = append(e.closers, y)
			go func() {
				if bugtrack.Enabled {
					defer bugtrack.Recover("cmd.options.reader.func1()")
				}
				io.Copy(y, r)
				y.Close()
			}()
		}
	}
	var (
		v, n uintptr = winapi.CurrentProcess, 0
		err  error
	)
	if e.parent > 0 {
		v = e.parent
	}
	if err = winapi.DuplicateHandle(winapi.CurrentProcess, h, v, &n, 0, true, 0x2); err != nil {
		return 0, xerr.Wrap("DuplicateHandle", err)
	}
	e.closers = append(e.closers, closer(n))
	return n, nil
}
func (e *executable) SetParent(f *filter.Filter, p *Process) {
	if e.filter = f; f != nil {
		e.SetNewConsole(true, p)
	}
}
func (e *executable) start(x context.Context, p *Process, sus bool) error {
	r, err := exec.LookPath(p.Args[0])
	if err != nil {
		return err
	}
	v, y, err := e.startInfo()
	if err != nil {
		return err
	}
	if v != nil && v.AttributeList != nil {
		e.closers = append(e.closers, closeFunc(func() error {
			return winapi.DeleteProcThreadAttributeList(v.AttributeList)
		}))
	}
	if p.Stderr != nil || p.Stdout != nil || p.Stdin != nil {
		if v.StartupInfo.StdInput, err = e.reader(p.Stdin); err != nil {
			return err
		}
		if v.StartupInfo.StdOutput, err = e.writer(p.Stdout); err != nil {
			return err
		}
		if p.Stderr == nil || p.Stderr == p.Stdout {
			v.StartupInfo.StdErr = v.StartupInfo.StdOutput
		} else {
			if v.StartupInfo.StdErr, err = e.writer(p.Stderr); err != nil {
				return err
			}
		}
		v.StartupInfo.Flags |= 0x100
	}
	u := e.token
	if u == 0 && e.parent == 0 {
		// NOTE(dij): Handle threads that currently have an impersonated Token
		//            set. This will trigger this function call to use
		//            'CreateProcessWithToken' instead of 'CreateProcess'.
		//            This is only called IF there is no parent Process set, as
		//            Windows permissions cause some fucky stuff to happen.
		//
		//            Failing silently is fine.
		winapi.OpenThreadToken(winapi.CurrentThread, 0xF01FF, true, &u)
	}
	if sus {
		p.flags |= 0x4
	}
	if e.r != nil {
		e.r.Close()
		e.r = nil
	}
	if z := createEnvBlock(p.Env, p.split); u > 0 {
		err = winapi.CreateProcessWithToken(u, 0x2, r, strings.Join(p.Args, " "), p.flags, z, p.Dir, y, v, &e.i)
	} else {
		err = winapi.CreateProcess(r, strings.Join(p.Args, " "), nil, nil, true, p.flags, z, p.Dir, y, v, &e.i)
	}
	if err != nil {
		return err
	}
	e.closers = append(e.closers, closer(e.i.Thread))
	if e.closers = append(e.closers, closer(e.i.Process)); sus {
		return nil
	}
	go e.wait(x, p)
	return nil
}
func (e *executable) startInfo() (*winapi.StartupInfoEx, *winapi.StartupInfo, error) {
	var (
		x   winapi.StartupInfoEx
		err error
	)
	e.close()
	x.StartupInfo.XSize, x.StartupInfo.YSize = e.w, e.h
	x.StartupInfo.Flags, x.StartupInfo.ShowWindow = e.sf, e.mode
	if x.StartupInfo.X, x.StartupInfo.Y = e.x, e.y; len(e.title) > 0 {
		if x.StartupInfo.Title, err = winapi.UTF16PtrFromString(e.title); err != nil {
			return nil, nil, xerr.Wrap("cannot convert title", err)
		}
	}
	if x.StartupInfo.Cb = uint32(unsafe.Sizeof(x)); !checkVersion() {
		return nil, &x.StartupInfo, nil
	}
	if e.filter != nil && !e.filter.Empty() {
		if e.parent, err = e.filter.HandleFunc(0x100CC1, nil); err != nil {
			return nil, nil, err
		}
		e.closers = append(e.closers, closer(e.parent))
	}
	if !protectEnable && e.parent == 0 {
		return nil, &x.StartupInfo, nil
	}
	var (
		s, w uint64
		c    uint32
	)
	if e.parent > 0 && protectEnable {
		w, c = 72, 2
	} else {
		w, c = 48, 1
	}
	if err = winapi.InitializeProcThreadAttributeList(nil, c, &s, w); err != nil {
		return nil, nil, xerr.Wrap("InitializeProcThreadAttributeList", err)
	}
	x.AttributeList = new(winapi.StartupAttributes)
	if err = winapi.InitializeProcThreadAttributeList(x.AttributeList, c, &s, 0); err != nil {
		return nil, nil, xerr.Wrap("InitializeProcThreadAttributeList", err)
	}
	if x.StartupInfo.Cb = uint32(unsafe.Sizeof(x)); e.parent > 0 {
		if err = winapi.UpdateProcThreadAttribute(x.AttributeList, 0x20000, unsafe.Pointer(&e.parent), uint64(unsafe.Sizeof(e.parent)), nil, nil); err != nil {
			winapi.DeleteProcThreadAttributeList(x.AttributeList)
			return nil, nil, xerr.Wrap("UpdateProcThreadAttribute", err)
		}
		c--
	}
	if c == 1 {
		if err = winapi.UpdateProcThreadAttribute(x.AttributeList, 0x20007, unsafe.Pointer(&secProtect), uint64(unsafe.Sizeof(secProtect)), nil, nil); err != nil {
			winapi.DeleteProcThreadAttributeList(x.AttributeList)
			return nil, nil, xerr.Wrap("UpdateProcThreadAttribute", err)
		}
	}
	return &x, nil, nil
}
