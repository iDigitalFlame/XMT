package rest

import (
	"context"
	"net/http"

	"github.com/PurpleSec/routex"
)

func (s *Server) httpSessionGet(_ context.Context, w http.ResponseWriter, r *routex.Request) {
	if x := s.cache.device(r.Values.StringDefault("session", "")); x != nil {
		routex.JSON(w, http.StatusOK, x)
		return
	}
	errors(http.StatusNotFound, "", w, r)
}
func (s *Server) httpSessionList(_ context.Context, w http.ResponseWriter, _ *routex.Request) {
	w.WriteHeader(http.StatusOK)
	if s.cache.devLock.RLock(); len(s.cache.dev) == 0 {
		s.cache.devLock.RUnlock()
		w.Write([]byte(`[]`))
		return
	}
	var c int
	w.Write([]byte("["))
	for _, x := range s.cache.dev {
		if c > 0 {
			w.Write([]byte(","))
		}
		x.JSON(w)
		c++
	}
	s.cache.devLock.RUnlock()
	w.Write([]byte("]"))
}
func (s *Server) httpSessionDelete(_ context.Context, w http.ResponseWriter, r *routex.Request) {
	d, err := r.Content()
	if err != nil && err != routex.ErrNoBody {
		errors(http.StatusBadRequest, err.Error(), w, r)
		return
	}
	if !s.cache.remove(r.Values.StringDefault("session", ""), d != nil && d.BoolDefault("shutdown", false)) {
		errors(http.StatusNotFound, "", w, r)
		return
	}
	w.WriteHeader(http.StatusOK)
}
