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

package wc2

import (
	"net/http"
	"testing"

	"github.com/iDigitalFlame/xmt/util/text"
)

func TestTargets(t *testing.T) {
	var i Target
	i.Host = text.Matcher("%10fs-Test123")
	i.Agent = text.Matcher("Test123-%10fs")
	i.Header("Content-Type", text.String("text/html"))
	i.Header("Content-Host", text.Matcher("host-%10fs-%h"))
	var (
		r = i.Rule()
		h http.Request
	)
	h.Host = i.Host.String()
	h.Header = http.Header{
		"User-Agent":   []string{i.Agent.String()},
		"Content-Type": []string{"text/html"},
		"Content-Host": []string{text.Matcher("host-%10fs-%h").String()},
	}
	if !r.match(&h) {
		t.Fatalf("Rule does not match generated http.Request!")
	}
}
