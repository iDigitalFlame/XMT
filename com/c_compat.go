//go:build !go1.18
// +build !go1.18

// Copyright (C) 2020 - 2023 iDigitalFlame
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

package com

import (
	"net"

	"github.com/iDigitalFlame/xmt/device"
)

type udpAddr struct {
	p uint16
	device.Address
}
type udpSockInternal interface{}

func (u udpAddr) IsValid() bool {
	return u.IsValid()
}
func (u udpAddr) String() string {
	return u.Address.String()
}
func (u *udpCompat) ReadPacket(p []byte) (int, udpAddr, error) {
	var (
		o         udpAddr
		n, a, err = u.udpSock.ReadFrom(p)
	)
	if a != nil {
		switch t := a.(type) {
		case *net.IPAddr:
			o.Set(t.IP)
		case *net.UDPAddr:
			o.Set(t.IP)
			o.p = uint16(t.Port)
		}
	}
	return n, o, err
}
func (u *udpCompat) WritePacket(p []byte, a udpAddr) (int, error) {
	var o net.Addr
	if a.p > 0 {
		o = &net.UDPAddr{IP: a.IP(), Port: int(a.p)}
	} else {
		o = &net.IPAddr{IP: a.IP()}
	}
	n, err := u.udpSock.WriteTo(p, o)
	o = nil
	return n, err
}
