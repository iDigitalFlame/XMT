package limits

import (
	"os/signal"
	"syscall"
)

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
