package pipe

import (
	"context"
	"net"
	"time"
)

// Pipe is the default Connector with the default timeout of 15 seconds.
const Pipe = Piper(time.Second * 15)

// Piper is a Connector that can be used with the 'c2' package to make Pipe
// connections for C2.
type Piper time.Duration

// Connect fulfills the Connector interface.
func (p Piper) Connect(x context.Context, a string) (net.Conn, error) {
	return DialContext(x, Format(a))
}

// Listen fulfills the Connector interface.
func (Piper) Listen(x context.Context, a string) (net.Listener, error) {
	return ListenPermsContext(x, Format(a), PermEveryone)
}
