//go:build !windows
// +build !windows

package cmd

import "github.com/iDigitalFlame/xmt/device"

// Start will attempt to start the Zombie and will return an errors that occur
// while starting the Process.
//
// This function will return 'ErrEmptyCommand' if the 'Args', the 'Data' or
// the 'Path; parameters are empty and 'ErrAlreadyStarted' if attempting to
// start a Zombie that already has been started previously.
//
// Always returns 'ErrNoWindows' on non-Windows devices.
func (Zombie) Start() error {
	return device.ErrNoWindows
}
