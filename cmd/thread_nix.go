//go:build !windows
// +build !windows

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

	"github.com/iDigitalFlame/xmt/device"
)

type thread struct {
	_ [0]func()
}

func (thread) Pid() uint32 {
	return 0
}
func (thread) Wait() error {
	return device.ErrNoWindows
}
func (thread) Stop() error {
	return device.ErrNoWindows
}
func (thread) Running() bool {
	return false
}
func (thread) Resume() error {
	return device.ErrNoWindows
}
func (thread) Suspend() error {
	return device.ErrNoWindows
}
func (thread) Release() error {
	return nil
}
func (thread) SetSuspended(_ bool) {}
func (thread) Done() <-chan struct{} {
	return nil
}
func (thread) ExitCode() (int32, error) {
	return 0, device.ErrNoWindows
}
func (thread) Handle() (uintptr, error) {
	return 0, device.ErrNoWindows
}
func (thread) Location() (uintptr, error) {
	return 0, device.ErrNoWindows
}
func threadInit(_ context.Context) thread {
	return thread{}
}
