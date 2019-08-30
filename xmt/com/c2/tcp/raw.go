package tcp

import (
	"fmt"
	"net"
	"time"

	"github.com/iDigitalFlame/xmt/xmt/com/c2"
)

var (
	// Raw is the TCP Raw connection profile.  This profile uses raw TCP
	// connections without any encoding or Transports.  Used for debugging mostly.
	Raw = &rawProfile{
		Dialer: &net.Dialer{
			Timeout: RawDefaultTimeout,
		},
	}

	// RawDefaultTimeout is the default timeout used for the Raw TCP profile.
	// The default is 5 seconds.
	RawDefaultTimeout = time.Duration(5) * time.Second
)

type rawProfile struct {
	Dialer *net.Dialer
}
type rawListener struct {
	l *net.TCPListener
}
type rawConnection struct {
	net.Conn
}

func (r *rawProfile) Size() int {
	return -1
}
func (r *rawListener) Close() error {
	return r.l.Close()
}
func (r *rawConnection) IP() string {
	return r.RemoteAddr().String()
}
func (r *rawListener) String() string {
	return fmt.Sprintf("TCP Raw: %s", r.l.Addr().String())
}
func (r *rawProfile) Wrapper() c2.Wrapper {
	return nil
}
func (r *rawProfile) Transport() c2.Transport {
	return nil
}
func (r *rawListener) Accept() (c2.Connection, error) {
	c, err := r.l.AcceptTCP()
	if err != nil {
		return nil, err
	}
	return &rawConnection{c}, nil
}
func (r *rawProfile) Listen(s string) (c2.Listener, error) {
	b, err := net.ResolveTCPAddr("tcp", s)
	if err != nil {
		return nil, err
	}
	l, err := net.ListenTCP("tcp", b)
	if err != nil {
		return nil, err
	}
	return &rawListener{l: l}, nil
}
func (r *rawProfile) Connect(s string) (c2.Connection, error) {
	c, err := r.Dialer.Dial("tcp", s)
	if err != nil {
		return nil, err
	}
	return &rawConnection{c}, nil
}
