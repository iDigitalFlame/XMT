package cmd

import (
	"context"
	"strconv"
	"time"
)

// DLL is a struct that can be used to reflectively load a DLL into the memory of a selected process.
// Similar to the Code struct, this struct can only be used on Windows devices and will return 'ErrNoWindows'
// on non-Windows devices.
type DLL struct {
	ctx context.Context
	err error
	ch  chan finished

	Path string
	base
	handle uintptr

	Timeout time.Duration
	exit    uint32
}

// Run will start the DLL and wait until it completes. This function will return the same errors as the 'Start'
// function if they occur or the 'Wait' function if any errors occur during DLL runtime.
func (d *DLL) Run() error {
	if err := d.Start(); err != nil {
		return err
	}
	return d.Wait()
}

// Error returns any errors that may have occurred during the DLL operation.
func (d DLL) Error() error {
	return d.err
}

// Wait will block until the DLL completes or is terminated by a call to Stop. This function will return
// 'ErrNotCompleted' if the DLL has not been started
func (d *DLL) Wait() error {
	if d.handle == 0 {
		return ErrNotCompleted
	}
	<-d.ch
	return d.err
}

// NewDll creates a new DLL instance that uses the supplied string as the DLL file path. Similar to '&DLL{Path: p}'.
func NewDll(p string) *DLL {
	return &DLL{Path: p}
}

// Running returns true if the current DLL is running, false otherwise.
func (d *DLL) Running() bool {
	if d.handle == 0 {
		return false
	}
	select {
	case <-d.ch:
		return false
	default:
		return true
	}
}

// String returns the formatted size/handle of the DLL data.
func (d DLL) String() string {
	if d.handle > 0 {
		return "DLL[" + strconv.FormatUint(uint64(d.handle), 16) + ", " + d.base.String() + "] " + d.Path
	}
	return "DLL " + d.Path
}

// ExitCode returns the Exit Code of the process. If the DLL is still running or has not been started, this
// function returns an 'ErrNotCompleted' error.
func (d DLL) ExitCode() (int32, error) {
	if d.handle > 0 && d.Running() {
		return 0, ErrNotCompleted
	}
	return int32(d.exit), nil
}

// Handle returns the handle of the current running DLL. The return is a uintptr that can converted into a
// Handle. This function returns an error if the DLL was not started. The handle is not expected to be valid
// after the DLL exits or is terminated.
func (d DLL) Handle() (uintptr, error) {
	if d.handle == 0 {
		return 0, ErrNotCompleted
	}
	return d.handle, nil
}

// NewDllContext creates a new DLL instance that uses the supplied string as the DLL file path.
// This function accepts a context that can be used to control the cancelation of this DLL.
func NewDllContext(x context.Context, p string) *DLL {
	return &DLL{Path: p, ctx: x}
}
