//go:build js
// +build js

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

package cmd

import (
	"context"
	"io"
	"os"
	"syscall"

	"github.com/iDigitalFlame/xmt/cmd/filter"
)

type executable struct {
	r       *os.File
	closers []io.Closer
}

func (executable) close() {}
func (executable) Pid() uint32 {
	return 0
}
func (executable) Resume() error {
	return syscall.EINVAL
}
func (executable) Suspend() error {
	return syscall.EINVAL
}

// ResumeProcess will attempt to resume the process via its PID. This will
// attempt to resume the process using an OS-dependent syscall.
//
// This will not affect already running processes.
func ResumeProcess(_ uint32) error {
	return syscall.EINVAL
}
func (executable) Handle() uintptr {
	return 0
}
func (executable) isStarted() bool {
	return false
}
func (executable) isRunning() bool {
	return false
}

// SuspendProcess will attempt to suspend the process via its PID. This will
// attempt to suspend the process using an OS-dependent syscall.
//
// This will not affect already suspended processes.
func SuspendProcess(_ uint32) error {
	return syscall.EINVAL
}
func (executable) SetToken(_ uintptr) {
}
func (executable) SetFullscreen(_ bool) {
}
func (executable) SetWindowDisplay(_ int) {
}
func (executable) SetWindowTitle(_ string) {
}
func (executable) SetLogin(_, _, _ string) {}
func (executable) SetWindowSize(_, _ uint32) {
}
func (executable) SetUID(_ int32, _ *Process) {
}
func (executable) SetGID(_ int32, _ *Process) {
}
func (executable) SetWindowPosition(_, _ uint32) {
}
func (executable) SetChroot(_ string, _ *Process) {
}
func (executable) SetNoWindow(_ bool, _ *Process) {
}
func (executable) SetDetached(_ bool, _ *Process) {
}
func (executable) SetSuspended(_ bool, _ *Process) {
}
func (executable) kill(_ uint32, _ *Process) error {
	return syscall.EINVAL
}
func (executable) SetNewConsole(_ bool, _ *Process) {
}
func (executable) SetParent(_ *filter.Filter, _ *Process) {
}
func (executable) start(_ context.Context, _ *Process, _ bool) error {
	return syscall.EINVAL
}
