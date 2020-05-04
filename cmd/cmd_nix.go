// +build !windows

package cmd

import (
	"os/exec"
	"os/user"
	"strconv"
	"syscall"
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

func (options) close() {
}
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
	}
	if p.exit != 0 {
		p.stopWith(&ExitError{Exit: p.exit})
		return
	}
	p.stopWith(nil)
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
	if err := p.opts.Process.Kill(); err != nil {
		return err
	}
	return nil
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
	p.opts.Dir = p.Dir
	p.opts.Env = p.Env
	p.opts.Stdin = p.Stdin
	p.opts.Stdout = p.Stdout
	p.opts.Stderr = p.Stderr
	if p.flags > 0 {
		p.opts.Cmd.SysProcAttr = &syscall.SysProcAttr{Chroot: p.root}
		if p.flags&flagSetUID != 0 && p.flags&flagSetGID != 0 {
			p.opts.Cmd.SysProcAttr.Credential = &syscall.Credential{Uid: p.uid, Gid: p.gid}
		} else {
			c, err := user.Current()
			if err != nil {
				return err
			}
			var (
				u, _ = strconv.Atoi(c.Uid)
				g, _ = strconv.Atoi(c.Gid)
			)
			p.opts.Cmd.SysProcAttr.Credential = new(syscall.Credential)
			if p.flags&flagSetUID != 0 {
				p.opts.Cmd.SysProcAttr.Credential.Uid = uint32(u)
			}
			if p.flags&flagSetGID != 0 {
				p.opts.Cmd.SysProcAttr.Credential.Gid = uint32(g)
			}
		}
	}
	if err := p.opts.Start(); err != nil {
		return err
	}
	go p.wait()
	return nil
}

// SetFlags will set the startup Flag values used for Windows programs. This function overrites many
// of the 'Set*' functions.
func (p *Process) SetFlags(f uint32) {
	p.flags = f
}

// SetUID will set the process UID at runtime. This function takes the numerical UID value. Use '-1' to disable this
// setting. The UID value is validated at runtime. This function has no effect on Windows devices and will return
// 'ErrNotSupportedOS'.
func (p *Process) SetUID(i int32) error {
	if i < 0 {
		p.flags, p.uid = p.flags&^flagSetUID, 0
		return nil
	}
	p.uid = uint32(i)
	p.flags |= flagSetUID
	return nil
}

// SetGID will set the process GID at runtime. This function takes the numerical GID value. Use '-1' to disable this
// setting. The GID value is validated at runtime. This function has no effect on Windows devices and will return
// 'ErrNotSupportedOS'.
func (p *Process) SetGID(i int32) error {
	if i < 0 {
		p.flags, p.gid = p.flags&^flagSetGID, 0
		return nil
	}
	p.gid = uint32(i)
	p.flags |= flagSetGID
	return nil
}

// SetChroot will set the process Chroot directory at runtime. This function takes the directory path as a string
// value. Use an empty string "" to disable this setting. The specified Path value is validated at runtime. This
// function has no effect on Windows devices and will return 'ErrNotSupportedOS'.
func (p *Process) SetChroot(s string) error {
	if len(s) == 0 {
		p.flags, p.root = p.flags&^flagSetChroot, ""
		return nil
	}
	p.flags |= flagSetChroot
	p.root = s
	return nil
}

// SetParent will instruct the Process to choose a parent with the supplied process name. If this string is empty,
// this will use the current process (default). Always returns 'ErrNotSupportedOS' if the device is not running Windows.
func (*Process) SetParent(_ string) error {
	return ErrNotSupportedOS
}

// SetNoWindow will hide or show the window of the newly spawned process. This function has no effect
// on commands that do not generate windows. Always returns 'ErrNotSupportedOS' if the device is not running Windows.
func (*Process) SetNoWindow(_ bool) error {
	return ErrNotSupportedOS
}

// SetDetached will detach or detach the console of the newly spawned process from the parent. This function
// has no effect on non-console commands. Setting this to true disables SetNewConsole. Always returns 'ErrNotSupportedOS'
// if the device is not running Windows.
func (*Process) SetDetached(_ bool) error {
	return ErrNotSupportedOS
}

// SetSuspended will delay the execution of this Process and will put the process in a suspended state until it
// is resumed using a Resume call. Always returns 'ErrNotSupportedOS' if the device is not running Windows.
func (*Process) SetSuspended(_ bool) error {
	return ErrNotSupportedOS
}

// SetNewConsole will allocate a new console for the newly spawned process. This console output will be
// independent of the parent process. Always returns 'ErrNotSupportedOS' if the device is not running Windows.
func (*Process) SetNewConsole(_ bool) error {
	return ErrNotSupportedOS
}

// SetParentPID will instruct the Process to choose a parent with the supplied process ID. If this number is
// zero, this will use the current process (default) and if < 0 this Process will choose a parent from a list
// of writable processes. Always returns 'ErrNotSupportedOS' if the device is not running Windows.
func (*Process) SetParentPID(_ int32) error {
	return ErrNotSupportedOS
}

// SetFullscreen will set the window fullscreen state of the newly spawned process. This function has no effect
// on commands that do not generate windows. Always returns 'ErrNotSupportedOS' if the device is not running Windows.
func (*Process) SetFullscreen(_ bool) error {
	return ErrNotSupportedOS
}

// SetWindowDisplay will set the window display mode of the newly spawned process. This function has no effect
// on commands that do not generate windows. Always returns 'ErrNotSupportedOS' if the device is not running Windows.
func (*Process) SetWindowDisplay(_ int) error {
	return ErrNotSupportedOS
}

// SetWindowTitle will set the title of the new spawned window to the the specified string. This function
// has no effect on commands that do not generate windows. Always returns 'ErrNotSupportedOS' if the device is not running Windows.
func (*Process) SetWindowTitle(_ string) error {
	return ErrNotSupportedOS
}

// SetWindowSize will set the window display size of the newly spawned process. This function has no effect
// on commands that do not generate windows. Always returns 'ErrNotSupportedOS' if the device is not running Windows.
func (*Process) SetWindowSize(_, _ uint32) error {
	return ErrNotSupportedOS
}

// SetParentRandom will set instruct the Process to choose a parent from the supplied string list on runtime. If this
// list is empty or nil, there is no limit to the name of the chosen process. Always returns 'ErrNotSupportedOS' if
// the device is not running Windows.
func (*Process) SetParentRandom(_ []string) error {
	return ErrNotSupportedOS
}

// SetWindowPosition will set the window postion of the newly spawned process. This function has no effect
// on commands that do not generate windows.
func (*Process) SetWindowPosition(_, _ uint32) error {
	return ErrNotSupportedOS
}
