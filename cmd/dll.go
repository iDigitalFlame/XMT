// Copyright (C) 2020 - 2022 iDigitalFlame
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
	"time"
)

// DLL is a struct that can be used to reflectively load a DLL into the memory
// of a selected process. Similar to the Assembly struct, this struct can only
// be used on Windows devices and will return 'ErrNoWindows' on non-Windows devices.
//
// The 'SetParent*' function will attempt to set the target that loads the DLL.
// If none are specified, the DLL will be loaded into the current process.
type DLL struct {
	Path    string
	t       thread
	Timeout time.Duration
}

// Run will start the DLL thread and wait until it completes. This function
// will return the same errors as the 'Start' function if they occur or the
// 'Wait' function if any errors occur during thread runtime.
//
// Always returns nil on non-Windows devices.
func (d *DLL) Run() error {
	if err := d.Start(); err != nil {
		return err
	}
	return d.Wait()
}

// Stop will attempt to terminate the currently running thread.
//
// Always returns nil on non-Windows devices.
func (d *DLL) Stop() error {
	return d.t.Stop()
}

// Wait will block until the thread completes or is terminated by a call to
// Stop.
//
// This function will return 'ErrNotStarted' if the thread has not been started.
func (d *DLL) Wait() error {
	if !d.t.Running() {
		if err := d.Start(); err != nil {
			return err
		}
	}
	return d.t.Wait()
}

// NewDLL creates a new DLL instance that uses the supplied string as the DLL
// file path. Similar to '&DLL{Path: p}'.
func NewDLL(p string) *DLL {
	return &DLL{Path: p}
}

// Running returns true if the current thread is running, false otherwise.
func (d *DLL) Running() bool {
	return d.t.Running()
}

// Release will attempt to release the resources for this DLL instance,
// including handles.
//
// After the first call to this function, all other function calls will fail
// with errors. Repeated calls to this function return nil and are a NOP.
func (d *DLL) Release() error {
	return d.t.Release()
}

// SetSuspended will delay the execution of this thread and will put the
// thread in a suspended state until it is resumed using a Resume call.
//
// This function has no effect if the device is not running Windows.
func (d *DLL) SetSuspended(s bool) {
	d.t.SetSuspended(s)
}

// Done returns a channel that's closed when this DLL completes
//
// This can be used to monitor a DLL's status using a select statement.
func (d *DLL) Done() <-chan struct{} {
	return d.t.Done()
}

// ExitCode returns the Exit Code of the thread. If the thread is still running or
// has not been started, this function returns an 'ErrNotCompleted' error.
func (d *DLL) ExitCode() (int32, error) {
	return d.t.ExitCode()
}

// Handle returns the handle of the current running thread. The return is a uintptr
// that can converted into a Handle.
//
// This function returns an error if the thread was not started. The handle is
// not expected to be valid after the thread exits or is terminated.
func (d *DLL) Handle() (uintptr, error) {
	return d.t.Handle()
}

// NewDLLBytes creates a new DLL instance that uses the supplied raw bytes as
// the binary data to construct the DLL on disk to be executed.
//
// NOTE(dij): This function does a write to disk.
// TODO(dij): In a future release, make this into a reflective loader.
//             Use 'NewAssembly(DLLtoASM(b))' func to bypass this.
func NewDLLBytes(b []byte) (*DLL, error) {
	return NewDLLBytesContext(context.Background(), b)
}

// NewDLLContext creates a new DLL instance that uses the supplied string as
// the DLL file path.
//
// This function accepts a context that can be used to control the cancelation
// of the thread.
func NewDLLContext(x context.Context, p string) *DLL {
	return &DLL{Path: p, t: thread{ctx: x}}
}

// NewDLLBytesContext creates a new DLL instance that uses the supplied raw bytes
// as the binary data to construct the DLL on disk to be executed.
//
// NOTE(dij): This function does a write to disk.
// TODO(dij): In a future release, make this into a reflective loader.
//
// This function accepts a context that can be used to control the cancelation
// of the thread.
func NewDLLBytesContext(x context.Context, b []byte) (*DLL, error) {
	return nil, nil
}
