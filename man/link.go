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

package man

import (
	"bytes"
	"context"
	"crypto/sha512"
	"net"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/com/pipe"
	"github.com/iDigitalFlame/xmt/util"
	"github.com/iDigitalFlame/xmt/util/bugtrack"
)

const (
	// TCP is a Linker type that can be used with a Guardian.
	// This Linker uses raw TCP sockets to determine Guardian status.
	TCP = netSync(false)
	// Pipe is a Linker type that can be used with a Guardian.
	// This Linker uses Unix Domain Sockets in *nix devices and Named Pipes
	// in Windows devices.
	//
	// Pipe names are prefixed with the appropriate namespace before being
	// checked or created (if it doesn't exist already).
	//
	// This is the default Linker used if a nil Linker is used.
	Pipe = netSync(true)

	// Mutex is a Linker type that can be used with a Guardian.
	// This Linker uses Windows Mutexes to determine Guardian status.
	//
	// This Linker type is only available on Windows devices.
	// non-Windows devices will always return a 'device.ErrNoWindows' error.
	Mutex = objSync(0)
	// Event is a Linker type that can be used with a Guardian.
	// This Linker uses Windows Events to determine Guardian status.
	//
	// This Linker type is only available on Windows devices.
	// non-Windows devices will always return a 'device.ErrNoWindows' error.
	Event = objSync(1)
	// Semaphore is a Linker type that can be used with a Guardian.
	// This Linker uses Windows Semaphores to determine Guardian status.
	//
	// This Linker type is only available on Windows devices.
	// non-Windows devices will always return a 'device.ErrNoWindows' error.
	Semaphore = objSync(2)
	// Mailslot is a Linker type that can be used with a Guardian.
	// This Linker uses Windows Mailslots to determine Guardian status.
	//
	// This Linker type is only available on Windows devices.
	// non-Windows devices will always return a 'device.ErrNoWindows' error.
	Mailslot = objSync(3)
)

type netSync bool
type objSync uint8

// Linker us an interface that specifies an object that can be used to check
// for a Guardian instance.
type Linker interface {
	check(s string) (bool, error)
	create(s string) (listener, error)
}
type netListener struct {
	_    [0]func()
	l    net.Listener
	done uint32
}
type listener interface {
	Listen()
	Close() error
}

func (n *netListener) Listen() {
	for atomic.LoadUint32(&n.done) == 0 {
		c, err := n.l.Accept()
		if err != nil {
			e, ok := err.(net.Error)
			if ok && e.Timeout() {
				continue
			}
			if ok && !e.Timeout() {
				break
			}
			continue
		}
		go netHandleConn(c)
	}
	n.l.Close()
}
func netHandleConn(c net.Conn) {
	var (
		b      [65]byte
		n, err = c.Read(b[:])
	)
	if err == nil && n == 65 && b[0] == 0xFF {
		if bugtrack.Enabled {
			bugtrack.Track("man.handleConnSock(): Connection from %q handled.", c.RemoteAddr())
		}
		h := sha512.New()
		h.Write(b[1:])
		copy(b[1:], h.Sum(nil))
		b[0], h = 0xA0, nil
		c.Write(b[:])
	}
	c.Close()
}
func (n netSync) String() string {
	if n {
		return com.NamePipe
	}
	return com.NameTCP
}
func (n *netListener) Close() error {
	atomic.StoreUint32(&n.done, 1)
	return n.l.Close()
}
func formatTCPName(s string) string {
	if i := strings.IndexByte(s, ':'); i >= 0 || len(s) == 0 {
		return s
	}
	if _, err := strconv.ParseUint(s, 10, 16); err == nil {
		return local + s
	}
	h := uint32(2166136261)
	for x := range s {
		h *= 16777619
		h ^= uint32(s[x])
	}
	v := uint16(h)
	if v < 1024 {
		v += 1024
	}
	return local + strconv.FormatUint(uint64(v), 10)
}

// LinkerFromName will attempt to map the name provided to an appropriate Linker
// interface.
//
// If no linker is found, the 'Pipe' Linker will be returned.
func LinkerFromName(n string) Linker {
	if len(n) == 0 {
		return Pipe
	}
	if len(n) == 1 {
		switch n[0] {
		case 't', 'T':
			return TCP
		case 'p', 'P':
			return Pipe
		case 'e', 'E':
			return Event
		case 'm', 'M':
			return Mutex
		case 'n', 'N':
			return Mailslot
		case 's', 'S':
			return Semaphore
		}
		return Pipe
	}
	switch {
	case len(n) == 3 && (n[0] == 't' || n[0] == 'T'):
		return TCP
	case len(n) == 4 && (n[0] == 'p' || n[0] == 'P'):
		return Pipe
	case len(n) == 5 && (n[0] == 'e' || n[0] == 'E'):
		return Event
	case len(n) == 5 && (n[0] == 'm' || n[0] == 'M'):
		return Mutex
	case len(n) == 8 && (n[0] == 'm' || n[0] == 'M'):
		return Mailslot
	case len(n) == 9 && (n[0] == 's' || n[0] == 'S'):
		return Semaphore
	}
	return Pipe
}
func netCheckConn(c net.Conn) (bool, error) {
	var (
		b    [65]byte
		_, _ = util.Rand.Read(b[1:])
		v    = sha512.New()
		_, _ = v.Write(b[1:])
		h    = v.Sum(nil)
	)
	b[0], v = 0xFF, nil
	c.SetDeadline(time.Now().Add(timeout))
	if _, err := c.Write(b[:]); err != nil {
		return false, err
	}
	if n, err := c.Read(b[:]); err != nil || n != 65 {
		return false, err
	}
	return b[0] == 0xA0 && bytes.Equal(b[1:], h), nil
}
func (n netSync) check(s string) (bool, error) {
	var (
		c   net.Conn
		err error
	)
	if n {
		c, err = pipe.DialTimeout(pipe.Format(s), timeout)
	} else {
		c, err = net.DialTimeout(com.NameTCP, formatTCPName(s), timeout)
	}
	if err != nil {
		return false, nil
	}
	v, err := netCheckConn(c)
	c.Close()
	return v, err
}
func (n netSync) create(s string) (listener, error) {
	var (
		l   net.Listener
		err error
	)
	if n {
		l, err = pipe.ListenPerms(pipe.Format(s), pipe.PermEveryone)
	} else {
		l, err = com.ListenConfig.Listen(context.Background(), com.NameTCP, formatTCPName(s))
	}
	if err != nil {
		return nil, err
	}
	return &netListener{l: l}, nil
}
