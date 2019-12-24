package web

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const (
	network      = "tcp"
	networkEmpty = addr("")
)

var bufs = &sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

type addr string
type conn struct {
	close  chan bool
	reader *http.Request
	writer io.Writer
}
type client struct {
	gen    *Generator
	buf    *bytes.Buffer
	host   string
	reader *http.Response
	client *http.Client
	parent *Server
}
type listener struct {
	new      chan *conn
	ctx      context.Context
	host     *url.URL
	cancel   context.CancelFunc
	parent   *Server
	listener net.Listener
	*http.Server
}

func (l *listener) listen() {
	if l.Server.TLSConfig != nil {
		l.Server.ServeTLS(l.listener, "", "")
	} else {
		l.Server.Serve(l.listener)
	}
	l.cancel()
}
func (addr) Network() string {
	return network
}
func (c *conn) Close() error {
	err := c.reader.Body.Close()
	c.close <- true
	return err
}
func (a addr) String() string {
	return string(a)
}
func (c *client) Close() error {
	if c.reader == nil {
		return nil
	}
	return c.reader.Body.Close()
}
func (l *listener) Close() error {
	defer func() { recover() }()
	l.cancel()
	close(l.new)
	return l.listener.Close()
}
func (conn) LocalAddr() net.Addr {
	return networkEmpty
}
func (l listener) String() string {
	return fmt.Sprintf("Web: %s", l.listener.Addr().String())
}
func (l listener) Addr() net.Addr {
	return l.listener.Addr()
}
func (client) LocalAddr() net.Addr {
	return networkEmpty
}
func (c conn) RemoteAddr() net.Addr {
	return addr(c.reader.RemoteAddr)
}
func (c client) RemoteAddr() net.Addr {
	return addr(c.host)
}
func (conn) SetDeadline(_ time.Time) error {
	return nil
}
func (c *conn) Read(b []byte) (int, error) {
	return c.reader.Body.Read(b)
}
func (c *conn) Write(b []byte) (int, error) {
	return c.writer.Write(b)
}
func (c *client) Read(b []byte) (int, error) {
	if c.reader == nil {
		var err error
		c.reader, err = c.request()
		c.buf.Reset()
		bufs.Put(c.buf)
		c.buf = nil
		if err != nil {
			return 0, err
		}
	}
	return c.reader.Body.Read(b)
}
func (client) SetDeadline(_ time.Time) error {
	return nil
}
func (c *client) Write(b []byte) (int, error) {
	if c.buf == nil {
		c.buf = bufs.Get().(*bytes.Buffer)
	}
	return c.buf.Write(b)
}
func (l *listener) Accept() (net.Conn, error) {
	select {
	case <-l.ctx.Done():
		return nil, l.ctx.Err()
	case n := <-l.new:
		return n, nil
	}
}
func (conn) SetReadDeadline(_ time.Time) error {
	return nil
}
func (conn) SetWriteDeadline(_ time.Time) error {
	return nil
}
func (client) SetReadDeadline(_ time.Time) error {
	return nil
}
func (c *client) request() (*http.Response, error) {
	r, err := http.NewRequest(http.MethodPost, "", c.buf)
	if err != nil {
		return nil, err
	}
	if c.parent != nil {
		r = r.WithContext(c.parent.ctx)
	}
	if i, err := url.Parse(c.host); err == nil {
		r.URL = i
	}
	if len(r.URL.Host) == 0 {
		r.URL.Host = c.host
		r.URL.Scheme = ""
	}
	if len(r.URL.Scheme) == 0 {
		if c.parent != nil && c.parent.tls != nil {
			r.URL.Scheme = "https"
		} else {
			r.URL.Scheme = "http"
		}
	}
	var u string
	if c.gen != nil && c.gen.URL != nil {
		u = c.gen.URL.String()
	}
	if len(u) > 1 && u[0] != '/' {
		r.URL.Path = fmt.Sprintf("/%s", u)
	} else {
		r.URL.Path = u
	}
	if c.gen != nil && c.gen.Host != nil {
		r.Host = c.gen.Host.String()
	}
	if c.gen != nil && c.gen.Agent != nil {
		r.Header.Set("User-Agent", c.gen.Agent.String())
	}
	var o *http.Response
	if c.client != nil {
		o, err = c.client.Do(r)
	} else if c.parent != nil {
		o, err = c.parent.Client.Do(r)
	} else {
		o, err = DefaultClient.Do(r)
	}
	if err != nil {
		return nil, err
	}
	if o.Body == nil {
		return nil, io.ErrUnexpectedEOF
	}
	return o, nil
}
func (client) SetWriteDeadline(_ time.Time) error {
	return nil
}
func (l *listener) context(_ net.Listener) context.Context {
	return l.ctx
}
func (l *listener) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil && l.parent.checkMatch(r) {
		c := &conn{
			close:  make(chan bool),
			reader: r,
			writer: w,
		}
		l.new <- c
		<-c.close
	} else {
		l.parent.handler.ServeHTTP(w, r)
	}
}
