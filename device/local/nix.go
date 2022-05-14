//go:build !windows && !js && !plan9

package local

import "golang.org/x/sys/unix"

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
