//go:build windows && !noservice
// +build windows,!noservice

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
		s.UpdateState(svc.Running, svc.AcceptStop|svc.AcceptShutdown)
		go func() {
			e <- f(x)
			close(e)
		}()
		select {
		case err = <-e:
		case <-x.Done():
		}
		if s.UpdateState(svc.StopPending); err != nil && err != ErrQuit {
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
		for s.UpdateState(svc.Running, svc.AcceptStop|svc.AcceptShutdown); ; {
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
		v.Stop()
		if s.UpdateState(svc.StopPending); err != nil && err != ErrQuit {
			return 1
		}
		return 0
	})
}
