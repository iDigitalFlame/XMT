// +build !windows

package pipe

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	// PermEveryone is the Linux permission string used in sockets to allow anyone to write and read
	// to the listening socket. This can be used for socket communcation between privilege boundaries.
	// This can be applied to the ListenPerm function.
	PermEveryone = "0766"

	network = "unix"
)

var dialer = new(net.Dialer)

// Format will ensure the path for this Pipe socket fits the proper OS based pathname. Valid pathnames will be
// returned without any changes.
func Format(s string) string {
	if !filepath.IsAbs(s) {
		return fmt.Sprintf("/tmp/%s", s)
	}
	return s
}

// Dial connects to the specified Pipe path. This function will return a net.Conn instance or any errors that may
// occur during the connection attempt. Pipe names are in the form of "/<path>". This function blocks indefinitely.
// Use the DialTimeout or DialContext to specify a control method.
func Dial(path string) (net.Conn, error) {
	return dialer.Dial(network, path)
}

// Listen returns a net.Listener that will listen for new connections on the Named Pipe path specified or any
// errors that may occur during listener creation. Pipe names are in the form of "/<path>".
func Listen(path string) (net.Listener, error) {
	return net.Listen(network, path)
}
func stringToDec(s string) (os.FileMode, error) {
	if v, err := strconv.ParseInt(s, 8, 32); err == nil {
		return os.FileMode(v), nil
	}
	var p os.FileMode
	for i, c := range strings.ToLower(s) {
		switch {
		case i < 3 && c == 'u':
			p |= os.ModeSetuid
		case i < 3 && c == 'g':
			p |= os.ModeSetgid
		case i == 0 && c == 't':
			p |= os.ModeSticky
		case i < 3 && c == 'r':
			p |= 0400
		case i < 3 && c == 'w':
			p |= 0200
		case i < 3 && c == 'x':
			p |= 0100
		case i >= 3 && i < 6 && c == 'r':
			p |= 0040
		case i >= 3 && i < 6 && c == 'w':
			p |= 0020
		case i >= 3 && i < 6 && c == 'x':
			p |= 0010
		case i >= 6 && c == 'r':
			p |= 0004
		case i >= 6 && c == 'w':
			p |= 0002
		case i >= 6 && c == 'x':
			p |= 0001
		case c == '-' || c == ' ':
		case c != 'r' && c != 'x' && c != 'w':
			return 0, fmt.Errorf("invalid permission string %q", s)
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
		return 0, -1, -1, fmt.Errorf("invalid permission size %d: %q", len(v), s)
	}
	var (
		u, g   = -1, -1
		p, err = stringToDec(v[0])
	)
	if err != nil {
		return 0, -1, -1, err
	}
	if len(v) == 3 {
		if g, err = strconv.Atoi(v[2]); err != nil {
			return 0, -1, -1, fmt.Errorf("invalid GID: %q: %w", v[2], err)
		}
	}
	if len(v) == 2 {
		if u, err = strconv.Atoi(v[1]); err != nil {
			return 0, -1, -1, fmt.Errorf("invalid UID: %q: %w", v[1], err)
		}
	}
	return p, u, g, nil
}

// ListenPerms returns a net.Listener that will listen for new connections on the Named Pipe path specified or any
// errors that may occur during listener creation. Pipe names are in the form of "/<path>". This function allows
// for specifying a SDDL string used to set the permissions of the listeneing Pipe.
func ListenPerms(path, perms string) (net.Listener, error) {
	l, err := net.Listen(network, path)
	if err != nil {
		return nil, err
	}
	if len(perms) == 0 {
		return l, err
	}
	m, u, g, err := getPerms(perms)
	if err != nil {
		l.Close()
		return nil, err
	}
	if m > 0 {
		if err := os.Chmod(path, m); err != nil {
			l.Close()
			return nil, fmt.Errorf("unable to set permissions on %q: %w", path, err)
		}
	}
	if err := os.Chown(path, u, g); err != nil {
		l.Close()
		return nil, fmt.Errorf("unable to set ownership on %q: %w", path, err)
	}
	return l, nil
}

// DialTimeout connects to the specified Pipe path. This function will return a net.Conn instance or any errors that
// may occur during the connection attempt. Pipe names are in the form of "/<path>". This function blocks for the
//specified amount of time and will return 'Errtimeout' if the timeout is reached.
func DialTimeout(path string, t time.Duration) (net.Conn, error) {
	return net.DialTimeout(network, path, t)
}

// DialContext connects to the specified Pipe path. This function will return a net.Conn instance or any errors that
// may occur during the connection attempt. Pipe names are in the form of "/<path>. This function blocks until the
// supplied context is cancled and will return the context's Err() if the cancel occurs before the connection.
func DialContext(x context.Context, path string) (net.Conn, error) {
	return dialer.DialContext(x, network, path)
}
