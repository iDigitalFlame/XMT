//go:build !implant

package c2

import (
	"io"
	"strconv"
	"time"

	"github.com/PurpleSec/escape"
	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/data"
)

const maxEvents = 2048

type proxyState struct {
	proxies []proxyData
}
type stringer interface {
	String() string
}

func (*Server) close() {}
func (s *Server) count() int {
	return len(s.events)
}
func (s *Session) name() string {
	return s.ID.String()
}
func (s status) String() string {
	switch s {
	case StatusError:
		return "error"
	case StatusWaiting:
		return "waiting"
	case StatusAccepted:
		return "accepted"
	case StatusReceiving:
		return "receiving"
	case StatusCompleted:
		return "completed"
	}
	return "invalid"
}

// String returns the details of this Session as a string.
func (s *Session) String() string {
	switch {
	case s.s == nil && s.sleep == 0:
		return "[" + s.ID.String() + "] -> " + s.host.String() + " " + s.Last.Format(time.RFC1123)
	case s.s == nil && (s.jitter == 0 || s.jitter > 100):
		return "[" + s.ID.String() + "] " + s.sleep.String() + " -> " + s.host.String()
	case s.s == nil:
		return "[" + s.ID.String() + "] " + s.sleep.String() + "/" + strconv.Itoa(int(s.jitter)) + "% -> " + s.host.String()
	case s.s != nil && (s.jitter == 0 || s.jitter > 100):
		return "[" + s.ID.String() + "] " + s.sleep.String() + " -> " + s.host.String() + " " + s.Last.Format(time.RFC1123)
	}
	return "[" + s.ID.String() + "] " + s.sleep.String() + "/" + strconv.Itoa(int(s.jitter)) + "% -> " + s.host.String() + " " + s.Last.Format(time.RFC1123)
}

// JSON returns the data of this Job as a JSON blob.
func (j *Job) JSON(w io.Writer) error {
	if j == nil {
		_, err := w.Write([]byte(`{}`))
		return err
	}
	if _, err := w.Write([]byte(`{"id":` + strconv.Itoa(int(j.ID)) + `,` +
		`"type":` + strconv.Itoa(int(j.Type)) + `,` +
		`"error":` + escape.JSON(j.Error) + `,` +
		`"status":"` + j.Status.String() + `",` +
		`"start":"` + j.Start.Format(time.RFC3339) + `"`,
	)); err != nil {
		return err
	}
	if j.Session != nil && !j.Session.ID.Empty() {
		if _, err := w.Write([]byte(`,"host":"` + j.Session.ID.String() + `"`)); err != nil {
			return err
		}
	}
	if !j.Complete.IsZero() {
		if _, err := w.Write([]byte(`,"complete":"` + j.Complete.Format(time.RFC3339) + `"`)); err != nil {
			return err
		}
	}
	if j.Result != nil {
		if _, err := w.Write([]byte(`,"result":` + strconv.Itoa(j.Result.Size()))); err != nil {
			return err
		}
	}
	_, err := w.Write([]byte{'}'})
	return err
}

// JSON returns the data of this Server as a JSON blob.
func (s *Server) JSON(w io.Writer) error {
	if _, err := w.Write([]byte(`{"listeners":{`)); err != nil {
		return err
	}
	i := 0
	for k, v := range s.active {
		if i > 0 {
			if _, err := w.Write([]byte{','}); err != nil {
				return err
			}
		}
		if _, err := w.Write([]byte(escape.JSON(k) + `:`)); err != nil {
			return err
		}
		if err := v.JSON(w); err != nil {
			return err
		}
		i++
	}
	i = 0
	if _, err := w.Write([]byte(`},"sessions":[`)); err != nil {
		return err
	}
	s.lock.RLock()
	for _, v := range s.sessions {
		if i > 0 {
			if _, err := w.Write([]byte{','}); err != nil {
				s.lock.RUnlock()
				return err
			}
		}
		if err := v.JSON(w); err != nil {
			s.lock.RUnlock()
			return err
		}
		i++
	}
	s.lock.RUnlock()
	_, err := w.Write([]byte(`]}`))
	return err
}
func (l *Listener) oneshot(n *com.Packet) {
	if l.s == nil || l.s.Oneshot == nil {
		return
	}
	l.m.queue(event{p: n, pf: l.s.Oneshot})
}

// JSON returns the data of this Session as a JSON blob.
func (s *Session) JSON(w io.Writer) error {
	if _, err := w.Write([]byte(`{` +
		`"id":"` + s.ID.String() + `",` +
		`"hash":"` + strconv.Itoa(int(s.ID.Hash())) + `",` +
		`"channel":` + strconv.FormatBool(s.InChannel()) + `,` +
		`"device":{` +
		`"id":"` + s.ID.Full() + `",` +
		`"user":` + escape.JSON(s.Device.User) + `,` +
		`"hostname":` + escape.JSON(s.Device.Hostname) + `,` +
		`"version":` + escape.JSON(s.Device.Version) + `,` +
		`"arch":"` + s.Device.Arch().String() + `",` +
		`"os":` + escape.JSON(s.Device.OS().String()) + `,` +
		`"elevated":` + strconv.FormatBool(s.Device.IsElevated()) + `,` +
		`"domain":` + strconv.FormatBool(s.Device.IsDomainJoined()) + `,` +
		`"pid":` + strconv.Itoa(int(s.Device.PID)) + `,` +
		`"ppid":` + strconv.Itoa(int(s.Device.PPID)) + `,` +
		`"network":[`,
	)); err != nil {
		return err
	}
	for i := range s.Device.Network {
		if i > 0 {
			if _, err := w.Write([]byte{','}); err != nil {
				return err
			}
		}
		if _, err := w.Write([]byte(
			`{"name":` + escape.JSON(s.Device.Network[i].Name) + `,` +
				`"mac":"` + s.Device.Network[i].Mac.String() + `","ip":[`,
		)); err != nil {
			return err
		}
		for x := range s.Device.Network[i].Address {
			if x > 0 {
				if _, err := w.Write([]byte{','}); err != nil {
					return err
				}
			}
			if _, err := w.Write([]byte(`"` + s.Device.Network[i].Address[x].String() + `"`)); err != nil {
				return err
			}
		}
		if _, err := w.Write([]byte("]}")); err != nil {
			return err
		}
	}
	_, err := w.Write([]byte(
		`]},"created":"` + s.Created.Format(time.RFC3339) + `",` +
			`"last":"` + s.Last.Format(time.RFC3339) + `",` +
			`"via":` + escape.JSON(s.host.String()) + `,` +
			`"sleep":` + strconv.Itoa(int(s.sleep)) + `,` +
			`"jitter":` + strconv.Itoa(int(s.jitter)),
	))
	if err != nil {
		return err
	}
	if s.parent != nil {
		if _, err = w.Write([]byte(`,"connector_name":` + escape.JSON(s.parent.name))); err != nil {
			return err
		}
	}
	if t, ok := s.parent.listener.(stringer); ok {
		if _, err = w.Write([]byte(`,"connector":` + escape.JSON(t.String()))); err != nil {
			return err
		}
	}
	if s.s != nil && len(s.proxies) > 0 {
		if _, err = w.Write([]byte(`,"proxy":[`)); err != nil {
			return err
		}
		for i := range s.proxies {
			if i > 0 {
				if _, err = w.Write([]byte{','}); err != nil {
					return err
				}
			}
			_, err = w.Write([]byte(
				`{"name":` + escape.JSON(s.proxies[i].n) + `,"address": ` + escape.JSON(s.proxies[i].b) + `}`,
			))
			if err != nil {
				return err
			}
		}
		if _, err = w.Write([]byte{']'}); err != nil {
			return err
		}
	}
	_, err = w.Write([]byte{'}'})
	return err
}

// JSON returns the data of this Listener as a JSON blob.
func (l *Listener) JSON(w io.Writer) error {
	if _, err := w.Write([]byte(`{"name":` + escape.JSON(l.name) + `,"address":` + escape.JSON(l.Address()))); err != nil {
		return err
	}
	var n uint64
	if len(l.s.sessions) > 0 {
		l.s.lock.RLock()
		for _, s := range l.s.sessions {
			if s.parent == l {
				n++
			}
		}
		l.s.lock.RUnlock()
	}
	if _, err := w.Write([]byte(`,"count":` + strconv.FormatUint(n, 10))); err != nil {
		return err
	}
	if t, ok := l.listener.(stringer); ok {
		if _, err := w.Write([]byte(`,"type":` + escape.JSON(t.String()))); err != nil {
			return err
		}
	}
	_, err := w.Write([]byte(`}`))
	return err
}

// MarshalJSON fulfils the JSON Marshaler interface.
func (j *Job) MarshalJSON() ([]byte, error) {
	b := buffers.Get().(*data.Chunk)
	j.JSON(b)
	d := b.Payload()
	returnBuffer(b)
	return d, nil
}

// MarshalJSON fulfils the JSON Marshaler interface.
func (s *Server) MarshalJSON() ([]byte, error) {
	b := buffers.Get().(*data.Chunk)
	s.JSON(b)
	d := b.Payload()
	returnBuffer(b)
	return d, nil
}

// MarshalJSON fulfils the JSON Marshaler interface.
func (s *Session) MarshalJSON() ([]byte, error) {
	b := buffers.Get().(*data.Chunk)
	s.JSON(b)
	d := b.Payload()
	returnBuffer(b)
	return d, nil
}

// MarshalJSON fulfils the JSON Marshaler interface.
func (l *Listener) MarshalJSON() ([]byte, error) {
	b := buffers.Get().(*data.Chunk)
	l.JSON(b)
	d := b.Payload()
	returnBuffer(b)
	return d, nil
}
func (s *Session) updateProxyInfo(v []proxyData) {
	if s.parent == nil {
		return
	}
	s.proxies = v
}
