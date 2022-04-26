//go:build !windows

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
	net.Listener
}

// Dial connects to the specified Pipe path. This function will return a 'net.Conn'
// instance or any errors that may occur during the connection attempt.
//
// Pipe names are in the form of "/<path>".
//
// This function blocksindefinitely. Use the DialTimeout or DialContext to specify
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
			p |= 0400
		case i < 3 && (c == 'w' || c == 'W'):
			p |= 0200
		case i < 3 && (c == 'x' || c == 'X'):
			p |= 0100
		case i >= 3 && i < 6 && (c == 'r' || c == 'R'):
			p |= 0040
		case i >= 3 && i < 6 && (c == 'w' || c == 'W'):
			p |= 0020
		case i >= 3 && i < 6 && (c == 'x' || c == 'X'):
			p |= 0010
		case i >= 6 && (c == 'r' || c == 'R'):
			p |= 0004
		case i >= 6 && (c == 'w' || c == 'W'):
			p |= 0002
		case i >= 6 && (c == 'x' || c == 'X'):
			p |= 0001
		case c == '-' || c == ' ':
		case c != 'r' && c != 'R' && c != 'x' && c != 'X' && c != 'w' && c != 'W':
			if xerr.Concat {
				return 0, xerr.Sub(`invalid permission "`+s+`"`, 0x3D)
			}
			return 0, xerr.Sub("invalid permissions", 0x3D)
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
		if xerr.Concat {
			return 0, -1, -1, xerr.Sub(`invalid permission "`+s+`" size `+strconv.Itoa(len(v)), 0x3E)
		}
		return 0, -1, -1, xerr.Sub("invalid permission size", 0x3E)
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
// permissions of the listeneing Pipe.
func ListenPerms(path, perms string) (net.Listener, error) {
	return ListenPermsContext(context.Background(), path, perms)
}

// DialTimeout connects to the specified Pipe path. This function will return a
// net.Conn instance or any errors that may occur during the connection attempt.
//
// Pipe names are in the form of "/<path>".
//
// This function blocks for the specified amount of time and will return 'Errtimeout'
// if the timeout is reached.
func DialTimeout(path string, t time.Duration) (net.Conn, error) {
	return net.DialTimeout(com.NameUnix, path, t)
}

// DialContext connects to the specified Pipe path. This function will return a
// net.Conn instance or any errors that may occur during the connection attempt.
//
// Pipe names are in the form of "/<path>".
//
// This function blocks until the supplied context is cancled and will return the
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
// permissions of the listeneing Pipe.
//
// The provided Context can be used to cancel the Listener.
func ListenPermsContext(x context.Context, path, perms string) (net.Listener, error) {
	l, err := com.ListenConfig.Listen(x, com.NameUnix, path)
	if err != nil {
		return nil, err
	}
	if len(perms) == 0 {
		return &listener{l}, err
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
	return &listener{l}, nil
}
