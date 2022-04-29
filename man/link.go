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
	// This Linker type is only avaliable on Windows devices.
	// non-Windows devices will always return a 'device.ErrNoWindows' error.
	Mutex = objSync(0)
	// Event is a Linker type that can be used with a Guardian.
	// This Linker uses Windows Events to determine Guardian status.
	//
	// This Linker type is only avaliable on Windows devices.
	// non-Windows devices will always return a 'device.ErrNoWindows' error.
	Event = objSync(1)
	// Semaphore is a Linker type that can be used with a Guardian.
	// This Linker uses Windows Semaphores to determine Guardian status.
	//
	// This Linker type is only avaliable on Windows devices.
	// non-Windows devices will always return a 'device.ErrNoWindows' error.
	Semaphore = objSync(2)
	// Mailslot is a Linker type that can be used with a Guardian.
	// This Linker uses Windows MailslotS to determine Guardian status.
	//
	// This Linker type is only avaliable on Windows devices.
	// non-Windows devices will always return a 'device.ErrNoWindows' error.
	Mailslot = objSync(3)
)

type netSync bool
type objSync uint8

// Linker us an interface that specifies an object that can be used to check
// for a Guardian instance.
type Linker interface {
	String() string
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
