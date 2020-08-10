// +build !windows

package cmd

import (
	"os"
	"os/exec"
	"sync/atomic"
	"syscall"

	"github.com/iDigitalFlame/xmt/device/devtools"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

const (
	flagSetUID    = 1 << 1
	flagSetGID    = 1 << 2
	flagSetChroot = 1 << 3
)

type options struct {
	*exec.Cmd
}
type container struct {
	uid  uint32
	gid  uint32
	root string
}

func (options) close() {}
func (p *Process) wait() {
	err := p.opts.Wait()
	if _, ok := err.(*exec.ExitError); !ok && err != nil {
		p.stopWith(err)
		return
	}
	if p.ctx.Err() != nil {
		return
	}
	if p.opts.ProcessState != nil {
		p.exit = uint32(p.opts.ProcessState.ExitCode())
		atomic.StoreUint32(&p.once, 2)
	}
	if p.exit != 0 {
		p.stopWith(&ExitError{Exit: p.exit})
		return
	}
	p.stopWith(nil)
}

// Fork will attempt to use built-in system utilities to fork off the process into a separate, but similar process.
// If successful, this function will return the PID of the new process.
func Fork() (uint32, error) {
	return 0, xerr.New("currently unimplemented on *nix systems (WIP)")
}

// Pid returns the current process PID. This function returns zero if the process has not been started.
func (p Process) Pid() uint64 {
	if !p.isStarted() {
		return 0
	}
	return uint64(p.opts.Process.Pid)
}
func (p *Process) kill() error {
	p.exit = exitStopped
	if p.opts.Process == nil {
		return nil
	}
	if err := p.opts.Process.Kill(); err != nil {
		return err
	}
	return nil
}

// SetUID will set the process UID at runtime. This function takes the numerical UID value. Use '-1' to disable this
// setting. The UID value is validated at runtime. This function has no effect on Windows devices.
func (p *Process) SetUID(i int32) {
	if i < 0 {
		p.flags, p.uid = p.flags&^flagSetUID, 0
	} else {
		p.uid = uint32(i)
		p.flags |= flagSetUID
	}
}

// SetGID will set the process GID at runtime. This function takes the numerical GID value. Use '-1' to disable this
// setting. The GID value is validated at runtime. This function has no effect on Windows devices.
func (p *Process) SetGID(i int32) {
	if i < 0 {
		p.flags, p.gid = p.flags&^flagSetGID, 0
	} else {
		p.gid = uint32(i)
		p.flags |= flagSetGID
	}
}
func (p Process) isStarted() bool {
	return p.opts != nil
}
func startProcess(p *Process) error {
	p.opts = new(options)
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
		p.opts.Cmd.SysProcAttr = &syscall.SysProcAttr{Chroot: p.root}
		switch {
		case p.flags&flagSetUID != 0 && p.flags&flagSetGID != 0:
			p.opts.Cmd.SysProcAttr.Credential = &syscall.Credential{Uid: p.uid, Gid: p.gid}
		case p.flags&flagSetUID != 0 && p.flags&flagSetGID == 0:
			p.opts.Cmd.SysProcAttr.Credential = &syscall.Credential{Uid: p.uid, Gid: uint32(os.Getgid())}
		case p.flags&flagSetUID == 0 && p.flags&flagSetGID != 0:
			p.opts.Cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(os.Getuid()), Gid: p.gid}
		}
	}
	if err := p.opts.Start(); err != nil {
		return err
	}
	go p.wait()
	return nil
}

// SetParent will instruct the Process to choose a parent with the supplied process name. If this string is empty,
// this will use the current process (default). This function has no effect if the device is not running Windows.
// Setting the Parent process will automatically set 'SetNewConsole' to true.
func (*Process) SetParent(_ string) {}

// SetNoWindow will hide or show the window of the newly spawned process. This function has no effect
// on commands that do not generate windows. This function has no effect if the device is not running Windows.
func (*Process) SetNoWindow(_ bool) {}

// SetDetached will detach or detach the console of the newly spawned process from the parent. This function
// has no effect on non-console commands. Setting this to true disables SetNewConsole. This function has no effect
// if the device is not running Windows.
func (*Process) SetDetached(_ bool) {}

// SetFlags will set the startup Flag values used for Windows programs. This function overrites many
// of the 'Set*' functions.
func (p *Process) SetFlags(f uint32) {
	p.flags = f
}

// SetSuspended will delay the execution of this Process and will put the process in a suspended state until it
// is resumed using a Resume call. This function has no effect if the device is not running Windows.
func (*Process) SetSuspended(_ bool) {}

// SetChroot will set the process Chroot directory at runtime. This function takes the directory path as a string
// value. Use an empty string "" to disable this setting. The specified Path value is validated at runtime. This
// function has no effect on Windows devices.
func (p *Process) SetChroot(s string) {
	if len(s) == 0 {
		p.flags, p.root = p.flags&^flagSetChroot, ""
	} else {
		p.flags |= flagSetChroot
		p.root = s
	}
}

// SetNewConsole will allocate a new console for the newly spawned process. This console output will be
// independent of the parent process. This function has no effect if the device is not running Windows.
func (*Process) SetNewConsole(_ bool) {}

// SetParentPID will instruct the Process to choose a parent with the supplied process ID. If this number is
// zero, this will use the current process (default) and if < 0 this Process will choose a parent from a list
// of writable processes. This function has no effect if the device is not running Windows. Setting the Parent
// process will automatically set 'SetNewConsole' to true.
func (*Process) SetParentPID(_ int32) {}

// SetFullscreen will set the window fullscreen state of the newly spawned process. This function has no effect
// on commands that do not generate windows. This function has no effect if the device is not running Windows.
func (*Process) SetFullscreen(_ bool) {}

// SetWindowDisplay will set the window display mode of the newly spawned process. This function has no effect
// on commands that do not generate windows. This function has no effect if the device is not running Windows.
// See the 'SW_*' values in winuser.h or the Golang windows package documentation for more details.
func (*Process) SetWindowDisplay(_ int) {}

// SetWindowTitle will set the title of the new spawned window to the the specified string. This function
// has no effect on commands that do not generate windows. This function has no effect if the device is not
// running Windows.
func (*Process) SetWindowTitle(_ string) {}

// Handle returns the handle of the current running Process. The return is a uintptr that can converted into a Handle.
// This function returns an error if the Process was not started. The handle is not expected to be valid after the
// Process exits or is terminated. This function always returns 'ErrNoWindows' on non-Windows devices.
func (Process) Handle() (uintptr, error) {
	return 0, devtools.ErrNoWindows
}

// SetWindowSize will set the window display size of the newly spawned process. This function has no effect
// on commands that do not generate windows. This function has no effect if the device is not running Windows.
func (*Process) SetWindowSize(_, _ uint32) {}

// SetParentRandom will set instruct the Process to choose a parent from the supplied string list on runtime. If this
// list is empty or nil, there is no limit to the name of the chosen process. This function has no effect if the
// device is not running Windows. Setting the Parent process will automatically set 'SetNewConsole' to true.
func (*Process) SetParentRandom(_ []string) {}

// SetParentEx will instruct the Process to choose a parent with the supplied process name. If this string
// is empty, this will use the current process (default). This function has no effect if the device is not running
// Windows. Setting the Parent process will automatically set 'SetNewConsole' to true.
//
// If the specified bool is true, this function will attempt to choose a high integrity process and will fail if
// none can be opened or found.
func (*Process) SetParentEx(_ string, _ bool) {}

// SetWindowPosition will set the window postion of the newly spawned process. This function has no effect
// on commands that do not generate windows. This function has no effect if the device is not running Windows.
func (*Process) SetWindowPosition(_, _ uint32) {}

// SetParentRandomEx will set instruct the Process to choose a parent from the supplied string list on runtime.
// If this list is empty or nil, there is no limit to the name of the chosen process. This function has no effect if
// the device is not running Windows. Setting the Parent process will automatically set 'SetNewConsole' to true.
//
// If the specified bool is true, this function will attempt to choose a high integrity process and will fail if
// none can be opened or found.
func (*Process) SetParentRandomEx(_ []string, _ bool) {}
