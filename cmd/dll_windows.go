// +build windows

package cmd

import (
	"context"
	"sync/atomic"
	"unsafe"

	"github.com/iDigitalFlame/xmt/util/xerr"
	"golang.org/x/sys/windows"
)

// If DLL loading doesn't work correctly, you can switch "LoadLibraryW" (unicode) for "LoadLibraryA" (ANSI/UTF-8)
const loadLibFunc = "LoadLibraryW" // LoadLibraryA

func (d *DLL) wait() {
	var (
		x   = make(chan error)
		err error
	)
	go waitFunc(windows.Handle(d.handle), windows.INFINITE, x)
	select {
	case err = <-x:
	case <-d.ctx.Done():
	}
	if err != nil {
		d.stopWith(err)
		return
	}
	if d.ctx.Err() != nil {
		d.stopWith(d.ctx.Err())
		return
	}
	funcGetExitCodeThread.Call(uintptr(d.handle), uintptr(unsafe.Pointer(&d.exit)))
	atomic.StoreUint32(&d.once, 2)
	if d.handle = 0; d.exit != 0 {
		d.stopWith(&ExitError{Exit: d.exit})
		return
	}
	d.stopWith(nil)
}
func (d *DLL) close() {
	if d.owner == 0 {
		return
	}
	if d.loc > 0 {
		freeMemory(d.owner, d.loc)
	}
	windows.Close(windows.Handle(d.handle))
	windows.CloseHandle(d.owner)
	d.handle, d.owner, d.loc = 0, 0, 0
}

// Stop will attempt to terminate the currently running DLL instance.
// Always returns nil on non-Windows devices.
func (d *DLL) Stop() error {
	if d.handle == 0 {
		return nil
	}
	return d.stopWith(d.kill())
}
func (d *DLL) kill() error {
	d.exit = exitStopped
	if r, _, err := funcTerminateThread.Call(uintptr(d.handle), uintptr(exitStopped)); r == 0 {
		return err
	}
	return nil
}

// Start will attempt to start the DLL and will return an errors that occur while starting the DLL.
// This function will return 'ErrEmptyCommand' if the 'Data' parameter is empty or nil and 'ErrAlreadyStarted'
// if attempting to start a DLL that already has been started previously. Always returns 'ErrNoWindows'
// on non-Windows devices.
func (d *DLL) Start() error {
	if d.Running() || d.handle > 0 {
		return ErrAlreadyStarted
	}
	if len(d.Path) == 0 {
		return ErrEmptyCommand
	}
	var b []byte
	if loadLibFunc == "LoadLibraryW" {
		p, err := windows.UTF16FromString(d.Path)
		if err != nil {
			return xerr.Wrap(`could not convert "`+d.Path+`" to UTF16 string`, err)
		}
		b := make([]byte, len(p)*2)
		for i := 0; i < len(b); i += 2 {
			b[i], b[i+1] = byte(p[i/2]), byte(p[i/2]>>8)
		}
	} else {
		b = append([]byte(d.Path), 0)
	}
	if d.ctx == nil {
		d.ctx = context.Background()
	}
	if d.cancel == nil {
		if d.Timeout > 0 {
			d.ctx, d.cancel = context.WithTimeout(d.ctx, d.Timeout)
		} else {
			d.ctx, d.cancel = context.WithCancel(d.ctx)
		}
	}
	var err error
	atomic.StoreUint32(&d.once, 0)
	d.ch = make(chan finished)
	if d.owner = windows.CurrentProcess(); !d.container.empty() {
		if d.owner, err = d.container.getParent(secCode); err != nil {
			return d.stopWith(err)
		}
	}
	if d.loc, err = allocateMemory(d.owner, uint32(len(b)), windows.PAGE_READWRITE); err != nil {
		return d.stopWith(err)
	}
	if _, err = writeMemory(d.owner, d.loc, b); err != nil {
		return d.stopWith(err)
	}
	if d.handle, err = createThread(d.owner, funcLoadLibrary.Addr(), d.loc); err != nil {
		return d.stopWith(err)
	}
	go d.wait()
	return nil
}

// SetParent will instruct the DLL to choose a parent with the supplied process name. If this string is empty,
// this will use the current process (default). This function has no effect if the device is not running Windows.
func (d *DLL) SetParent(n string) {
	if d.container.clear(); len(n) > 0 {
		d.container.name = n
	}
}

// SetParentPID will instruct the DLL to choose a parent with the supplied process ID. If this number is
// zero, this will use the current process (default) and if < 0 this Process will choose a parent from a list
// of writable processes. This function has no effect if the device is not running Windows.
func (d *DLL) SetParentPID(i int32) {
	if d.container.clear(); i != 0 {
		d.container.pid = i
	}
}
func (d *DLL) stopWith(e error) error {
	if atomic.LoadUint32(&d.once) != 1 {
		s := d.once
		atomic.StoreUint32(&d.once, 1)
		if d.handle > 0 && s != 2 {
			d.kill()
			d.close()
		}
		if s != 2 && d.ctx.Err() != nil && d.exit == 0 {
			d.err = d.ctx.Err()
			d.exit = exitStopped
		}
		close(d.ch)
	}
	if d.cancel(); d.err == nil && d.ctx.Err() != nil {
		if e != nil {
			d.err = e
			return e
		}
		return nil
	}
	return d.err
}

// SetParentRandom will set instruct the DLL to choose a parent from the supplied string list on runtime.
// If this list is empty or nil, there is no limit to the name of the chosen process. This function has no effect if the
// device is not running Windows.
func (d *DLL) SetParentRandom(s []string) {
	if len(s) == 0 {
		d.SetParentPID(-1)
	} else {
		d.container.clear()
		d.container.choices = s
	}
}

// SetParentEx will instruct the DLL to choose a parent with the supplied process name. If this string
// is empty, this will use the current process (default). This function has no effect if the device is not running
// Windows.
//
// If the specified bool is true, this function will attempt to choose a high integrity process and will fail if
// none can be opened or found.
func (d *DLL) SetParentEx(n string, e bool) {
	d.container.elevated = e
	d.SetParent(n)
}

// SetParentRandomEx will set instruct the DLL to choose a parent from the supplied string list on runtime.
// If this list is empty or nil, there is no limit to the name of the chosen process. This function has no effect if
// the device is not running Windows.
//
// If the specified bool is true, this function will attempt to choose a high integrity process and will fail if
// none can be opened or found.
func (d *DLL) SetParentRandomEx(s []string, e bool) {
	d.container.elevated = e
	d.SetParentRandom(s)
}
