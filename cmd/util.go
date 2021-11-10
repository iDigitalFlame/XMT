package cmd

import (
	"strconv"

	"github.com/iDigitalFlame/xmt/util/xerr"
)

var (
	// ErrNotStarted is an error returned by multiple functions functions when attempting to access a
	// Runnable function that requires the Runnable to be started first.
	ErrNotStarted = xerr.New("process has not been started")
	// ErrEmptyCommand is an error returned when attempting to start a Runnable that has empty arguments.
	ErrEmptyCommand = xerr.New("process arguments are empty")
	// ErrStillRunning is returned when attempting to access the exit code on a Runnable.
	ErrStillRunning = xerr.New("process is still running")
	// ErrAlreadyStarted is an error returned by the 'Start' or 'Run' functions when attempting to start
	// a Runnable that has already been started via a 'Start' or 'Run' function call.
	ErrAlreadyStarted = xerr.New("process has already been started")
	// ErrNoProcessFound is returned by the SetParent* functions on Windows devices when a specified parent
	// could not be found.
	ErrNoProcessFound = xerr.New("could not find a suitable parent")

	errStdinSet  = xerr.New("process Stdin already set")
	errStderrSet = xerr.New("process Stderr already set")
	errStdoutSet = xerr.New("process Stdout already set")
)

// ExitError is a type of error that is returned by the Wait and Run functions when a function
// returns an error code other than zero.
type ExitError struct {
	Exit uint32
}
type finished interface{}

// Runnable is an interface that helps support the type of structs that can be used for execution, such as
// Assembly, DLL and Process, which share the same methods as this interface.
type Runnable interface {
	Run() error
	Pid() uint32
	Wait() error
	Stop() error
	Start() error
	Running() bool
	SetParent(*Filter)
	ExitCode() (int32, error)
}

// Split will attempt to split the specified string based on the escape characters and spaces while attempting
// to preserve anything that is not a splitting space. This will automatically detect quotes and backslashes. The
// return result is a string array that can be used as args.
//
// TODO(dij): Refactor
func Split(s string) []string {
	var (
		b       []rune
		r       []string
		l, e, i rune
	)
	for _, c := range s {
		switch {
		case c == ' ' && i == 0 && e == 0 && len(b) > 0:
			r, b = append(r, string(b)), nil
		case c == '"' && i == 0 && e == 0:
			fallthrough
		case c == '\'' && i == 0 && e == 0:
			i = c
		case c == '"' && i == c && e == 0:
			fallthrough
		case c == '\'' && i == c && e == 0:
			if len(b) > 0 {
				r, b = append(r, string(b)), nil
			}
			i = 0
		case c == ' ' && i == 0 && e > 0:
			fallthrough
		case c == '"' && i == 0 && e > 0:
			fallthrough
		case c == '\'' && i == 0 && e > 0:
			b, e = append(b, c), 0
		case c == '\\' && e > 0:
			e = 0
		case c == '\\' && e == 0:
			e = c
		case c == i && e > 0:
			e = 0
			fallthrough
		default:
			if e > 0 {
				b = append(b, e)
				e = 0
			}
			b = append(b, c)
		}
		l = c
	}
	if i > 0 && (l == '"' || l == '\'') {
		b = append(b, l)
	}
	if len(b) > 0 {
		r = append(r, string(b))
	}
	return r
}

// Error fulfills the error interface and retruns a formatted string that specifies the Process Exit Code.
func (e ExitError) Error() string {
	return "process exit: " + strconv.Itoa(int(e.Exit))
}
