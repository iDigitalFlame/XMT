package devtools

import (
	"net/http"
	"net/url"
	"sync"
)

var (
	proxy     func(reqURL *url.URL) (*url.URL, error)
	proxyOnce sync.Once
)

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
// BUG(dij): I don't have handeling of "<local>" (Windows specific) bypass
//           rules in place. I would have to re-implement "httpproxy" code
//           and might not be worth it.
func Proxy(r *http.Request) (*url.URL, error) {
	proxyOnce.Do(func() {
		if p := proxyInit(); p != nil {
			proxy = p.ProxyFunc()
		}
	})
	if proxy == nil {
		return r.URL, nil
	}
	return proxy(r.URL)
}
