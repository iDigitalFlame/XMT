package c2

import (
	"context"
	"net"
	"sort"
	"time"

	"github.com/iDigitalFlame/xmt/c2/cout"
	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/util/bugtrack"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

var wake struct{}

type event struct {
	s  *Session
	p  *com.Packet
	j  *Job
	jf func(*Job)
	sf func(*Session)
	pf func(*com.Packet)
	af func(*com.Packet) bool
	hf func(*Session, *com.Packet) bool
}

type cluster struct {
	data []*com.Packet
	max  uint16
}
type proxyData struct {
	n, b string
	p    []byte
}
type eventer chan event
type connection struct {
	s   *Server
	w   Wrapper
	t   Transform
	m   messager
	p   Profile
	ctx context.Context
	log *cout.Log
}
type messager interface {
	close()
	count() int
	queue(event)
}
type runnable interface {
	Pid() uint32
	Start() error
	Release() error
}
type notifier interface {
	accept(uint16)
	frag(uint16, uint16, uint16)
}
type marshaler interface {
	MarshalBinary() (data []byte, err error)
}
type readerTimeout struct {
	_ [0]func()
	c net.Conn
	t time.Duration
	i bool
}

func (e eventer) close() {
	close(e)
}
func (c *cluster) Len() int {
	return len(c.data)
}
func (e eventer) count() int {
	return len(e)
}
func (e eventer) queue(x event) {
	e <- x
}
func (c *cluster) Swap(i, j int) {
	c.data[i], c.data[j] = c.data[j], c.data[i]
}
func (e eventer) listen(s *Session) {
	if bugtrack.Enabled {
		defer bugtrack.Recover("c2.eventer.listen()")
	}
	if cout.Enabled {
		s.log.Info("Client-side event processing thread started!")
	}
	for {
		select {
		case <-s.ctx.Done():
			s.Remove()
			s.Close()
			return
		case v := <-e:
			v.process(s.log)
		}
	}
}
func (e event) process(l *cout.Log) {
	defer func() {
		if err := recover(); err != nil {
			if cout.Enabled {
				l.Error("Server event processing function recovered from a panic: %s!", err)
			}
		}
	}()
	switch {
	case e.af != nil && e.p != nil && e.s != nil: // Direct client side packet handler.
		if e.af(e.p) && e.s.Receive == nil { // Handled and Receive is nil, clear and break.
			break
		}
		if e.s.Receive != nil { // If Receive is not nil, call it then break.
			e.s.Receive(e.s, e.p)
			break
		}
		if e.s.recv != nil && e.s.state.CanRecv() { // If there's a receive channel setup, use that.
			select {
			case e.s.recv <- e.p:
			default:
				if cout.Enabled {
					l.Warning("[%s] Packet %q was dropped during receive!", e.s.ID, e.p)
				}
			}
			break
		}
		e.p.Clear() // Otherwise, clear the packet.
	case e.hf != nil && e.p != nil && e.s != nil: // Mux packet handler.
		if e.hf(e.s, e.p) && e.s.Receive == nil {
			break
		}
		if e.s.Receive != nil {
			e.s.Receive(e.s, e.p)
			break
		}
		if e.s.recv != nil && e.s.state.CanRecv() {
			select {
			case e.s.recv <- e.p:
			default:
				if cout.Enabled {
					l.Warning("[%s] Packet %q was dropped during receive!", e.s.ID, e.p)
				}
			}
			break
		}
		e.p.Clear()
	case e.pf != nil && e.p != nil: // Oneshot handler
		e.pf(e.p)
	case e.sf != nil && e.s != nil: // Session New or Shutdown handler.
		e.sf(e.s)
	case e.jf != nil && e.j != nil: // Job Update handler
		e.jf(e.j)
	}
}
func (c *cluster) done() *com.Packet {
	if len(c.data) == 0 {
		return nil
	}
	if uint16(len(c.data)) >= c.max {
		sort.Sort(c)
		n := c.data[0]
		for x := 1; x < len(c.data); x++ {
			n.Add(c.data[x])
			c.data[x].Clear()
			c.data[x] = nil
		}
		c.data = nil
		n.Flags.Clear()
		return n
	}
	return nil
}
func (c *cluster) Less(i, j int) bool {
	return c.data[i].Flags.Position() < c.data[j].Flags.Position()
}
func (c *cluster) add(p *com.Packet) error {
	if p == nil || p.Empty() || p.Flags.Len() <= c.max {
		return nil
	}
	if len(c.data) > 0 && !c.data[0].Belongs(p) {
		return xerr.Sub("packet ID does not match the supplied ID", 0xD)
	}
	if p.Flags.Len() > c.max {
		c.max = p.Flags.Len()
	}
	if p.Empty() {
		return nil
	}
	c.data = append(c.data, p)
	return nil
}
func (r *readerTimeout) Read(b []byte) (int, error) {
	if r.i {
		r.c.SetReadDeadline(time.Now().Add(r.t))
	} else {
		r.i = true
	}
	return r.c.Read(b)
}
