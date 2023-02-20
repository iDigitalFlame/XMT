//go:build !go1.13
// +build !go1.13

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
	"bytes"
	"context"
	"net"
	"syscall"
	"time"

	// Importing unsafe to use linkname
	_ "unsafe"
)

type netConfig struct{}

//go:linkname parseIPv4 net.parseIPv4
func parseIPv4(s string) net.IP

func newListenConfig(_ time.Duration) netConfig {
	return netConfig{}
}

//go:linkname parseIPv6 net.parseIPv6
func parseIPv6(s string, a bool) (net.IP, string)
func resolveAddrList(x context.Context, n, s string) ([]net.Addr, error) {
	a, _, err := parseNetwork(x, n, true)
	if err != nil {
		return nil, err
	}
	if len(a) == 0 {
		return nil, syscall.EINVAL
	}
	if (a[0] == 'u' || a[0] == 'U') && len(a) >= 4 && (a[3] == 'x' || a[3] == 'X') {
		v, err := net.ResolveUnixAddr(a, s)
		if err != nil {
			return nil, err
		}
		return []net.Addr{v}, nil
	}
	return internetAddrList(x, a, s)
}
func internetAddrList(x context.Context, n, a string) ([]net.Addr, error) {
	var (
		v    int
		err  error
		h, p string
	)
	switch n[0] {
	case 'i', 'I':
		if len(a) == 0 {
			break
		}
		h = a
	case 't', 'T', 'u', 'U':
		if len(a) == 0 {
			break
		}
		if h, p, err = net.SplitHostPort(a); err != nil {
			return nil, err
		}
		if v, err = net.DefaultResolver.LookupPort(x, n, p); err != nil {
			return nil, err
		}
	default:
		return nil, syscall.EINVAL
	}
	if len(h) == 0 {
		switch n[0] {
		case 't', 'T':
			return []net.Addr{&net.TCPAddr{Port: v}}, nil
		case 'u', 'U':
			return []net.Addr{&net.UDPAddr{Port: v}}, nil
		case 'i', 'I':
			return []net.Addr{&net.IPAddr{}}, nil
		}
		return nil, syscall.EINVAL
	}
	var i []net.IPAddr
	if k := parseIPv4(h); k != nil {
		i = []net.IPAddr{{IP: k}}
	} else if k, z := parseIPv6(h, true); k != nil {
		if i = []net.IPAddr{{IP: k, Zone: z}}; len(k) == len(net.IPv6unspecified) && bytes.Compare(net.IPv6unspecified, k) == 0 {
			i = append(i, net.IPAddr{IP: net.IPv4zero})
		}
	} else if i, err = net.DefaultResolver.LookupIPAddr(x, h); err != nil {
		return nil, err
	}
	var (
		o    = make([]net.Addr, 0, len(i))
		y, u = n[len(n)-1] == '4', n[len(n)-1] == '6'
	)
loop:
	for r := range i {
		switch {
		case !y && !u:
		case i[r].IP.To4() != nil:
		case len(i[r].IP) == net.IPv6len && i[r].IP.To4() == nil:
		default:
			continue loop
		}
		switch n[0] {
		case 't', 'T':
			o = append(o, &net.TCPAddr{IP: i[r].IP, Port: v, Zone: i[r].Zone})
		case 'u', 'U':
			o = append(o, &net.UDPAddr{IP: i[r].IP, Port: v, Zone: i[r].Zone})
		case 'i', 'I':
			o = append(o, &net.IPAddr{IP: i[r].IP, Zone: i[r].Zone})
		default:
			return nil, syscall.EINVAL
		}
	}
	if len(o) == 0 {
		return nil, syscall.EADDRNOTAVAIL
	}
	return o, nil
}

//go:linkname parseNetwork net.parseNetwork
func parseNetwork(x context.Context, n string, p bool) (string, int, error)
func (netConfig) Listen(x context.Context, n, a string) (net.Listener, error) {
	v, err := resolveAddrList(x, n, a)
	if err != nil {
		return nil, err
	}
	for i := range v {
		switch k := v[i].(type) {
		case *net.TCPAddr:
			switch {
			case len(n) == 3 && (n[0] == 't' || n[0] == 'T') && (n[1] == 'c' || n[1] == 'C') && (n[2] == 'p' || n[2] == 'P'):
			case len(n) == 4 && (n[0] == 't' || n[0] == 'T') && (n[1] == 'c' || n[1] == 'C') && (n[2] == 'p' || n[2] == 'P') && (n[3] == '4' || n[3] == '6'):
			default:
				return nil, syscall.EINVAL
			}
			l, err := listenTCP(x, n, k)
			if err != nil {
				return nil, err
			}
			return l, nil
		case *net.UnixAddr:
			switch {
			case len(n) == 4 && (n[0] == 'u' || n[0] == 'U') && (n[3] == 'x' || n[3] == 'X'):
			case len(n) == 8 && (n[0] == 'u' || n[0] == 'U') && (n[7] == 'm' || n[7] == 'M'):
			default:
				return nil, syscall.EINVAL
			}
			l, err := listenUnix(x, n, k)
			if err != nil {
				return nil, err
			}
			return l, nil
		}
	}
	return nil, syscall.EADDRNOTAVAIL
}

//go:linkname listenIP net.listenIP
func listenIP(x context.Context, n string, a *net.IPAddr) (*net.IPConn, error)

//go:linkname listenUDP net.listenUDP
func listenUDP(x context.Context, n string, a *net.UDPAddr) (*net.UDPConn, error)

//go:linkname listenTCP net.listenTCP
func listenTCP(x context.Context, n string, a *net.TCPAddr) (*net.TCPListener, error)
func (netConfig) ListenPacket(x context.Context, n, a string) (net.PacketConn, error) {
	v, err := resolveAddrList(x, n, a)
	if err != nil {
		return nil, err
	}
	for i := range v {
		switch k := v[i].(type) {
		case *net.IPAddr:
			l, err := listenIP(x, n, k)
			if err != nil {
				return nil, err
			}
			return l, nil
		case *net.UDPAddr:
			switch {
			case len(n) == 3 && (n[0] == 'u' || n[0] == 'U') && (n[1] == 'd' || n[1] == 'D') && (n[2] == 'p' || n[2] == 'P'):
			case len(n) == 4 && (n[0] == 'u' || n[0] == 'U') && (n[1] == 'd' || n[1] == 'D') && (n[2] == 'p' || n[2] == 'P') && (n[3] == '4' || n[3] == '6'):
			default:
				return nil, syscall.EINVAL
			}
			l, err := listenUDP(x, n, k)
			if err != nil {
				return nil, err
			}
			return l, nil
		case *net.UnixAddr:
			if len(n) != 8 || (n[0] != 'u' && n[0] != 'U') || (n[7] != 'm' && n[7] != 'M') {
				return nil, syscall.EINVAL
			}
			l, err := listenUnixgram(x, n, k)
			if err != nil {
				return nil, err
			}
			return l, nil
		}
	}
	return nil, syscall.EADDRNOTAVAIL
}

//go:linkname listenUnix net.listenUnix
func listenUnix(x context.Context, n string, a *net.UnixAddr) (*net.UnixListener, error)

//go:linkname listenUnixgram net.listenUnixgram
func listenUnixgram(x context.Context, n string, a *net.UnixAddr) (*net.UnixConn, error)
