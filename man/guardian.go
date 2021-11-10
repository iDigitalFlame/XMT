package man

import (
	"context"

	"github.com/iDigitalFlame/xmt/util/xerr"
)

// Guardian is a struct that is used to maintain a running process that will re-establish itself
// if it is not detected running. Guardian instances use Linker interfaces to determine status.
type Guardian struct {
	ch   chan struct{}
	sock listener
}

// Wait will block until the Guardian is closed.
func (g *Guardian) Wait() {
	<-g.ch
}

// Close will close the Guardian and stoppings the listener. Any errors during listener close will be returned.
// This function will block until the Guardian fully closes.
func (g *Guardian) Close() error {
	if g.sock == nil {
		return nil
	}
	err := g.sock.Close()
	close(g.ch)
	g.sock = nil
	return err
}

// MustGuard returns a Guardian instance that watches on the name provided. This function must complete
// and will panic if an error occurs. Otherwise a Guardian instance is returned.
//
// This function defaults to the 'Pipe' Linker if a nil Linker is specified.
func MustGuard(l Linker, n string) *Guardian {
	g, err := GuardContext(context.Background(), l, n)
	if err != nil {
		panic(err)
	}
	return g
}

// Guard will attempt to create a Guardian instance on the provided name using the specified Linker.
// This function will return an error if the name is already being listened on.
//
// This function defaults to the 'Pipe' Linker if a nil Linker is specified.
func Guard(l Linker, n string) (*Guardian, error) {
	return GuardContext(context.Background(), l, n)
}

// MustGuardContext returns a Guardian instance that watches on the name provided. This function must complete
// and will panic if an error occurs. Otherwise a Guardian instance is returned. This function also takes a
// context.Context to be used for resource control.
func MustGuardContext(x context.Context, l Linker, n string) *Guardian {
	g, err := GuardContext(x, l, n)
	if err != nil {
		panic(err)
	}
	return g
}

// GuardContext will attempt to create a Guardian instance on the provided name. This function will return an error
// if the name is already being listened on. This function will choose the proper connection method based on the host
// operating system. This function also takes a context.Context to be used for resource control.
func GuardContext(x context.Context, l Linker, n string) (*Guardian, error) {
	if len(n) == 0 {
		return nil, xerr.New("name cannot be empty")
	}
	if l == nil {
		l = Pipe
	}
	v, err := l.create(n)
	if err != nil {
		return nil, err
	}
	g := &Guardian{ch: make(chan struct{}, 1), sock: v}
	if x != nil && x != context.Background() {
		go func() {
			select {
			case <-g.ch:
			case <-x.Done():
				g.Close()
			}
		}()
	}
	go v.Listen()
	return g, nil
}
