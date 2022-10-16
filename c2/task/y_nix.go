//go:build !windows

// Copyright (C) 2020 - 2022 iDigitalFlame
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

package task

import (
	"context"

	"github.com/iDigitalFlame/xmt/cmd"
	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/device"
)

func taskTroll(_ context.Context, _ data.Reader, _ data.Writer) error {
	return device.ErrNoWindows
}
func taskCheck(_ context.Context, _ data.Reader, _ data.Writer) error {
	return device.ErrNoWindows
}
func taskPatch(_ context.Context, _ data.Reader, _ data.Writer) error {
	return device.ErrNoWindows
}
func taskInject(_ context.Context, _ data.Reader, _ data.Writer) error {
	return device.ErrNoWindows
}
func taskZombie(_ context.Context, _ data.Reader, _ data.Writer) error {
	return device.ErrNoWindows
}
func taskUntrust(_ context.Context, _ data.Reader, _ data.Writer) error {
	return device.ErrNoWindows
}
func taskRegistry(_ context.Context, _ data.Reader, _ data.Writer) error {
	return device.ErrNoWindows
}
func taskInteract(_ context.Context, _ data.Reader, _ data.Writer) error {
	return device.ErrNoWindows
}
func taskLoginsAct(_ context.Context, _ data.Reader, _ data.Writer) error {
	return device.ErrNoWindows
}
func taskWindowList(_ context.Context, _ data.Reader, _ data.Writer) error {
	return device.ErrNoWindows
}
func taskLoginsProc(_ context.Context, _ data.Reader, _ data.Writer) error {
	return device.ErrNoWindows
}

// DLLUnmarshal will read this DLL's struct data from the supplied reader and
// returns a DLL runnable struct along with the wait and delete status booleans.
//
// This function returns an error if building or reading fails or if the device
// is not running Windows.
func DLLUnmarshal(_ context.Context, _ data.Reader) (*cmd.DLL, bool, bool, error) {
	return nil, false, false, device.ErrNoWindows
}

// ZombieUnmarshal will read this Zombies's struct data from the supplied reader
// and returns a Zombie runnable struct along with the wait and delete status
// booleans.
//
// This function returns an error if building or reading fails or if the device
// is not running Windows.
func ZombieUnmarshal(_ context.Context, _ data.Reader) (*cmd.Zombie, bool, error) {
	return nil, false, device.ErrNoWindows
}
