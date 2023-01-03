//go:build windows

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
	if t.loc > 0 {
		if t.owner == winapi.CurrentProcess {
			winapi.NtFreeVirtualMemory(t.owner, t.loc)
		} else {
			freeMemory(t.owner, t.loc)
		}
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
	if atomic.SwapUint32(&t.cookie, 2) == 2 {
		return nil
	}
	if t.m > 0 {
		winapi.SetEvent(t.m)
	}
	winapi.CloseHandle(t.hwnd)
	winapi.CloseHandle(t.owner)
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
	go func() {
		if bugtrack.Enabled {
			defer bugtrack.Recover("cmd.(*thread).wait():func1()")
		}
		var e error
		if e = wait(t.hwnd, t.m); p > 0 && i > 0 {
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
		}
		x <- e
		close(x)
	}()
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
	if atomic.SwapUint32(&t.cookie, 2) == 2 {
		t.stopWith(0, nil)
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
	var (
		// 0x20 - PAGE_EXECUTE_READ
		z = uint32(0x20)
		l = uint64(len(b))
	)
	if a > 0 {
		// 0x2 - PAGE_READONLY
		z = 0x2
	}
	if t.owner == winapi.CurrentProcess || t.owner == 0 {
		// 0x4 - PAGE_READWRITE
		if t.loc, err = winapi.NtAllocateVirtualMemory(t.owner, uint32(l), 0x4); err != nil {
			return t.stopWith(exitStopped, err)
		}
		for i := range b {
			(*(*[1]byte)(unsafe.Pointer(t.loc + uintptr(i))))[0] = b[i]
		}
		if _, err = winapi.NtProtectVirtualMemory(t.owner, t.loc, uint32(l), z); err != nil {
			return t.stopWith(exitStopped, err)
		}
	} else if t.loc, err = writeMemory(t.owner, z, l, b); err != nil {
		return t.stopWith(exitStopped, err)
	}
	if a > 0 {
		if t.hwnd, err = winapi.NtCreateThreadEx(t.owner, a, t.loc, t.suspended); err != nil {
			return t.stopWith(exitStopped, err)
		}
		return nil
	}
	if t.hwnd, err = winapi.NtCreateThreadEx(t.owner, t.loc, 0, t.suspended); err != nil {
		return t.stopWith(exitStopped, err)
	}
	return nil
}
