//go:build !windows && crypt

package cmd

import (
	"os"
	"sort"
	"strconv"

	"github.com/iDigitalFlame/xmt/util/crypt"
)

func Processes() ([]ProcessInfo, error) {
	l, err := os.ReadDir(crypt.Get(16)) // /proc/
	if err != nil {
		return nil, err
	}
	if len(l) == 0 {
		return nil, nil
	}
	var (
		n string
		b []byte
		v uint64
		p uint32
		r = make(processList, 0, len(l)/2)
	)
	for i := range l {
		if !l[i].IsDir() {
			continue
		}
		if n = l[i].Name(); len(n) < 1 || n[0] < 48 || n[0] > 57 {
			continue
		}
		if v, err = strconv.ParseUint(n, 10, 32); err != nil {
			continue
		}
		b, err = os.ReadFile(
			crypt.Get(16) + // /proc/
				n +
				crypt.Get(17), // /status
		)
		if err != nil {
			continue
		}
		u := getProcUser(crypt.Get(16) + n) // /proc/
		if n, p = readProcStats(b); len(n) == 0 {
			continue
		}
		r = append(r, ProcessInfo{Name: n, User: u, PID: uint32(v), PPID: p})
	}
	sort.Sort(r)
	return r, nil
}
