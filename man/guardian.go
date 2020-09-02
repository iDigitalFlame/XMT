package man

import (
	"context"
	"net"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/iDigitalFlame/xmt/com/pipe"
	"github.com/iDigitalFlame/xmt/data/crypto"
)

var bufs = sync.Pool{
	New: func() interface{} {
		b := make([]byte, 65)
		return &b
	},
}

type waker struct{}

// Guardian is a struct that is used to maintain a running process that will re-establish itself
// if it is not detected running. Guardian instances take advantage of local sockets (*NIX), Named Pipes (Windows) or
// TCP sockets (if specified). This struct will maintain a thread that will 'heartbeat' back to the Sentinel
// connections made.
type Guardian struct {
	ch     chan waker
	ctx    context.Context
	sock   net.Listener
	done   uint32
	cancel context.CancelFunc
}

// Wait will block until the Guardian is closed.
func (g *Guardian) Wait() {
	<-g.ch
}
func (g *Guardian) listen() {
	go func() {
		<-g.ctx.Done()
		g.sock.Close()
	}()
	for atomic.LoadUint32(&g.done) == 0 {
		c, err := g.sock.Accept()
		if err != nil {
			e, ok := err.(net.Error)
			if ok && e.Timeout() {
				continue
			}
			if ok && !e.Timeout() && !e.Temporary() {
				break
			}
			continue
		}
		go handleSock(c)
	}
	g.sock.Close()
	g.cancel()
	close(g.ch)
}
func handleSock(c net.Conn) {
	var (
		b      = *bufs.Get().(*[]byte)
		n, err = c.Read(b)
	)
	if err == nil && n == 65 && b[0] == 0xFF {
		copy(b[1:], crypto.SHA512(b[1:]))
		b[0] = 0xA0
		c.Write(b)
	}
	bufs.Put(&b)
	c.Close()
}

// Close will close the Guardian and stoppings the listener. Any errors during listener close will be returned.
func (g *Guardian) Close() error {
	atomic.StoreUint32(&g.done, 1)
	g.cancel()
	return g.sock.Close()
}

// MustGuard returns a Guardian instance that watches on the name provided. This function must complete
// and will panic if an error occurs. Otherwise a Guardian instance is returned.
func MustGuard(n string) *Guardian {
	g, err := Guard(n)
	if err != nil {
		panic(err)
	}
	return g
}

// Guard will attempt to create a Guardian instance on the provided name. This function will return an error if the
// name is already being listened on. This function will choose the proper connection method based on the host
// operating system.
func Guard(n string) (*Guardian, error) {
	return GuardContext(context.Background(), n)
}

// GuardTCP will attempt to create a Guardian instance on the provided localhost TCP port. This function will return an
// error if the port is already being listened on.
func GuardTCP(p uint16) (*Guardian, error) {
	return GuardContextTCP(context.Background(), p)
}

// GuardContext will attempt to create a Guardian instance on the provided name. This function will return an error
// if the name is already being listened on. This function will choose the proper connection method based on the host
// operating system. This function also takes a context.Context to be used for resource control.
func GuardContext(x context.Context, n string) (*Guardian, error) {
	l, err := pipe.ListenPerms(pipe.Format(n), pipe.PermEveryone)
	if err != nil {
		return nil, err
	}
	g := &Guardian{ch: make(chan waker, 1), sock: l}
	g.ctx, g.cancel = context.WithCancel(x)
	go g.listen()
	return g, nil
}

// GuardContextTCP will attempt to create a Guardian instance on the provided localhost TCP port. This function will
// return an error if the port is already being listened on. This function also takes a context.Context to be used for
// resource control.
func GuardContextTCP(x context.Context, p uint16) (*Guardian, error) {
	l, err := net.Listen("tcp", "127.0.0.1"+strconv.Itoa(int(p)))
	if err != nil {
		return nil, err
	}
	g := &Guardian{ch: make(chan waker, 1), sock: l}
	g.ctx, g.cancel = context.WithCancel(x)
	go g.listen()
	return g, nil
}
