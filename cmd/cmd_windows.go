// +build windows

package cmd

import (
	"errors"
	"io"
	"os/exec"
	"strings"
	"syscall"

	"golang.org/x/sys/windows"
)

const (
	// ModeNone resets the window display mode to default.
	ModeNone = 0
	// ModeHide hides the window and activates another window.
	ModeHide = 0
	// ModeShow activates the window and displays it in its current size and position.
	ModeShow = 5
	// ModeDefault activates and displays a window. If the window is minimized or maximized, the system restores it to its original size and position.
	// An application should specify this flag when displaying the window for the first time.
	ModeDefault = 1
	// ModeRestore activates and displays the window. If the window is minimized or maximized, the system restores it to its
	// original size and position. An application should specify this flag when restoring a minimized window.
	ModeRestore = 9
	// ModeMaximize maximizes the specified window.
	ModeMaximize = 3
	// ModeMinimize minimizes the specified window and activates the next top-level window in the Z order.
	ModeMinimize = 6
	// ModeShowNoActive displays the window in its current size and position. This value is similar to SW_SHOW, except that the window is not activated.
	ModeShowNoActive = 8
	// ModeForceMinimize minimizes a window, even if the thread that owns the window is not responding.
	// This flag should only be used when minimizing windows from a different thread.
	ModeForceMinimize = 11
	// ModeShowMaximized activates the window and displays it as a maximized window.
	ModeShowMaximized = 3
	// ModeShowMinimized activates the window and displays it as a minimized window.
	ModeShowMinimized = 2
	// ModeRestoreNoActive displays a window in its most recent size and position. This value is similar to SW_SHOWNORMAL, except that the window is not activated.
	ModeRestoreNoActive = 4
	// ModeMinimizeNoActive displays the window as a minimized window. This value is similar to SW_SHOWMINIMIZED, except the window is not activated.
	ModeMinimizeNoActive = 7
)
const (
	flagSize        = 0x00000002
	flagTitle       = 0x00001000
	flagPosition    = 0x00000004
	flagDetached    = 0x00000008
	flagSuspended   = 0x00000004
	flagFullscreen  = 0x00000020
	flagNewConsole  = 0x00000010
	flagDisplayMode = 0x00000001

	secStandard uintptr = 0x0001 | 0x0010 | 0x0400 | 0x0040 | 0x0080
)

// ErrNoStartupInfo is an error that is returned when there is no StartupInfo structs passed to the
// startProcess function.
var ErrNoStartupInfo = errors.New("startup info is missing")

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
	pid     int32
	name    string
	choices []string
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
	if r, _, err := funcTerminateProcess.Call(uintptr(p.opts.info.Process), uintptr(exitStopped)); r == 0 {
		return err
	}
	return nil
}
func (c container) empty() bool {
	return c.pid == 0 && len(c.name) == 0 && len(c.choices) == 0
}
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
	v, err := createEnv(p.Env)
	if err != nil {
		return err
	}
	if !p.container.empty() {
		i, err := p.container.getPid()
		if err != nil {
			return err
		}
		if p.opts.parent, err = openProcess(i, secStandard); err != nil {
			return err
		}
	}
	if s.StdInput, err = p.opts.readHandle(p.Stdin); err != nil {
		return err
	}
	if s.StdOutput, err = p.opts.writeHandle(p.Stdout); err != nil {
		return err
	}
	if p.Stdout == p.Stderr {
		s.StdErr = s.StdOutput
	} else if s.StdErr, err = p.opts.writeHandle(p.Stderr); err != nil {
		return err
	}
	if s.StdInput > 0 || s.StdOutput > 0 || s.StdErr > 0 {
		s.Flags |= syscall.STARTF_USESTDHANDLES
	}
	var (
		a string
		e *startupInfoEx
	)
	if p.opts.parent > 0 {
		if e, err = newParentEx(p.opts.parent, s); err != nil {
			return err
		}
	}
	if len(p.Args) > 1 {
		a = strings.Join(p.Args[1:], " ")
	}
	if err = run(x, a, p.Dir, nil, nil, p.flags, v, s, e, nil, &p.opts.info); err != nil {
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
func (*Process) SetUID(_ int32) error {
	return ErrNotSupportedOS
}

// SetGID will set the process GID at runtime. This function takes the numerical GID value. Use '-1' to disable this
// setting. The GID value is validated at runtime. This function has no effect on Windows devices and will return
// 'ErrNotSupportedOS'.
func (*Process) SetGID(_ int32) error {
	return ErrNotSupportedOS
}

// SetChroot will set the process Chroot directory at runtime. This function takes the directory path as a string
// value. Use an empty string "" to disable this setting. The specified Path value is validated at runtime. This
// function has no effect on Windows devices and will return 'ErrNotSupportedOS'.
func (*Process) SetChroot(_ string) error {
	return ErrNotSupportedOS
}

// SetParent will instruct the Process to choose a parent with the supplied process name. If this string is empty,
// this will use the current process (default). Always returns 'ErrNotSupportedOS' if the device is not running Windows.
func (p *Process) SetParent(n string) error {
	p.container.clear()
	if len(n) == 0 {
		return nil
	}
	p.container.name = n
	return nil
}

// SetNoWindow will hide or show the window of the newly spawned process. This function has no effect
// on commands that do not generate windows. Always returns 'ErrNotSupportedOS' if the device is not running Windows.
func (p *Process) SetNoWindow(h bool) error {
	if h {
		p.flags |= syscall.STARTF_USESHOWWINDOW
	} else {
		p.flags = p.flags &^ syscall.STARTF_USESHOWWINDOW
	}
	return nil
}

// SetDetached will detach or detach the console of the newly spawned process from the parent. This function
// has no effect on non-console commands. Setting this to true disables SetNewConsole. Always returns 'ErrNotSupportedOS'
// if the device is not running Windows.
func (p *Process) SetDetached(d bool) error {
	if d {
		p.flags |= flagDetached
		p.SetNewConsole(false)
	} else {
		p.flags = p.flags &^ flagDetached
	}
	return nil
}

// SetSuspended will delay the execution of this Process and will put the process in a suspended state until it
// is resumed using a Resume call. Always returns 'ErrNotSupportedOS' if the device is not running Windows.
func (p *Process) SetSuspended(s bool) error {
	if s {
		p.flags |= flagSuspended
	} else {
		p.flags = p.flags &^ flagSuspended
	}
	return nil
}

// SetNewConsole will allocate a new console for the newly spawned process. This console output will be
// independent of the parent process. Always returns 'ErrNotSupportedOS' if the device is not running Windows.
func (p *Process) SetNewConsole(c bool) error {
	if c {
		p.flags |= flagNewConsole
	} else {
		p.flags = p.flags &^ flagNewConsole
	}
	return nil
}

// SetParentPID will instruct the Process to choose a parent with the supplied process ID. If this number is
// zero, this will use the current process (default) and if < 0 this Process will choose a parent from a list
// of writable processes. Always returns 'ErrNotSupportedOS' if the device is not running Windows.
func (p *Process) SetParentPID(i int32) error {
	p.container.clear()
	if i == 0 {
		return nil
	}
	p.container.pid = i
	return nil
}

// SetFullscreen will set the window fullscreen state of the newly spawned process. This function has no effect
// on commands that do not generate windows. Always returns 'ErrNotSupportedOS' if the device is not running Windows.
func (p *Process) SetFullscreen(f bool) error {
	if p.opts == nil {
		p.opts = new(options)
	}
	if f {
		p.opts.Flags |= flagFullscreen
	} else {
		p.opts.Flags = p.opts.Flags &^ flagFullscreen
	}
	return nil
}

// SetWindowDisplay will set the window display mode of the newly spawned process. This function has no effect
// on commands that do not generate windows. Always returns 'ErrNotSupportedOS' if the device is not running Windows.
func (p *Process) SetWindowDisplay(m int) error {
	if p.opts == nil {
		p.opts = new(options)
	}
	if m > 0 {
		p.opts.Flags |= flagDisplayMode
	} else {
		p.opts.Flags = p.opts.Flags &^ flagDisplayMode
	}
	p.opts.Mode = uint16(m)
	return nil
}

// SetWindowTitle will set the title of the new spawned window to the the specified string. This function
// has no effect on commands that do not generate windows. Always returns 'ErrNotSupportedOS' if the device is not running Windows.
func (p *Process) SetWindowTitle(s string) error {
	if p.opts == nil {
		p.opts = new(options)
	}
	p.opts.Title = s
	p.opts.Flags |= flagTitle
	return nil
}

// SetWindowSize will set the window display size of the newly spawned process. This function has no effect
// on commands that do not generate windows. Always returns 'ErrNotSupportedOS' if the device is not running Windows.
func (p *Process) SetWindowSize(w, h uint32) error {
	if p.opts == nil {
		p.opts = new(options)
	}
	p.opts.W, p.opts.H = w, h
	p.opts.Flags |= flagSize
	return nil
}

// SetParentRandom will set instruct the Process to choose a parent from the supplied string list on runtime. If this
// list is empty or nil, there is no limit to the name of the chosen process. Always returns 'ErrNotSupportedOS' if
// the device is not running Windows.
func (p *Process) SetParentRandom(c []string) error {
	if len(c) == 0 {
		return p.SetParentPID(-1)
	}
	p.container.clear()
	p.container.choices = c
	return nil
}

// SetWindowPosition will set the window postion of the newly spawned process. This function has no effect
// on commands that do not generate windows.
func (p *Process) SetWindowPosition(x, y uint32) error {
	if p.opts == nil {
		p.opts = new(options)
	}
	p.opts.Flags |= flagPosition
	p.opts.X = x
	p.opts.Y = y
	return nil
}
