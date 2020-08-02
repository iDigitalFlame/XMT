// +build windows

package cmd

import (
	"io"
	"os"
	"os/exec"
	"strings"
	"unsafe"

	"github.com/iDigitalFlame/xmt/util/xerr"
	"golang.org/x/sys/windows"
)

const (
	flagSize       = 0x00000002
	flagTitle      = 0x00001000
	flagPosition   = 0x00000004
	flagFullscreen = 0x00000020
)

type options struct {
	X, Y  uint32
	W, H  uint32
	Mode  uint16
	Flags uint32
	Title string

	info    windows.ProcessInformation
	parent  windows.Handle
	closers []io.Closer
}
type container struct {
	pid      int32
	name     string
	choices  []string
	elevated bool
}

// Fork will attempt to use built-in system utilities to fork off the process into a separate, but similar process.
// If successful, this function will return the PID of the new process.
func Fork() (uint32, error) {
	var (
		i         processInfo
		r, _, err = funcRtlCloneUserProcess.Call(
			0x00000001|0x00000002, 0, 0, 0, uintptr(unsafe.Pointer(&i)),
		)
	)
	switch r {
	case 0:
		h, err := windows.OpenThread(0x000F|0x00100000|0xffff, false, uint32(i.ClientID.UniqueThread))
		if err != nil {
			return 0, xerr.Wrap("winapi OpenThread error", err)
		}
		if _, err := windows.ResumeThread(h); err != nil {
			return 0, xerr.Wrap("winapi ResumeThread error", err)
		}
		return uint32(i.ClientID.UniqueProcess), windows.CloseHandle(h)
	case 297:
		if r, _, err = funcAllocConsole.Call(); r == 0 {
			return 0, xerr.Wrap("winapi AllocConsole error", err)
		}
		return 0, nil
	}
	return 0, xerr.Wrap("winapi RtlCloneUserProcess error", err)
}

// Pid returns the current process PID. This function returns zero if the process has not been started.
func (p Process) Pid() uint64 {
	if !p.isStarted() {
		return 0
	}
	return uint64(p.opts.info.ProcessId)
}
func (p *Process) kill() error {
	p.exit = exitStopped
	if err := windows.TerminateProcess(p.opts.info.Process, exitStopped); err != nil {
		return err
	}
	return nil
}
func (c container) empty() bool {
	return c.pid == 0 && len(c.name) == 0 && len(c.choices) == 0
}

// SetUID will set the process UID at runtime. This function takes the numerical UID value. Use '-1' to disable this
// setting. The UID value is validated at runtime. This function has no effect on Windows devices.
func (*Process) SetUID(_ int32) {}

// SetGID will set the process GID at runtime. This function takes the numerical GID value. Use '-1' to disable this
// setting. The GID value is validated at runtime. This function has no effect on Windows devices.
func (*Process) SetGID(_ int32) {}

func (p Process) isStarted() bool {
	return p.opts != nil && p.opts.info.Process > 0
}
func startProcess(p *Process) error {
	x, err := exec.LookPath(p.Args[0])
	if err != nil {
		return err
	}
	if p.opts == nil {
		p.opts = new(options)
	}
	s, err := p.opts.startInfo()
	if err != nil {
		return err
	}
	var v *uint16
	if len(p.Env) == 0 && !p.split {
		v, err = createEnv(windows.Environ()[4:])
	} else {
		var (
			f bool
			e = p.Env
		)
		if !p.split {
			z := os.Environ()
			if e == nil {
				e = make([]string, 0, len(z))
			}
			for n := range z {
				if !f && strings.HasPrefix(strings.ToLower(z[n]), "systemroot=") {
					f = true
				}
				e = append(e, z[n])
			}
		}
		for i := 0; !f && i < len(e); i++ {
			if strings.HasPrefix(strings.ToLower(e[i]), "systemroot=") {
				f = true
				break
			}
		}
		if !f {
			v, err = createEnv(append(e, "SYSTEMROOT="+os.Getenv("SYSTEMROOT")))
		} else {
			v, err = createEnv(e)
		}
	}
	if err != nil {
		return err
	}
	if !p.container.empty() {
		if p.opts.parent, err = p.getParent(secStandard); err != nil {
			return err
		}
	}
	m := p.Stderr != nil || p.Stdout != nil || p.Stdin != nil
	if s.StdInput, err = p.opts.readHandle(p.Stdin, m); err != nil {
		return err
	}
	if s.StdOutput, err = p.opts.writeHandle(p.Stdout, m); err != nil {
		return err
	}
	if p.Stdout == p.Stderr {
		s.StdErr = s.StdOutput
	} else if s.StdErr, err = p.opts.writeHandle(p.Stderr, m); err != nil {
		return err
	}
	if m {
		s.Flags |= windows.STARTF_USESTDHANDLES
	}
	var e *startupInfoEx
	if p.opts.parent > 0 {
		if e, err = newParentEx(p.opts.parent, s); err != nil {
			return err
		}
	}
	if err = run(x, strings.Join(p.Args, " "), p.Dir, nil, nil, p.flags, v, s, e, nil, &p.opts.info); err != nil {
		return err
	}
	go p.wait()
	return nil
}

// SetChroot will set the process Chroot directory at runtime. This function takes the directory path as a string
// value. Use an empty string "" to disable this setting. The specified Path value is validated at runtime. This
// function has no effect on Windows devices.
func (*Process) SetChroot(_ string) {}

// SetFlags will set the startup Flag values used for Windows programs. This function overrites many
// of the 'Set*' functions.
func (p *Process) SetFlags(f uint32) {
	p.flags = f
}

// SetParent will instruct the Process to choose a parent with the supplied process name. If this string is empty,
// this will use the current process (default). This function has no effect if the device is not running Windows.
// Setting the Parent process will automatically set 'SetNewConsole' to true.
func (p *Process) SetParent(n string) {
	p.container.clear()
	if len(n) > 0 {
		p.container.name = n
		p.SetNewConsole(true)
	}
}

// SetNoWindow will hide or show the window of the newly spawned process. This function has no effect
// on commands that do not generate windows. This function has no effect if the device is not running Windows.
func (p *Process) SetNoWindow(h bool) {
	if h {
		p.flags |= windows.CREATE_NO_WINDOW
	} else {
		p.flags = p.flags &^ windows.CREATE_NO_WINDOW
	}
}

// SetDetached will detach or detach the console of the newly spawned process from the parent. This function
// has no effect on non-console commands. Setting this to true disables SetNewConsole. This function has no effect
// if the device is not running Windows.
func (p *Process) SetDetached(d bool) {
	if d {
		p.flags |= windows.DETACHED_PROCESS
		p.SetNewConsole(false)
	} else {
		p.flags = p.flags &^ windows.DETACHED_PROCESS
	}
}

// SetSuspended will delay the execution of this Process and will put the process in a suspended state until it
// is resumed using a Resume call. This function has no effect if the device is not running Windows.
func (p *Process) SetSuspended(s bool) {
	if s {
		p.flags |= windows.CREATE_SUSPENDED
	} else {
		p.flags = p.flags &^ windows.CREATE_SUSPENDED
	}
}

// SetNewConsole will allocate a new console for the newly spawned process. This console output will be
// independent of the parent process. This function has no effect if the device is not running Windows.
func (p *Process) SetNewConsole(c bool) {
	if c {
		p.flags |= windows.CREATE_NEW_CONSOLE
	} else {
		p.flags = p.flags &^ windows.CREATE_NEW_CONSOLE
	}
}

// SetParentPID will instruct the Process to choose a parent with the supplied process ID. If this number is
// zero, this will use the current process (default) and if < 0 this Process will choose a parent from a list
// of writable processes. This function has no effect if the device is not running Windows. Setting the Parent
// process will automatically set 'SetNewConsole' to true.
func (p *Process) SetParentPID(i int32) {
	p.container.clear()
	if i != 0 {
		p.container.pid = i
		p.SetNewConsole(true)
	}
}

// SetFullscreen will set the window fullscreen state of the newly spawned process. This function has no effect
// on commands that do not generate windows. This function has no effect if the device is not running Windows.
func (p *Process) SetFullscreen(f bool) {
	if p.opts == nil {
		p.opts = new(options)
	}
	if f {
		p.opts.Flags |= flagFullscreen
	} else {
		p.opts.Flags = p.opts.Flags &^ flagFullscreen
	}
}

// SetWindowDisplay will set the window display mode of the newly spawned process. This function has no effect
// on commands that do not generate windows. This function has no effect if the device is not running Windows.
// See the 'SW_*' values in winuser.h or the Golang windows package documentation for more details.
func (p *Process) SetWindowDisplay(m int) {
	if p.opts == nil {
		p.opts = new(options)
	}
	if m < 0 {
		p.opts.Flags = p.opts.Flags &^ windows.STARTF_USESHOWWINDOW
	} else {
		p.opts.Flags |= windows.STARTF_USESHOWWINDOW
		p.opts.Mode = uint16(m)
	}
}

// SetWindowTitle will set the title of the new spawned window to the the specified string. This function
// has no effect on commands that do not generate windows. This function has no effect if the device is not
// running Windows.
func (p *Process) SetWindowTitle(s string) {
	if p.opts == nil {
		p.opts = new(options)
	}
	if len(s) > 0 {
		p.opts.Flags |= flagTitle
		p.opts.Title = s
	} else {
		p.opts.Flags, p.opts.Title = p.opts.Flags&^flagTitle, ""
	}
}

// Handle returns the handle of the current running Process. The return is a uintptr that can converted into a Handle.
// This function returns an error if the Process was not started. The handle is not expected to be valid after the
// Process exits or is terminated. This function always returns 'ErrNoWindows' on non-Windows devices.
func (p Process) Handle() (uintptr, error) {
	if !p.isStarted() || p.opts == nil {
		return 0, ErrNotCompleted
	}
	return uintptr(p.opts.info.Process), nil
}

// SetWindowSize will set the window display size of the newly spawned process. This function has no effect
// on commands that do not generate windows. This function has no effect if the device is not running Windows.
func (p *Process) SetWindowSize(w, h uint32) {
	if p.opts == nil {
		p.opts = new(options)
	}
	p.opts.Flags |= flagSize
	p.opts.W, p.opts.H = w, h
}

// SetParentRandom will set instruct the Process to choose a parent from the supplied string list on runtime. If this
// list is empty or nil, there is no limit to the name of the chosen process. This function has no effect if the
// device is not running Windows. Setting the Parent process will automatically set 'SetNewConsole' to true.
func (p *Process) SetParentRandom(c []string) {
	if len(c) > 0 {
		p.container.clear()
		p.container.choices = c
		p.SetNewConsole(true)
	} else {
		p.SetParentPID(-1)
	}
}

// SetParentEx will instruct the Process to choose a parent with the supplied process name. If this string
// is empty, this will use the current process (default). This function has no effect if the device is not running
// Windows. Setting the Parent process will automatically set 'SetNewConsole' to true.
//
// If the specified bool is true, this function will attempt to choose a high integrity process and will fail if
// none can be opened or found.
func (p *Process) SetParentEx(n string, e bool) {
	p.container.elevated = e
	p.SetParent(n)
}

// SetWindowPosition will set the window postion of the newly spawned process. This function has no effect
// on commands that do not generate windows. This function has no effect if the device is not running Windows.
func (p *Process) SetWindowPosition(x, y uint32) {
	if p.opts == nil {
		p.opts = new(options)
	}
	p.opts.Flags |= flagPosition
	p.opts.X, p.opts.Y = x, y
}

// SetParentRandomEx will set instruct the Process to choose a parent from the supplied string list on runtime.
// If this list is empty or nil, there is no limit to the name of the chosen process. This function has no effect if
// the device is not running Windows. Setting the Parent process will automatically set 'SetNewConsole' to true.
//
// If the specified bool is true, this function will attempt to choose a high integrity process and will fail if
// none can be opened or found.
func (p *Process) SetParentRandomEx(s []string, e bool) {
	p.container.elevated = e
	p.SetParentRandom(s)
}
