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

// Package man is the implementation of the Guardian and Sentinel structs. These
// can be used to guard against accidental launching of multiple processes and
// can determine if targets are 'alive'.
//
// Windows clients have many options using built-in API calls, while other
// options such as TCP or Sockets (Named Pipes on Windows, UDS on *nix) to use
// generic control structs.
//
// Sentinel is a struct that can be Marshaled/Unmashaled from a file or network
// stream with optional encryption capabilities. Sentinels can launch
// applications in may different ways, including downloading, injecting or
// directly executing.
package man

import (
	"context"

	"github.com/iDigitalFlame/xmt/util/xerr"
)

// Guardian is a struct that is used to maintain a running process that will
// re-establish itself if it is not detected running.
//
// Guardian instances use Linker interfaces to determine status.
type Guardian struct {
	_    [0]func()
	ch   chan struct{}
	sock listener
}

// Wait will block until the Guardian is closed.
func (g *Guardian) Wait() {
	<-g.ch
}

// Close will close the Guardian and stops the listener.
//
// Any errors during listener close will be returned.
//
// This function will block until the Guardian fully closes.
func (g *Guardian) Close() error {
	if g.sock == nil {
		return nil
	}
	err := g.sock.Close()
	close(g.ch)
	g.sock = nil
	return err
}

// Done returns a channel that's closed when this Guardian is closed.
//
// This can be used to monitor a Guardian's status using a select statement.
func (g *Guardian) Done() <-chan struct{} {
	return g.ch
}
func (g *Guardian) wait(x context.Context) {
	select {
	case <-g.ch:
	case <-x.Done():
		g.Close()
	}
}

// MustGuard returns a Guardian instance that watches on the name provided.
//
// This function must complete and will panic if an error occurs, otherwise a
// Guardian instance is returned.
//
// This function defaults to the 'Pipe' Linker if a nil Linker is specified.
func MustGuard(l Linker, n string) *Guardian {
	g, err := GuardContext(context.Background(), l, n)
	if err != nil {
		panic(err)
	}
	return g
}

// Guard will attempt to create a Guardian instance on the provided name using
// the specified Linker.
//
// This function will return an error if the name is already being listened on.
//
// This function defaults to the 'Pipe' Linker if a nil Linker is specified.
func Guard(l Linker, n string) (*Guardian, error) {
	return GuardContext(context.Background(), l, n)
}

// MustGuardContext returns a Guardian instance that watches on the name provided.
//
// This function must complete and will panic if an error occurs, otherwise a
// Guardian instance is returned.
//
// This function also takes a context.Context to be used for resource control.
func MustGuardContext(x context.Context, l Linker, n string) *Guardian {
	g, err := GuardContext(x, l, n)
	if err != nil {
		panic(err)
	}
	return g
}

// GuardContext will attempt to create a Guardian instance on the provided name.
//
// This function will return an error if the name is already being listened on.
//
// This function will choose the proper connection method based on the host
// operating system.
//
// This function also takes a context.Context to be used for resource control.
func GuardContext(x context.Context, l Linker, n string) (*Guardian, error) {
	if len(n) == 0 {
		return nil, xerr.Sub("empty or invalid Guardian name", 0x10)
	}
	if l == nil {
		l = Pipe
	}
	v, err := l.create(n)
	if err != nil {
		return nil, err
	}
	g := &Guardian{ch: make(chan struct{}), sock: v}
	if x != nil && x != context.Background() {
		go g.wait(x)
	}
	go v.Listen()
	return g, nil
}
