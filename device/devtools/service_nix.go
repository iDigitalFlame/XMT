// +build !windows

package devtools

import (
	"context"
	"time"
)

// Service is a struct that assists in running a Windows service. This struct can be created and given functions
// to run (Exec - the function to run for each Timeout when greater than zero, Start - function to run on service start,
// End - function to run on service shutdown.) Trigger the service to start by using the 'Service.Run' function.
// The 'Run' function always returns 'ErrNoWindows' on non-Windows devices.
type Service struct {
	Start, End, Exec func()
	Name             string
	Interval         time.Duration
}

// Run will trigger the service to start and will block until the service completes. Will always returns
// 'ErrNoWindows' on non-Windows devices.
func (s *Service) Run() error {
	return ErrNoWindows
}

// RunContext will trigger the service to start and will block until the service completes. Will always returns
// 'ErrNoWindows' on non-Windows devices. This function allows to pass a Context to cancel the running service.
func (s *Service) RunContext(_ context.Context) error {
	return ErrNoWindows
}
