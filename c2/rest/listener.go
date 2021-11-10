package rest

import (
	"context"
	"net/http"

	"github.com/PurpleSec/routex"
)

func (s *Server) httpListenerGet(_ context.Context, w http.ResponseWriter, r *routex.Request) {
	if l := s.c.Listener(r.Values.StringDefault("listener", "")); l != nil {
		routex.JSON(w, http.StatusOK, l)
		return
	}
	errors(http.StatusNotFound, "", w, r)
}
func (s *Server) httpListenerList(_ context.Context, w http.ResponseWriter, _ *routex.Request) {
	l := s.c.Listeners()
	if w.WriteHeader(http.StatusOK); len(l) == 0 {
		w.Write([]byte(`[]`))
		return
	}
	w.Write([]byte("["))
	for i := range l {
		if i > 0 {
			w.Write([]byte(","))
		}
		l[i].JSON(w)
	}
	w.Write([]byte("]"))
}
