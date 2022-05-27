//go:build !windows && !crypt

package cmd

import (
	"os"
	"sort"
	"strconv"
)

// Processes attempts to gather the current running Processes and returns them
// as a slice of ProcessInfo structs, otherwise any errors are returned.
func Processes() ([]ProcessInfo, error) {
	l, err := os.ReadDir("/proc/")
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
		if b, err = os.ReadFile("/proc/" + n + "/status"); err != nil {
			continue
		}
		u := getProcUser("/proc/" + n)
		if n, p = readProcStats(b); len(n) == 0 {
			continue
		}
		r = append(r, ProcessInfo{Name: n, User: u, PID: uint32(v), PPID: p})
	}
	sort.Sort(r)
	return r, nil
}
