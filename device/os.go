// Copyright (C) 2020 - 2022 iDigitalFlame
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

// Package device contains many function that provide access to Operating System
// functions and resources. Many of these are OS agnostic and might not work
// as intended on some systems.
//
package device

import (
	"os"
	"sync"
	"syscall"
	"time"

	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/device/arch"
	"github.com/iDigitalFlame/xmt/util"
)

// Arch represents the current device Architecture type.
const Arch = arch.Current

const (
	// Windows represents the Windows family of Operating Systems.
	Windows OSType = 0x0
	// Linux represents the Linux family of Operating Systems
	Linux OSType = 0x1
	// Unix represents the Unix/BSD family of Operating Systems
	Unix OSType = 0x2
	// Mac represents the macOS family of Operating Systems
	Mac OSType = 0x3
	// IOS represents the iOS family of Operating Systems
	// Technically is Mac, but deserves its own type for any special actions.
	IOS OSType = 0x4
	// Android represents the Android family of Operating Systems
	// Technically is Linux, but deserves its own type for any special actions.
	Android OSType = 0x5
	// Plan9 represents the Plan9 family of Operating Systems
	Plan9 OSType = 0x6
	// Unsupported represents a device type that does not have direct support
	// any may not work properly.
	Unsupported OSType = 0xF
)

var builders = sync.Pool{
	New: func() any {
		return new(util.Builder)
	},
}

// OSType is a numerical representation of the device Operating System type.
type OSType uint8

// Login is a struct that represents a current user Session on the device.
type Login struct {
	_         [0]func()
	User      string
	Host      string
	Login     time.Time
	LastInput time.Time
	ID        uint32
	From      Address
	Status    uint8
}

func init() {
	t := os.TempDir()
	syscall.Setenv("tmp", t)
	syscall.Setenv("temp", t)
}

// Expand attempts to determine environment variables from the current session
// and translate them from the supplied string.
//
// This function supports both Windows (%var%) and *nix ($var or ${var})
// variable substitutions.
func Expand(s string) string {
	if len(s) == 0 {
		return s
	}
	if len(s) >= 2 && s[0] == '~' && s[1] == '/' && len(home) > 0 {
		// Account for shell expansion. (Except JS/WASM)
		s = home + s[1:]
	}
	var (
		l  = -1
		b  = builders.Get().(*util.Builder)
		c  byte
		v  string
		ok bool
	)
	for i := range s {
		switch {
		case s[i] == '$':
			if c > 0 {
				if c == '{' {
					b.WriteString(s[l-1 : i])
				} else {
					b.WriteString(s[l:i])
				}
			}
			c, l = s[i], i
		case s[i] == '%' && c == '%' && i != l:
			if v, ok = syscall.Getenv(s[l+1 : i]); ok {
				b.WriteString(v)
			} else {
				b.WriteString(s[l:i])
			}
			c, l = 0, 0
		case s[i] == '%':
			c, l = s[i], i
		case s[i] == '}' && c == '{':
			if v, ok = syscall.Getenv(s[l+1 : i]); ok {
				b.WriteString(v)
			} else {
				b.WriteString(s[l-1 : i])
			}
			c, l = 0, 0
		case s[i] == '{' && i > 0 && c == '$':
			c, l = s[i], i
		case s[i] >= 'a' && s[i] <= 'z':
			fallthrough
		case s[i] >= 'A' && s[i] <= 'Z':
			fallthrough
		case s[i] == '_':
			if c == 0 {
				b.WriteByte(s[i])
			}
		case s[i] >= '0' && s[i] <= '9':
			if c > 0 && i > l && i-l == 1 {
				c, l = 0, 0
			}
			if c == 0 {
				b.WriteByte(s[i])
			}
		default:
			if c == '$' {
				if v, ok = syscall.Getenv(s[l+1 : i]); ok {
					b.WriteString(v)
				} else {
					b.WriteString(s[l:i])
				}
				c, l = 0, 0
			} else if c > 0 {
				if c == '{' {
					b.WriteString(s[l-1 : i])
				} else {
					b.WriteString(s[l:i])
				}
				c, l = 0, 0
			}
			b.WriteByte(s[i])
		}
	}
	if l == -1 {
		b.Reset()
		builders.Put(b)
		return s
	}
	if l < len(s) && c > 0 {
		if c == '$' {
			if v, ok = syscall.Getenv(s[l+1:]); ok {
				b.WriteString(v)
			} else {
				b.WriteString(s[l:])
			}
		} else if c == '{' {
			b.WriteString(s[l-1:])
		} else {
			b.WriteString(s[l:])
		}
	}
	v = b.Output()
	b.Reset()
	builders.Put(b)
	return v
}

// MarshalStream writes the data of this c to the supplied Writer.
func (l Login) MarshalStream(w data.Writer) error {
	if err := w.WriteUint32(l.ID); err != nil {
		return err
	}
	if err := w.WriteUint8(l.Status); err != nil {
		return err
	}
	if err := w.WriteInt64(l.Login.Unix()); err != nil {
		return err
	}
	if err := w.WriteInt64(l.LastInput.Unix()); err != nil {
		return err
	}
	if err := l.From.MarshalStream(w); err != nil {
		return err
	}
	if err := w.WriteString(l.User); err != nil {
		return err
	}
	return w.WriteString(l.Host)
}

// UnmarshalStream reads the data of this Login from the supplied Reader.
func (l *Login) UnmarshalStream(r data.Reader) error {
	if err := r.ReadUint32(&l.ID); err != nil {
		return err
	}
	if err := r.ReadUint8(&l.Status); err != nil {
		return err
	}
	v, err := r.Int64()
	if err != nil {
		return err
	}
	i, err := r.Int64()
	if err != nil {
		return err
	}
	l.Login, l.LastInput = time.Unix(v, 0), time.Unix(i, 0)
	if err = l.From.UnmarshalStream(r); err != nil {
		return err
	}
	if err = r.ReadString(&l.User); err != nil {
		return err
	}
	return r.ReadString(&l.Host)
}
