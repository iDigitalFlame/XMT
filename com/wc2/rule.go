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

// Header adds the matcher too the Rule's header set.
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
		return false
	}
	for i := range s {
		if !s[i].match(r) {
			return false
		}
	}
	return true
}
