//go:build windows
// +build windows

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

package pipe

import (
	"context"
	"io"
	"net"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/device/winapi"
	"github.com/iDigitalFlame/xmt/util/bugtrack"
)

const retry = 100 * time.Millisecond

// ErrClosed is an error returned by the 'Accept' function when the
// underlying Pipe was closed.
var ErrClosed = &errno{m: io.ErrClosedPipe.Error()}

type addr string

// Conn is a struct that implements a Windows Pipe connection. This is similar
// to the 'net.Conn' interface except it adds the 'Impersonate' function, which
// is only from the 'AcceptPipe' function.
type Conn struct {
	_           [0]func()
	read, write time.Time
	addr        addr
	handle      uintptr
}
type wait struct {
	_   [0]func()
	err error
	n   uint32
}
type errno struct {
	e error
	m string
	t bool
}

// Listener is a struct that fulfils the 'net.Listener' interface, but used for
// Windows named pipes.
type Listener struct {
	_              [0]func()
	overlap        *winapi.Overlapped
	perms          *winapi.SecurityAttributes
	addr           addr
	active, handle uintptr
	done           uint32
}

// Fd returns the file descriptor handle for this PipeCon.
func (c *Conn) Fd() uintptr {
	return c.handle
}
func (addr) Network() string {
	return com.NamePipe
}

// Close releases the associated Pipe's resources. The connection is no longer
// considered valid after a call to this function.
func (c *Conn) Close() error {
	winapi.CancelIoEx(c.handle, nil)
	winapi.DisconnectNamedPipe(c.handle)
	err := winapi.CloseHandle(c.handle)
	c.handle = 0
	return err
}
func (a addr) String() string {
	return string(a)
}
func (e errno) Timeout() bool {
	return e.t
}
func (e errno) Error() string {
	if len(e.m) == 0 && e.e != nil {
		return e.e.Error()
	}
	return e.m
}
func (e errno) Unwrap() error {
	return e.e
}
func (e errno) Temporary() bool {
	return e.t
}

// Close closes the listener. Any blocked Accept operations will be unblocked
// and return errors.
func (l *Listener) Close() error {
	if atomic.LoadUint32(&l.done) == 1 {
		return nil
	}
	var err error // We shouldn't let errors stop us from cleaning up!
	if atomic.StoreUint32(&l.done, 1); l.handle > 0 {
		if err = winapi.DisconnectNamedPipe(l.handle); err != nil {
			if bugtrack.Enabled {
				bugtrack.Track("com.(*Listener).Close(): DisconnectNamedPipe error: %s!", err.Error())
			}
		}
		if err = winapi.CloseHandle(l.handle); err != nil {
			if bugtrack.Enabled {
				bugtrack.Track("com.(*Listener).Close(): CloseHandle(handle) error: %s!", err.Error())
			}
		}
		l.handle = 0
	}
	if l.overlap != nil && l.active > 0 {
		if err = winapi.CancelIoEx(l.active, l.overlap); err != nil {
			if bugtrack.Enabled {
				bugtrack.Track("com.(*Listener).Close(): CancelIoEx error: %s!", err.Error())
			}
		}
		if l.active > 0 { // Extra check as it can be reset here sometimes.
			if err = winapi.CloseHandle(l.active); err != nil {
				if bugtrack.Enabled {
					bugtrack.Track("com.(*Listener).Close(): CloseHandle(active) error: %s!", err.Error())
				}
			}
		}
		l.active = 0
	}
	return err
}

// Addr returns the listener's network address.
func (l *Listener) Addr() net.Addr {
	return l.addr
}

// Impersonate will attempt to call 'ImpersonatePipeToken' which, if successful,
// will set the token of this Thread to the Pipe's connected client token.
//
// A call to 'device.RevertToSelf()' will reset the token.
func (c *Conn) Impersonate() error {
	if c.handle == 0 {
		return nil
	}
	return winapi.ImpersonatePipeToken(c.handle)
}

// LocalAddr returns the Pipe's local endpoint address.
func (c *Conn) LocalAddr() net.Addr {
	return c.addr
}

// RemoteAddr returns the Pipe's remote endpoint address.
func (c *Conn) RemoteAddr() net.Addr {
	return c.addr
}

// Dial connects to the specified Pipe path. This function will return a 'net.Conn'
// instance or any errors that may occur during the connection attempt.
//
// Pipe names are in the form of "\\<computer>\pipe\<path>".
//
// This function blocks indefinitely. Use the DialTimeout or DialContext to specify
// a control method.
func Dial(path string) (net.Conn, error) {
	return DialContext(context.Background(), path)
}

// Read implements the 'net.Conn' interface.
func (c *Conn) Read(b []byte) (int, error) {
	if c.handle == 0 {
		return 0, ErrClosed
	}
	var (
		a   uint32
		o   = new(winapi.Overlapped)
		err error
	)
	if o.Event, err = winapi.CreateEvent(nil, true, true, ""); err != nil {
		return 0, &errno{m: "could not create event", e: err}
	}
	return c.finish(winapi.ReadFile(c.handle, b, &a, o), int(a), c.read, o)
}
func (l *Listener) wait(x context.Context) {
	<-x.Done()
	l.Close()
}

// Write implements the 'net.Conn' interface.
func (c *Conn) Write(b []byte) (int, error) {
	var (
		a   uint32
		o   = new(winapi.Overlapped)
		err error
	)
	if o.Event, err = winapi.CreateEvent(nil, true, true, ""); err != nil {
		return 0, &errno{m: "could not create event", e: err}
	}
	return c.finish(winapi.WriteFile(c.handle, b, &a, o), int(a), c.write, o)
}

// Listen returns a 'net.Listener' that will listen for new connections on the
// Named Pipe path specified or any errors that may occur during listener
// creation.
//
// Pipe names are in the form of "\\<computer>\pipe\<path>".
func Listen(path string) (*Listener, error) {
	return ListenSecurityContext(context.Background(), path, nil)
}

// Accept waits for and returns the next connection to the listener.
func (l *Listener) Accept() (net.Conn, error) {
	return l.AcceptPipe()
}

// SetDeadline implements the 'net.Conn' interface.
func (c *Conn) SetDeadline(t time.Time) error {
	c.read, c.write = t, t
	return nil
}

// AcceptPipe waits for and returns the next connection to the listener.
//
// This function returns the real type of 'Conn' that can be used with the
// 'Impersonate' function.
func (l *Listener) AcceptPipe() (*Conn, error) {
	if atomic.LoadUint32(&l.done) == 1 {
		return nil, ErrClosed
	}
	var (
		h   uintptr
		err error
	)
	if l.handle == 0 {
		if h, err = create(l.addr, l.perms, 50, 512, false); err != nil {
			return nil, &errno{e: err}
		}
	} else {
		h, l.handle = l.handle, 0
	}
	o := new(winapi.Overlapped)
	if o.Event, err = winapi.CreateEvent(nil, true, true, ""); err != nil {
		winapi.CloseHandle(h)
		return nil, &errno{m: "could not create event", e: err}
	}
	if err = winapi.ConnectNamedPipe(h, o); err == winapi.ErrIoPending || err == winapi.ErrIoIncomplete {
		l.overlap, l.active = o, h
		_, err = complete(h, o)
	}
	if l.active = 0; atomic.LoadUint32(&l.done) == 1 {
		winapi.CloseHandle(o.Event)
		winapi.CloseHandle(h)
		return nil, ErrClosed
	}
	winapi.CancelIoEx(l.active, l.overlap)
	if winapi.CloseHandle(o.Event); err == winapi.ErrOperationAborted {
		winapi.CloseHandle(h)
		return nil, ErrClosed
	}
	if err == winapi.ErrNoData {
		winapi.CloseHandle(h)
		return nil, ErrEmptyConn
	}
	if err != nil && err != winapi.ErrPipeConnected {
		winapi.CloseHandle(h)
		return nil, &errno{m: "could not connect", e: err}
	}
	return &Conn{addr: l.addr, handle: h}, nil
}

// SetReadDeadline implements the 'net.Conn' interface.
func (c *Conn) SetReadDeadline(t time.Time) error {
	c.read = t
	return nil
}

// SetWriteDeadline implements the 'net.Conn' interface.
func (c *Conn) SetWriteDeadline(t time.Time) error {
	c.write = t
	return nil
}
func connect(path string, t uint32) (*Conn, error) {
	if len(path) == 0 || len(path) > 255 {
		return nil, &errno{m: "invalid path length"}
	}
	if err := winapi.WaitNamedPipe(path, t); err != nil {
		return nil, err
	}
	// 0xC0000000 - FILE_FLAG_OVERLAPPED | FILE_FLAG_WRITE_THROUGH
	// 0x3        - FILE_SHARE_READ | FILE_SHARE_WRITE
	// 0x3        - OPEN_EXISTING
	// 0x40000000 - FILE_FLAG_OVERLAPPED
	h, err := winapi.CreateFile(path, 0xC0000000, 0x3, nil, 0x3, 0x40000000, 0)
	if err != nil {
		return nil, err
	}
	return &Conn{addr: addr(path), handle: h}, nil
}

// ListenPerms returns a Listener that will listen for new connections on the
// Named Pipe path specified or any errors that may occur during listener
// creation.
//
// Pipe names are in the form of "\\<computer>\pipe\<path>".
//
// This function allows for specifying a SDDL string used to set the permissions
// of the listening Pipe.
func ListenPerms(path, perms string) (*Listener, error) {
	return ListenPermsContext(context.Background(), path, perms)
}
func complete(h uintptr, o *winapi.Overlapped) (uint32, error) {
	if _, err := winapi.WaitForSingleObject(o.Event, -1); err != nil {
		return 0, err
	}
	var (
		n   uint32
		err = winapi.GetOverlappedResult(h, o, &n, true)
	)
	return n, err
}

// DialTimeout connects to the specified Pipe path. This function will return a
// net.Conn instance or any errors that may occur during the connection attempt.
//
// Pipe names are in the form of "\\<computer>\pipe\<path>".
//
// This function blocks for the specified amount of time and will return 'ErrTimeout'
// if the timeout is reached.
func DialTimeout(path string, t time.Duration) (net.Conn, error) {
	for n, d := time.Now(), time.Now().Add(t); n.Before(d); n = time.Now() {
		c, err := connect(path, uint32(d.Sub(n)/time.Millisecond))
		if err == nil {
			return c, nil
		}
		if err == winapi.ErrSemTimeout {
			return nil, ErrTimeout
		}
		if err == winapi.ErrBadPathname {
			return nil, &errno{m: `invalid path "` + path + `"`, e: err}
		}
		if err == winapi.ErrFileNotFound || err == winapi.ErrPipeBusy {
			if l := time.Until(d); l < retry {
				time.Sleep(l - time.Millisecond)
			} else {
				time.Sleep(retry)
			}
			continue
		}
		return nil, &errno{m: err.Error(), e: err}
	}
	return nil, ErrTimeout
}

// DialContext connects to the specified Pipe path. This function will return a
// net.Conn instance or any errors that may occur during the connection attempt.
//
// Pipe names are in the form of "\\<computer>\pipe\<path>".
//
// This function blocks until the supplied context is canceled and will return the
// context's Err() if the cancel occurs before the connection.
func DialContext(x context.Context, path string) (net.Conn, error) {
	for {
		if x != nil {
			select {
			case <-x.Done():
				return nil, x.Err()
			default:
			}
		}
		c, err := connect(path, 250)
		if err == nil {
			return c, nil
		}
		if err == winapi.ErrSemTimeout {
			return nil, ErrTimeout
		}
		if err == winapi.ErrBadPathname {
			return nil, &errno{m: `invalid path "` + path + `"`, e: err}
		}
		if err == winapi.ErrFileNotFound || err == winapi.ErrPipeBusy {
			time.Sleep(retry)
			continue
		}
		return nil, &errno{m: err.Error(), e: err}
	}
}

// ListenContext returns a 'net.Listener' that will listen for new connections
// on the Named Pipe path specified or any errors that may occur during listener
// creation.
//
// Pipe names are in the form of "\\<computer>\pipe\<path>".
//
// The provided Context can be used to cancel the Listener.
func ListenContext(x context.Context, path string) (*Listener, error) {
	return ListenSecurityContext(x, path, nil)
}
func waitComplete(w chan<- wait, h uintptr, o *winapi.Overlapped, s *uint32) {
	if n, err := complete(h, o); atomic.LoadUint32(s) == 0 {
		w <- wait{n: n, err: err}
	}
}

// ListenSecurity returns a net.Listener that will listen for new connections on
// the Named Pipe path specified or any errors that may occur during listener
// creation.
//
// Pipe names are in the form of "\\<computer>\pipe\<path>".
//
// This function allows for specifying a SecurityAttributes object used to set
// the permissions of the listening Pipe.
func ListenSecurity(path string, p *winapi.SecurityAttributes) (*Listener, error) {
	return ListenSecurityContext(context.Background(), path, p)
}

// ListenPermsContext returns a Listener that will listen for new connections on
// the Named Pipe path specified or any errors that may occur during listener
// creation.
//
// Pipe names are in the form of "\\<computer>\pipe\<path>".
//
// This function allows for specifying a SDDL string used to set the permissions
// of the listening Pipe.
//
// The provided Context can be used to cancel the Listener.
func ListenPermsContext(x context.Context, path, perms string) (*Listener, error) {
	var (
		s   = winapi.SecurityAttributes{InheritHandle: 0}
		err error
	)
	if len(perms) > 0 {
		if s.SecurityDescriptor, err = winapi.SecurityDescriptorFromString(perms); err != nil {
			return nil, err
		}
		s.Length = uint32(unsafe.Sizeof(s))
	}
	return ListenSecurityContext(x, path, &s)
}
func (c *Conn) finish(e error, a int, t time.Time, o *winapi.Overlapped) (int, error) {
	if e == winapi.ErrBrokenPipe {
		winapi.CloseHandle(o.Event)
		return a, io.EOF
	}
	if e != winapi.ErrIoIncomplete && e != winapi.ErrIoPending {
		winapi.CloseHandle(o.Event)
		return a, e
	}
	var (
		z *time.Timer
		s uint32
		f <-chan time.Time
		w = make(chan wait, 1)
	)
	if !t.IsZero() && time.Now().Before(t) {
		z = time.NewTimer(time.Until(t))
		f = z.C
	}
	go waitComplete(w, c.handle, o, &s)
	select {
	case <-f:
		winapi.CancelIoEx(c.handle, o)
		a, e = 0, ErrTimeout
	case d := <-w:
		a, e = int(d.n), d.err
	}
	atomic.StoreUint32(&s, 1)
	if close(w); e == winapi.ErrBrokenPipe {
		e = io.EOF
	}
	if winapi.CloseHandle(o.Event); z != nil {
		z.Stop()
	}
	return a, e
}
func create(path addr, p *winapi.SecurityAttributes, t, l uint32, f bool) (uintptr, error) {
	// 0x40040003 - PIPE_ACCESS_DUPLEX | WRITE_DAC | FILE_FLAG_OVERLAPPED
	m := uint32(0x40040003)
	if f {
		// 0x80000 - FILE_FLAG_FIRST_PIPE_INSTANCE
		m |= 0x80000
	}
	h, err := winapi.CreateNamedPipe(string(path), m, 0, 0xFF, l, l, t, p)
	if err != nil {
		return 0, err
	}
	return h, nil
}

// ListenSecurityContext returns a net.Listener that will listen for new connections
// on the Named Pipe path specified or any errors that may occur during listener
// creation.
//
// Pipe names are in the form of "\\<computer>\pipe\<path>".
//
// This function allows for specifying a SecurityAttributes object used to set
// the permissions of the listening Pipe.
//
// The provided Context can be used to cancel the Listener.
func ListenSecurityContext(x context.Context, path string, p *winapi.SecurityAttributes) (*Listener, error) {
	var (
		a      = addr(path)
		l, err = create(a, p, 50, 512, true)
	)
	if err != nil {
		if err == winapi.ErrInvalidName {
			return nil, &errno{m: `invalid path "` + path + `"`, e: err}
		}
		return nil, &errno{m: err.Error(), e: err}
	}
	n := &Listener{addr: a, handle: l, perms: p}
	if x != context.Background() {
		go n.wait(x)
	}
	return n, nil
}
