package cmd

import (
	"context"
	"fmt"
	"time"
)

// Code is a struct that can be used to contain and run shellcode on Windows devices. This struct has many of the
// functionallies of the standard 'cmd.Program' function. The 'SetParent*' function will attempt to set the target
// that runs the shellcode. If none are specified, the shellcode will be injected into the current process.
// This struct only works on Windows devices. All calls on non-Windows devices will return 'ErrNotSupportedOS'.
type Code struct {
	Data    []byte
	Timeout time.Duration

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

// NewCode creates a new Code thread instance that uses the supplied byte array as the Data buffer .
// Similar to '&Code{Data: b}'.
func NewCode(b []byte) *Code {
	return &Code{Data: b}
}

// Running returns true if the current Code thread is running, false otherwise.
func (c Code) Running() bool {
	return c.handle > 0 && c.ctx != nil && c.ctx.Err() == nil
}

// String returns the formatted size of the Code thread data.
func (c Code) String() string {
	if c.handle > 0 {
		return fmt.Sprintf("Code[0x%X, %s] %d bytes", c.handle, c.base.String(), len(c.Data))
	}
	return fmt.Sprintf("Code %d bytes", len(c.Data))
}

// ExitCode returns the Exit Code of the process. If the Process is still running or has not been started, this
// function returns an 'ErrNotCompleted' error.
func (c Code) ExitCode() (int32, error) {
	if c.handle > 0 && c.Running() {
		return 0, ErrNotCompleted
	}
	return int32(c.exit), nil
}

// NewCodeContext creates a new Code thread instance that uses the supplied byte array as the Data buffer.
// This function accepts a context that can be used to control the cancelation of this Code thread.
func NewCodeContext(x context.Context, b []byte) *Code {
	return &Code{Data: b, ctx: x}
}
