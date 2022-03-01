//go:build !windows && !js && !wasm && crypt
// +build !windows,!js,!wasm,crypt

package device

import (
	"io"
	"os"
	"strings"

	"github.com/iDigitalFlame/xmt/util/crypt"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

var (
	// Shell is the default machine specific command shell.
	Shell = crypt.Get(34) // /bin/sh
	// ShellArgs is the default machine specific command shell arguments to run
	// commands.
	ShellArgs = crypt.Get(35) // -c
	// PowerShell is the path to the PowerShell binary, which is based on the
	// underlying OS type.
	PowerShell = crypt.Get(36) // pwsh
)

// Mounts attempts to get the mount points on the local device.
//
// On Windows devices, this is the drive letters avaliable, otherwise on nix*
// systems, this will be the mount points on the system.
//
// The return result (if no errors occurred) will be a string list of all the
// mount points (or Windows drive letters).
func Mounts() ([]string, error) {
	f, err := os.Open(crypt.Get(210)) // /proc/self/mounts
	if err != nil {
		if f, err = os.Open(crypt.Get(211)); err != nil { // /etc/mtab
			return nil, xerr.Wrap("cannot find mounts", err)
		}
	}
	b, err := io.ReadAll(f)
	if f.Close(); err != nil {
		return nil, err
	}
	var (
		e = strings.Split(string(b), "\n")
		m = make([]string, 0, len(e))
	)
	for _, v := range e {
		x, l := 0, 0
		for s := 0; s < 2; s++ {
			for l = x; x < len(v)-1 && v[x] != ' ' && v[x] != 9; x++ {
			}
			if x < len(v)-1 && s == 0 {
				x++
			}
		}
		if l == x {
			continue
		}
		m = append(m, v[l:x])
	}
	return m, nil
}
