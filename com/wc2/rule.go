package wc2

import (
	"fmt"
	"net/http"

	"github.com/iDigitalFlame/xmt/util"
)

var (
	// DefaultGenerator is the generator used if no generator is provided when a client attempts a
	// connection. The default values are a URL for an news post, Windows host and a Firefox version 70 user agent.
	DefaultGenerator = Generator{
		URL:   util.Matcher("/news/post/%d/"),
		Agent: util.Matcher("Mozilla/5.0 (Windows NT 10; WOW64; rv:79.0) Gecko/20101%100d Firefox/79.0"),
	}
)

// Rule is a struct that represents a rule set used by the Web server to determine
// the difference between normal and C2 traffic.
type Rule struct {
	URL, Host, Agent matcher
}

// Generator is a struct that is composed of three separate fmt.Stringer interfaces. These are called
// via their 'String' function to specify the User-Agent, URL and Host string values. They can be set to
// static strings using the 'util.String' wrapper. This struct can be used as a C2 client connector. If
// the Client property is not set, the DefaultClient value will be used.
type Generator struct {
	URL, Host, Agent fmt.Stringer
}
type matcher interface {
	MatchString(string) bool
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
		} else if m, ok := g.URL.(util.Matcher); ok {
			r.URL = m.Match()
		} else {
			r.URL = util.Matcher(g.URL.String()).Match()
		}
	}
	if g.Host != nil {
		if m, ok := g.Host.(matcher); ok {
			r.Host = m
		} else if m, ok := g.Host.(util.Matcher); ok {
			r.Host = m.Match()
		} else {
			r.Host = util.Matcher(g.Host.String()).Match()
		}
	}
	if g.Agent != nil {
		if m, ok := g.Agent.(matcher); ok {
			r.Agent = m
		} else if m, ok := g.Agent.(util.Matcher); ok {
			r.Agent = m.Match()
		} else {
			r.Agent = util.Matcher(g.Agent.String()).Match()
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
			r.URL.Path = fmt.Sprintf("/%s", s)
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
