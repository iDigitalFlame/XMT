//go:build !windows && !js && !plan9 && !darwin && !ios && crypt

package local

import (
	"io"
	"io/fs"
	"os"
	"strings"

	"github.com/iDigitalFlame/xmt/util/crypt"
	"golang.org/x/sys/unix"
)

func uname() string {
	var (
		u   unix.Utsname
		err = unix.Uname(&u)
	)
	if err != nil {
		return ""
	}
	var (
		v = make([]byte, 65)
		i int
	)
	for ; i < 65; i++ {
		if u.Release[i] == 0 {
			break
		}
		v[i] = byte(u.Release[i])
	}
	return string(v[:i])
}
func release() map[string]string {
	var (
		f      = os.DirFS(crypt.Get(79))   // /etc
		e, err = fs.Glob(f, crypt.Get(80)) // *release*
	)
	if err != nil || len(e) == 0 {
		return nil
	}
	m := make(map[string]string)
	for i := range e {
		d, err := f.Open(e[i])
		if err != nil {
			continue
		}
		b, err := io.ReadAll(d)
		if d.Close(); err != nil || len(b) == 0 {
			continue
		}
		for _, v := range strings.Split(string(b), "\n") {
			x := strings.IndexByte(v, '=')
			if x < 1 || len(v)-x < 2 {
				continue
			}
			c, s := len(v)-1, x+1
			for ; c > x && v[c] == '"'; c-- {
			}
			for ; s < c && v[s] == '"'; s++ {
			}
			m[strings.ToUpper(v[:x])] = v[s : c+1]
		}
	}
	return m
}
