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

package cfg

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/iDigitalFlame/xmt/util"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

const (
	// SelectorLastValid is the default selection that will keep using the last
	// Group unless it fails. On a failure (or the first call), this will act
	// similar to 'SelectorRoundRobin'.
	//
	// Takes effect only if there are multiple Groups in this Config.
	// This value is GLOBAL and can be present in any Group!
	SelectorLastValid = cBit(0xAA)
	// SelectorRoundRobin is a selection option that will simply select the NEXT
	// Group on every connection attempt. This option is affected by the Group
	// weights set on each addition and will prefer higher numbered options in
	// order.
	//
	// Takes effect only if there are multiple Groups in this Config.
	// This value is GLOBAL and can be present in any Group!
	SelectorRoundRobin = cBit(0xAB)
	// SelectorRandom is a selection option that will ignore all weights and order
	// and will select an entry from the list randomly.
	//
	// Takes effect only if there are multiple Groups in this Config.
	// This value is GLOBAL and can be present in any Group!
	SelectorRandom = cBit(0xAC)
	// SelectorSemiRoundRobin is a selection option that will potentially select
	// the NEXT Group dependent on a random (25%) chance on every connection
	// attempt. This option is affected by the Group weights set on each addition
	// and will prefer higher numbered options in order. Otherwise, the last
	// group used is kept.
	//
	// Takes effect only if there are multiple Groups in this Config.
	// This value is GLOBAL and can be present in any Group!
	SelectorSemiRoundRobin = cBit(0xAD)
	// SelectorSemiRandom is a selection option that will ignore all weights and
	// order and will select an entry from the list randomly dependent on a
	// random (25%) chance on every connection attempt. Otherwise, the last
	// group used is kept.
	//
	// Takes effect only if there are multiple Groups in this Config.
	// This value is GLOBAL and can be present in any Group!
	SelectorSemiRandom = cBit(0xAE)
)

// Group is a struct that allows for using multiple connections for a single
// Session.
//
// Groups are automatically created when a Config is built into a Profile
// that contains multiple Profile 'Groups'.
type Group struct {
	lock sync.Mutex

	cur     *profile
	entries []*profile

	src []byte
	sel uint8
}
type profile struct {
	w    Wrapper
	t    Transform
	kill *time.Time
	work *WorkHours
	conn any

	src   []byte
	hosts []string
	sleep time.Duration

	weight uint8
	jitter int8
}

func (g *Group) init() {
	if g.cur == nil {
		g.Switch(false)
	}
}

// Len implements the 'sort.Interface' interface, this allows for a Group to
// be sorted.
func (g *Group) Len() int {
	return len(g.entries)
}

// Jitter returns a value that represents a percentage [0-100] that will be taken
// into account by a Session in order to skew its connection timeframe.
//
// The value zero (0) is used to signify that Jitter is disabled. Other values
// greater than one hundred (100) are ignored, as well as values below zero.
//
// The special value '-1' indicates that this Profile does not set a Jitter value
// and to use the system default '10%'.
func (g *Group) Jitter() int8 {
	if g.init(); g.cur == nil {
		return -1
	}
	return g.cur.jitter
}

// Swap implements the 'sort.Interface' interface, this allows for a Group to be
// sorted.
func (g *Group) Swap(i, j int) {
	g.entries[i], g.entries[j] = g.entries[j], g.entries[i]
}
func (p *profile) Jitter() int8 {
	return p.jitter
}
func (profile) Switch(_ bool) bool {
	return false
}

// Less implements the 'sort.Interface' interface, this allows for a Group to be
// sorted.
func (g *Group) Less(i, j int) bool {
	return g.entries[i].weight > g.entries[j].weight
}

// Switch is function that will indicate to the caller if the 'Next' function
// needs to be called. Calling this function has the potential to advance the
// Profile group, if available.
//
// The supplied boolean must be true if the last call to 'Connect' ot 'Listen'
// resulted in an error or if a forced switch if warranted.
// This indicates to the Profile is "dirty" and a switchover must be done.
//
// It is recommended to call the 'Next' function after if the result of this
// function is true.
//
// Static Profile variants may always return 'false' to prevent allocations.
func (g *Group) Switch(e bool) bool {
	if (g.cur != nil && !e && g.sel == 0) || len(g.entries) == 0 {
		return false
	}
	if g.sel == 0 && !e && g.cur != nil {
		return false
	}
	if g.cur != nil && (g.sel == 3 || g.sel == 4) && util.FastRandN(4) != 0 {
		return false
	}
	if g.lock.Lock(); g.sel == 2 || g.sel == 4 {
		if n := g.entries[util.FastRandN(len(g.entries))]; g.cur != n {
			g.cur = n
			g.lock.Unlock()
			return true
		}
		g.lock.Unlock()
		return false
	}
	if g.cur == nil {
		g.cur = g.entries[0]
		g.lock.Unlock()
		return true
	}
	var f bool
	for i := range g.entries {
		if g.entries[i] == g.cur {
			f = true
			continue
		}
		if f {
			g.cur = g.entries[i]
			g.lock.Unlock()
			return true
		}
	}
	if f && g.cur == g.entries[0] {
		g.lock.Unlock()
		return false
	}
	g.cur = g.entries[0]
	g.lock.Unlock()
	return true
}

// Sleep returns a value that indicates the amount of time a Session should wait
// before attempting communication again, modified by Jitter (if enabled).
//
// Sleep MUST be greater than zero (0), any value that is zero or less is ignored
// and indicates that this profile does not set a Sleep value and will use the
// system default '60s'.
func (g *Group) Sleep() time.Duration {
	if g.init(); g.cur == nil {
		return -1
	}
	return g.cur.sleep
}

// KillDate returns a value that indicates the date and time when the Session will
// stop functioning. If this value is nil, there is no 'KillDate'.
func (g *Group) KillDate() *time.Time {
	if g.init(); g.cur == nil {
		return nil
	}
	return g.cur.kill
}
func (p *profile) KillDate() *time.Time {
	return p.kill
}
func (p *profile) Sleep() time.Duration {
	return p.sleep
}

// WorkHours returns a value that indicates when the Session should be active
// and communicate with the C2 Server. The returned struct can be used to
// determine which days the Session can connect.
//
// If the returned value is nil, there are no Working hours restrictions.
func (g *Group) WorkHours() *WorkHours {
	if g.init(); g.cur == nil {
		return nil
	}
	return g.cur.work
}
func (p *profile) WorkHours() *WorkHours {
	return p.work
}

// MarshalBinary allows the source of this Group to be retrieved to be reused
// again.
//
// This function returns an error if the source is not available.
func (g *Group) MarshalBinary() ([]byte, error) {
	if len(g.src) > 0 {
		return g.src, nil
	}
	return nil, xerr.Sub("binary source not available", 0x60)
}
func (p *profile) MarshalBinary() ([]byte, error) {
	if len(p.src) > 0 {
		return p.src, nil
	}
	return nil, xerr.Sub("binary source not available", 0x60)
}

// Next is a function call that can be used to grab the Profile's current target
// along with the appropriate Wrapper and Transform.
//
// Implementations of a Profile are recommend to ensure that this function does
// not affect how the Profile currently works until a call to 'Switch' as this
// WILL be called on startup of a Session.
func (g *Group) Next() (string, Wrapper, Transform) {
	if g.init(); g.cur == nil {
		return "", nil, nil
	}
	return g.cur.Next()
}
func (p *profile) Next() (string, Wrapper, Transform) {
	if len(p.hosts) == 0 {
		return "", p.w, p.t
	}
	if len(p.hosts) == 1 {
		return p.hosts[0], p.w, p.t
	}
	return p.hosts[util.FastRandN(len(p.hosts))], p.w, p.t
}

// Connect is a function that will preform a Connection attempt against the
// supplied address string. This function may return an error if a connection
// could not be made or if this Profile does not support Client-side connections.
//
// It is recommended for implementations to implement using the passed Context
// to stop in-flight calls.
func (g *Group) Connect(x context.Context, s string) (net.Conn, error) {
	if g.init(); g.cur == nil {
		return nil, ErrNotAConnector
	}
	return g.cur.Connect(x, s)
}
func (p *profile) Connect(x context.Context, s string) (net.Conn, error) {
	c, ok := p.conn.(Connector)
	if !ok {
		return nil, ErrNotAConnector
	}
	return c.Connect(x, s)
}

// Listen is a function that will attempt to create a listening connection on
// the supplied address string. This function may return an error if a listener
// could not be created or if this Profile does not support Server-side connections.
//
// It is recommended for implementations to implement using the passed Context
// to stop running Listeners.
func (g *Group) Listen(x context.Context, s string) (net.Listener, error) {
	if g.init(); g.cur == nil {
		return nil, ErrNotAListener
	}
	return g.cur.Listen(x, s)
}
func (p *profile) Listen(x context.Context, s string) (net.Listener, error) {
	l, ok := p.conn.(Accepter)
	if !ok {
		return nil, ErrNotAListener
	}
	return l.Listen(x, s)
}
