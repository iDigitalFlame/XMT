package rpc

import (
	"context"
	"net/http"

	"github.com/PurpleSec/routex"
	"github.com/iDigitalFlame/xmt/data"
)

func (s *Server) httpListenerGet(x context.Context, w http.ResponseWriter, r *routex.Request) {
	n, err := r.Values.String("listener")
	if w.Header().Set("Content-Type", "application/json; charset=utf-8"); err != nil {
		errors(http.StatusBadRequest, err.Error(), w, r)
		return
	}
	l := s.c.Listener(n)
	if l == nil {
		errors(http.StatusNotFound, http.StatusText(http.StatusNotFound), w, r)
		return
	}
	b, _ := l.MarshalJSON()
	w.Write(b)
}
func (s *Server) httpListenersGet(x context.Context, w http.ResponseWriter, r *routex.Request) {
	l := s.c.Listeners()
	if w.Header().Set("Content-Type", "application/json; charset=utf-8"); len(l) == 0 {
		w.Write([]byte(`[]`))
		return
	}
	b := buffers.Get().(*data.Chunk)
	b.WriteUint8(uint8('['))
	for i := range l {
		if i > 0 {
			b.WriteUint8(uint8(','))
		}
		l[i].JSON(b)
	}
	b.WriteUint8(uint8(']'))
	w.Write(b.Payload())
	b.Reset()
	buffers.Put(b)
}
