// +build linux freebsd netbsd openbsd dragonfly solaris darwin

package cmd

import (
	"errors"
	"os/exec"
)

var exitError = new(exec.ExitError)

type options struct {
	*exec.Cmd
}
type container struct{}

func (p *Process) wait() {
	err := p.opts.Wait()
	if err != nil && !errors.Is(err, exitError) {
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
func (o options) close() {
}

// Stop will attempt to terminate the currently running Process instance. Stopping a Process may prevent the
// ability to read the Stdout/Stderr and any proper exit codes.
func (p *Process) Stop() error {
	if !p.isStarted() || !p.Running() {
		return nil
	}
	p.exit = exitStopped
	if err := p.opts.Process.Kill(); err != nil {
		return p.stopWith(err)
	}
	return p.stopWith(nil)
}
func (p Process) isStarted() bool {
	return p.opts != nil
}
func startProcess(p *Process) error {
	p.opts = &options{}
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
	if err := p.opts.Start(); err != nil {
		return err
	}
	go p.wait()
	return nil
}

// SetParent will instruct the Process to choose a parent with the supplied process name. If this string is empty,
// this will use the current process (default). Always returns 'ErrNotSupportedOS' if the device is not running Windows.
func (p *Process) SetParent(n string) error {
	return ErrNotSupportedOS
}

// SetNoWindow will hide or show the window of the newly spawned process. This function has no effect
// on commands that do not generate windows. Always returns 'ErrNotSupportedOS' if the device is not running Windows.
func (p *Process) SetNoWindow(h bool) error {
	return ErrNotSupportedOS
}

// SetDetached will detach or detach the console of the newly spawned process from the parent. This function
// has no effect on non-console commands. Setting this to true disables SetNewConsole. Always returns 'ErrNotSupportedOS'
// if the device is not running Windows.
func (p *Process) SetDetached(d bool) error {
	return ErrNotSupportedOS
}

// SetSuspended will delay the execution of this Process and will put the process in a suspended state until it
// is resumed using a Resume call. Always returns 'ErrNotSupportedOS' if the device is not running Windows.
func (p *Process) SetSuspended(s bool) error {
	return ErrNotSupportedOS
}

// SetNewConsole will allocate a new console for the newly spawned process. This console output will be
// independent of the parent process. Always returns 'ErrNotSupportedOS' if the device is not running Windows.
func (p *Process) SetNewConsole(c bool) error {
	return ErrNotSupportedOS
}

// SetParentPID will instruct the Process to choose a parent with the supplied process ID. If this number is
// zero, this will use the current process (default) and if < 0 this Process will choose a parent from a list
// of writable processes. Always returns 'ErrNotSupportedOS' if the device is not running Windows.
func (p *Process) SetParentPID(i int32) error {
	return ErrNotSupportedOS
}

// SetFullscreen will set the window fullscreen state of the newly spawned process. This function has no effect
// on commands that do not generate windows. Always returns 'ErrNotSupportedOS' if the device is not running Windows.
func (p *Process) SetFullscreen(f bool) error {
	return ErrNotSupportedOS
}

// SetWindowDisplay will set the window display mode of the newly spawned process. This function has no effect
// on commands that do not generate windows. Always returns 'ErrNotSupportedOS' if the device is not running Windows.
func (p *Process) SetWindowDisplay(m int) error {
	return ErrNotSupportedOS
}

// SetWindowTitle will set the title of the new spawned window to the the specified string. This function
// has no effect on commands that do not generate windows. Always returns 'ErrNotSupportedOS' if the device is not running Windows.
func (p *Process) SetWindowTitle(s string) error {
	return ErrNotSupportedOS
}

// SetWindowSize will set the window display size of the newly spawned process. This function has no effect
// on commands that do not generate windows. Always returns 'ErrNotSupportedOS' if the device is not running Windows.
func (p *Process) SetWindowSize(w, h uint32) error {
	return ErrNotSupportedOS
}

// SetParentRandom will set instruct the Process to choose a parent from the supplied string list on runtime. If this
// list is empty or nil, there is no limit to the name of the chosen process. Always returns 'ErrNotSupportedOS' if
// the device is not running Windows.
func (p *Process) SetParentRandom(c []string) error {
	return ErrNotSupportedOS
}

// SetWindowPosition will set the window postion of the newly spawned process. This function has no effect
// on commands that do not generate windows.
func (p *Process) SetWindowPosition(x, y uint32) error {
	return ErrNotSupportedOS
}
