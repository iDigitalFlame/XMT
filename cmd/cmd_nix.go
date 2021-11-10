//go:build !windows
// +build !windows

package cmd

import (
	"os"
	"os/exec"
	"sync/atomic"
	"syscall"

	"github.com/iDigitalFlame/xmt/device/devtools"
)

const (
	flagSetUID    = 1 << 1
	flagSetGID    = 1 << 2
	flagSetChroot = 1 << 3
)

type options struct {
	*exec.Cmd
	root     string
	uid, gid uint32
}

func (options) close() {}
func (p *Process) wait() {
	err := p.opts.Wait()
	if _, ok := err.(*exec.ExitError); err != nil && !ok {
		p.stopWith(exitStopped, err)
		return
	}
	if err2 := p.ctx.Err(); err2 != nil {
		p.stopWith(exitStopped, err2)
		return
	}
	if atomic.StoreUint32(&p.cookie, 2); p.opts.ProcessState != nil {
		p.exit = uint32(p.opts.ProcessState.ExitCode())
	}
	if p.exit != 0 {
		p.stopWith(p.exit, &ExitError{Exit: p.exit})
		return
	}
	p.stopWith(p.exit, nil)
}

// Pid returns the current process PID. This function returns zero if the
// process has not been started.
func (p *Process) Pid() uint32 {
	if !p.isStarted() {
		return 0
	}
	if p.opts.ProcessState != nil {
		return uint32(p.opts.ProcessState.Pid())
	}
	return uint32(p.opts.Process.Pid)
}

// Resume will attempt to resume this process. This will attempt to resume
// the process using an OS-dependent syscall.
//
// This will not affect already running processes.
func (p *Process) Resume() error {
	if !p.isStarted() || p.opts.Process == nil {
		return ErrNotStarted
	}
	if !p.Running() {
		return nil
	}
	return p.opts.Process.Signal(syscall.SIGCONT)
}

// SetUID will set the process UID at runtime. This function takes the numerical
// UID value. Use '-1' to disable this setting. The UID value is validated at
// runtime.
//
// This function has no effect on Windows devices.
func (p *Process) SetUID(i int32) {
	if i < 0 {
		p.flags, p.opts.uid = p.flags&^flagSetUID, 0
	} else {
		p.opts.uid = uint32(i)
		p.flags |= flagSetUID
	}
}

// SetGID will set the process GID at runtime. This function takes the numerical
// GID value. Use '-1' to disable this setting. The GID value is validated at runtime.
//
//This function has no effect on Windows devices.
func (p *Process) SetGID(i int32) {
	if i < 0 {
		p.flags, p.opts.gid = p.flags&^flagSetGID, 0
	} else {
		p.opts.gid = uint32(i)
		p.flags |= flagSetGID
	}
}

// Suspend will attempt to suspend this process. This will attempt to suspend
// the process using an OS-dependent syscall.
//
// This will not affect already suspended processes.
func (p *Process) Suspend() error {
	if !p.isStarted() || p.opts.Process == nil {
		return ErrNotStarted
	}
	if !p.Running() {
		return nil
	}
	return p.opts.Process.Signal(syscall.SIGSTOP)
}
func (p *Process) isStarted() bool {
	return p.opts.Cmd != nil && p.opts.Cmd.Process != nil
}

// ResumeProcess will attempt to resume the process via it's PID. This will
// attempt to resume the process using an OS-dependent syscall.
//
// This will not affect already running processes.
func ResumeProcess(p uint32) error {
	return syscall.Kill(int(p), syscall.SIGCONT)
}

// SuspendProcess will attempt to suspend the process via it's PID. This will
// attempt to suspend the process using an OS-dependent syscall.
//
// This will not affect already suspended processes.
func SuspendProcess(p uint32) error {
	return syscall.Kill(int(p), syscall.SIGSTOP)
}

// SetNoWindow will hide or show the window of the newly spawned process.
//
// This function has no effect on commands that do not generate windows or
// if the device is not running Windows.
func (Process) SetNoWindow(_ bool) {}

// SetDetached will detach or detach the console of the newly spawned process
// from the parent. This function has no effect on non-console commands. Setting
// this to true disables SetNewConsole.
//
// This function has no effect if the device is not running Windows.
func (Process) SetDetached(_ bool) {}

// SetParent will instruct the Process to choose a parent with the supplied
// process Filter. If the Filter is nil this will use the current process (default).
// Setting the Parent process will automatically set 'SetNewConsole' to true
//
// This function has no effect if the device is not running Windows.
func (Process) SetParent(_ *Filter) {}

// SetFlags will set the startup Flag values used for Windows programs. This
// function overrites many of the 'Set*' functions.
func (p *Process) SetFlags(f uint32) {
	p.flags = f
}

// SetSuspended will delay the execution of this Process and will put the
// process in a suspended state until it is resumed using a Resume call.
//
// This function has no effect if the device is not running Windows.
func (Process) SetSuspended(_ bool) {}

// SetNewConsole will allocate a new console for the newly spawned process.
// This console output will be independent of the parent process.
//
// This function has no effect if the device is not running Windows.
func (Process) SetNewConsole(_ bool) {}

// SetFullscreen will set the window fullscreen state of the newly spawned process.
// This function has no effect on commands that do not generate windows.
//
// This function has no effect if the device is not running Windows.
func (Process) SetFullscreen(_ bool) {}

// SetChroot will set the process Chroot directory at runtime. This function
// takes the directory path as a string value. Use an empty string "" to
// disable this setting. The specified Path value is validated at runtime.
//
// This function has no effect on Windows devices.
func (p *Process) SetChroot(s string) {
	if len(s) == 0 {
		p.flags, p.opts.root = p.flags&^flagSetChroot, ""
	} else {
		p.flags |= flagSetChroot
		p.opts.root = s
	}
}
func (p *Process) start(_ bool) error {
	if len(p.Args) == 1 {
		p.opts.Cmd = exec.CommandContext(p.ctx, p.Args[0])
	} else {
		p.opts.Cmd = exec.CommandContext(p.ctx, p.Args[0], p.Args[1:]...)
	}
	p.opts.Dir, p.opts.Env = p.Dir, p.Env
	p.opts.Stdin, p.opts.Stdout, p.opts.Stderr = p.Stdin, p.Stdout, p.Stderr
	if !p.split {
		z := os.Environ()
		if p.opts.Env == nil {
			p.opts.Env = make([]string, 0, len(z))
		}
		for n := range z {
			p.opts.Env = append(p.opts.Env, z[n])
		}
	}
	if p.flags > 0 {
		p.opts.Cmd.SysProcAttr = &syscall.SysProcAttr{Chroot: p.opts.root}
		switch {
		case p.flags&flagSetUID != 0 && p.flags&flagSetGID != 0:
			p.opts.Cmd.SysProcAttr.Credential = &syscall.Credential{Uid: p.opts.uid, Gid: p.opts.gid}
		case p.flags&flagSetUID != 0 && p.flags&flagSetGID == 0:
			p.opts.Cmd.SysProcAttr.Credential = &syscall.Credential{Uid: p.opts.uid, Gid: uint32(os.Getgid())}
		case p.flags&flagSetUID == 0 && p.flags&flagSetGID != 0:
			p.opts.Cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(os.Getuid()), Gid: p.opts.gid}
		}
	}
	if err := p.opts.Start(); err != nil {
		return err
	}
	go p.wait()
	return nil
}
func (p *Process) kill(e uint32) error {
	if p.exit = e; p.opts.Process == nil {
		return nil
	}
	if err := p.opts.Process.Kill(); err != nil {
		return err
	}
	return nil
}

// SetWindowDisplay will set the window display mode of the newly spawned process.
// This function has no effect on commands that do not generate windows.
//
// See the 'SW_*' values in winuser.h or the Golang windows package documentation for more details.
//
// This function has no effect if the device is not running Windows.
func (Process) SetWindowDisplay(_ int) {}

// SetWindowTitle will set the title of the new spawned window to the the
// specified string. This function has no effect on commands that do not
// generate windows. Setting the value to an empty string will unset this
// setting.
//
// This function has no effect if the device is not running Windows.
func (Process) SetWindowTitle(_ string) {}

// Handle returns the handle of the current running Process. The return is a
// uintptr that can converted into a Handle.
//
// This function returns an error if the Process was not started. The handle
// is not expected to be valid after the Process exits or is terminated.
//
// This function always returns 'ErrNoWindows' on non-Windows devices.
func (Process) Handle() (uintptr, error) {
	return 0, devtools.ErrNoWindows
}

// SetWindowSize will set the window display size of the newly spawned process.
// This function has no effect on commands that do not generate windows.
//
// This function has no effect if the device is not running Windows.
func (Process) SetWindowSize(_, _ uint32) {}

// SetWindowPosition will set the window postion of the newly spawned process.
// This function has no effect on commands that do not generate windows.
//
// This function has no effect if the device is not running Windows.
func (Process) SetWindowPosition(_, _ uint32) {}
