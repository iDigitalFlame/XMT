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

package cfg

const (
	// ConnectTCP will provide a TCP connection setting to the generated Profile.
	//
	// If multiple connections are contained in the current Config Group, a
	// 'ErrMultipleConnections' error will be returned during a build.
	ConnectTCP = cBit(0xC0)
	// ConnectTLS will provide a TLS over TCP connection setting to the generated
	// Profile.
	//
	// If multiple connections are contained in the current Config Group, a
	// 'ErrMultipleConnections' error will be returned during a build.
	//
	// This hint cannot be used as a Listener.
	ConnectTLS = cBit(0xC1)
	// ConnectUDP will provide a UCO connection setting to the generated Profile.
	//
	// If multiple connections are contained in the current Config Group, a
	// 'ErrMultipleConnections' error will be returned during a build.
	ConnectUDP = cBit(0xC2)
	// ConnectICMP will provide a ICMP connection setting to the generated Profile.
	//
	// If multiple connections are contained in the current Config Group, a
	// 'ErrMultipleConnections' error will be returned during a build.
	ConnectICMP = cBit(0xC3)
	// ConnectPipe will provide a Pipe connection setting to the generated Profile.
	//
	// If multiple connections are contained in the current Config Group, a
	// 'ErrMultipleConnections' error will be returned during a build.
	ConnectPipe = cBit(0xC4)
	// ConnectTLSNoVerify will provide a TLS over TCP connection setting to the
	// generated Profile.
	//
	// If multiple connections are contained in the current Config Group, a
	// 'ErrMultipleConnections' error will be returned during a build.
	//
	// This hint cannot be used as a Listener.
	ConnectTLSNoVerify = cBit(0xC5)
)

const (
	valIP      = cBit(0xB0)
	valWC2     = cBit(0xB1)
	valTLSx    = cBit(0xB2)
	valMuTLS   = cBit(0xB3)
	valTLSxCA  = cBit(0xB4)
	valTLSCert = cBit(0xB5)
)

// ConnectIP will provide an IP connection setting to the generated Profile with
// the specified protocol number.
//
// If multiple connections are contained in the current Config Group, a
// 'ErrMultipleConnections' error will be returned during a build.
func ConnectIP(p uint) Setting {
	return cBytes{byte(valIP), byte(p)}
}

// ConnectTLSEx will provide a TLS connection setting to the generated Profile
// with the specified TLS minimum version specified. Using the version value '0'
// will use the system default (same as the ConnectTLS option).
//
// If multiple connections are contained in the current Config Group, a
// 'ErrMultipleConnections' error will be returned during a build.
//
// This hint cannot be used as a Listener.
func ConnectTLSEx(ver uint16) Setting {
	return cBytes{byte(valTLSx), byte(ver & 0xFF)}
}

// ConnectTLSExCA will provide a TLS connection setting to the generated Profile
// with the specified TLS minimum version and will use the specified PEM bytes
// as the Root CA to trust when connecting.
//
// Using the version value '0' will use the system default (same as the ConnectTLS
// option). Empty PEM blocks will default to system root CAs.
//
// If multiple connections are contained in the current Config Group, a
// 'ErrMultipleConnections' error will be returned during a build.
//
// This hint cannot be used as a Listener.
func ConnectTLSExCA(ver uint16, ca []byte) Setting {
	a := len(ca)
	if a > 0xFFFF {
		a = 0xFFFF
	}
	c := make(cBytes, 4+a)
	_ = c[3+a]
	c[0], c[1] = byte(valTLSxCA), byte(ver&0xFF)
	c[2], c[3] = byte(a>>8), byte(a)
	copy(c[4:], ca[:a])
	return c
}

// ConnectTLSCerts will provide a TLS connection setting to the generated Profile
// with the specified TLS config that will allow for a Listener to use the
// specified PEM and Private Key data in PEM format for listening.
//
// This will also work as a Connector and can use the specified certificate for
// TLS authentication.
//
// Using the version value '0' will use the system default (same as the ConnectTLS
// option). Empty PEM blocks will render and error on build.
//
// If multiple connections are contained in the current Config Group, a
// 'ErrMultipleConnections' error will be returned during a build.
func ConnectTLSCerts(ver uint16, pem, key []byte) Setting {
	p, k := len(pem), len(key)
	if p > 0xFFFF {
		p = 0xFFFF
	}
	if k > 0xFFFF {
		k = 0xFFFF
	}
	c := make(cBytes, 6+p+k)
	_ = c[5+p+k]
	c[0], c[1] = byte(valTLSCert), byte(ver&0xFF)
	c[2], c[3] = byte(p>>8), byte(p)
	c[4], c[5] = byte(k>>8), byte(k)
	n := copy(c[6:], pem[:p]) + 6
	copy(c[n:], key[:p])
	return c
}

// ConnectMuTLS will provide a TLS connection setting to the generated Profile
// with the specified TLS config that will allow for a complete mTLS setup.
//
// This can be used for Listeners and Connectors, but the CA PEM data provided
// MUST be able to validate the client certificates, otherwise connections will
// fail.
//
// Using the version value '0' will use the system default (same as the ConnectTLS
// option). Empty PEM blocks will render and error on build.
//
// If multiple connections are contained in the current Config Group, a
// 'ErrMultipleConnections' error will be returned during a build.
func ConnectMuTLS(ver uint16, ca, pem, key []byte) Setting {
	a, p, k := len(ca), len(pem), len(key)
	if a > 0xFFFF {
		a = 0xFFFF
	}
	if p > 0xFFFF {
		p = 0xFFFF
	}
	if k > 0xFFFF {
		k = 0xFFFF
	}
	c := make(cBytes, 8+a+p+k)
	_ = c[7+a+p+k]
	c[0], c[1] = byte(valMuTLS), byte(ver&0xFF)
	c[2], c[3] = byte(a>>8), byte(a)
	c[4], c[5] = byte(p>>8), byte(p)
	c[6], c[7] = byte(k>>8), byte(k)
	n := copy(c[8:], ca[:a]) + 8
	n += copy(c[n:], pem[:p])
	copy(c[n:], key)
	return c
}

// ConnectWC2 will provide a WebC2 connection setting to the generated Profile
// with the specified User-Agent, URL and Host Matcher strings (strings can be empty).
//
// If multiple connections are contained in the current Config Group, a
// 'ErrMultipleConnections' error will be returned during a build.
//
// This hint cannot be used as a Listener.
func ConnectWC2(url, host, agent string, headers map[string]string) Setting {
	if len(url) > 0xFFFF {
		url = url[:0xFFFF]
	}
	if len(host) > 0xFFFF {
		host = host[:0xFFFF]
	}
	if len(agent) > 0xFFFF {
		agent = agent[:0xFFFF]
	}
	c := make(cBytes, len(url)+len(host)+len(agent)+8)
	_ = c[len(url)+len(host)+len(agent)+7]
	c[0] = byte(valWC2)
	c[1], c[2] = byte(len(url)>>8), byte(len(url))
	c[3], c[4] = byte(len(host)>>8), byte(len(host))
	c[5], c[6] = byte(len(agent)>>8), byte(len(agent))
	c[7] = byte(len(headers))
	n := copy(c[8:], url) + 8
	n += copy(c[n:], host)
	if copy(c[n:], agent); len(headers) == 0 {
		return c
	}
	i := 0
	for k, v := range headers {
		if i >= 0xFF {
			break
		}
		if len(k) > 0xFF {
			k = k[:0xFF]
		}
		if len(v) > 0xFF {
			v = v[:0xFF]
		}
		c = append(c, byte(len(k)), byte(len(v)))
		c = append(c, []byte(k)...)
		c = append(c, []byte(v)...)
		i++
	}
	return c
}
