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

package cmd

import (
	"context"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/iDigitalFlame/xmt/cmd/filter"
	"github.com/iDigitalFlame/xmt/device/winapi"
	"github.com/iDigitalFlame/xmt/util"
	"github.com/iDigitalFlame/xmt/util/bugtrack"
)

type thread struct {
	ctx                 context.Context
	err                 error
	callback            func()
	ch                  chan struct{}
	cancel              context.CancelFunc
	filter              *filter.Filter
	hwnd, loc, owner, m uintptr
	exit, cookie        uint32
	suspended           bool
}

func (t *thread) close() {
	if t.hwnd == 0 || t.owner == 0 {
		return
	}
	if t.loc > 0 && atomic.LoadUint32(&t.cookie)&cookieRelease == 0 {
		if t.owner == winapi.CurrentProcess {
			winapi.NtFreeVirtualMemory(t.owner, t.loc, 0)
		} else {
			freeMemory(t.owner, t.loc)
		}
	}
	if t.callback != nil {
		t.callback()
	} else {
		// NOTE(dij): We only need to close the owner if there is no callback as
		//            it's usually a Zombie that'll handle it's own handle closure.
		winapi.CloseHandle(t.owner)
	}
	winapi.CloseHandle(t.hwnd)
	t.hwnd, t.owner, t.loc = 0, 0, 0
}
func (t *thread) kill() error {
	if t.hwnd == 0 {
		return t.err
	}
	t.exit = exitStopped
	atomic.StoreUint32(&t.cookie, atomic.LoadUint32(&t.cookie)|cookieStopped)
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
	if atomic.SwapUint32(&t.cookie, atomic.LoadUint32(&t.cookie)|cookieStopped|cookieRelease)&cookieStopped != 0 {
		return nil
	}
	if t.m > 0 {
		winapi.SetEvent(t.m)
	}
	return t.stopWith(0, nil)
}
func (t *thread) Suspend() error {
	if !t.Running() || t.hwnd == 0 {
		return ErrNotStarted
	}
	_, err := winapi.SuspendThread(t.hwnd)
	return err
}
func (t *thread) wait(p, i uint32) {
	var (
		x   = make(chan error)
		err error
	)
	if t.m, err = winapi.CreateEvent(nil, false, false, ""); err != nil {
		if bugtrack.Enabled {
			bugtrack.Track("cmd.(*thread).wait(): Creating Event failed, falling back to single wait: %s", err.Error())
		}
	}
	go t.waitInner(x, p, i)
	select {
	case err = <-x:
	case <-t.ctx.Done():
	}
	if t.m > 0 {
		winapi.CloseHandle(t.m)
		t.m = 0
	}
	if err != nil {
		t.stopWith(exitStopped, err)
		return
	}
	if err2 := t.ctx.Err(); err2 != nil {
		t.stopWith(exitStopped, err2)
		return
	}
	if atomic.SwapUint32(&t.cookie, atomic.LoadUint32(&t.cookie)|cookieStopped)&cookieStopped != 0 {
		t.stopWith(0, nil)
		return
	}
	if err = winapi.GetExitCodeThread(t.hwnd, &t.exit); err != nil {
		t.stopWith(exitStopped, err)
		return
	}
	if t.exit != 0 {
		t.stopWith(t.exit, &ExitError{Exit: t.exit})
		return
	}
	t.stopWith(t.exit, nil)
}
func (t *thread) SetSuspended(s bool) {
	t.suspended = s
}
func (t *thread) Done() <-chan struct{} {
	return t.ch
}
func nextNonThread(p, i uint32) uintptr {
	var (
		n   uintptr
		err error
	)
	winapi.EnumThreads(p, func(e winapi.ThreadEntry) error {
		if e.TID == i {
			return nil
		}
		// 0x120043 - READ_CONTROL | SYNCHRONIZE | THREAD_QUERY_INFORMATION |
		//            THREAD_SET_INFORMATION | THREAD_SUSPEND_RESUME | THREAD_TERMINATE
		if n, err = e.Handle(0x120043); err == nil {
			return winapi.ErrNoMoreFiles
		}
		return nil
	})
	return n
}
func threadInit(x context.Context) thread {
	return thread{ctx: x}
}
func (t *thread) Handle() (uintptr, error) {
	if t.hwnd == 0 {
		return 0, ErrNotStarted
	}
	return t.hwnd, nil
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
	return t.loc, nil
}
func (t *thread) stopWith(c uint32, e error) error {
	if !t.Running() {
		return e
	}
	if atomic.LoadUint32(&t.cookie)&cookieFinal == 0 {
		if atomic.SwapUint32(&t.cookie, t.cookie|cookieStopped|cookieFinal)&cookieStopped == 0 && t.hwnd > 0 {
			t.kill()
		}
		if err := t.ctx.Err(); err != nil && t.exit == 0 {
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
func (t *thread) waitInner(x chan<- error, p, i uint32) {
	if bugtrack.Enabled {
		defer bugtrack.Recover("cmd.(*thread).waitInner()")
	}
	e := wait(t.hwnd, t.m)
	if p == 0 || i == 0 {
		x <- e
		close(x)
		return
	}
	// If we have more threads (that are not our zombie thread) switch
	// to watch that one until we have none left.
	if n := nextNonThread(p, i); n > 0 {
		winapi.CloseHandle(t.hwnd)
		for t.hwnd = n; t.hwnd > 0; {
			e = wait(t.hwnd, t.m)
			if n = nextNonThread(p, i); n == 0 {
				break
			}
			winapi.CloseHandle(t.hwnd)
			t.hwnd = n
		}
	}
	x <- e
	close(x)
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
		// (old 0x47B - PROCESS_CREATE_THREAD | PROCESS_QUERY_INFORMATION | PROCESS_VM_READ |
		//               PROCESS_VM_WRITE | PROCESS_VM_OPERATION | PROCESS_DUP_HANDLE | PROCESS_TERMINATE)
		//
		// 0x43A - PROCESS_CREATE_THREAD | PROCESS_QUERY_INFORMATION | PROCESS_VM_READ |
		//         PROCESS_VM_WRITE | PROCESS_VM_OPERATION
		// NOTE(dij): Not adding 'PROCESS_QUERY_LIMITED_INFORMATION' here as we
		//            need more permissions here such as 'PROCESS_VM_*' stuff.
		if t.owner, err = t.filter.HandleFunc(0x43A, nil); err != nil {
			return t.stopWith(exitStopped, err)
		}
	}
	// 0x20 - PAGE_EXECUTE_READ
	z := uint32(0x20)
	if a > 0 {
		// 0x2 - PAGE_READONLY
		z = 0x2
	}
	// Add a bit of "randomness" to where we start and end for funsies.
	var (
		s, e = uint64(util.FastRandN(2048)), uint64(util.FastRandN(2048))
		v    = make([]byte, len(b)+int(s+e))
		l    = uint64(len(v))
	)
	// NOTE(dij): We use the for loops here instead of 'Rand.Read' for readability
	//            and to prevent 'v' from escaping.
	for i := uint64(0); i < s; i++ { // If 's' is zero, this should be a NOP.
		v[i] = byte(util.FastRandN(256))
	}
	if copy(v[s:], b); e > 0 {
		for i := uint64(len(b)); i < l; i++ {
			v[i] = byte(util.FastRandN(256))
		}
	}
	if t.owner == winapi.CurrentProcess || t.owner == 0 {
		// 0x4 - PAGE_READWRITE
		if t.loc, err = winapi.NtAllocateVirtualMemory(t.owner, uint32(l), 0x4); err != nil {
			return t.stopWith(exitStopped, err)
		}
		for i := range v {
			(*(*[1]byte)(unsafe.Pointer(t.loc + uintptr(i))))[0] = v[i]
		}
		if _, err = winapi.NtProtectVirtualMemory(t.owner, t.loc, uint32(l), z); err != nil {
			return t.stopWith(exitStopped, err)
		}
	} else if t.loc, err = writeMemory(t.owner, z, l, v); err != nil {
		return t.stopWith(exitStopped, err)
	}
	if a > 0 {
		if t.hwnd, err = winapi.NtCreateThreadEx(t.owner, a, t.loc+uintptr(s), t.suspended); err != nil {
			return t.stopWith(exitStopped, err)
		}
		return nil
	}
	if t.hwnd, err = winapi.NtCreateThreadEx(t.owner, t.loc+uintptr(s), 0, t.suspended); err != nil {
		return t.stopWith(exitStopped, err)
	}
	return nil
}
