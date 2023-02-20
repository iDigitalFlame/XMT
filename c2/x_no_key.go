//go:build nokeyset
// +build nokeyset

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

package c2

import "github.com/iDigitalFlame/xmt/com"

func (*Session) keyCheckRevert() {
}
func (*Session) keyCheckSync() error {
}
func (*Session) keyNextSync() *com.Packet {
}
func (*Session) keySessionGenerate(_ *com.Packet) {
}
func (*Session) keySessionSync(_ *com.Packet) error {
	return nil
}
