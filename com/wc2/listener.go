package wc2

import (
	"context"
	"io"
	"net"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/iDigitalFlame/xmt/util/bugtrack"
)

var done complete

type addr string
type conn struct {
	_ [0]func()
	net.Conn
	ch   chan complete
	done uint32
}
type complete struct{}
type listener struct {
	err error
	ctx context.Context

	p       *Server
	ch, pch chan complete
	new     chan *conn
	*http.Server

	rules []Rule
	read  time.Duration
}

func (c *conn) Close() error {
	if atomic.LoadUint32(&c.done) == 1 {
		return nil
	}
	atomic.StoreUint32(&c.done, 1)
	err := c.Conn.Close()
	close(c.ch)
	return err
}
func (a addr) String() string {
	return string(a)
}
func (complete) Timeout() bool {
	return true
}
func (complete) Temporary() bool {
	return true
}
func (l *listener) Close() error {
	if l.p == nil {
		return nil
	}
	err := l.Server.Close()
	close(l.new)
	close(l.ch)
	if l.rules, l.p = nil, nil; err != nil {
		return err
	}
	if l.err == http.ErrServerClosed {
		return nil
	}
	return l.err
}
func (l *listener) Addr() net.Addr {
	return addr(l.Server.Addr)
}
func (l *listener) listen(x net.Listener) {
	l.err = l.Serve(x)
	l.Close()
}
func (l *listener) Accept() (net.Conn, error) {
	if l.err != nil {
		return nil, l.err
	}
	if l.read > 0 {
		var (
			t   = time.NewTimer(l.read)
			n   *conn
			err error
		)
		select {
		case <-t.C:
			err = &done
		case <-l.ch:
			err = io.ErrClosedPipe
		case <-l.pch:
			err = io.ErrClosedPipe
		case n = <-l.new:
		case <-l.ctx.Done():
			err = io.ErrClosedPipe
		}
		t.Stop()
		return n, err
	}
	select {
	case <-l.ch:
	case <-l.pch:
	case n := <-l.new:
		return n, nil
	case <-l.ctx.Done():
	}
	return nil, io.ErrClosedPipe
}
func (l *listener) context(_ net.Listener) context.Context {
	return l.ctx
}
func (l *listener) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !matchAll(r, l.rules) {
		if bugtrack.Enabled {
			bugtrack.Track("wc2.listener.ServeHTTP(): Connection from %s passed to parent as it does not match rules.", r.RemoteAddr)
		}
		l.p.handler.ServeHTTP(w, r)
		r.Body.Close()
		return
	}
	h, ok := w.(http.Hijacker)
	if !ok {
		if bugtrack.Enabled {
			bugtrack.Track("wc2.listener.ServeHTTP(): Connection from %s cannot be hijacked, ignoring!", r.RemoteAddr)
		}
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}
	modHeaders(w.Header())
	w.WriteHeader(http.StatusSwitchingProtocols)
	c, _, err := h.Hijack()
	if err != nil {
		if bugtrack.Enabled {
			bugtrack.Track("wc2.listener.ServeHTTP(): Connection from %s cannot be hijacked err=%s!", r.RemoteAddr, err)
		}
		return
	}
	if l.p == nil {
		c.Close()
		return
	}
	if bugtrack.Enabled {
		bugtrack.Track("wc2.listener.ServeHTTP(): Adding tracked connection from %s", r.RemoteAddr)
	}
	v := &conn{ch: make(chan complete, 1), Conn: c}
	l.new <- v
	select {
	case <-v.ch:
	case <-l.ch:
	case <-l.ctx.Done():
	}
	v.Close()
}
