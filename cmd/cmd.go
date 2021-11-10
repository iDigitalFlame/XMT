package cmd

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/iDigitalFlame/xmt/util/xerr"
)

const exitStopped uint32 = 0x1337

// Process is a struct that represents an executable command and allows for setting
// options in order change the operating functions.
type Process struct {
	Stdin          io.Reader
	Stdout, Stderr io.Writer
	err            error
	ctx            context.Context
	reader         *os.File
	cancel         context.CancelFunc
	ch             chan finished

	Dir       string
	Env, Args []string
	closers   []*os.File
	opts      options

	Timeout             time.Duration
	flags, exit, cookie uint32
	split               bool
}

// Run will start the process and wait until it completes. This function will
// return the same errors as the 'Start' function if they occur or the 'Wait'
// function if any errors occur during Process runtime.
func (p *Process) Run() error {
	if err := p.Start(); err != nil {
		return err
	}
	return p.Wait()
}

// Wait will block until the Process completes or is terminated by a call to Stop.
// This will start the process if not already started.
func (p *Process) Wait() error {
	if !p.isStarted() {
		return p.Start()
	}
	<-p.ch
	return p.err
}

// Stop will attempt to terminate the currently running Process instance.
// Stopping a Process may prevent the ability to read the Stdout/Stderr and any
// proper exit codes.
func (p *Process) Stop() error {
	if !p.isStarted() || !p.Running() {
		return nil
	}
	return p.stopWith(exitStopped, p.kill(exitStopped))
}

// Start will attempt to start the Process and will return an errors that occur
// while starting the Process.
//
// This function will return 'ErrEmptyCommand' if the 'Args' parameter is empty
// and 'ErrAlreadyStarted' if attempting to start a Process that already has
// been started previously.
func (p *Process) Start() error {
	if p.Running() || p.isStarted() {
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
	if atomic.StoreUint32(&p.cookie, 0); p.reader != nil {
		p.reader.Close()
		p.reader = nil
	}
	p.ch = make(chan finished)
	if err := p.start(true); err != nil {
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
	if !p.isStarted() {
		return false
	}
	select {
	case <-p.ch:
		return false
	default:
	}
	return true
}

// String returns the command and arguments that this Process will execute.
func (p *Process) String() string {
	return strings.Join(p.Args, " ")
}

// NewProcess creates a new process instance that uses the supplied string
// vardict as the command line arguments. Similar to '&Process{Args: s}'.
func NewProcess(s ...string) *Process {
	return &Process{Args: s}
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

// Output runs the Process and returns its standard output. Any returned error
// will usually be of type *ExitError.
func (p *Process) Output() ([]byte, error) {
	if p.Stdout != nil {
		return nil, errStdoutSet
	}
	var b bytes.Buffer
	p.Stdout = &b
	err := p.Run()
	return b.Bytes(), err
}

// ExitCode returns the Exit Code of the process. If the Process is still
// running or has not been started, this function returns an 'ErrStillRunning' error.
func (p *Process) ExitCode() (int32, error) {
	if p.isStarted() && p.Running() {
		return 0, ErrStillRunning
	}
	return int32(p.exit), nil
}

// CombinedOutput runs the Process and returns its combined standard output
// and standard error. Any returned error will usually be of type *ExitError.
func (p *Process) CombinedOutput() ([]byte, error) {
	if p.Stdout != nil {
		return nil, errStdoutSet
	}
	if p.Stderr != nil {
		return nil, errStderrSet
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
			p.kill(exitStopped)
		}
		if err := p.ctx.Err(); s != 2 && err != nil && p.exit == 0 {
			p.err, p.exit = err, c
		}
		if p.opts.close(); len(p.closers) > 0 {
			for i := range p.closers {
				p.closers[i].Close()
				p.closers[i] = nil
			}
			p.closers = nil
		}
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
	if p.isStarted() {
		return nil, ErrAlreadyStarted
	}
	if p.Stdin != nil {
		return nil, errStdinSet
	}
	var err error
	if p.Stdin, p.reader, err = os.Pipe(); err != nil {
		return nil, xerr.Wrap("unable to create Pipe", err)
	}
	return p.reader, nil
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
	if p.isStarted() {
		return nil, ErrAlreadyStarted
	}
	if p.Stdout != nil {
		return nil, errStdoutSet
	}
	r, w, err := os.Pipe()
	if err != nil {
		return nil, xerr.Wrap("unable to create Pipe", err)
	}
	p.Stdout = w
	p.closers = append(p.closers, w)
	return r, nil
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
	if p.isStarted() {
		return nil, ErrAlreadyStarted
	}
	if p.Stdout != nil {
		return nil, errStderrSet
	}
	r, w, err := os.Pipe()
	if err != nil {
		return nil, xerr.Wrap("unable to create Pipe", err)
	}
	p.Stderr = w
	p.closers = append(p.closers, w)
	return r, nil
}

// NewProcessContext creates a new process instance that uses the supplied
// string vardict as the command line arguments.
//
// This function accepts a context that can be used to control the cancelation
// of this process.
func NewProcessContext(x context.Context, s ...string) *Process {
	return &Process{Args: s, ctx: x}
}
