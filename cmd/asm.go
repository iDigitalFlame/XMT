package cmd

import (
	"context"
	"time"
)

// Assembly is a struct that can be used to contain and run shellcode on Windows devices.
// This struct has many of the functionallies of the standard 'cmd.Program' function.
//
// The 'SetParent*' function will attempt to set the target that runs the shellcode.
// If none are specified, the shellcode will be injected into the current process.
//
// This struct only works on Windows devices.
// All calls on non-Windows devices will return 'ErrNoWindows'.
//
// TODO(dij): Add Linux shellcode execution support.
type Assembly struct {
	Data    []byte
	t       thread
	Timeout time.Duration
}

// Run will start the Assembly thread and wait until it completes. This function
// will return the same errors as the 'Start' function if they occur or the
// 'Wait' function if any errors occur during thread runtime.
//
// Always returns nil on non-Windows devices.
func (a *Assembly) Run() error {
	if err := a.Start(); err != nil {
		return err
	}
	return a.Wait()
}

// Stop will attempt to terminate the currently running thread.
//
// Always returns nil on non-Windows devices.
func (a *Assembly) Stop() error {
	return a.t.Stop()
}

// Wait will block until the thread completes or is terminated by a call to
// Stop.
//
// This function will return 'ErrNotStarted' if the thread has not been started.
func (a *Assembly) Wait() error {
	if !a.t.Running() {
		if err := a.Start(); err != nil {
			return err
		}
	}
	return a.t.Wait()
}

// NewAsm creates a new Assembly thread instance that uses the supplied byte
// array as the Data buffer. Similar to '&Assembly{Data: b}'.
func NewAsm(b []byte) *Assembly {
	return &Assembly{Data: b}
}

// Running returns true if the current thread is running, false otherwise.
func (a *Assembly) Running() bool {
	return a.t.Running()
}

// String returns a string representation of the thread's data, such as the pointer
// and memory addresses.
func (a *Assembly) String() string {
	return "ASM" + a.t.String()
}

// SetSuspended will delay the execution of this thread and will put the
// thread in a suspended state until it is resumed using a Resume call.
//
// This function has no effect if the device is not running Windows.
func (a *Assembly) SetSuspended(s bool) {
	a.t.SetSuspended(s)
}

// ExitCode returns the Exit Code of the thread. If the thread is still running or
// has not been started, this function returns an 'ErrNotCompleted' error.
func (a *Assembly) ExitCode() (int32, error) {
	return a.t.ExitCode()
}

// Handle returns the handle of the current running thread. The return is a uintptr
// that can converted into a Handle.
//
// This function returns an error if the thread was not started. The handle is
// not expected to be valid after the thread exits or is terminated.
func (a *Assembly) Handle() (uintptr, error) {
	return a.t.Handle()
}

// Location returns the in-memory Location of the current Assembly thread, if running.
// The return is a uintptr that can converted into a Handle.
//
// This function returns an error if the Assembly thread was not started. The
// handle is not expected to be valid after the thread exits or is terminated.
func (a *Assembly) Location() (uintptr, error) {
	return a.t.Location()
}

// NewAsmContext creates a new Code thread instance that uses the supplied byte
// array as the Data buffer.
//
// This function accepts a context that can be used to control the cancelation
// of the thread.
func NewAsmContext(x context.Context, b []byte) *Assembly {
	return &Assembly{Data: b, t: thread{ctx: x}}
}
