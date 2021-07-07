package wc2

import (
	"net/http"

	"github.com/iDigitalFlame/xmt/util/text"
)

// DefaultGenerator is the generator used if no generator is provided when a client attempts a
// connection. The default values are a URL for an news post, Windows host and a Firefox version 86 user agent.
var DefaultGenerator = Generator{
	URL:   text.Matcher("/news/post/%d/"),
	Agent: text.Matcher("Mozilla/5.0 (Windows NT 10; WOW64; rv:86.0) Gecko/20101%100d Firefox/86.0"),
}

// Rule is a struct that represents a rule set used by the Web server to determine
// the difference between normal and C2 traffic.
type Rule struct {
	URL, Host, Agent matcher
}

// Generator is a struct that is composed of three separate Stringer interfaces. These are called
// via their 'String' function to specify the User-Agent, URL and Host string values. They can be set to
// static strings using the 'text.String' wrapper. This struct can be used as a C2 client connector. If
// the Client property is not set, the DefaultClient value will be used.
type Generator struct {
	URL, Host, Agent stringer
}
type matcher interface {
	MatchString(string) bool
}
type stringer interface {
	String() string
}

// Reset sets all the Generator values to nil. This allows for an empty Generator to be used.
func (g *Generator) Reset() {
	g.URL, g.Host, g.Agent = nil, nil, nil
}

// Rule will attempt to generate a Rule that matches this generator using the current configuration.
// Rules will only be added if the settings implement the 'MatchString(string) bool' function. Otherwise, the
// specified rule configuration will be empty.
func (g Generator) Rule() Rule {
	var r Rule
	if g.URL != nil {
		if m, ok := g.URL.(matcher); ok {
			r.URL = m
		} else if m, ok := g.URL.(text.Matcher); ok {
			r.URL = m.Match()
		} else {
			r.URL = text.Matcher(g.URL.String()).Match()
		}
	}
	if g.Host != nil {
		if m, ok := g.Host.(matcher); ok {
			r.Host = m
		} else if m, ok := g.Host.(text.Matcher); ok {
			r.Host = m.Match()
		} else {
			r.Host = text.Matcher(g.Host.String()).Match()
		}
	}
	if g.Agent != nil {
		if m, ok := g.Agent.(matcher); ok {
			r.Agent = m
		} else if m, ok := g.Agent.(text.Matcher); ok {
			r.Agent = m.Match()
		} else {
			r.Agent = text.Matcher(g.Agent.String()).Match()
		}
	}
	return r
}
func (g Generator) empty() bool {
	return g.Agent == nil && g.Host == nil && g.URL == nil
}
func (r Rule) checkMatch(c *http.Request) bool {
	if r.Host == nil && r.URL == nil && r.Agent == nil {
		return false
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
	return true
}
func (g Generator) prepRequest(r *http.Request) {
	if g.URL != nil {
		s := g.URL.String()
		if len(s) > 0 && s[0] != '/' {
			r.URL.Path = "/" + s
		} else {
			r.URL.Path = s
		}
	}
	if g.Host != nil {
		r.Host = g.Host.String()
	}
	if g.Agent != nil {
		r.Header.Set("User-Agent", g.Agent.String())
	}
}
