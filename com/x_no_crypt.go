//go:build !crypt

package com

// Named Network Constants
const (
	NameIP   = "ip"
	NameTCP  = "tcp"
	NameUDP  = "udp"
	NamePipe = "pipe"
	NameUnix = "unix"
	NameHTTP = "http"
)

func (udpErr) Error() string {
	return "deadline exceeded"
}
