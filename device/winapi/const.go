//go:build windows
// +build windows

// Copyright (C) 2020 - 2023 iDigitalFlame
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.
//

package winapi

import "syscall"

// Windows API Specific syscall error values.
const (
	ErrNoData             syscall.Errno = 232
	ErrPipeBusy           syscall.Errno = 231
	ErrIoPending          syscall.Errno = 997
	ErrBrokenPipe         syscall.Errno = 109
	ErrSemTimeout         syscall.Errno = 121
	ErrBadPathname        syscall.Errno = 161
	ErrInvalidName        syscall.Errno = 123
	ErrNoMoreFiles        syscall.Errno = 18
	ErrIoIncomplete       syscall.Errno = 996
	ErrFileNotFound       syscall.Errno = 2
	ErrPipeConnected      syscall.Errno = 535
	ErrOperationAborted   syscall.Errno = 995
	ErrInsufficientBuffer syscall.Errno = 122
)

const (
	// CurrentThread returns the handle for the current thread. It is a pseudo
	// handle that does not need to be closed.
	CurrentThread = ^uintptr(2 - 1)
	// CurrentProcess returns the handle for the current process. It is a pseudo
	// handle that does not need to be closed.
	CurrentProcess = ^uintptr(0)
	layeredPtr     = ^uintptr(19)
	invalid        = ^uintptr(0)
)
