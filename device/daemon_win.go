//go:build windows && !noservice
// +build windows,!noservice

package device

import (
	"context"
	"time"

	"github.com/iDigitalFlame/xmt/device/winapi/svc"
)

// Daemon starts a "Service" (on Windows devices) and will run the function
// until interrupted. This function will block while running the function and
// can be interrupted via the Windows service control manager or SIGNALS (on
// Linux).
//
// Any errors during runtime or returned from the functions will be returned.
//
// NOTE: The 'name' argument is the service name on Windows, but is ignored
// on *nix systems.
func Daemon(name string, f DaemonFunc) error {
	return svc.Run(name, func(x context.Context, s svc.Service, _ []string) uint32 {
		var (
			e   = make(chan error)
			err error
		)
		go func() {
			e <- f(x)
			close(e)
		}()
		select {
		case err = <-e:
		case <-x.Done():
		}
		if err != nil && err != ErrQuit {
			return 1
		}
		return 0
	})
}

// DaemonTicker starts a "Service" (on Windows devices) and will run the function
// every 't' duration until interrupted. This function will block while running
// and can be interrupted via the Windows service control manager or SIGNALS (on
// Linux).
//
// Returning the error 'ErrQuit' will break the loop with a non-error.
//
// Any errors during runtime or returned from the functions will be returned.
// Non-nil (non- ErrQuit) error returns will break the loop with an error.
//
// NOTE: The 'name' argument is the service name on Windows, but is ignored
// on *nix systems.
func DaemonTicker(name string, t time.Duration, f DaemonFunc) error {
	return svc.Run(name, func(x context.Context, s svc.Service, _ []string) uint32 {
		var (
			v   = time.NewTimer(t)
			err error
		)
	loop:
		for {
			select {
			case <-v.C:
				if err = f(x); err != nil {
					break loop
				}
				v.Reset(t)
			case <-x.Done():
				break loop
			}
		}
		if v.Stop(); err != nil && err != ErrQuit {
			return 1
		}
		return 0
	})
}
