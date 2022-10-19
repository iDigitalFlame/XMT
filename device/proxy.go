// Copyright (C) 2020 - 2022 iDigitalFlame
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

package device

import (
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"syscall"
)

var proxySync struct {
	f func(reqURL *url.URL) (*url.URL, error)
	sync.Once
}

type config struct {
	HTTPProxy  string
	HTTPSProxy string
	NoProxy    string
	CGI        bool

	http   *url.URL
	https  *url.URL
	ip     []matcher
	domain []matcher
}
type matchIP struct {
	i net.IP
	p string
}
type matchAll struct{}
type matchCIDR struct {
	_ [0]func()
	*net.IPNet
}
type matcher interface {
	match(string, string, net.IP) bool
}
type matchDomain struct {
	_ [0]func()
	h string
	p string
	x bool
}

func (c *config) parse() {
	if u, err := parse(c.HTTPProxy); err == nil {
		c.http = u
	}
	if u, err := parse(c.HTTPSProxy); err == nil {
		c.https = u
	}
	for _, v := range strings.Split(c.NoProxy, ",") {
		if v = strings.ToLower(strings.TrimSpace(v)); len(v) == 0 {
			continue
		}
		if v == "*" {
			c.ip, c.domain = []matcher{matchAll{}}, []matcher{matchAll{}}
			return
		}
		if _, r, err := net.ParseCIDR(v); err == nil {
			c.ip = append(c.ip, matchCIDR{IPNet: r})
			continue
		}
		h, p, err := net.SplitHostPort(v)
		if err == nil {
			if len(h) == 0 {
				continue
			}
			if h[0] == '[' && h[len(h)-1] == ']' {
				h = h[1 : len(h)-1]
			}
		} else {
			h = v
		}
		if i := net.ParseIP(h); i != nil {
			c.ip = append(c.ip, matchIP{i: i, p: p})
			continue
		}
		if len(h) == 0 {
			continue
		}
		if strings.HasPrefix(h, "*.") {
			h = h[1:]
		}
		if h[0] != '.' {
			c.domain = append(c.domain, matchDomain{h: "." + h, p: p, x: true})
		} else {
			c.domain = append(c.domain, matchDomain{h: h, p: p, x: false})
		}
	}
}
func realAddr(u *url.URL) string {
	var (
		a = u.Hostname()
		p = u.Port()
	)
	if len(p) == 0 && len(u.Scheme) >= 4 {
		if u.Scheme[0] == 'h' || u.Scheme[0] == 'H' {
			if len(u.Scheme) == 5 && (u.Scheme[4] == 's' || u.Scheme[4] == 'S') {
				p = "443"
			} else {
				p = "80"
			}
		} else if u.Scheme[0] == 's' || u.Scheme[0] == 'S' {
			p = "1080"
		}
	}
	return net.JoinHostPort(a, p)
}
func (c *config) usable(u string) bool {
	if len(u) == 0 {
		return true
	}
	h, p, err := net.SplitHostPort(u)
	if err != nil {
		return false
	}
	if h == "localhost" {
		return false
	}
	i := net.ParseIP(h)
	if i != nil && i.IsLoopback() {
		return false
	}
	a := strings.ToLower(strings.TrimSpace(h))
	if i != nil {
		for x := range c.ip {
			if c.ip[x].match(a, p, i) {
				return false
			}
		}
	}
	for x := range c.domain {
		if c.domain[x].match(a, p, i) {
			return false
		}
	}
	return true
}
func parse(u string) (*url.URL, error) {
	if len(u) == 0 {
		return nil, nil
	}
	v, err := url.Parse(u)
	if err != nil || (v.Scheme != "http" && v.Scheme != "https" && v.Scheme != "socks5") {
		if v, err := url.Parse("http://" + u); err == nil {
			return v, nil
		}
	}
	if err != nil {
		return nil, syscall.EINVAL
	}
	return v, nil
}

// Proxy returns the URL of the proxy to use for a given request, as
// indicated by the on-device settings.
//
// Unix/Linux/BSD devices use the environment variables
// HTTP_PROXY, HTTPS_PROXY and NO_PROXY (or the lowercase versions
// thereof). HTTPS_PROXY takes precedence over HTTP_PROXY for https
// requests.
//
// Windows devices will query the Windows API and resolve the system setting
// values.
//
// The environment values may be either a complete URL or a
// "host[:port]", in which case the "http" scheme is assumed.
// The schemes "http", "https", and "socks5" are supported.
// An error is returned if the value is a different form.
//
// A nil URL and nil error are returned if no proxy is defined in the
// environment, or a proxy should not be used for the given request,
// as defined by NO_PROXY or ProxyBypass.
//
// As a special case, if req.URL.Host is "localhost" (with or without
// a port number), then a nil URL and nil error will be returned.
//
// NOTE(dij): I don't have handling of "<local>" (Windows specific) bypass
//            rules in place. I would have to re-implement "httpproxy" code
//            and might not be worth it.
func Proxy(r *http.Request) (*url.URL, error) {
	proxySync.Do(func() {
		if p := proxyInit(); p != nil {
			proxySync.f = p.ProxyFunc()
		}
	})
	if proxySync.f == nil {
		return nil, nil
	}
	return proxySync.f(r.URL)
}
func (matchAll) match(_, _ string, _ net.IP) bool {
	return true
}
func (m matchIP) match(_, p string, i net.IP) bool {
	if m.i.Equal(i) {
		return len(m.p) == 0 || m.p == p
	}
	return false
}
func (m matchCIDR) match(_, _ string, i net.IP) bool {
	return m.Contains(i)
}
func (m matchDomain) match(h, p string, _ net.IP) bool {
	if strings.HasSuffix(h, m.h) || (m.x && h == m.h[1:]) {
		return len(m.p) == 0 || m.p == p
	}
	return false
}
func (c *config) proxyForURL(r *url.URL) (*url.URL, error) {
	if !c.usable(realAddr(r)) {
		return nil, nil
	}
	if len(r.Scheme) == 5 && (r.Scheme[4] == 's' || r.Scheme[4] == 'S') {
		return c.https, nil
	}
	if c.http != nil && c.CGI {
		return nil, syscall.EINVAL
	}
	return c.http, nil
}
func (c *config) ProxyFunc() func(*url.URL) (*url.URL, error) {
	c.parse()
	return c.proxyForURL
}
