//go:build js
// +build js

package limits

import (
	"os/signal"
	"syscall"
)

// Reset will undo all the signals ignored by the 'Ignore' function.
func Reset() {
	signal.Reset(syscall.SIGQUIT, syscall.SIGTRAP)
}

// Ignore is a simple helper method that can be used to ignore signals
// that can be used to abort or generate a stack-trace.
//
// Used for anti-debugging measures.
func Ignore() {
	signal.Ignore(syscall.SIGQUIT, syscall.SIGTRAP)
}
