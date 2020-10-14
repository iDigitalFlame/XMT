package wc2

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/com/limits"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

const (
	empty  = addr("")
	netWeb = "tcp"
)

// Server is a C2 profile that mimics a standard web server and client setup. This struct
// inherits the http.Server struct and can be used to serve real files and pages. Use the
// Mapper struct to provide a URL mapping that can be used by clients to access the C2 functions.
type Server struct {
	Client    *http.Client
	Generator Generator

	ctx     context.Context
	tls     *tls.Config
	lock    sync.RWMutex
	rules   []Rule
	dialer  *net.Dialer
	cancel  context.CancelFunc
	handler *http.ServeMux
}
type fileHandler string

// Close terminates this Web instance and signals all current listeners and clients to disconnect. This
// will close all connections related to this struct.
func (s *Server) Close() error {
	s.cancel()
	return nil
}

// Rule adds the specified rules to the Web instance to assist in determing real and C2 traffic.
func (s *Server) Rule(r ...Rule) {
	if len(r) == 0 {
		return
	}
	s.lock.Lock()
	if s.rules == nil {
		s.rules = make([]Rule, 0, len(r))
	}
	s.rules = append(s.rules, r...)
	s.lock.Unlock()
}

// New creates a new Web C2 server instance. This can be passed to the Listen function of a controller to
// serve a Web Server that also acts as a C2 instance. This struct supports all the default Golang http.Server
// functions and can be used to serve real web pages. Rules must be defined using the 'Rule' function to allow
// the server to differentiate between C2 and real traffic.
func New(t time.Duration) *Server {
	return NewTLS(t, nil)
}

// Serve attempts to serve the specified filesystem path 'f' at the URL mapped path 'p'. This function will
// determine if the path represents a file or directory and will call 'ServeFile' or 'ServeDirectory' depending
// on the path result. This function will return an error if the filesystem path does not exist or is invalid.
func (s *Server) Serve(p, f string) error {
	i, err := os.Stat(f)
	if err != nil {
		return err
	}
	if i.IsDir() {
		s.handler.Handle(p, http.FileServer(http.Dir(f)))
		return nil
	}
	s.handler.Handle(p, http.FileServer(fileHandler(f)))
	return nil
}

// ServeFile attempts to serve the specified filesystem path 'f' at the URL mapped path 'p'. This function is used
// to serve files and will return an error if the filesystem path does not exist or the path destination is not
// a file.
func (s *Server) ServeFile(p, f string) error {
	i, err := os.Stat(f)
	if err != nil {
		return err
	}
	if !i.IsDir() {
		s.handler.Handle(p, http.FileServer(fileHandler(f)))
		return nil
	}
	return xerr.New(`path "` + f + `" is not a file`)
}

// Handle registers the handler for the given pattern. If a handler already exists for pattern, Handle panics.
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

// ServeDirectory attempts to serve the specified filesystem path 'f' at the URL mapped path 'p'. This function is used
// to serve directories and will return an error if the filesystem path does not exist or the path destination is not
// a directory.
func (s *Server) ServeDirectory(p, f string) error {
	i, err := os.Stat(f)
	if err != nil {
		return err
	}
	if i.IsDir() {
		s.handler.Handle(p, http.FileServer(http.Dir(f)))
		return nil
	}
	return xerr.New(`path "` + f + `" is not a directory`)
}

// NewTLS creates a new TLS wrapped Web C2 server instance. This can be passed to the Listen function of a Controller
// to serve a Web Server that also acts as a C2 instance. This struct supports all the default Golang http.Server
// functions and can be used to serve real web pages. Rules must be defined using the 'Rule' function to allow the
// server to differentiate between C2 and real traffic.
func NewTLS(t time.Duration, c *tls.Config) *Server {
	w := &Server{
		tls:     c,
		dialer:  &net.Dialer{Timeout: t, KeepAlive: t, DualStack: true},
		handler: new(http.ServeMux),
	}
	w.ctx, w.cancel = context.WithCancel(context.Background())
	w.Client = &http.Client{
		Timeout: t,
		Transport: &http.Transport{
			Dial:                  w.dialer.Dial,
			Proxy:                 http.ProxyFromEnvironment,
			DialContext:           w.dialer.DialContext,
			MaxIdleConns:          limits.SmallLimit(),
			IdleConnTimeout:       w.dialer.Timeout,
			TLSHandshakeTimeout:   w.dialer.Timeout,
			ExpectContinueTimeout: w.dialer.Timeout,
			ResponseHeaderTimeout: w.dialer.Timeout,
		},
	}
	return w
}

// Connect creates a C2 client connector that uses the same properties of the Web struct parent.
func (s *Server) Connect(a string) (net.Conn, error) {
	c := &client{gen: s.Generator, host: a, parent: s}
	if c.gen.empty() {
		c.gen = DefaultGenerator
	}
	return c, nil
}
func (f fileHandler) Open(_ string) (http.File, error) {
	return os.OpenFile(string(f), os.O_RDONLY, 0)
}

// Listen returns a new C2 listener for this Web instance. This function creates a separate server, but still
// shares the handler for the base Web instance that it's created from.
func (s *Server) Listen(a string) (net.Listener, error) {
	if s.tls != nil && (len(s.tls.Certificates) == 0 || s.tls.GetCertificate == nil) {
		return nil, com.ErrInvalidTLSConfig
	}
	c, err := com.ListenConfig.Listen(context.Background(), netWeb, a)
	if err != nil {
		return nil, err
	}
	if s.tls != nil {
		c = tls.NewListener(c, s.tls)
	}
	l := &listener{
		new:    make(chan *conn, limits.SmallLimit()),
		parent: s,
		socket: c,
		Server: &http.Server{
			TLSConfig:         s.tls,
			ReadTimeout:       s.dialer.Timeout,
			IdleTimeout:       s.dialer.Timeout,
			WriteTimeout:      s.dialer.Timeout,
			ReadHeaderTimeout: s.dialer.Timeout,
		},
	}
	l.ctx, l.cancel = context.WithCancel(s.ctx)
	l.Server.Handler, l.Server.BaseContext = l, l.context
	go l.listen()
	return l, nil
}

// HandleFunc registers the handler function for the given pattern.
func (s *Server) HandleFunc(p string, h func(http.ResponseWriter, *http.Request)) {
	s.handler.HandleFunc(p, h)
}
