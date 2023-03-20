//go:build !amd64 && !386 && !windows && !crypt
// +build !amd64,!386,!windows,!crypt

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

package device

import (
	"bytes"
	"os"
	"strings"

	"github.com/iDigitalFlame/xmt/data"
)

func isVirtual() bool {
	// Check for a container first
	// OpenVZ
	if _, err := os.Lstat("/proc/vz"); err == nil {
		if _, err := os.Lstat("/proc/bc"); err != nil {
			return true
		}
	}
	// Docker
	if _, err := os.Lstat("/.dockerenv"); err == nil {
		return true
	}
	if _, err := os.Lstat("/run/.containerenv"); err == nil {
		return true
	}
	// systemd-nspawn
	if _, err := os.Lstat("/run/systemd/container"); err == nil {
		return true
	}
	// WSL
	if b, _ := data.ReadFile("/proc/sys/kernel/osrelease"); len(b) > 0 {
		if x := bytes.IndexByte(b, 'M'); x > -1 && x+8 < len(b) {
			// Microsoft
			if (b[x+1] == 'I' || b[x+1] == 'i') && (b[x+2] == 'C' || b[x+2] == 'c') && (b[x+3] == 'R' || b[x+3] == 'r') && (b[x+4] == 'O' || b[x+4] == 'o') {
				return true
			}
		}
		if x := bytes.IndexByte(b, 'W'); x > -1 && x+2 < len(b) {
			// WSL
			if (b[x+1] == 'S' || b[x+1] == 's') && (b[x+2] == 'L' || b[x+2] == 'l') {
				return true
			}
		}
	}
	// PROOT
	var n string
	for _, v := range data.ReadSplit("/proc/self/status", "\n") {
		if len(v) == 0 || (v[0] != 'T' && v[0] != 't') {
			continue
		}
		// TracerPid
		if (v[1] != 'R' && v[1] != 'r') || (v[3] != 'C' && v[3] != 'c') || (v[6] != 'P' && v[6] != 'p') || (v[8] != 'D' && v[8] != 'd') {
			continue
		}
		x := strings.IndexByte(v, ':')
		if x < 8 {
			continue
		}
		n = strings.TrimSpace(v[x+1:])
		break
	}
	if len(n) > 0 {
		if p, _ := data.ReadFile("/proc/" + n + "/comm"); len(p) > 0 {
			// proot
			if (p[0] == 'P' || p[0] == 'p') && (p[1] == 'R' || p[1] == 'r') && (p[2] == 'O' || p[2] == 'o') && (p[4] == 'T' || p[8] == 't') {
				return true
			}
		}
	}
	if os.Getpid() == 1 {
		if k, ok := os.LookupEnv("CONTAINER"); ok && len(k) > 0 {
			return true
		}
	}
	if b, _ := data.ReadFile("/proc/1/environ"); len(b) > 0 {
		for s, e, n := 0, 0, 0; e < len(b); e++ {
			if b[e] != 0 {
				continue
			}
			if e-s > 9 {
				// CONTAINER=
				if b[s] == 'C' && b[s+1] == 'O' && b[s+2] == 'N' && b[s+7] == 'N' && b[s+9] == 'R' && b[s+10] == '=' {
					return true
				}
			}
			s, n = e+1, n+1
		}
	}
	// User Mode Linux
	for _, v := range data.ReadSplit("/proc/cpuinfo", "\n") {
		if len(v) == 0 || (v[0] != 'V' && v[0] != 'v') {
			continue
		}
		// vendor_id
		if (v[1] != 'E' && v[1] != 'e') || (v[3] != 'D' && v[3] != 'd') || v[6] != '_' || (v[7] != 'I' && v[7] != 'i') {
			continue
		}
		x := strings.IndexByte(v, ':')
		if x < 8 {
			continue
		}
		if n := strings.TrimSpace(v[x+1:]); len(n) >= 15 && (n[0] == 'U' || n[0] == 'u') && (n[5] == 'M' || n[5] == 'm') && (n[8] == 'E' || n[8] == 'e') && (n[10] == 'L' || n[10] == 'l') {
			return true
		}
	}
	// Check DMI
	return checkVendorFile("/sys/class/dmi/id/sys_vendor") ||
		checkVendorFile("/sys/class/dmi/id/board_vendor") ||
		checkVendorFile("/sys/class/dmi/id/bios_vendor") ||
		checkVendorFile("/sys/class/dmi/id/product_version")
}
func checkVendorFile(s string) bool {
	if b, _ := data.ReadFile(s); len(b) > 0 {
		return isKnownVendor(bytes.TrimSpace(b))
	}
	return false
}
