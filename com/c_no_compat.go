//go:build go1.18
// +build go1.18

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
	"net/netip"
	"syscall"
)

var emptyAddr netip.AddrPort

type udpAddr netip.AddrPort
type udpSockInternal interface {
	WriteToUDPAddrPort([]byte, netip.AddrPort) (int, error)
	ReadFromUDPAddrPort([]byte) (int, netip.AddrPort, error)
}

func (u udpAddr) IsValid() bool {
	return netip.AddrPort(u).IsValid()
}
func (u udpAddr) String() string {
	// NOTE(dij): This causes IPv4 addresses to weirdly be wrapped as an IPv6
	//            address. This doesn't seem to affect how it works on IPv4, but
	//            we'll watch it. It only makes IPv4 addresses print out as IPv6
	//            formatted addresses.
	return netip.AddrPort(u).String()
}
func (u *udpCompat) ReadPacket(p []byte) (int, udpAddr, error) {
	n, a, err := u.udpSock.ReadFromUDPAddrPort(p)
	return n, udpAddr(a), err
}
func (u *udpCompat) WritePacket(p []byte, a udpAddr) (int, error) {
	return u.udpSock.WriteToUDPAddrPort(p, netip.AddrPort(a))
}
func (i *ipPacketConn) ReadFromUDPAddrPort(b []byte) (int, netip.AddrPort, error) {
	// NOTE(dij): Have to add this as there's no support for the netip
	//            package for IPConns.
	n, a, err := i.ReadFrom(b)
	if a == nil {
		return n, emptyAddr, err
	}
	v, ok := a.(*net.IPAddr)
	if !ok {
		if err != nil {
			return n, emptyAddr, err
		}
		return n, emptyAddr, syscall.EINVAL
	}
	x, _ := netip.AddrFromSlice(v.IP)
	return n, netip.AddrPortFrom(x, 0), err
}
func (i *ipPacketConn) WriteToUDPAddrPort(b []byte, a netip.AddrPort) (int, error) {
	return i.WriteTo(b, &net.IPAddr{IP: a.Addr().AsSlice()})
}
