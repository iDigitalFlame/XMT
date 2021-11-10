package pipe

import (
	"net"
	"time"
)

// Pipe is the default Connector with the default timeout of 15 seconds.
const Pipe = Piper(time.Second * 15)

// Piper is a Connector that can be used with the 'c2' package to make Pipe
// connections for C2.
type Piper time.Duration

// Connect fulfills the Connector interface.
func (p Piper) Connect(a string) (net.Conn, error) {
	return DialTimeout(Format(a), time.Duration(p))
}

// Listen fulfills the Connector interface.
func (Piper) Listen(a string) (net.Listener, error) {
	return ListenPerms(Format(a), PermEveryone)
}
