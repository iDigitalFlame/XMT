package rest

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PurpleSec/escape"
	"github.com/PurpleSec/routex"
	"github.com/iDigitalFlame/xmt/c2"
	"github.com/iDigitalFlame/xmt/device"
)

const (
	prefix = `^/api/v1`

	expire  = time.Hour * 2
	timeout = time.Second * 10
)

// Server is the struct that handles the C2 RPC server interface and submits the control requests.
type Server struct {
	Auth string
	http.Server

	ctx   context.Context
	cache *cache
	c     *c2.Server

	cancel context.CancelFunc
	mux    *routex.Mux

	Key     string
	Timeout time.Duration
}

// New creates a new RPC server instance using the resulting C2 Server.
//
// The provided key can be used to authenticate to the Rest service with the
// 'X-RestAuth' HTTP header containing the supplied key.
// If empty, authentication is disabled.
func New(c *c2.Server, key string) *Server {
	return NewContext(context.Background(), c, key)
}

// Listen will bind to the specified address and begin serving requests.
// This function will return when the server is closed.
func (s *Server) Listen(addr string) error {
	return s.ListenTLS(addr, "", "")
}
func configureRoutes(s *Server, m *routex.Mux) {
	m.Must(
		prefix+`/listener$`, routex.Func(s.httpListenerList),
		http.MethodGet,
	)
	m.Must(
		prefix+`/listener/(?P<name>[a-zA-Z0-9\-._]+)$`, routex.Func(s.httpListenerGet),
		http.MethodGet,
	)
	m.Must(
		prefix+`/session$`, routex.Func(s.httpSessionList),
		http.MethodGet,
	)
	m.Must(
		prefix+`/session/(?P<session>[a-zA-Z0-9]+)$`, routex.Func(s.httpSessionGet),
		http.MethodGet,
	)
	m.Must(
		prefix+`/session/(?P<session>[a-zA-Z0-9]+)$`, routex.Func(s.httpSessionDelete),
		http.MethodDelete,
	)
	m.Must(
		prefix+`/session/(?P<session>[a-zA-Z0-9]+)/cmd$`,
		routex.Marshal(valTasklet, tasklet{}, routex.MarshalFunc(s.httpTaskCmdPut)),
		http.MethodPut,
	)
	m.Must(
		prefix+`/session/(?P<session>[a-zA-Z0-9]+)/pull$`,
		routex.Wrap(valUpload, routex.WrapFunc(s.httpTaskPullPut)),
		http.MethodPut,
	)
	m.Must(
		prefix+`/session/(?P<session>[a-zA-Z0-9]+)/pullex$`,
		routex.Marshal(valTasklet, tasklet{}, routex.MarshalFunc(s.httpTaskPullExecPut)),
		http.MethodPut,
	)
	m.Must(
		prefix+`/session/(?P<session>[a-zA-Z0-9]+)/upload$`,
		routex.Wrap(valUpload, routex.WrapFunc(s.httpTaskUploadPut)),
		http.MethodPut,
	)
	m.Must(
		prefix+`/session/(?P<session>[a-zA-Z0-9]+)/download$`,
		routex.Wrap(valDownload, routex.WrapFunc(s.httpTaskDownloadPut)),
		http.MethodPut,
	)
	m.Must(
		prefix+`/session/(?P<session>[a-zA-Z0-9]+)/assembly$`,
		routex.Marshal(valTasklet, tasklet{}, routex.MarshalFunc(s.httpTaskAssemblyPut)),
		http.MethodPut,
	)
	m.Must(
		prefix+`/session/(?P<session>[a-zA-Z0-9]+)/job$`,
		routex.Func(s.httpJobList),
		http.MethodGet, http.MethodDelete,
	)
	m.Must(
		prefix+`/session/(?P<session>[a-zA-Z0-9]+)/job/(?P<job>[0-9]+)$`,
		routex.Func(s.httpJobGetDelete),
		http.MethodGet, http.MethodDelete,
	)
	m.Must(
		prefix+`/session/(?P<session>[a-zA-Z0-9]+)/job/(?P<job>[0-9]+)/result$`,
		routex.Func(s.httpJobResultGet),
		http.MethodGet,
	)
}

// ListenTLS will bind to the specified address and use the provided certificate and key file paths
// to listen using a secure TLS tunnel. This function will return when the server is closed.
func (s *Server) ListenTLS(addr, cert, key string) error {
	if s.Timeout == 0 {
		s.Timeout = timeout
	}
	s.Addr = addr
	s.ReadTimeout, s.IdleTimeout = s.Timeout, s.Timeout
	s.WriteTimeout, s.ReadHeaderTimeout = s.Timeout, s.Timeout
	go s.cache.prune(s.ctx)
	if len(cert) == 0 || len(key) == 0 {
		return s.ListenAndServe()
	}
	s.TLSConfig = &tls.Config{
		NextProtos: []string{"h2", "http/1.1"},
		MinVersion: tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
		CurvePreferences:         []tls.CurveID{tls.CurveP256, tls.X25519},
		PreferServerCipherSuites: true,
	}
	return s.ListenAndServeTLS(cert, key)
}
func (s *Server) context(_ net.Listener) context.Context {
	return s.ctx
}

// NewContext creates a new RPC server instance using the resulting C2 Server. This function
// allows specifying a Context to aid in cancelation.
func NewContext(x context.Context, c *c2.Server, key string) *Server {
	s := &Server{
		c: c,
		cache: &cache{
			dev: make(map[string]*c2.Session),
			job: make(map[device.ID]map[uint16]*c2.Job),
		},
		Auth:    key,
		Timeout: timeout,
	}
	s.c.New, s.c.Oneshot = s.cache.new, s.cache.catch
	s.ctx, s.cancel = context.WithCancel(x)
	s.BaseContext = s.context
	s.mux = routex.NewContext(x)
	s.mux.Middleware(encoding)
	s.mux.Middleware(s.auth)
	s.mux.Error = routex.ErrorFunc(errors)
	configureRoutes(s, s.mux)
	s.Handler = s.mux
	return s
}
func errors(c int, e string, w http.ResponseWriter, _ *routex.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(c)
	w.Write([]byte(`{"source": "xmt_rest", "code": ` + strconv.Itoa(c) + `, "error": `))
	if len(e) > 0 {
		w.Write([]byte(escape.JSON(e)))
	} else {
		w.Write([]byte(`""`))
	}
	w.Write([]byte(`}`))
}
func encoding(_ context.Context, w http.ResponseWriter, _ *routex.Request) bool {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return true
}
func (s *Server) auth(_ context.Context, w http.ResponseWriter, r *routex.Request) bool {
	if len(s.Auth) == 0 {
		return true
	}
	if !strings.EqualFold(r.Header.Get("X-RestAuth"), s.Auth) {
		w.WriteHeader(http.StatusUnauthorized)
		return false
	}
	return true
}
