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

// Package pipe contains a cross-device compatable Pipes/NamedPipes connection
// interface. This package differs from the standard library as it allows for
// setting permissions on the Pipes without any OS-specific functions.
package pipe

import (
	"context"
	"net"
	"time"
)

// Pipe is the default Connector with the default timeout of 15 seconds.
const Pipe = Piper(time.Second * 15)

// Piper is a Connector that can be used with the 'c2' package to make Pipe
// connections for C2.
type Piper time.Duration

// Connect fulfills the Connector interface.
func (p Piper) Connect(x context.Context, a string) (net.Conn, error) {
	return DialContext(x, Format(a))
}

// Listen fulfills the Connector interface.
func (Piper) Listen(x context.Context, a string) (net.Listener, error) {
	return ListenPermsContext(x, Format(a), PermEveryone)
}
