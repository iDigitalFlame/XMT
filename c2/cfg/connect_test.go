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

package cfg

import (
	"testing"
)

func TestConnect(t *testing.T) {
	c := Pack(
		ConnectIP(1),
		ConnectTLSEx(10),
		ConnectTLSExCA(12, []byte("derp1234")),
		ConnectTLSCerts(13, []byte("testing123"), []byte("456testing")),
		ConnectMuTLS(14, []byte("derp1234"), []byte("testing123"), []byte("456testing")),
		ConnectWC2("http://localhost", "host", "", map[string]string{"Test": "", "Content-Type": "text/html"}),
	)

	if _, err := c.Build(); err == nil {
		t.Fatalf("TestConnect(): Invalid build should have failed!")
	}

	if n := c.Len(); n != 135 {
		t.Fatalf(`TestConnect(): Len returned invalid size "%d" should ne "135"!`, n)
	}
	if c[0] != byte(valIP) || c[1] != 1 {
		t.Fatalf(`TestConnect(): Invalid byte at position "0:1"!`)
	}
	if c[2] != byte(valTLSx) || c[3] != 10 {
		t.Fatalf(`TestConnect(): Invalid byte at position "2:3"!`)
	}
	if c[4] != byte(valTLSxCA) || c[5] != 12 {
		t.Fatalf(`TestConnect(): Invalid byte at position "4:5"!`)
	}
	if c[16] != byte(valTLSCert) {
		t.Fatalf(`TestConnect(): Invalid byte at position "16"!`)
	}
	if c[42] != byte(valMuTLS) {
		t.Fatalf(`TestConnect(): Invalid byte at position "42"!`)
	}
	if c[78] != byte(valWC2) {
		t.Fatalf(`TestConnect(): Invalid byte at position "78"!`)
	}
}
