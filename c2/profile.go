package c2

import (
	"io"
	"net"
	"time"
)

const (
	// DefaultSleep is the default sleep Time when the provided sleep value is empty or negative.
	DefaultSleep = time.Duration(60) * time.Second

	// DefaultJitter is the default Jitter value when the provided jitter value is negative.
	DefaultJitter uint8 = 5
)

// Static is a simple static Profile implementation.
// If 'S' or 'J' are omitted or zero values, they will be replaced with the
// DefaultJitter and DefaultSleep values.
//
// This struct DOES NOT fill any hinter interfaces!
type Static struct {
	// W is the Wrapper
	W Wrapper
	// T is the Transform
	T Transform
	// S is the Sleep duration
	S time.Duration
	// J is the Jitter percentage
	J uint8
}
type hinter interface {
	Host() string
	Listener() Accepter
	Connector() Connector
}

// Profile is an interface that defines a C2 profile. This is used for defining the specifics that will
// be used to listen by servers and for connections by clients.
type Profile interface {
	Jitter() uint8
	Wrapper() Wrapper
	Sleep() time.Duration
	Transform() Transform
}

// Wrapper is an interface that wraps the binary streams into separate stream types. This allows for using
// encryption or compression (or both!).
type Wrapper interface {
	Unwrap(io.Reader) (io.Reader, error)
	Wrap(io.WriteCloser) (io.WriteCloser, error)
}
type stackCloser struct {
	s io.WriteCloser
	io.WriteCloser
}

// Accepter is an interface that can be used to create listening sockets. This interface
// defines a single function that returns a listener based on a single accept address string.
type Accepter interface {
	Listen(string) (net.Listener, error)
}

// Connector is an interface that can be used to connect to listening sockets. This interface defines
// a single function that returns a Connected socket based on the single connection string.
type Connector interface {
	Connect(string) (net.Conn, error)
}

// Transform is an interface that can modify the data BEFORE it is written or AFTER is read from a Connection.
// Transforms may be used to mask and unmask communications as benign protocols such as DNS, FTP or HTTP.
type Transform interface {
	Read([]byte, io.Writer) error
	Write([]byte, io.Writer) error
}

// MultiWrapper is an alias for an array of Wrappers. This will preform the wrapper/unwrapping operations in the
// order of the array. This is automatically created by a Config instance when multiple Wrappers are present.
type MultiWrapper []Wrapper

// ConnectFunc is a wrapper alias that will fulfil the client interface and allow using a single function
// instead of creating a struct to create connections. This can be used in all Server 'Connect' function calls.
type ConnectFunc func(string) (net.Conn, error)

// ListenerFunc is a wrapper alias that will fulfil the listener interface and allow using a single function
// instead of creating a struct to create listeners. This can be used in all Server 'Listen' function calls.
type ListenerFunc func(string) (net.Listener, error)

// Jitter fulfils the Profile interface.
func (s Static) Jitter() uint8 {
	if s.J == 0 || s.J > 100 {
		return DefaultJitter
	}
	return s.J
}

// Wrapper fulfils the Profile interface.
func (s Static) Wrapper() Wrapper {
	return s.W
}
func (s *stackCloser) Close() error {
	if err := s.WriteCloser.Close(); err != nil {
		return err
	}
	return s.s.Close()
}

// Sleep fulfils the Profile interface.
func (s Static) Sleep() time.Duration {
	if s.S <= 0 {
		return DefaultSleep
	}
	return s.S
}

// Transform fulfils the Profile interface.
func (s Static) Transform() Transform {
	return s.T
}

// Connect fulfills the serverClient interface.
func (c ConnectFunc) Connect(a string) (net.Conn, error) {
	return c(a)
}

// Listen fulfills the serverListener interface.
func (l ListenerFunc) Listen(a string) (net.Listener, error) {
	return l(a)
}

// Unwrap satisfies the Wrapper interface.
func (m MultiWrapper) Unwrap(r io.Reader) (io.Reader, error) {
	var (
		o   = r
		err error
	)
	for x := len(m) - 1; x >= 0; x-- {
		if o, err = m[x].Unwrap(o); err != nil {
			return nil, err
		}
	}
	return o, nil
}

// Wrap satisfies the Wrapper interface.
func (m MultiWrapper) Wrap(w io.WriteCloser) (io.WriteCloser, error) {
	var (
		o   = w
		k   io.WriteCloser
		err error
	)
	for x := len(m) - 1; x >= 0; x-- {
		if k, err = m[x].Wrap(o); err != nil {
			return nil, err
		}
		o = &stackCloser{s: o, WriteCloser: k}
	}
	return o, nil
}
