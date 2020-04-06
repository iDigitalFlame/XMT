package wc2

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

var complete = finished{}

type addr string
type conn struct {
	w    io.Writer
	in   *http.Request
	done chan finished
}
type finished struct{}
type listener struct {
	new    chan *conn
	ctx    context.Context
	cancel context.CancelFunc
	parent *Server
	socket net.Listener
	*http.Server
}

func (l *listener) listen() {
	if l.Server.TLSConfig != nil {
		l.Server.ServeTLS(l.socket, "", "")
	} else {
		l.Server.Serve(l.socket)
	}
	l.cancel()
}
func (addr) Network() string {
	return netWeb
}
func (c *conn) Close() error {
	err := c.in.Body.Close()
	c.in = nil
	c.done <- complete
	return err
}
func (a addr) String() string {
	return string(a)
}
func (l *listener) Close() error {
	if l.new == nil {
		return nil
	}
	l.cancel()
	close(l.new)
	err := l.socket.Close()
	l.socket, l.new = nil, nil
	return err
}
func (conn) LocalAddr() net.Addr {
	return empty
}
func (l listener) String() string {
	return fmt.Sprintf("WC2[%s]", l.socket.Addr().String())
}
func (l listener) Addr() net.Addr {
	if l.socket == nil {
		return empty
	}
	return l.socket.Addr()
}
func (c conn) RemoteAddr() net.Addr {
	return addr(c.in.RemoteAddr)
}
func (conn) SetDeadline(_ time.Time) error {
	return nil
}
func (c *conn) Read(b []byte) (int, error) {
	n, err := c.in.Body.Read(b)
	if err != nil && n > 0 && err != io.EOF {
		return n, nil
	}
	return n, err
}
func (c *conn) Write(b []byte) (int, error) {
	return c.w.Write(b)
}
func (conn) SetReadDeadline(_ time.Time) error {
	return nil
}
func (conn) SetWriteDeadline(_ time.Time) error {
	return nil
}
func (l *listener) Accept() (n net.Conn, err error) {
	if l.ReadTimeout > 0 {
		x, f := context.WithTimeout(l.ctx, l.ReadTimeout)
		select {
		case <-x.Done():
			err = &net.OpError{Op: "read", Net: "wc2", Err: x.Err()}
		case n = <-l.new:
		case <-l.ctx.Done():
			err = l.ctx.Err()
		}
		f()
		return
	}
	select {
	case n = <-l.new:
		return
	case <-l.ctx.Done():
		err = l.ctx.Err()
		return
	}
}
func (l *listener) context(_ net.Listener) context.Context {
	return l.ctx
}
func (l *listener) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil && l.parent.checkMatch(r) {
		c := &conn{w: w, in: r, done: make(chan finished)}
		l.new <- c
		<-c.done
	} else {
		l.parent.handler.ServeHTTP(w, r)
	}
}
