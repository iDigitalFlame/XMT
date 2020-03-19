// build windows

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

	secStandard uintptr = 0x0001 | 0x00100000 | 0x0400 | 0x0040 | 0x0080
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
	PID     int32
	Name    string
	Choices []string
}

// Stop will attempt to terminate the currently running Process instance. Stopping a Process may prevent the
// ability to read the Stdout/Stderr and any proper exit codes.
func (p *Process) Stop() error {
	if !p.isStarted() || !p.Running() {
		return nil
	}
	p.exit = exitStopped
	if r, _, err := funcTerminateProcess.Call(uintptr(p.opts.info.Process), uintptr(exitStopped)); r == 0 {
		return p.stopWith(err)
	}
	return p.stopWith(nil)
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
	if p.parent != nil {
		i, err := p.parent.pid()
		if err != nil {
			return err
		}
		if p.opts.parent, err = openProcess(i, secStandard); err != nil {
			return err
		}
	}
	if s.StdInput, err = p.opts.readHandle(p.Stdin); err != nil {
		//p.opts.close()
		//p.cancel()
		return err
	}
	if s.StdOutput, err = p.opts.writeHandle(p.Stdout); err != nil {
		//p.opts.close()
		//p.cancel()
		return err
	}
	if p.Stdout == p.Stderr {
		s.StdErr = s.StdOutput
	} else if s.StdErr, err = p.opts.writeHandle(p.Stderr); err != nil {
		//p.opts.close()
		//p.cancel()
		return err
	}
	if s.StdInput > 0 || s.StdOutput > 0 || s.StdErr > 0 {
		s.Flags |= syscall.STARTF_USESTDHANDLES
	}
	var a string
	var e *startupInfoEx
	if p.opts.parent > 0 {
		if e, err = newParentEx(p.opts.parent, s); err != nil {
			//p.opts.close()
			//p.cancel()
			return err
		}
	}
	if len(p.Args) > 1 {
		a = strings.Join(p.Args[1:], " ")
	}
	if err = run(x, a, p.Dir, nil, nil, p.flags, v, s, e, nil, &p.opts.info); err != nil {
		//p.opts.close()
		//p.cancel()
		return err
	}
	go p.wait()
	return nil
}

// SetParent will instruct the Process to choose a parent with the supplied process name. If this string is empty,
// this will use the current process (default). Always returns 'ErrNotSupportedOS' if the device is not running Windows.
func (p *Process) SetParent(n string) error {
	if len(n) == 0 {
		p.parent = nil
		return nil
	}
	if p.parent != nil {
		p.parent.Name = n
		p.parent.PID, p.parent.Choices = 0, nil
	} else {
		p.parent = &container{Name: n}
	}
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
	if i == 0 {
		p.parent = nil
		return nil
	}
	if p.parent != nil {
		p.parent.PID = i
		p.parent.Name, p.parent.Choices = "", nil
	} else {
		p.parent = &container{PID: i}
	}
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
	if p.parent != nil {
		p.parent.Choices = c
		p.parent.PID, p.parent.Name = 0, ""
	} else {
		p.parent = &container{Choices: c}
	}
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
