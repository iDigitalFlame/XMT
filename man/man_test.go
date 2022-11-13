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

package man

import (
	"bytes"
	"testing"

	"github.com/iDigitalFlame/xmt/cmd/filter"
	"github.com/iDigitalFlame/xmt/data/crypto"
)

func TestRawParse(t *testing.T) {
	if _, err := rawParse("google.com"); err != nil {
		t.Fatalf(`TestRawParse(): Raw URL Parse "google.com" failed with error: %s!`, err.Error())
	}
	if _, err := rawParse("https://google.com"); err != nil {
		t.Fatalf(`TestRawParse(): Raw URL Parse "https://google.com" failed with error: %s!`, err.Error())
	}
	if _, err := rawParse("/google.com"); err != nil {
		t.Fatalf(`TestRawParse(): Raw URL Parse "/google.com" failed with error: %s!`, err.Error())
	}
	if _, err := rawParse("\\\\google.com"); err == nil {
		t.Fatalf(`TestRawParse(): Raw URL Parse "\\google.com" should have failed!`)
	}
	if _, err := rawParse("\\google.com"); err == nil {
		t.Fatalf(`TestRawParse(): Raw URL Parse "\google.com" should have failed!`)
	}
	if _, err := rawParse("derp:google.com"); err == nil {
		t.Fatalf(`TestRawParse(): Raw URL Parse "\google.com" should have failed!`)
	}
}
func TestParseHeaders(t *testing.T) {
	// DLL parsers can be 1 or 2 dependent on the build constant 'cmd.LoaderEnabled'
	// So we test them separately.
	if r := ParseDownloadHeader(map[string][]string{"Content-Type": {"app/dll"}}); r != 1 && r != 2 {
		t.Fatalf(`TestParseHeaders(): ParseDownloadHeader "app/dll" returned "%d", expected 1 or 2!`, r)
	}
	if r := ParseDownloadHeader(map[string][]string{"Content-Type": {"x-application/dynamic"}}); r != 1 && r != 2 {
		t.Fatalf(`TestParseHeaders(): ParseDownloadHeader "x-application/dynamic" returned "%d", expected 1 or 2!`, r)
	}
	if r := ParseDownloadHeader(map[string][]string{"Content-Type": {"x-application/derp"}}); r != 1 && r != 2 {
		t.Fatalf(`TestParseHeaders(): ParseDownloadHeader "x-application/derp" returned "%d", expected 1 or 2!`, r)
	}
	v := [...]struct {
		Type   string
		Result uint8
	}{
		{"abcdef/asm", 2},
		{"derp123/bin", 2},
		{"ahdfkjahs/shell", 3},
		{"ahdfkjahs/shell", 3},
		{"hello/cmd", 3},
		{"testing-123/xexec", 3},
		{"text/com", 3},
		{"application/pwsh", 4},
		{"text/pwn", 4},
		{"x-icon/po", 4},
		{"y-app/shellcode", 2},
		{"testing/code", 2},
		{"application/javascript", 0},
		{"text/html", 0},
		{"invalid", 0},
	}
	for i := range v {
		if r := ParseDownloadHeader(map[string][]string{"Content-Type": {v[i].Type}}); r != v[i].Result {
			t.Fatalf(`TestParseHeaders(): ParseDownloadHeader "%s" returned %d, expected %d!`, v[i].Type, r, v[i].Result)
		}
	}
}
func TestLoadSaveSentinel(t *testing.T) {
	var s Sentinel
	s.AddDownload("google.com/1", "")
	s.AddDownload("google.com/2", "agent1")
	s.AddDownload("google.com/3", "")
	s.AddDownload("google.com/4", "agent2")
	s.AddDownload("google.com/5", "")
	s.AddExecute("cmd.exe")
	s.AddExecute("explorer.exe")
	s.Include = []string{"svchost.exe", "rundll32.exe"}
	s.Elevated = filter.True
	c, err := crypto.NewAes([]byte("0123456789ABCDEF"))
	if err != nil {
		t.Fatalf("TestLoadSaveSentinel(): Generating AWS cipher failed: %s!", err)
	}
	var b bytes.Buffer
	if err = s.Write(c, &b); err != nil {
		t.Fatalf("TestLoadSaveSentinel(): Writing Sentinel failed: %s!", err)
	}
	var n Sentinel
	if err = n.Read(c, bytes.NewReader(b.Bytes())); err != nil {
		t.Fatalf("TestLoadSaveSentinel(): Reading Sentinel failed: %s!", err)
	}
	if len(s.paths) != len(n.paths) {
		t.Fatalf(`TestLoadSaveSentinel(): New Sentinel path count "%d" does not match the original count "%d"!`, len(n.paths), len(s.paths))
	}
	if s.Elevated != n.Elevated {
		t.Fatalf(`TestLoadSaveSentinel(): New Sentinel 'filter.Elevated' "%d" does not match the original 'filter.Elevated' "%d"!`, n.Elevated, s.Elevated)
	}
	if len(s.Include) != len(n.Include) || s.Include[0] != n.Include[0] || s.Include[1] != n.Include[1] {
		t.Fatalf(`TestLoadSaveSentinel(): New Sentinel 'filter.Include' "%s" does not match the original 'filter.Include' "%s"!`, n.Include, s.Include)
	}
}
