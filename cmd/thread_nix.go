//go:build !windows

package cmd

import "github.com/iDigitalFlame/xmt/device"

type thread struct {
	ctx any
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
