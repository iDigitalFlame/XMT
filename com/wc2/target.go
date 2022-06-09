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

// TargetEmpty is a Target that does not have anything set and will not mutate a
// Request.
var TargetEmpty Target

// Target is a struct that is composed of three separate Stringer interfaces.
// These are called via their 'String' function to specify the User-Agent, URL
// and Host string values. They can be set to static strings using the
// 'text.String' wrapper. This struct can be used as a C2 client connector. If
// the Client property is not set, the DefaultClient value will be used.
type Target struct {
	_                [0]func()
	URL, Host, Agent Stringer
	Headers          map[string]Stringer
}

// Stringer is a utility interface that takes a single 'String() string'
// function for returning a string to be used as a Target.
type Stringer interface {
	String() string
}

// Reset sets all the Target values to nil. This allows for an empty Target to
// be used.
func (t *Target) Reset() {
	if t.URL, t.Host, t.Agent = nil, nil, nil; t.Headers != nil && len(t.Headers) > 0 {
		for k := range t.Headers {
			delete(t.Headers, k)
		}
	}
}

// Rule will attempt to generate a Rule that matches this generator using the
// current configuration.
//
// Rules will only be added if the settings implement the 'MatchString(string) bool'
// function. Otherwise, the specified rule will attempt to match using a Text
// Matcher.
//
// Empty Target return an empty rule.
func (t Target) Rule() Rule {
	var r Rule
	if t.empty() {
		return r
	}
	if t.URL != nil {
		if m, ok := t.URL.(Matcher); ok {
			r.URL = m
		} else if m, ok := t.URL.(text.Matcher); ok {
			r.URL = m.Match()
		} else {
			r.URL = text.Matcher(t.URL.String()).Match()
		}
	}
	if t.Host != nil {
		if m, ok := t.Host.(Matcher); ok {
			r.Host = m
		} else if m, ok := t.Host.(text.Matcher); ok {
			r.Host = m.Match()
		} else {
			r.Host = text.Matcher(t.Host.String()).Match()
		}
	}
	if t.Agent != nil {
		if m, ok := t.Agent.(Matcher); ok {
			r.Agent = m
		} else if m, ok := t.Agent.(text.Matcher); ok {
			r.Agent = m.Match()
		} else {
			r.Agent = text.Matcher(t.Agent.String()).Match()
		}
	}
	if t.Headers != nil && len(t.Headers) > 0 {
		r.Headers = make(map[string]Matcher, len(t.Headers))
		for k, v := range t.Headers {
			if m, ok := v.(Matcher); ok {
				r.Headers[k] = m
			} else if m, ok := v.(text.Matcher); ok {
				r.Headers[k] = m.Match()
			} else {
				r.Headers[k] = text.Matcher(v.String()).Match()
			}
		}
	}
	return r
}
func (t Target) empty() bool {
	return t.Agent == nil && t.Host == nil && t.URL == nil && len(t.Headers) == 0
}
func (t *Target) mutate(r *http.Request) {
	if t.URL != nil {
		s := t.URL.String()
		if len(s) > 0 && s[0] != '/' {
			r.URL.Path = "/" + s
		} else {
			r.URL.Path = s
		}
	}
	if t.Host != nil {
		r.Host = t.Host.String()
	}
	if t.Agent != nil {
		r.Header.Set(userAgent, t.Agent.String())
	}
	for k, v := range t.Headers {
		r.Header.Set(k, v.String())
	}
}

// Header adds the stringer too the Target's header set.
//
// This function will create the headers map if it's nil.
func (t *Target) Header(k string, v Stringer) {
	if t.Headers == nil {
		t.Headers = make(map[string]Stringer)
	}
	t.Headers[k] = v
}
