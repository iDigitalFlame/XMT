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

package xerr

import (
	"errors"
	"io"
	"testing"
)

func TestErrorWithSub(t *testing.T) {
	if v := Sub("test error", 0xFA).Error(); v != "test error" && v != "0xFA" {
		t.Fatalf(`Error string "%s" did not match the given string value!`, v)
	}
}

func TestErrorWithWrap(t *testing.T) {
	if e := Wrap("test error", io.EOF); !errors.Is(e, io.EOF) && !errors.Is(errors.Unwrap(e), io.EOF) {
		t.Fatalf(`Wrapped error "%s" did not match the given wrapped value!`, e)
	}
}
