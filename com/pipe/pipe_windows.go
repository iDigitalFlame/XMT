// +build windows

package pipe

import (
	"context"
	"io"
	"net"
	"sync/atomic"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	// PermEveryone is the SDDL string used in Windows Pipes to allow anyone to write and read
	// to the listening Pipe. This can be used for Pipe communcation between privilege boundaries.
	// This can be applied to the ListenPerm function.
	PermEveryone = "D:PAI(A;;FA;;;WD)(A;;FA;;;SY)"

	retry   = 100 * time.Millisecond
	network = "pipe"
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
	funcCreateNamedPipe     = dllKernel32.NewProc("CreateNamedPipeW")
	funcConnectNamedPipe    = dllKernel32.NewProc("ConnectNamedPipe")
	funcDisconnectNamedPipe = dllKernel32.NewProc("DisconnectNamedPipe")
)

type addr string
type conn struct {
	read, write time.Time
	addr        addr
	handle      windows.Handle
}
type wait struct {
	err error
	n   uint32
}
type errno struct {
	e error
	m string
	t bool
}
type listener struct {
	overlap        *windows.Overlapped
	addr           addr
	active, handle windows.Handle
	done           uint32
}

// Format will ensure the path for this Pipe socket fits the proper OS based pathname. Valid pathnames will be
// returned without any changes.
func Format(s string) string {
	if len(s) > 2 && s[0] == '\\' && s[1] == '\\' {
		return s
	}
	return `\\.\pipe\` + s
}
func (e errno) Cause() error {
	return e.e
}
func (addr) Network() string {
	return network
}
func (c *conn) Close() error {
	return windows.CloseHandle(c.handle)
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
	return e.m
}
func (e errno) Temporary() bool {
	return e.t
}
func (l *listener) Close() error {
	if atomic.LoadUint32(&l.done) == 1 {
		return nil
	}
	atomic.StoreUint32(&l.done, 1)
	if l.handle > 0 {
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
func (l listener) Addr() net.Addr {
	return l.addr
}
func (c conn) LocalAddr() net.Addr {
	return c.addr
}
func (c conn) RemoteAddr() net.Addr {
	return c.addr
}

// Dial connects to the specified Pipe path. This function will return a net.Conn instance or any errors that may
// occur during the connection attempt. Pipe names are in the form of "\\<computer>\pipe\<path>". This function
// blocks indefinitely. Use the DialTimeout or DialContext to specify a control method.
func Dial(path string) (net.Conn, error) {
	return DialContext(context.Background(), path)
}
func (c *conn) Read(b []byte) (int, error) {
	var (
		a   uint32
		o   = new(windows.Overlapped)
		err error
	)
	if o.HEvent, err = windows.CreateEvent(nil, 1, 1, nil); err != nil {
		return 0, &errno{m: "could not create pipe event", e: err}
	}
	return c.finish(windows.ReadFile(c.handle, b, &a, o), int(a), c.read, o)
}
func (c *conn) Write(b []byte) (int, error) {
	var (
		a   uint32
		o   = new(windows.Overlapped)
		err error
	)
	if o.HEvent, err = windows.CreateEvent(nil, 1, 1, nil); err != nil {
		return 0, &errno{m: "could not create pipe event", e: err}
	}
	return c.finish(windows.WriteFile(c.handle, b, &a, o), int(a), c.write, o)
}
func (l *listener) Accept() (net.Conn, error) {
	if atomic.LoadUint32(&l.done) == 1 {
		return nil, ErrClosed
	}
	var (
		h   windows.Handle
		err error
	)
	if l.handle == 0 {
		if h, err = createPipe(l.addr, nil, 50, 512, false); err != nil {
			return nil, &errno{m: "could not create pipe", e: err}
		}
	} else {
		h, l.handle = l.handle, 0
	}
	o := new(windows.Overlapped)
	if o.HEvent, err = windows.CreateEvent(nil, 1, 1, nil); err != nil {
		windows.CloseHandle(h)
		return nil, &errno{m: "could not create pipe event", e: err}
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
		return nil, &errno{m: "could not connect pipe: " + err.Error(), e: err}
	}
	return &conn{addr: l.addr, handle: h}, nil
}
func (c *conn) SetDeadline(t time.Time) error {
	c.read, c.write = t, t
	return nil
}

// Listen returns a net.Listener that will listen for new connections on the Named Pipe path specified or any
// errors that may occur during listener creation. Pipe names are in the form of "\\<computer>\pipe\<path>".
func Listen(path string) (net.Listener, error) {
	return ListenSecurity(path, nil)
}
func (c *conn) SetReadDeadline(t time.Time) error {
	c.read = t
	return nil
}
func (c *conn) SetWriteDeadline(t time.Time) error {
	c.write = t
	return nil
}
func connectPipe(path string, t uint32) (*conn, error) {
	n, err := windows.UTF16PtrFromString(string(path))
	if err != nil {
		return nil, err
	}
	r, _, err := funcWaitNamedPipe.Call(uintptr(unsafe.Pointer(n)), uintptr(t))
	if r == 0 {
		return nil, err
	}
	h, err := windows.CreateFile(n,
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
	return &conn{addr: addr(path), handle: h}, nil

}

// ListenPerms returns a net.Listener that will listen for new connections on the Named Pipe path specified or any
// errors that may occur during listener creation. Pipe names are in the form of "\\<computer>\pipe\<path>".
// This function allows for specifying a SDDL string used to set the permissions of the listeneing Pipe.
func ListenPerms(path, perms string) (net.Listener, error) {
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
		c, err := connectPipe(path, uint32(d.Sub(n)/time.Millisecond))
		if err == nil {
			return c, nil
		}
		if err == windows.ERROR_BAD_PATHNAME {
			return nil, &errno{m: `invalid pipe path "` + path + `"`, e: err}
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
		c, err := connectPipe(path, windows.INFINITE)
		if err == nil {
			return c, nil
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
func waitComplete(w chan<- wait, h windows.Handle, o *windows.Overlapped) {
	n, err := complete(h, o)
	w <- wait{n: n, err: err}
}

// ListenSecurity returns a net.Listener that will listen for new connections on the Named Pipe path specified or any
// errors that may occur during listener creation. Pipe names are in the form of "\\<computer>\pipe\<path>".
// This function allows for specifying a SecurityAttributes object used to set the permissions of the listeneing Pipe.
func ListenSecurity(path string, p *windows.SecurityAttributes) (net.Listener, error) {
	var (
		a      = addr(path)
		l, err = createPipe(a, p, 50, 512, true)
	)
	if err != nil {
		if err == windows.ERROR_INVALID_NAME {
			return nil, &errno{m: `invalid pipe path "` + path + `"`, e: err}
		}
		return nil, &errno{m: err.Error(), e: err}
	}
	return &listener{addr: a, handle: l}, nil
}
func (c *conn) finish(e error, a int, t time.Time, o *windows.Overlapped) (int, error) {
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
		f <-chan time.Time
		w = make(chan wait, 1)
	)
	if !t.IsZero() && time.Now().Before(t) {
		z = time.NewTimer(time.Until(t))
		f = z.C
	}
	go waitComplete(w, c.handle, o)
	select {
	case <-f:
		windows.CancelIoEx(c.handle, o)
		a, e = 0, ErrTimeout
	case d := <-w:
		a, e = int(d.n), d.err
	}
	if close(w); e == windows.ERROR_BROKEN_PIPE {
		e = io.EOF
	}
	if z != nil {
		z.Stop()
	}
	windows.CloseHandle(o.HEvent)
	return a, e
}
func createPipe(path addr, p *windows.SecurityAttributes, t, l uint32, f bool) (windows.Handle, error) {
	var (
		m      = 0x00000003 | 0x00040000 | windows.FILE_FLAG_OVERLAPPED
		n, err = windows.UTF16PtrFromString(string(path))
	)
	if err != nil {
		return 0, err
	}
	if f {
		m |= 0x00080000
	}
	r, _, err := funcCreateNamedPipe.Call(
		uintptr(unsafe.Pointer(n)), uintptr(m), 0, 0xFF,
		uintptr(l), uintptr(l), uintptr(t),
		uintptr(unsafe.Pointer(p)),
	)
	if r == uintptr(windows.InvalidHandle) {
		return 0, err
	}
	return windows.Handle(r), nil
}
