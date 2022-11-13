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

package text

import (
	"strconv"
	"testing"
)

func TestMatcher(t *testing.T) {
	for i := 0; i < 32; i++ {
		m := Matcher("test1-%d-%h-%5n-%5c-%5u-%5l-%5s-%d-%h-test1-" + strconv.Itoa(i))
		if !m.Match().MatchString(m.String()) {
			t.Fatalf(`TestMatcher(): Matcher "%s" did not match it's generated Regexp!`, m)
		}
	}
	for i := 0; i < 32; i++ {
		m := Matcher("test1-%5l-%5d-%5fl-%5fu-%5fn-test2-" + strconv.Itoa(i))
		if !m.Match().MatchString(m.String()) {
			t.Fatalf(`TestMatcher(): Matcher "%s" did not match it's generated Regexp!`, m)
		}
	}
	for i := 0; i < 32; i++ {
		m := Matcher("test1-%5c-%5fc-%5fh-%5fd-%5n-%5fn-test3-" + strconv.Itoa(i))
		if !m.Match().MatchString(m.String()) {
			t.Fatalf(`TestMatcher(): Matcher "%s" did not match it's generated Regexp!`, m)
		}
	}
}
func TestMatcherAny(t *testing.T) {
	for i := 0; i < 32; i++ {
		m := Matcher("%s-test1-%5h-%5n-%5c-%5u-%5l-%5s-%d-%h-test1-" + strconv.Itoa(i))
		if !MatchAny.MatchString(m.String()) {
			t.Fatalf(`TestMatcherAny(): Matcher "%s" did not match MatchAny!`, m)
		}
	}
}
func TestMatcherNone(t *testing.T) {
	for i := 0; i < 32; i++ {
		m := Matcher("%s-test1-%5h-%5n-%5c-%5u-%5l-%5s-%d-%h-test1-" + strconv.Itoa(i))
		if MatchNone.MatchString(m.String()) {
			t.Fatalf(`TestMatcherNone(): Matcher "%s" matched MatchNone!`, m)
		}
	}
}
func TestMatcherInverse(t *testing.T) {
	for i := 0; i < 32; i++ {
		m := Matcher("%s-test1-%5h-%5n-%5c-%5u-%5l-%5s-%d-%h-test1-" + strconv.Itoa(i))
		if m.MatchEx(false).MatchString(m.String()) {
			t.Fatalf(`TestMatcherInverse(): Matcher "%s" matched it's generated inverse Regexp!`, m)
		}
	}
}
