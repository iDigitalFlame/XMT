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

	"github.com/iDigitalFlame/xmt/util/text"
)

// RuleAny is a Rule that can be used to match any Request that comes in.
//
// Useful for debugging.
var RuleAny = Rule{URL: text.MatchAny, Host: text.MatchAny, Agent: text.MatchAny}

// Rule is a struct that represents a rule set used by the Web server to determine
// the difference between normal and C2 traffic.
type Rule struct {
	URL, Host, Agent Matcher
	Headers          map[string]Matcher
}

// Matcher is a utility interface that takes a single 'MatchString(string) bool'
// function and reports true if the string matches.
type Matcher interface {
	MatchString(string) bool
}

func (r Rule) match(c *http.Request) bool {
	if r.Host == nil && r.URL == nil && r.Agent == nil {
		return true
	}
	if r.Host != nil && !r.Host.MatchString(c.Host) {
		return false
	}
	if r.URL != nil && !r.URL.MatchString(c.URL.EscapedPath()) {
		return false
	}
	if r.Agent != nil && !r.Agent.MatchString(c.UserAgent()) {
		return false
	}
	if r.Headers != nil && len(r.Headers) > 0 {
		if len(c.Header) == 0 {
			return false
		}
		for k, v := range r.Headers {
			if h, ok := c.Header[k]; !ok || len(h) == 0 || !v.MatchString(h[0]) {
				return false
			}
		}
	}
	return true
}

// Header adds the matcher to the Rule's header set.
//
// This function will create the headers map if it's nil.
func (r *Rule) Header(k string, v Matcher) {
	if r.Headers == nil {
		r.Headers = make(map[string]Matcher)
	}
	r.Headers[k] = v
}
func matchAll(r *http.Request, s []Rule) bool {
	if len(s) == 0 {
		return true
	}
	for i := range s {
		if s[i].match(r) {
			return true
		}
	}
	return false
}
