//go:build windows
// +build windows

package winapi

import "syscall"

const (
	// ErrNoMoreFiles Windows API Error
	//   There are no more files
	ErrNoMoreFiles syscall.Errno = 18

	ErrNoData           syscall.Errno = 232
	ErrPipeBusy         syscall.Errno = 231
	ErrIoPending        syscall.Errno = 997
	ErrBrokenPipe       syscall.Errno = 109
	ErrSemTimeout       syscall.Errno = 121
	ErrBadPathname      syscall.Errno = 161
	ErrInvalidName      syscall.Errno = 123
	ErrIoIncomplete     syscall.Errno = 996
	ErrFileNotFound     syscall.Errno = 2
	ErrPipeConnected    syscall.Errno = 535
	ErrOperationAborted syscall.Errno = 995

	// CurrentThread returns the handle for the current thread. It is a pseudo
	// handle that does not need to be closed.
	CurrentThread = ^uintptr(2 - 1)
	// CurrentProcess returns the handle for the current process. It is a pseudo
	// handle that does not need to be closed.
	CurrentProcess = ^uintptr(0)

	invalid = ^uintptr(0)
)
