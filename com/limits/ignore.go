//go:build !plan9 && !js

package limits

import (
	"os"
	"os/signal"
	"syscall"
)

// Reset will undo all the signals ignored by the 'Ignore' function.
func Reset() {
	signal.Reset(
		syscall.SIGQUIT, syscall.Signal(27), syscall.SIGSEGV, syscall.SIGHUP,
		syscall.SIGABRT, syscall.SIGTRAP,
	)
}

// Ignore is a simple helper method that can be used to ignore signals
// that can be used to abort or generate a stack-trace.
//
// Used for anti-debugging measures.
func Ignore() {
	signal.Ignore(
		syscall.SIGQUIT, syscall.Signal(27), syscall.SIGSEGV, syscall.SIGHUP,
		syscall.SIGABRT, syscall.SIGTRAP,
	)
}
func convertSignal(s uint32) os.Signal {
	return syscall.Signal(s)
}
