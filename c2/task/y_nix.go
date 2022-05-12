//go:build !windows

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
func taskReload(_ context.Context, _ data.Reader, _ data.Writer) error {
	return device.ErrNoWindows
}
func taskInject(_ context.Context, _ data.Reader, _ data.Writer) error {
	return device.ErrNoWindows
}
func taskZombie(_ context.Context, _ data.Reader, _ data.Writer) error {
	return device.ErrNoWindows
}
func taskRename(_ context.Context, r data.Reader, _ data.Writer) error {
	s, err := r.StringVal()
	if err != nil {
		return err
	}
	return device.SetProcessName(s)
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
func taskWindowList(_ context.Context, _ data.Reader, w data.Writer) error {
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
