//go:build windows
// +build windows

package cmd

import (
	"github.com/iDigitalFlame/xmt/util/xerr"
	"golang.org/x/sys/windows"
)

// If DLL loading doesn't work correctly, you can switch "LoadLibraryW" (unicode)
// for "LoadLibraryA" (ANSI/UTF-8).
const loadLibFunc = "LoadLibraryW" // LoadLibraryA

// Pid retruns the process ID of the owning process (the process running
// the thread.)
//
// This may return zero if the thread has not yet been started.
func (d *DLL) Pid() uint32 {
	return d.t.Pid()
}

// Start will attempt to start the DLL and will return an errors that occur while
// starting the DLL.
//
// This function will return 'ErrEmptyCommand' if the 'Path' parameter is empty
// and 'ErrAlreadyStarted' if attempting to start a DLL that already has been
// started previously.
//
// Always returns 'ErrNoWindows' on non-Windows devices.
func (d *DLL) Start() error {
	if len(d.Path) == 0 {
		return ErrEmptyCommand
	}
	if d.Running() {
		return ErrAlreadyStarted
	}
	var b []byte
	if loadLibFunc == "LoadLibraryW" {
		p, err := windows.UTF16FromString(d.Path)
		if err != nil {
			return xerr.Wrap("could not convert path", err)
		}
		b = make([]byte, len(p)*2)
		for i := 0; i < len(b); i += 2 {
			b[i], b[i+1] = byte(p[i/2]), byte(p[i/2]>>8)
		}
	} else {
		b = append([]byte(d.Path), 0)
	}
	if err := d.t.Start(0, d.Timeout, windows.Handle(funcLoadLibrary.Addr()), b); err != nil {
		return err
	}
	go d.t.wait()
	return nil
}

// SetParent will instruct the DLL to choose a parent with the supplied process
// Filter. If the Filter is nil this will use the current process (default).
//
// This function has no effect if the device is not running Windows.
func (d *DLL) SetParent(f *Filter) {
	d.t.filter = f
}
