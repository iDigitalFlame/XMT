package com

import (
	"context"
	"io"
	"net"
	"sync"
	"time"

	"github.com/iDigitalFlame/xmt/device"
	"github.com/iDigitalFlame/xmt/util/bugtrack"
)

const (
	udpLimit = 4096

	readOp  = time.Microsecond * 15
	writeOp = time.Microsecond * 35
)

var (
	empty time.Time

	udpWake     udpWaker
	udpDeadline = new(udpErr)

	buffers = sync.Pool{
		New: func() interface{} {
			b := make([]byte, udpLimit)
			return &b
		},
	}
)

type udpErr struct{}
type udpConn struct {
	lock        sync.Mutex
	addr        net.Addr
	bufs        chan udpData
	sock        *udpListener
	wake        chan udpWaker
	buf         []byte
	dev         udpAddr
	read, write time.Duration
}
type udpData struct {
	b *[]byte
	n int
}
type udpAddr struct {
	device.Address
	p uint16
}
type udpWaker struct{}
type udpStream struct {
	net.Conn
	buf         []byte
	size        int
	read, write time.Duration
}
type udpListener struct {
	del      chan udpAddr
	err      error
	ctx      context.Context
	new      chan *udpConn
	cons     map[udpAddr]*udpConn
	sock     net.PacketConn
	lock     sync.RWMutex
	cancel   context.CancelFunc
	deadline time.Duration
}
type udpConnector struct {
	net.Dialer
}

func (udpErr) Error() string {
	return "deadline exceeded"
}
func (udpErr) Timeout() bool {
	return true
}
func (l *udpListener) purge() {
	for {
		select {
		case d := <-l.del:
			l.lock.Lock()
			if c, ok := l.cons[d]; ok {
				delete(l.cons, d)
				close(c.bufs)
				close(c.wake)
				c.bufs, c.wake = nil, nil
				c.sock = nil
				c.lock.Unlock()
			}
			l.lock.Unlock()
		case <-l.ctx.Done():
			return
		}
	}
}
func (udpErr) Temporary() bool {
	return true
}
func (l *udpListener) listen() {
loop:
	for l.sock.SetReadDeadline(empty); ; l.sock.SetReadDeadline(empty) {
		var (
			b         = buffers.Get().(*[]byte)
			n, a, err = l.sock.ReadFrom(*b)
		)
		if bugtrack.Enabled {
			bugtrack.Track("com.udpListener.listen(): Accept n=%d, a=%s, err=%s", n, a, err)
		}
		select {
		case <-l.ctx.Done():
			buffers.Put(b)
			break loop
		default:
			if err != nil && a == nil && n == 0 {
				buffers.Put(b)
				l.err = err
				break loop
			}
			if n == 0 || a == nil {
				buffers.Put(b)
				continue loop
			}
		}
		var d udpAddr
		switch v := a.(type) {
		case *net.IPAddr:
			d.Set(v.IP)
		case *net.UDPAddr:
			d.Set(v.IP)
			d.p = uint16(v.Port)
		default:
		}
		if d.IsZero() {
			buffers.Put(b)
			continue
		}
		l.lock.RLock()
		c, ok := l.cons[d]
		if l.lock.RUnlock(); ok {
			if c.lock.Lock(); c.bufs != nil {
				if bugtrack.Enabled {
					bugtrack.Track("com.udpListener.listen(): Pushing n=%d bytes to conn a=%s", n, a)
				}
				c.bufs <- udpData{n: n, b: b}
				c.lock.Unlock()
				continue
			}
			c.lock.Unlock()
			c = nil
		}
		if bugtrack.Enabled {
			bugtrack.Track("com.udpListener.listen(): New tracked conn a=%s", a)
		}
		c = &udpConn{dev: d, addr: a, sock: l, bufs: make(chan udpData, 256), wake: make(chan udpWaker, 1)}
		c.append(n, b, false)
		go c.receive(l.ctx)
		l.lock.Lock()
		l.cons[d] = c
		l.lock.Unlock()
		l.new <- c
	}
	l.cancel()
	if err := l.sock.Close(); err != nil && l.err == nil {
		l.err = err
	}
	l.lock.Lock()
	for _, c := range l.cons {
		c.Close()
	}
	l.lock.Unlock()
	close(l.del)
	close(l.new)
	l.cons = nil
}
func (c *udpConn) Close() error {
	if c.sock == nil {
		return nil
	}
	c.lock.Lock()
	c.sock.del <- c.dev
	c.sock = nil
	return nil
}
func (l *udpListener) Close() error {
	err := l.sock.Close()
	l.cancel()
	return err
}
func (l *udpListener) String() string {
	return "UDP/" + l.sock.LocalAddr().String()
}
func (l *udpListener) Addr() net.Addr {
	return l.sock.LocalAddr()
}
func (c *udpConn) LocalAddr() net.Addr {
	return c.addr
}

// NewUDP creates a new simple UDP based connector with the supplied timeout.
func NewUDP(t time.Duration) Connector {
	if t < 0 {
		t = DefaultTimeout
	}
	return &udpConnector{Dialer: net.Dialer{Timeout: t, KeepAlive: t, DualStack: true}}
}
func (c *udpConn) RemoteAddr() net.Addr {
	return c.addr
}
func (c *udpConn) receive(x context.Context) {
	for {
		select {
		case <-x.Done():
		case p, ok := <-c.bufs:
			if !ok {
				return
			}
			c.append(p.n, p.b, true)
		}
	}
}
func (c *udpConn) Read(b []byte) (int, error) {
	if len(c.buf) == 0 && c.bufs == nil {
		if bugtrack.Enabled {
			bugtrack.Track("com.udpCon.Read(): read on closed conn.")
		}
		return 0, io.ErrClosedPipe
	}
	var (
		t   *time.Timer
		n   int
		w   <-chan time.Time
		err error
	)
loop:
	for n < len(b) {
		if bugtrack.Enabled {
			bugtrack.Track("com.udpCon.Read(): n=%d, len(b)=%d, len(c.buf)=%d", n, len(b), len(c.buf))
		}
		if len(c.buf) > 0 {
			c.lock.Lock()
			v := copy(b[n:], c.buf)
			if bugtrack.Enabled {
				bugtrack.Track("com.udpCon.Read(): n=%d, v=%d, len(b)=%d, len(c.buf)=%d", n, v, len(b), len(c.buf))
			}
			if c.buf = c.buf[v:]; len(c.buf) == 0 {
				c.buf = nil
			}
			c.lock.Unlock()
			n += v
			continue
		}
		if n == 0 {
			if c.bufs == nil {
				err = io.EOF
				break
			}
			if t != nil {
				t.Stop()
				t, w = nil, nil
			}
			if c.read > 0 {
				t = time.NewTimer(c.read)
				w = t.C
			}
			select {
			case <-w:
				err = udpDeadline
				break loop
			case <-c.wake:
				continue loop
			case <-c.sock.ctx.Done():
				err = io.ErrClosedPipe
				break loop
			}
		}
		break
	}
	if t != nil {
		t.Stop()
	}
	if bugtrack.Enabled {
		bugtrack.Track("com.udpCon.Read(): return n=%d, err=%s", n, err)
	}
	return n, err
}
func (c *udpConn) Write(b []byte) (int, error) {
	if c.sock == nil {
		return 0, io.ErrShortWrite
	}
	var (
		n   int
		t   *time.Timer
		w   <-chan time.Time
		err error
	)
loop:
	for v, s, x := 0, 0, udpLimit; n < len(b) && s < len(b); {
		if t != nil {
			t.Stop()
			w, t = nil, nil
		}
		if x > len(b) {
			x = len(b)
		}
		if c.write > 0 {
			t = time.NewTimer(c.write)
			w = t.C
		}
		v, err = c.sock.sock.WriteTo(b[s:x], c.addr)
		s += v
		x += v
		if n += v; err != nil {
			break
		}
		select {
		case <-w:
			err = udpDeadline
			break loop
		case <-c.sock.ctx.Done():
			err = io.ErrClosedPipe
			break loop
		default:
			time.Sleep(writeOp)
		}
	}
	if t != nil {
		t.Stop()
	}
	return n, err
}
func (s *udpStream) Read(b []byte) (int, error) {
	if s.size == 0 || s.size < len(b) {
		var (
			n, c int
			err  error
		)
		for {
			if len(s.buf) == 0 || len(s.buf)-s.size < udpLimit {
				s.buf = append(s.buf, make([]byte, udpLimit)...)
			}
			if time.Sleep(readOp); s.read == 0 {
				if n > 0 {
					if bugtrack.Enabled {
						bugtrack.Track("com.udpStream.Read(): Implementing our own timeout for a Read operation.")
					}
					s.Conn.SetReadDeadline(time.Now().Add(time.Millisecond * 500))
				}
			} else {
				s.Conn.SetReadDeadline(time.Now().Add(s.read))
			}
			if bugtrack.Enabled {
				bugtrack.Track("com.udpStream.Read(): Pre-read s.size=%d, len(s.buf)=%d", s.size, len(s.buf))
			}
			n, err = s.Conn.Read(s.buf[s.size:])
			if s.size += n; err != nil {
				if e, ok := err.(net.Error); ok && (e.Temporary() || e.Timeout()) {
					err = nil
					if c++; c > 1 || s.size > 0 {
						if bugtrack.Enabled {
							bugtrack.Track("com.udpStream.Read(): Pre-read timeout hit, n=%d, s.size=%d, len(s.buf)=%d", n, s.size, len(s.buf))
						}
						break
					}
					continue
				}
				if err == io.EOF {
					err = nil
				}
				break
			}
		}
		if bugtrack.Enabled {
			bugtrack.Track("com.udpStream.Read(): Pre-read return n=%d, s.size=%d, len(s.buf)=%d, err=%s", n, s.size, len(s.buf), err)
		}
		if err != nil {
			return n, err
		}
		if s.size == 0 {
			return 0, io.EOF
		}
	}
	if bugtrack.Enabled {
		bugtrack.Track("com.udpStream.Read(): Read s.size=%d, len(s.buf)=%d, len(b)=%d", s.size, len(s.buf), len(b))
	}
	/* NOTE(idf): I think this causes false returns when data isn't 100% ready.
	if s.size == 0 {
		return 0, io.EOF
	}*/
	n := copy(b, s.buf[:s.size])
	s.buf = s.buf[n:]
	if s.size -= n; s.size <= 0 {
		s.buf = nil
	}
	if bugtrack.Enabled {
		bugtrack.Track("com.udpStream.Read(): Post-read n=%d, s.size=%d, len(s.buf)=%d, len(b)=%d", n, s.size, len(s.buf), len(b))
	}
	return n, nil
}
func (s *udpStream) Write(b []byte) (int, error) {
	var (
		t   *time.Timer
		w   <-chan time.Time
		n   int
		err error
	)
loop:
	for e, c, x := 0, 0, udpLimit; n < len(b) && e < len(b); {
		if t != nil {
			t.Stop()
			w, t = nil, nil
		}
		if x > len(b) {
			x = len(b)
		}
		if s.write > 0 {
			t = time.NewTimer(s.write)
			w = t.C
			s.Conn.SetWriteDeadline(time.Now().Add(s.write))
		}
		c, err = s.Conn.Write(b[e:x])
		if bugtrack.Enabled {
			bugtrack.Track("com.udpStream.Write(): e=%d, x=%d, c=%d, n=%d, len(b)=%d, err=%s", e, x, c, n, len(b), err)
		}
		e += c
		x += c
		if n += c; err != nil {
			break loop
		}
		select {
		case <-w:
			err = udpDeadline
			break loop
		default:
			time.Sleep(writeOp)
		}
	}
	return n, err
}
func (c *udpConn) SetDeadline(t time.Time) error {
	if t.IsZero() {
		c.read, c.write = 0, 0
		return nil
	}
	d := time.Until(t)
	if d <= 0 {
		c.read, c.write = 0, 0
		return nil
	}
	c.read, c.write = d, d
	return nil
}
func (l *udpListener) Accept() (net.Conn, error) {
	var (
		t *time.Timer
		w <-chan time.Time
	)
	if l.deadline > 0 {
		t = time.NewTimer(l.deadline)
		w = t.C
	}
loop:
	for l.err == nil {
		select {
		case <-w:
			return nil, udpDeadline
		case n := <-l.new:
			return n, nil
		case <-l.ctx.Done():
			break loop
		}
	}
	if t != nil {
		t.Stop()
	}
	return nil, l.err
}
func (c *udpConn) append(n int, b *[]byte, w bool) {
	if bugtrack.Enabled {
		bugtrack.Track("com.udpCon.append(): n=%d, w=%t, len(c.buf)=%d", n, w, len(c.buf))
	}
	c.lock.Lock()
	c.buf = append(c.buf, (*b)[:n]...)
	c.lock.Unlock()
	if buffers.Put(b); w {
		select {
		case c.wake <- udpWake:
			if bugtrack.Enabled {
				bugtrack.Track("com.udpCon.append(): Triggering wake.")
			}
		default:
		}
	}
}
func (s *udpStream) SetDeadline(t time.Time) error {
	if t.IsZero() {
		s.read, s.write = 0, 0
		return s.Conn.SetDeadline(t)
	}
	d := time.Until(t)
	if d <= 0 {
		s.read, s.write = 0, 0
		return s.Conn.SetDeadline(t)
	}
	s.read, s.write = d, d
	return s.Conn.SetDeadline(t)
}
func (c *udpConn) SetReadDeadline(t time.Time) error {
	if t.IsZero() {
		c.read = 0
		return nil
	}
	d := time.Until(t)
	if d <= 0 {
		c.read = 0
		return nil
	}
	c.read = d
	return nil
}
func (c *udpConn) SetWriteDeadline(t time.Time) error {
	if t.IsZero() {
		c.write = 0
		return nil
	}
	d := time.Until(t)
	if d <= 0 {
		c.write = 0
		return nil
	}
	c.write = d
	return nil
}
func (s *udpStream) SetReadDeadline(t time.Time) error {
	if t.IsZero() {
		s.read = 0
		return s.Conn.SetReadDeadline(t)
	}
	d := time.Until(t)
	if d <= 0 {
		s.read = 0
		return s.Conn.SetReadDeadline(t)
	}
	s.read = d
	return s.Conn.SetReadDeadline(t)
}
func (s *udpStream) SetWriteDeadline(t time.Time) error {
	if t.IsZero() {
		s.write = 0
		return s.Conn.SetWriteDeadline(t)
	}
	d := time.Until(t)
	if d <= 0 {
		s.write = 0
		return s.Conn.SetWriteDeadline(t)
	}
	s.write = d
	return s.Conn.SetWriteDeadline(t)
}
func (c *udpConnector) Connect(s string) (net.Conn, error) {
	x, err := c.Dial("udp", s)
	if err != nil {
		return nil, err
	}
	return &udpStream{Conn: x}, nil
}
func (*udpConnector) Listen(s string) (net.Listener, error) {
	c, err := ListenConfig.ListenPacket(context.Background(), "udp", s)
	if err != nil {
		return nil, err
	}
	l := &udpListener{
		new:  make(chan *udpConn, 16),
		del:  make(chan udpAddr, 16),
		cons: make(map[udpAddr]*udpConn),
		sock: c,
	}
	l.ctx, l.cancel = context.WithCancel(context.Background())
	go l.purge()
	go l.listen()
	return l, nil
}
