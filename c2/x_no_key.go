//go:build nokeyset

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

package c2

import "github.com/iDigitalFlame/xmt/com"

func (s *Session) keyCheck()  {}
func (s *Session) keyRevert() {}
func (s *Session) doNextKeySwap() bool {
	return false
}
func (s *Session) generateSessionKey(n *com.Packet) {
	n.Write(s.key[:])
}
