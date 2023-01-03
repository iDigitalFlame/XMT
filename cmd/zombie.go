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

import "context"

// Zombie is a struct that represents an Assembly backed process.
// This is similar to 'execute-assembly' and will launch a suspended process to be
// injected into.
//
// The 'Path' or 'Data' arguments can be used to specify a DLL path or shellcode
// to be ran by the zombie. The 'Data' argument takes precedence over 'Path'.
// At least one of them must be supplied or an 'ErrEmptyCommand' error will be
// returned on any calls to 'Start'.
//
// This struct shares many of the same methods as the 'Process' struct.
// The 'SetParent' function will affect the parent of the spawned process.
type Zombie struct {
	Data []byte
	t    thread
	Process
}

// Run will start the Zombie and wait until it completes. This function will
// return the same errors as the 'Start' function if they occur or the 'Wait'
// function if any errors occur during Process runtime.
func (z *Zombie) Run() error {
	if err := z.Start(); err != nil {
		return err
	}
	return z.t.Wait()
}

// Wait will block until the Zombie completes or is terminated by a call to Stop.
// This will start the process if not already started.
func (z *Zombie) Wait() error {
	if !z.t.Running() {
		if err := z.Start(); err != nil {
			return err
		}
	}
	err := z.t.Wait()
	z.Process.Wait()
	// NOTE(dij): Fix for threads that load a secondary thread to prevent
	//            premature exits.
	return err
}

// Stop will attempt to terminate the currently running Zombie instance.
// Stopping a Zombie may prevent the ability to read the Stdout/Stderr and any
// proper exit codes.
func (z *Zombie) Stop() error {
	if !z.t.Running() {
		return nil
	}
	if err := z.t.Stop(); err != nil {
		return err
	}
	return z.Process.Stop()
}

// Running returns true if the current Zombie is running, false otherwise.
func (z *Zombie) Running() bool {
	return z.t.Running()
}

// Resume will attempt to resume this process. This will attempt to resume
// the process using an OS-dependent syscall.
//
// This will not affect already running processes.
func (z *Zombie) Resume() error {
	return z.t.Resume()
}

// Suspend will attempt to suspend this process. This will attempt to suspend
// the process using an OS-dependent syscall.
//
// This will not affect already suspended processes.
func (z *Zombie) Suspend() error {
	return z.t.Suspend()
}

// SetSuspended will delay the execution of this thread and will put the
// thread in a suspended state until it is resumed using a Resume call.
//
// This function has no effect if the device is not running Windows.
func (z *Zombie) SetSuspended(s bool) {
	z.t.SetSuspended(s)
}

// ExitCode returns the Exit Code of the Zombie thread. If the Zombie is still
// running or has not been started, this function returns an 'ErrStillRunning'
// error.
func (z *Zombie) ExitCode() (int32, error) {
	return z.t.ExitCode()
}

// Handle returns the handle of the current running Zombie. The return is an
// uintptr that can converted into a Handle.
//
// This function returns an error if the Zombie was not started. The handle
// is not expected to be valid after the Process exits or is terminated.
//
// This function always returns 'ErrNoWindows' on non-Windows devices.
func (z *Zombie) Handle() (uintptr, error) {
	return z.t.Handle()
}

// Location returns the in-memory Location of the current Zombie thread, if running.
// The return is an uintptr that can converted into a Handle.
//
// This function returns an error if the Zombie thread was not started. The
// handle is not expected to be valid after the thread exits or is terminated.
func (z *Zombie) Location() (uintptr, error) {
	return z.t.Location()
}

// NewZombie creates a Zombie struct that can be used to spawn a sacrificial
// process specified in the args vardict that will execute the shellcode in the
// byte array.
func NewZombie(b []byte, s ...string) *Zombie {
	return NewZombieContext(context.Background(), b, s...)
}

// NewZombieContext creates a Zombie struct that can be used to spawn a sacrificial
// process specified in the args vardict that will execute the shellcode in the
// byte array.
//
// This function allows for specification of a Context for cancellation.
func NewZombieContext(x context.Context, b []byte, s ...string) *Zombie {
	return &Zombie{Data: b, Process: Process{Args: s, ctx: x}, t: thread{ctx: x}}
}
