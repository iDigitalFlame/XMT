package cmd

import (
	"bytes"
	"context"
	"io"
	"os"
	"sync/atomic"
	"time"

	"github.com/iDigitalFlame/xmt/cmd/filter"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

const exitStopped uint32 = 0x1337

// Process is a struct that represents an executable command and allows for setting
// options in order change the operating functions.
type Process struct {
	ctx            context.Context
	Stdout, Stderr io.Writer
	err            error

	Stdin  io.Reader
	ch     chan struct{}
	cancel context.CancelFunc

	Dir       string
	Args, Env []string
	x         executable

	Timeout             time.Duration
	flags, exit, cookie uint32
	split               bool
}

// Run will start the process and wait until it completes.
//
// This function will return the same errors as the 'Start' function if they
// occur or the 'Wait' function if any errors occur during Process runtime.
func (p *Process) Run() error {
	if err := p.Start(); err != nil {
		return err
	}
	return p.Wait()
}

// Pid returns the current process PID. This function returns zero if the
// process has not been started.
func (p *Process) Pid() uint32 {
	if !p.x.isStarted() {
		return 0
	}
	return p.x.Pid()
}

// Wait will block until the Process completes or is terminated by a call to
// Stop.
//
// This will start the process if not already started.
func (p *Process) Wait() error {
	if !p.x.isStarted() {
		return p.Start()
	}
	<-p.ch
	return p.err
}

// Stop will attempt to terminate the currently running Process instance.
//
// Stopping a Process may prevent the ability to read the Stdout/Stderr and any
// proper exit codes.
func (p *Process) Stop() error {
	if !p.x.isStarted() || !p.Running() {
		return nil
	}
	return p.stopWith(exitStopped, p.x.kill(exitStopped, p))
}

// Start will attempt to start the Process and will return an errors that occur
// while starting the Process.
//
// This function will return 'ErrEmptyCommand' if the 'Args' parameter is empty
// and 'ErrAlreadyStarted' if attempting to start a Process that already has
// been started previously.
func (p *Process) Start() error {
	if p.x.isStarted() || p.Running() || atomic.LoadUint32(&p.cookie) > 0 {
		return ErrAlreadyStarted
	}
	if len(p.Args) == 0 {
		return ErrEmptyCommand
	}
	if p.ctx == nil {
		p.ctx = context.Background()
	}
	if p.Timeout > 0 {
		p.ctx, p.cancel = context.WithTimeout(p.ctx, p.Timeout)
	} else {
		p.cancel = func() {}
	}
	p.ch = make(chan struct{})
	atomic.StoreUint32(&p.cookie, 0)
	if err := p.x.start(p.ctx, p, false); err != nil {
		return p.stopWith(exitStopped, err)
	}
	return nil
}

// Flags returns the current set flags value based on the configured options.
func (p *Process) Flags() uint32 {
	return p.flags
}

// Running returns true if the current Process is running, false otherwise.
func (p *Process) Running() bool {
	if !p.x.isStarted() {
		return false
	}
	select {
	case <-p.ch:
		return false
	default:
	}
	return true
}

// Release will attempt to release the resources for this Process, including
// handles.
//
// After the first call to this function, all other function calls will fail
// with errors. Repeated calls to this function return nil and are a NOP.
func (p *Process) Release() error {
	if !p.x.isStarted() {
		return ErrNotStarted
	}
	if atomic.SwapUint32(&p.cookie, 2) != 0 {
		return nil
	}
	atomic.StoreUint32(&p.cookie, 2)
	p.x.close()
	return nil
}

// Resume will attempt to resume this process. This will attempt to resume
// the process using an OS-dependent syscall.
//
// This will not affect already running processes.
func (p *Process) Resume() error {
	if !p.x.isStarted() {
		return ErrNotStarted
	}
	if !p.Running() {
		return nil
	}
	return p.x.Resume()
}

// Suspend will attempt to suspend this process. This will attempt to suspend
// the process using an OS-dependent syscall.
//
// This will not affect already suspended processes.
func (p *Process) Suspend() error {
	if !p.x.isStarted() {
		return ErrNotStarted
	}
	if !p.Running() {
		return nil
	}
	return p.x.Suspend()
}

// SetUID will set the process UID at runtime. This function takes the numerical
// UID value. Use '-1' to disable this setting. The UID value is validated at
// runtime.
//
// This function has no effect on Windows devices.
func (p *Process) SetUID(u int32) {
	p.x.SetUID(u, p)
}

// SetGID will set the process GID at runtime. This function takes the numerical
// GID value. Use '-1' to disable this setting. The GID value is validated at runtime.
//
//This function has no effect on Windows devices.
func (p *Process) SetGID(g int32) {
	p.x.SetGID(g, p)
}

// SetFlags will set the startup Flag values used for Windows programs. This
// function overrites many of the 'Set*' functions.
func (p *Process) SetFlags(f uint32) {
	p.flags = f
}

// NewProcess creates a new process instance that uses the supplied string
// vardict as the command line arguments. Similar to '&Process{Args: s}'.
func NewProcess(s ...string) *Process {
	return &Process{Args: s}
}

// SetToken will set the User or Process Token handle that this Process will
// run under.
//
// WARNING: This may cause issues when running with a parent process.
//
// This function has no effect on commands that do not generate windows or
// if the device is not running Windows.
func (p *Process) SetToken(t uintptr) {
	p.x.SetToken(t)
}

// SetNoWindow will hide or show the window of the newly spawned process.
//
// This function has no effect on commands that do not generate windows or
// if the device is not running Windows.
func (p *Process) SetNoWindow(h bool) {
	p.x.SetNoWindow(h, p)
}

// SetDetached will detach or detach the console of the newly spawned process
// from the parent. This function has no effect on non-console commands. Setting
// this to true disables SetNewConsole.
//
// This function has no effect if the device is not running Windows.
func (p *Process) SetDetached(d bool) {
	p.x.SetDetached(d, p)
}

// SetSuspended will delay the execution of this Process and will put the
// process in a suspended state until it is resumed using a Resume call.
//
// This function has no effect if the device is not running Windows.
func (p *Process) SetSuspended(s bool) {
	p.x.SetSuspended(s, p)
}

// SetInheritEnv will change the behavior of the Environment variable
// inheritance on startup. If true (the default), the current Environment
// variables will be filled in, even if 'Env' is not empty.
//
// If set to false, the current Environment variables will not be added into
// the Process's starting Environment.
func (p *Process) SetInheritEnv(i bool) {
	p.split = !i
}

// SetNewConsole will allocate a new console for the newly spawned process.
// This console output will be independent of the parent process.
//
// This function has no effect if the device is not running Windows.
func (p *Process) SetNewConsole(c bool) {
	p.x.SetNewConsole(c, p)
}

// SetFullscreen will set the window fullscreen state of the newly spawned process.
// This function has no effect on commands that do not generate windows.
//
// This function has no effect if the device is not running Windows.
func (p *Process) SetFullscreen(f bool) {
	p.x.SetFullscreen(f)
}

// SetWindowDisplay will set the window display mode of the newly spawned process.
// This function has no effect on commands that do not generate windows.
//
// See the 'SW_*' values in winuser.h or the Golang windows package documentation for more details.
//
// This function has no effect if the device is not running Windows.
func (p *Process) SetWindowDisplay(m int) {
	p.x.SetWindowDisplay(m)
}

// SetWindowTitle will set the title of the new spawned window to the the
// specified string. This function has no effect on commands that do not
// generate windows. Setting the value to an empty string will unset this
// setting.
//
// This function has no effect if the device is not running Windows.
func (p *Process) SetWindowTitle(s string) {
	p.x.SetWindowTitle(s)
}

// Output runs the Process and returns its standard output. Any returned error
// will usually be of type *ExitError.
func (p *Process) Output() ([]byte, error) {
	if p.Stdout != nil {
		return nil, xerr.Sub("stdout already set", 0x37)
	}
	var b bytes.Buffer
	p.Stdout = &b
	err := p.Run()
	return b.Bytes(), err
}

// Handle returns the handle of the current running Process. The return is a
// uintptr that can converted into a Handle.
//
// This function returns an error if the Process was not started. The handle
// is not expected to be valid after the Process exits or is terminated.
//
// This function always returns 'ErrNoWindows' on non-Windows devices.
func (p *Process) Handle() (uintptr, error) {
	if !p.x.isStarted() {
		return 0, ErrNotStarted
	}
	return p.x.Handle(), nil
}

// ExitCode returns the Exit Code of the process.
//
// If the Process is still running or has not been started, this function returns
// an 'ErrStillRunning' error.
func (p *Process) ExitCode() (int32, error) {
	if p.x.isStarted() && p.Running() {
		return 0, ErrStillRunning
	}
	return int32(p.exit), nil
}

// SetLogin will set the User credentials that this Process will run under.
//
// WARNING: This may cause issues when running with a parent process.
//
// Currently only supported on Windows devices.
func (p *Process) SetLogin(u, d, pw string) {
	p.x.SetLogin(u, d, pw)
}

// SetWindowSize will set the window display size of the newly spawned process.
// This function has no effect on commands that do not generate windows.
//
// This function has no effect if the device is not running Windows.
func (p *Process) SetWindowSize(w, h uint32) {
	p.x.SetWindowSize(w, h)
}

// SetParent will instruct the Process to choose a parent with the supplied
// process Filter. If the Filter is nil this will use the current process (default).
// Setting the Parent process will automatically set 'SetNewConsole' to true
//
// This function has no effect if the device is not running Windows.
func (p *Process) SetParent(f *filter.Filter) {
	p.x.SetParent(f, p)
}

// SetWindowPosition will set the window postion of the newly spawned process.
// This function has no effect on commands that do not generate windows.
//
// This function has no effect if the device is not running Windows.
func (p *Process) SetWindowPosition(x, y uint32) {
	p.x.SetWindowPosition(x, y)
}

// CombinedOutput runs the Process and returns its combined standard output
// and standard error. Any returned error will usually be of type *ExitError.
func (p *Process) CombinedOutput() ([]byte, error) {
	if p.Stdout != nil {
		return nil, xerr.Sub("stdout already set", 0x37)
	}
	if p.Stderr != nil {
		return nil, xerr.Sub("stderr already set", 0x38)
	}
	var b bytes.Buffer
	p.Stdout = &b
	p.Stderr = &b
	err := p.Run()
	return b.Bytes(), err
}
func (p *Process) stopWith(c uint32, e error) error {
	if !p.Running() {
		return e
	}
	if atomic.LoadUint32(&p.cookie) != 1 {
		s := p.cookie
		if atomic.StoreUint32(&p.cookie, 1); p.Running() && s != 2 {
			p.x.kill(exitStopped, p)
		}
		if err := p.ctx.Err(); s != 2 && err != nil && p.exit == 0 {
			p.err, p.exit = err, c
		}
		p.x.close()
		close(p.ch)
	}
	if p.cancel(); p.err == nil && p.ctx.Err() != nil {
		if e != nil {
			p.err = e
			return e
		}
		return nil
	}
	return p.err
}

// StdinPipe returns a pipe that will be connected to the Processes's standard
// input when the Process starts. The pipe will be closed automatically after
// the Processes starts. A caller need only call Close to force the pipe to
// close sooner.
func (p *Process) StdinPipe() (io.WriteCloser, error) {
	if p.x.isStarted() {
		return nil, ErrAlreadyStarted
	}
	if p.Stdin != nil {
		return nil, xerr.Sub("stdin already set", 0x39)
	}
	return p.x.StdinPipe(p)
}

// StdoutPipe returns a pipe that will be connected to the Processes's
// standard output when the Processes starts.
//
// The pipe will be closed after the Processe exits, so most callers need not
// close the pipe themselves. It is thus incorrect to call Wait before all
// reads from the pipe have completed. For the same reason, it is incorrect
// to use Run when using StderrPipe.
//
// See the stdlib StdoutPipe example for idiomatic usage.
func (p *Process) StdoutPipe() (io.ReadCloser, error) {
	if p.x.isStarted() {
		return nil, ErrAlreadyStarted
	}
	if p.Stdout != nil {
		return nil, xerr.Sub("stdout already set", 0x37)
	}
	return p.x.StdoutPipe(p)
}

// StderrPipe returns a pipe that will be connected to the Processes's
// standard error when the Processes starts.
//
// The pipe will be closed after the Processe exits, so most callers need
// not close the pipe themselves. It is thus incorrect to call Wait before all
// reads from the pipe have completed. For the same reason, it is incorrect
// to use Run when using StderrPipe.
//
// See the stdlib StdoutPipe example for idiomatic usage.
func (p *Process) StderrPipe() (io.ReadCloser, error) {
	if p.x.isStarted() {
		return nil, ErrAlreadyStarted
	}
	if p.Stdout != nil {
		return nil, xerr.Sub("stderr already set", 0x38)
	}
	return p.x.StderrPipe(p)
}

// NewProcessContext creates a new process instance that uses the supplied
// string vardict as the command line arguments.
//
// This function accepts a context that can be used to control the cancelation
// of this process.
func NewProcessContext(x context.Context, s ...string) *Process {
	return &Process{Args: s, ctx: x}
}
func (e *executable) StdinPipe(p *Process) (io.WriteCloser, error) {
	var err error
	if p.Stdin, e.r, err = os.Pipe(); err != nil {
		return nil, xerr.Wrap("unable to create Pipe", err)
	}
	e.closers = append(e.closers, p.Stdin.(io.Closer))
	return e.r, nil
}
func (e *executable) StdoutPipe(p *Process) (io.ReadCloser, error) {
	r, w, err := os.Pipe()
	if err != nil {
		return nil, xerr.Wrap("unable to create Pipe", err)
	}
	p.Stdout = w
	e.closers = append(e.closers, w)
	return r, nil
}
func (e *executable) StderrPipe(p *Process) (io.ReadCloser, error) {
	r, w, err := os.Pipe()
	if err != nil {
		return nil, xerr.Wrap("unable to create Pipe", err)
	}
	p.Stderr = w
	e.closers = append(e.closers, w)
	return r, nil
}
