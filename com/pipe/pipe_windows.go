//go:build windows
// +build windows

package pipe

import (
	"context"
	"io"
	"net"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/iDigitalFlame/xmt/device/devtools"
	"golang.org/x/sys/windows"
)

const (
	// PermEveryone is the SDDL string used in Windows Pipes to allow anyone to write and read
	// to the listening Pipe. This can be used for Pipe communcation between privilege boundaries.
	// This can be applied to the ListenPerm function.
	PermEveryone = "D:PAI(A;;FA;;;WD)(A;;FA;;;SY)"

	retry = 100 * time.Millisecond

	pipeBuffer  = 50
	pipeTimeout = 512
)

var (
	// ErrClosed is an error returned by the 'Accept' function when the underlying Pipe was closed.
	ErrClosed = &errno{m: "pipe was closed"}
	// ErrTimeout is an error returned by the 'Dial*' functions when the specified timeout was reached when attempting
	// to connect to a Pipe.
	ErrTimeout = &errno{m: "pipe connection timeout", t: true}
	// ErrEmptyConn is an error received when the 'Listen' function receives a shortly lived Pipe connection. This
	// error is only temporary and does not indicate any Pipe server failures.
	ErrEmptyConn = &errno{m: "received an empty connection", t: true}

	dllKernel32 = windows.NewLazySystemDLL("kernel32.dll")

	funcWaitNamedPipe       = dllKernel32.NewProc("WaitNamedPipeW")
	funcConnectNamedPipe    = dllKernel32.NewProc("ConnectNamedPipe")
	funcDisconnectNamedPipe = dllKernel32.NewProc("DisconnectNamedPipe")
)

type addr string
type wait struct {
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
	overlap        *windows.Overlapped
	addr           addr
	active, handle windows.Handle
	done           uint32
}

// PipeConn is a struct that implements a Windows Pipe connection. This is similar to the 'net.Conn'
// interface except it adds the 'Impersonate' function, which is only from the 'AcceptPipe' function.
type PipeConn struct {
	read, write time.Time
	addr        addr
	handle      windows.Handle
}

// Format will ensure the path for this Pipe socket fits the proper OS based pathname. Valid pathnames will be
// returned without any changes.
func Format(s string) string {
	if len(s) > 2 && s[0] == '\\' && s[1] == '\\' {
		return s
	}
	return `\\.\pipe` + "\\" + s
}
func (e errno) Cause() error {
	return e.e
}
func (addr) Network() string {
	return "pipe"
}
func (a addr) String() string {
	return string(a)
}
func (e errno) Timeout() bool {
	return e.t
}
func (e errno) Error() string {
	return e.m
}
func (e errno) Unwrap() error {
	return e.e
}
func (e errno) String() string {
	if e.e == nil {
		return e.m
	}
	return e.m + ": " + e.e.Error()
}
func (e errno) Temporary() bool {
	return e.t
}

// Close releases the associated Pipe's resources. The handle is no longer considered valid after
// a call to this function.
func (c *PipeConn) Close() error {
	err := windows.CloseHandle(c.handle)
	c.handle = 0
	return err
}

// Close closes the listener. Any blocked Accept operations will be unblocked and return errors.
func (l *Listener) Close() error {
	if atomic.LoadUint32(&l.done) == 1 {
		return nil
	}
	if atomic.StoreUint32(&l.done, 1); l.handle > 0 {
		if r, _, err := funcDisconnectNamedPipe.Call(uintptr(l.handle)); r == 0 {
			return err
		}
		if err := windows.CloseHandle(l.handle); err != nil {
			return err
		}
		l.handle = 0
	}
	if l.overlap != nil && l.active > 0 {
		if err := windows.CancelIoEx(l.active, l.overlap); err != nil {
			return err
		}
		if err := windows.CloseHandle(l.overlap.HEvent); err != nil {
			return err
		}
		if err := windows.CloseHandle(l.active); err != nil {
			return err
		}
		l.active = 0
	}
	return nil
}

// String returns a string representation of this listener.
func (l *Listener) String() string {
	return "PIPE/" + string(l.addr)
}

// Addr returns the listener's network address.
func (l *Listener) Addr() net.Addr {
	return l.addr
}

// Impersonate will attempt to call 'ImpersonatePipeToken' which, if successful, will set the token of this
// Thread to the Pipe's connected client token. A call to 'devtools.RevertToSelf()' will reset the token.
func (c *PipeConn) Impersonate() error {
	if c.handle == 0 {
		return nil
	}
	return devtools.ImpersonatePipeToken(uintptr(c.handle))
}

// LocalAddr returns the Pipe's local endpoint address.
func (c *PipeConn) LocalAddr() net.Addr {
	return c.addr
}

// RemoteAddr returns the Pipe's remote endpoint address.
func (c *PipeConn) RemoteAddr() net.Addr {
	return c.addr
}

// Dial connects to the specified Pipe path. This function will return a net.Conn instance or any errors that may
// occur during the connection attempt. Pipe names are in the form of "\\<computer>\pipe\<path>". This function
// blocks indefinitely. Use the DialTimeout or DialContext to specify a control method.
func Dial(path string) (net.Conn, error) {
	return DialContext(context.Background(), path)
}

// Listen returns a net.Listener that will listen for new connections on the Named Pipe path specified or any
// errors that may occur during listener creation. Pipe names are in the form of "\\<computer>\pipe\<path>".
func Listen(path string) (*Listener, error) {
	return ListenSecurity(path, nil)
}

// Accept waits for and returns the next connection to the listener.
func (l *Listener) Accept() (net.Conn, error) {
	return l.AcceptPipe()
}

// Read implements the 'net.Conn' interface.
func (c *PipeConn) Read(b []byte) (int, error) {
	if c.handle == 0 {
		return 0, ErrClosed
	}
	var (
		a   uint32
		o   = new(windows.Overlapped)
		err error
	)
	if o.HEvent, err = windows.CreateEvent(nil, 1, 1, nil); err != nil {
		return 0, &errno{m: "could not create Pipe event", e: err}
	}
	return c.finish(windows.ReadFile(c.handle, b, &a, o), int(a), c.read, o)
}

// Write implements the 'net.Conn' interface.
func (c *PipeConn) Write(b []byte) (int, error) {
	var (
		a   uint32
		o   = new(windows.Overlapped)
		err error
	)
	if o.HEvent, err = windows.CreateEvent(nil, 1, 1, nil); err != nil {
		return 0, &errno{m: "could not create Pipe event", e: err}
	}
	return c.finish(windows.WriteFile(c.handle, b, &a, o), int(a), c.write, o)
}

// SetDeadline implements the 'net.Conn' interface.
func (c *PipeConn) SetDeadline(t time.Time) error {
	c.read, c.write = t, t
	return nil
}

// AcceptPipe waits for and returns the next connection to the listener. This function returns a the real type
// of 'PipeConn' that can be used for the 'Impersonate' function.
func (l *Listener) AcceptPipe() (*PipeConn, error) {
	if atomic.LoadUint32(&l.done) == 1 {
		return nil, ErrClosed
	}
	var (
		h   windows.Handle
		err error
	)
	if l.handle == 0 {
		if h, err = create(l.addr, nil, pipeBuffer, pipeTimeout, false); err != nil {
			return nil, &errno{m: "could not create Pipe", e: err}
		}
	} else {
		h, l.handle = l.handle, 0
	}
	o := new(windows.Overlapped)
	if o.HEvent, err = windows.CreateEvent(nil, 1, 1, nil); err != nil {
		windows.CloseHandle(h)
		return nil, &errno{m: "could not create Pipe event", e: err}
	}
	r, _, err := funcConnectNamedPipe.Call(uintptr(h), uintptr(unsafe.Pointer(o)))
	if err == windows.ERROR_IO_PENDING || err == windows.ERROR_IO_INCOMPLETE {
		l.overlap, l.active = o, h
		_, err = complete(h, o)
	}
	if windows.CloseHandle(o.HEvent); err == windows.ERROR_OPERATION_ABORTED {
		windows.CloseHandle(h)
		return nil, ErrClosed
	}
	if err == windows.ERROR_NO_DATA {
		windows.CloseHandle(h)
		return nil, ErrEmptyConn
	}
	if r == 0 && err != nil && err != windows.ERROR_PIPE_CONNECTED {
		windows.CloseHandle(h)
		return nil, &errno{m: "could not connect Pipe", e: err}
	}
	return &PipeConn{addr: l.addr, handle: h}, nil
}

// SetReadDeadline implements the 'net.Conn' interface.
func (c *PipeConn) SetReadDeadline(t time.Time) error {
	c.read = t
	return nil
}

// SetWriteDeadline implements the 'net.Conn' interface.
func (c *PipeConn) SetWriteDeadline(t time.Time) error {
	c.write = t
	return nil
}
func connect(path string, t uint32) (*PipeConn, error) {
	if len(path) == 0 || len(path) > 255 {
		return nil, &errno{m: "invalid Pipe path length"}
	}
	n, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}
	r, _, err := funcWaitNamedPipe.Call(uintptr(unsafe.Pointer(n)), uintptr(t))
	if r == 0 {
		return nil, err
	}
	h, err := windows.CreateFile(
		n,
		windows.GENERIC_READ|windows.GENERIC_WRITE,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE,
		nil,
		windows.OPEN_EXISTING,
		windows.FILE_FLAG_OVERLAPPED,
		0,
	)
	if err != nil {
		return nil, err
	}
	return &PipeConn{addr: addr(path), handle: h}, nil
}

// ListenPerms returns a Listener that will listen for new connections on the Named Pipe path specified or any
// errors that may occur during listener creation. Pipe names are in the form of "\\<computer>\pipe\<path>".
// This function allows for specifying a SDDL string used to set the permissions of the listeneing Pipe.
func ListenPerms(path, perms string) (*Listener, error) {
	var (
		s   = windows.SecurityAttributes{InheritHandle: 1}
		err error
	)
	if len(perms) > 0 {
		if s.SecurityDescriptor, err = windows.SecurityDescriptorFromString(perms); err != nil {
			return nil, err
		}
		s.Length = uint32(unsafe.Sizeof(s))
	}
	return ListenSecurity(path, &s)
}

// DialTimeout connects to the specified Pipe path. This function will return a net.Conn instance or any errors that
// may occur during the connection attempt. Pipe names are in the form of "\\<computer>\pipe\<path>". This function
// blocks for the specified amount of time and will return 'Errtimeout' if the timeout is reached.
func DialTimeout(path string, t time.Duration) (net.Conn, error) {
	for n, d := time.Now(), time.Now().Add(t); n.Before(d); n = time.Now() {
		c, err := connect(path, uint32(d.Sub(n)/time.Millisecond))
		if err == nil {
			return c, nil
		}
		if err == windows.ERROR_BAD_PATHNAME {
			return nil, &errno{m: `invalid Pipe path "` + path + `"`, e: err}
		}
		if err == windows.ERROR_SEM_TIMEOUT {
			return nil, ErrTimeout
		}
		if err == windows.ERROR_FILE_NOT_FOUND || err == windows.ERROR_PIPE_BUSY {
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

// DialContext connects to the specified Pipe path. This function will return a net.Conn instance or any errors that
// may occur during the connection attempt. Pipe names are in the form of "\\<computer>\pipe\<path>". This function
// blocks until the supplied context is cancled and will return the context's Err() if the cancel occurs before the
// connection.
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
		if err == windows.ERROR_SEM_TIMEOUT {
			return nil, ErrTimeout
		}
		if err == windows.ERROR_BAD_PATHNAME {
			return nil, &errno{m: `invalid pipe path "` + path + `"`, e: err}
		}
		if err == windows.ERROR_FILE_NOT_FOUND || err == windows.ERROR_PIPE_BUSY {
			time.Sleep(retry)
			continue
		}
		return nil, &errno{m: err.Error(), e: err}
	}
}
func complete(h windows.Handle, o *windows.Overlapped) (uint32, error) {
	if _, err := windows.WaitForSingleObject(o.HEvent, windows.INFINITE); err != nil {
		return 0, err
	}
	var (
		n   uint32
		err = windows.GetOverlappedResult(h, o, &n, true)
	)
	return n, err
}
func waitComplete(w chan<- wait, h windows.Handle, o *windows.Overlapped, s *uint32) {
	if n, err := complete(h, o); atomic.LoadUint32(s) == 0 {
		w <- wait{n: n, err: err}
	}
}

// ListenSecurity returns a net.Listener that will listen for new connections on the Named Pipe path specified or any
// errors that may occur during listener creation. Pipe names are in the form of "\\<computer>\pipe\<path>".
// This function allows for specifying a SecurityAttributes object used to set the permissions of the listeneing Pipe.
func ListenSecurity(path string, p *windows.SecurityAttributes) (*Listener, error) {
	var (
		a      = addr(path)
		l, err = create(a, p, pipeBuffer, pipeTimeout, true)
	)
	if err != nil {
		if err == windows.ERROR_INVALID_NAME {
			return nil, &errno{m: `invalid Pipe path "` + path + `"`, e: err}
		}
		return nil, &errno{m: err.Error(), e: err}
	}
	return &Listener{addr: a, handle: l}, nil
}
func (c *PipeConn) finish(e error, a int, t time.Time, o *windows.Overlapped) (int, error) {
	if e == windows.ERROR_BROKEN_PIPE {
		windows.CloseHandle(o.HEvent)
		return a, io.EOF
	}
	if e != windows.ERROR_IO_INCOMPLETE && e != windows.ERROR_IO_PENDING {
		windows.CloseHandle(o.HEvent)
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
		windows.CancelIoEx(c.handle, o)
		a, e = 0, ErrTimeout
	case d := <-w:
		a, e = int(d.n), d.err
	}
	atomic.StoreUint32(&s, 1)
	if close(w); e == windows.ERROR_BROKEN_PIPE {
		e = io.EOF
	}
	if z != nil {
		z.Stop()
	}
	windows.CloseHandle(o.HEvent)
	return a, e
}
func create(path addr, p *windows.SecurityAttributes, t, l uint32, f bool) (windows.Handle, error) {
	var (
		m      uint32 = 0x00000003 | 0x00040000 | windows.FILE_FLAG_OVERLAPPED
		n, err        = windows.UTF16PtrFromString(string(path))
	)
	if err != nil {
		return 0, err
	}
	if f {
		m |= 0x00080000
	}
	h, err := windows.CreateNamedPipe(n, m, 0, 0xFF, l, l, t, p)
	if err != nil || h == windows.InvalidHandle {
		return 0, err
	}
	return h, nil
}
