//go:build !windows
// +build !windows

package cmd

import "github.com/iDigitalFlame/xmt/device/devtools"

type thread struct {
	ctx interface{}
}

func (thread) Pid() uint32 {
	return 0
}
func (thread) Wait() error {
	return devtools.ErrNoWindows
}
func (thread) Stop() error {
	return devtools.ErrNoWindows
}
func (thread) Running() bool {
	return false
}
func (thread) Resume() error {
	return devtools.ErrNoWindows
}
func (thread) String() string {
	return ""
}
func (thread) Suspend() error {
	return devtools.ErrNoWindows
}
func (thread) SetSuspended(_ bool) {}
func (thread) ExitCode() (int32, error) {
	return 0, devtools.ErrNoWindows
}
func (thread) Handle() (uintptr, error) {
	return 0, devtools.ErrNoWindows
}
func (thread) Location() (uintptr, error) {
	return 0, devtools.ErrNoWindows
}
