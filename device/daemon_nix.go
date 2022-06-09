//go:build !windows || noservice

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

package device

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"
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
func Daemon(_ string, f DaemonFunc) error {
	var (
		w    = make(chan os.Signal, 1)
		e    = make(chan error)
		x, y = context.WithCancel(context.Background())
		err  error
	)
	signal.Notify(w, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		e <- f(x)
		close(e)
	}()
	select {
	case err = <-e:
	case <-x.Done():
	}
	y()
	signal.Reset(syscall.SIGINT, syscall.SIGTERM)
	if close(w); err != nil && err != ErrQuit {
		return err
	}
	return nil
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
func DaemonTicker(_ string, t time.Duration, f DaemonFunc) error {
	var (
		w    = make(chan os.Signal, 1)
		v    = time.NewTimer(t)
		x, y = context.WithCancel(context.Background())
		err  error
	)
	signal.Notify(w, syscall.SIGINT, syscall.SIGTERM)
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
	y()
	v.Stop()
	signal.Reset(syscall.SIGINT, syscall.SIGTERM)
	if close(w); err != nil && err != ErrQuit {
		return err
	}
	return nil
}
