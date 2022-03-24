//go:build !windows && !js && !wasm

package device

import (
	"io"
	"os"
	"strconv"
)

func parseLine(e string, f *os.File, w io.Writer) error {
	var d, s int
	for ; d < len(e) && e[d] != '-'; d++ {
	}
	for s = d + 1; s < len(e) && e[s] != ' '; s++ {
	}
	if d >= len(e) || s-d < 4 {
		return nil
	}
	if len(e) < s+21 || e[s+1] != 'r' {
		return nil
	}
	x := s + 6
	for ; x < len(e) && e[x] != ' '; x++ {
	}
	for x++; x < len(e) && e[x] != ' '; x++ {
	}
	if e[x+1] == '0' && (e[x+2] == ' ' || e[x+2] == 9 || e[x+2] == '\t') {
		return nil
	}
	v, err := strconv.ParseUint(e[0:d], 16, 64)
	if err != nil {
		return err
	}
	g, err := strconv.ParseUint(e[d+1:s], 16, 64)
	if err != nil {
		return err
	}
	var b [4096]byte
	for i, k, q := v, uint64(0), 0; i < g; {
		if k = g - i; k > 4096 {
			k = 4096
		}
		if q, err = f.ReadAt(b[:k], int64(i)); err != nil && err != io.EOF {
			break
		}
		if _, err = w.Write(b[:q]); err != nil {
			break
		}
		if i += uint64(q); q == 0 || i >= g {
			break
		}
	}
	return err
}
