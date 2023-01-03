//go:build !windows && !plan9 && !js && !darwin && !linux && !android && !crypt

// Copyright (C) 2020 - 2023 iDigitalFlame
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.
//

package local

import (
	"bytes"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func sysID() []byte {
	switch runtime.GOOS {
	case "aix":
		// AIX specific support: https://github.com/denisbrodbeck/machineid/pull/16
		if b, err := exec.Command("lsattr", "-l", "sys0", "-a", "os_uuid", "-E").CombinedOutput(); err == nil {
			if i := bytes.IndexByte(b, ' '); i > 0 {
				return b[i+1:]
			}
			return b
		}
	case "openbsd":
		// Support get hardware UUID for OpenBSD: https://github.com/denisbrodbeck/machineid/pull/14
		if b, err := exec.Command("sysctl", "-n", "hw.uuid").CombinedOutput(); err == nil {
			return b
		}
	}
	if b, err := os.ReadFile("/etc/hostid"); err == nil {
		return b
	}
	o, _ := exec.Command("kenv", "-q", "smbios.system.uuid").CombinedOutput()
	return o
}
func version() string {
	var (
		ok      bool
		b, n, v string
	)
	if m := release(); len(m) > 0 {
		b = m["ID"]
		if n, ok = m["PRETTY_NAME"]; !ok {
			n = m["NAME"]
		}
		if v, ok = m["VERSION_ID"]; !ok {
			v = m["VERSION"]
		}
	}
	if len(b) == 0 && strings.Contains(runtime.GOOS, "bsd") {
		if o, err := exec.Command("freebsd-version", "-k").CombinedOutput(); err == nil {
			b = strings.ReplaceAll(string(o), "\n", "")
		}
	}
	if len(v) == 0 {
		v = uname()
	}
	switch {
	case len(n) == 0 && len(b) == 0 && len(v) == 0:
		return "BSD"
	case len(n) == 0 && len(b) > 0 && len(v) > 0:
		return "BSD (" + v + ", " + b + ")"
	case len(n) == 0 && len(b) == 0 && len(v) > 0:
		return "BSD (" + v + ")"
	case len(n) == 0 && len(b) > 0 && len(v) == 0:
		return "BSD (" + b + ")"
	case len(n) > 0 && len(b) > 0 && len(v) > 0:
		return n + " (" + v + ", " + b + ")"
	case len(n) > 0 && len(b) == 0 && len(v) > 0:
		return n + " (" + v + ")"
	case len(n) > 0 && len(b) > 0 && len(v) == 0:
		return n + " (" + b + ")"
	}
	return "BSD"
}
