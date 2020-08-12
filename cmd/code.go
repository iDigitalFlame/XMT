package cmd

import (
	"context"
	"strconv"
	"time"
)

// Code is a struct that can be used to contain and run shellcode on Windows devices. This struct has many of the
// functionallies of the standard 'cmd.Program' function. The 'SetParent*' function will attempt to set the target
// that runs the shellcode. If none are specified, the shellcode will be injected into the current process.
// This struct only works on Windows devices. All calls on non-Windows devices will return 'ErrNotSupportedOS'.
type Code struct {
	Data    []byte
	Timeout time.Duration

	ch     chan finished
	ctx    context.Context
	err    error
	exit   uint32
	handle uintptr
	base
}

// Run will start the Code thread and wait until it completes. This function will return the same errors as the 'Start'
// function if they occur or the 'Wait' function if any errors occur during Code thread runtime.
func (c *Code) Run() error {
	if err := c.Start(); err != nil {
		return err
	}
	return c.Wait()
}

// Error returns any errors that may have occurred during the Code thread operation.
func (c Code) Error() error {
	return c.err
}

// Wait will block until the Code thread completes or is terminated by a call to Stop. This function will return
// 'ErrNotCompleted' if the Code thread has not been started
func (c *Code) Wait() error {
	if c.handle == 0 {
		return ErrNotCompleted
	}
	<-c.ch
	return c.err
}

// NewCode creates a new Code thread instance that uses the supplied byte array as the Data buffer .
// Similar to '&Code{Data: b}'.
func NewCode(b []byte) *Code {
	return &Code{Data: b}
}

// Running returns true if the current Code thread is running, false otherwise.
func (c *Code) Running() bool {
	if c.handle == 0 {
		return false
	}
	select {
	case <-c.ch:
		return false
	default:
		return true
	}
}

// String returns the formatted size of the Code thread data.
func (c Code) String() string {
	if c.handle > 0 {
		return "Code[" + strconv.FormatUint(uint64(c.handle), 16) + ", " + c.base.String() + "] " + strconv.Itoa(len(c.Data)) + "B"
	}
	return "Code " + strconv.Itoa(len(c.Data)) + "B"
}

// ExitCode returns the Exit Code of the process. If the Code thread is still running or has not been started, this
// function returns an 'ErrNotCompleted' error.
func (c Code) ExitCode() (int32, error) {
	if c.handle > 0 && c.Running() {
		return 0, ErrNotCompleted
	}
	return int32(c.exit), nil
}

// Handle returns the handle of the current running Code thread. The return is a uintptr that can converted into a
// Handle. This function returns an error if the Code thread was not started. The handle is not expected to be valid
// after the Code thread exits or is terminated.
func (c Code) Handle() (uintptr, error) {
	if c.handle == 0 {
		return 0, ErrNotCompleted
	}
	return c.handle, nil
}

// NewCodeContext creates a new Code thread instance that uses the supplied byte array as the Data buffer.
// This function accepts a context that can be used to control the cancelation of this Code thread.
func NewCodeContext(x context.Context, b []byte) *Code {
	return &Code{Data: b, ctx: x}
}
