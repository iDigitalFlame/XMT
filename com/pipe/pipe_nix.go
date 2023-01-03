//go:build !windows

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

package pipe

import (
	"context"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

var dialer = new(net.Dialer)

type listener struct {
	_ [0]func()
	p string
	net.Listener
}

func (l *listener) Close() error {
	if err := l.Listener.Close(); err != nil {
		return err
	}
	return os.Remove(l.p)
}

// Dial connects to the specified Pipe path. This function will return a 'net.Conn'
// instance or any errors that may occur during the connection attempt.
//
// Pipe names are in the form of "/<path>".
//
// This function blocks indefinitely. Use the DialTimeout or DialContext to specify
// a control method.
func Dial(path string) (net.Conn, error) {
	return dialer.Dial(com.NameUnix, path)
}

// Listen returns a 'net.Listener' that will listen for new connections on the
// Named Pipe path specified or any errors that may occur during listener
// creation.
//
// Pipe names are in the form of "/<path>".
func Listen(path string) (net.Listener, error) {
	return ListenPermsContext(context.Background(), path, "")
}
func stringToDec(s string) (os.FileMode, error) {
	if v, err := strconv.ParseInt(s, 8, 32); err == nil {
		return os.FileMode(v), nil
	}
	var p os.FileMode
	for i, c := range s {
		switch {
		case i < 3 && (c == 'u' || c == 'U'):
			p |= os.ModeSetuid
		case i < 3 && (c == 'g' || c == 'G'):
			p |= os.ModeSetgid
		case i == 0 && (c == 't' || c == 'T'):
			p |= os.ModeSticky
		case i < 3 && (c == 'r' || c == 'R'):
			p |= 0o400
		case i < 3 && (c == 'w' || c == 'W'):
			p |= 0o200
		case i < 3 && (c == 'x' || c == 'X'):
			p |= 0o100
		case i >= 3 && i < 6 && (c == 'r' || c == 'R'):
			p |= 0o040
		case i >= 3 && i < 6 && (c == 'w' || c == 'W'):
			p |= 0o020
		case i >= 3 && i < 6 && (c == 'x' || c == 'X'):
			p |= 0o010
		case i >= 6 && (c == 'r' || c == 'R'):
			p |= 0o004
		case i >= 6 && (c == 'w' || c == 'W'):
			p |= 0o002
		case i >= 6 && (c == 'x' || c == 'X'):
			p |= 0o001
		case c == '-' || c == ' ':
		case c != 'r' && c != 'R' && c != 'x' && c != 'X' && c != 'w' && c != 'W':
			if xerr.ExtendedInfo {
				return 0, xerr.Sub(`invalid permission "`+s+`"`, 0x2E)
			}
			return 0, xerr.Sub("invalid permissions", 0x2E)
		}
	}
	return p, nil
}
func getPerms(s string) (os.FileMode, int, int, error) {
	if i := strings.IndexByte(s, 59); i == -1 {
		p, err := stringToDec(s)
		return p, -1, -1, err
	}
	v := strings.Split(s, ";")
	if len(v) > 3 {
		if xerr.ExtendedInfo {
			return 0, -1, -1, xerr.Sub(`invalid permission "`+s+`" size `+strconv.FormatUint(uint64(len(v)), 10), 0x2F)
		}
		return 0, -1, -1, xerr.Sub("invalid permission size", 0x2F)
	}
	var (
		u, g   = -1, -1
		p, err = stringToDec(v[0])
	)
	if err != nil {
		return 0, -1, -1, err
	}
	if len(v) == 3 && len(v[2]) > 0 {
		if g, err = strconv.Atoi(v[2]); err != nil {
			return 0, -1, -1, xerr.Wrap("invalid GID", err)
		}
	}
	if len(v) >= 2 && len(v[1]) > 0 {
		if u, err = strconv.Atoi(v[1]); err != nil {
			return 0, -1, -1, xerr.Wrap("invalid UID", err)
		}
	}
	return p, u, g, nil
}

// ListenPerms returns a Listener that will listen for new connections on the
// Named Pipe path specified or any errors that may occur during listener
// creation.
//
// Pipe names are in the form of "/<path>".
//
// This function allows for specifying a Linux permissions string used to set the
// permissions of the listening Pipe.
func ListenPerms(path, perms string) (net.Listener, error) {
	return ListenPermsContext(context.Background(), path, perms)
}

// DialTimeout connects to the specified Pipe path. This function will return a
// net.Conn instance or any errors that may occur during the connection attempt.
//
// Pipe names are in the form of "/<path>".
//
// This function blocks for the specified amount of time and will return 'ErrTimeout'
// if the timeout is reached.
func DialTimeout(path string, t time.Duration) (net.Conn, error) {
	return net.DialTimeout(com.NameUnix, path, t)
}

// DialContext connects to the specified Pipe path. This function will return a
// net.Conn instance or any errors that may occur during the connection attempt.
//
// Pipe names are in the form of "/<path>".
//
// This function blocks until the supplied context is canceled and will return the
// context's Err() if the cancel occurs before the connection.
func DialContext(x context.Context, path string) (net.Conn, error) {
	return dialer.DialContext(x, com.NameUnix, path)
}

// ListenContext returns a 'net.Listener' that will listen for new connections
// on the Named Pipe path specified or any errors that may occur during listener
// creation.
//
// Pipe names are in the form of "/<path>".
//
// The provided Context can be used to cancel the Listener.
func ListenContext(x context.Context, path string) (net.Listener, error) {
	return ListenPermsContext(x, path, "")
}

// ListenPermsContext returns a Listener that will listen for new connections on
// the Named Pipe path specified or any errors that may occur during listener
// creation.
//
// Pipe names are in the form of "/".
//
// This function allows for specifying a Linux permissions string used to set the
// permissions of the listening Pipe.
//
// The provided Context can be used to cancel the Listener.
func ListenPermsContext(x context.Context, path, perms string) (net.Listener, error) {
	l, err := com.ListenConfig.Listen(x, com.NameUnix, path)
	if err != nil {
		return nil, err
	}
	if len(perms) == 0 {
		return &listener{Listener: l, p: path}, err
	}
	m, u, g, err := getPerms(perms)
	if err != nil {
		l.Close()
		return nil, err
	}
	if m > 0 {
		if err := os.Chmod(path, m); err != nil {
			l.Close()
			return nil, err
		}
	}
	if err := os.Chown(path, u, g); err != nil {
		l.Close()
		return nil, err
	}
	return &listener{Listener: l, p: path}, nil
}
