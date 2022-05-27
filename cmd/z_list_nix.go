//go:build !windows && !plan9

package cmd

import (
	"os"
	"os/user"
	"strconv"
	"syscall"
)

func getProcUser(p string) string {
	f, err := os.Stat(p)
	if err != nil {
		return ""
	}
	s, ok := f.Sys().(*syscall.Stat_t)
	if !ok {
		return ""
	}
	u, err := user.LookupId(strconv.FormatUint(uint64(s.Uid), 10))
	if err != nil {
		return ""
	}
	return u.Username
}
