//go:build !windows

package cmd

import (
	"strconv"
	"strings"
)

func readProcStats(b []byte) (string, uint32) {
	var (
		n string
		p uint32
	)
	for _, e := range strings.Split(string(b), "\n") {
		if len(n) > 0 && p > 0 {
			return n, p
		}
		if len(e) > 6 && e[0] == 'N' && e[2] == 'm' && e[4] == ':' {
			for i := 5; i < len(e); i++ {
				if e[i] == ' ' || e[i] == 9 || e[i] == '\t' {
					continue
				}
				n = e[i:]
				break
			}
		}
		if len(e) > 5 && e[0] == 'P' && e[1] == 'P' && e[3] == 'd' && e[4] == ':' {
			for i := 5; i < len(e); i++ {
				if e[i] == ' ' || e[i] == 9 || e[i] == '\t' {
					continue
				}
				if v, err := strconv.ParseUint(e[i:], 10, 32); err == nil {
					p = uint32(v)
					break
				}
			}
		}
	}
	return n, p
}
