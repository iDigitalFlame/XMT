package rpc

import (
	"context"
	"encoding/base64"
	"net/http"
	"strconv"

	"github.com/PurpleSec/escape"
	"github.com/PurpleSec/routex"
	"github.com/iDigitalFlame/xmt/c2"
	"github.com/iDigitalFlame/xmt/c2/task"
	"github.com/iDigitalFlame/xmt/data"
)

func (r *Server) httpJob(_ context.Context, w http.ResponseWriter, x *routex.Request) {
	if w.Header().Set("Content-Type", "application/json; charset=utf-8"); !x.IsDelete() && !x.IsGet() {
		errors(http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed), w, x)
		return
	}
	n, err := x.Values.String("session")
	if err != nil {
		errors(http.StatusBadRequest, err.Error(), w, x)
		return
	}
	if len(n) == 0 {
		errors(http.StatusBadRequest, "empty session ID", w, x)
		return
	}
	i, err := x.Values.Uint64("job")
	if err != nil {
		errors(http.StatusBadRequest, err.Error(), w, x)
		return
	}
	if i == 0 || i > data.DataLimitMedium {
		errors(http.StatusBadRequest, "invalid job ID range", w, x)
	}
	s := r.session(n)
	if s == nil {
		errors(http.StatusNotFound, http.StatusText(http.StatusNotFound), w, x)
		return
	}
	var j *c2.Job
	if t, ok := r.jobs[s.ID.Hash()]; ok {
		if j, ok = t[uint16(i)]; ok && x.IsDelete() {
			delete(t, uint16(i))
		}
	}
	if j == nil {
		j = s.Job(uint16(i))
	}
	if j == nil {
		errors(http.StatusNotFound, http.StatusText(http.StatusNotFound), w, x)
		return
	}
	b, _ := j.MarshalJSON()
	w.Write(b)
}
func (r *Server) httpJobsGet(_ context.Context, w http.ResponseWriter, x *routex.Request) {
	n, err := x.Values.String("session")
	if w.Header().Set("Content-Type", "application/json; charset=utf-8"); err != nil {
		errors(http.StatusBadRequest, err.Error(), w, x)
		return
	}
	if len(n) == 0 {
		errors(http.StatusBadRequest, "empty session ID", w, x)
		return
	}
	s := r.session(n)
	if s == nil {
		errors(http.StatusNotFound, http.StatusText(http.StatusNotFound), w, x)
		return
	}
	j := s.Jobs()
	if t, ok := r.jobs[s.ID.Hash()]; ok && len(t) > 0 {
		for _, a := range t {
			j = append(j, a)
		}
	}
	if len(j) == 0 {
		w.Write([]byte(`{}`))
		return
	}
	b := buffers.Get().(*data.Chunk)
	b.WriteUint8(uint8('{'))
	for i := range j {
		if i > 0 {
			b.WriteUint8(uint8(','))
		}
		b.Write([]byte(`"` + strconv.Itoa(int(j[i].ID)) + `":`))
		j[i].JSON(b)
	}
	b.WriteUint8(uint8('}'))
	w.Write(b.Payload())
	b.Reset()
	buffers.Put(b)
}
func (r *Server) httpJobResultGet(_ context.Context, w http.ResponseWriter, x *routex.Request) {
	n, err := x.Values.String("session")
	if w.Header().Set("Content-Type", "application/json; charset=utf-8"); err != nil {
		errors(http.StatusBadRequest, err.Error(), w, x)
		return
	}
	if len(n) == 0 {
		errors(http.StatusBadRequest, "empty session ID", w, x)
		return
	}
	i, err := x.Values.Uint64("job")
	if err != nil {
		errors(http.StatusBadRequest, err.Error(), w, x)
		return
	}
	if i == 0 || i > data.DataLimitMedium {
		errors(http.StatusBadRequest, "invalid job ID range", w, x)
	}
	s := r.session(n)
	if s == nil {
		errors(http.StatusNotFound, http.StatusText(http.StatusNotFound), w, x)
		return
	}
	var j *c2.Job
	if t, ok := r.jobs[s.ID.Hash()]; ok {
		j = t[uint16(i)]
	}
	if j == nil {
		j = s.Job(uint16(i))
	}
	switch {
	case j == nil || j.Status < c2.Completed:
		errors(http.StatusNotFound, http.StatusText(http.StatusNotFound), w, x)
		return
	case j.Status == c2.Error:
		w.WriteHeader(http.StatusPartialContent)
		w.Write([]byte(`{"error": ` + escape.JSON(j.Error) + `}`))
		return
	case j.Result == nil || j.Result.Empty():
		w.WriteHeader(http.StatusNoContent)
		return
	}
	switch j.Type {
	case task.TvDownload:
		var (
			p, _ = j.Result.StringVal()
			d, _ = j.Result.Bool()
			b, _ = j.Result.Uint64()
		)
		w.Write([]byte(
			`{"path":` + escape.JSON(p) + `,"size":` + strconv.FormatUint(b, 10) + `,"dir":` +
				strconv.FormatBool(d) + `,"data":"`,
		))
		e := base64.NewEncoder(base64.StdEncoding, w)
		e.Write(j.Result.Payload())
		e.Close()
		w.Write([]byte(`","type":"download"}`))
		j.Result.Reset()
	case task.TvUpload:
		w.Write([]byte(`{"type":"upload"}`))
	case task.TvCode:
		w.Write([]byte(`{"type":"code"}`))
	case task.TvExecute:
		var (
			p, _ = j.Result.Uint64()
			c, _ = j.Result.Uint32()
		)
		w.Write([]byte(`{"pid":` + strconv.Itoa(int(p)) + `,"exit":` + strconv.Itoa(int(c)) + `,"data":"`))
		e := base64.NewEncoder(base64.StdEncoding, w)
		e.Write(j.Result.Payload())
		e.Close()
		w.Write([]byte(`","type":"execute"}`))
		j.Result.Seek(0, 0)
	default:
		w.Write(j.Result.Payload())
	}
}
