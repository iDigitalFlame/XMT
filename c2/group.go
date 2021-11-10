package c2

import (
	"net"
	"sort"

	"github.com/iDigitalFlame/xmt/util"
)

const (
	// GroupLastValid is the default selection that will keep using the last Group
	// connection unless it fails. On a failure (or the first call), this will act
	// similar to 'GroupRoundRobin'.
	GroupLastValid = selector(0)
	// GroupRoundRobin is a selection option that will simply select the NEXT
	// Group on every connection attempt. This option is affected by the Group
	// weights set on each addition and will perfer higher numbered options.
	GroupRoundRobin = selector(1)
	// GroupRandom is a selection option that will ignore all weights and order
	// and will select an entry from the list randomally.
	GroupRandom = selector(2)
)

type link struct {
	p Profile
	c Connector
	h string
	w uint8
}
type selector uint8

// Group is a struct that allows for using multiple connections for a single
// Session.
//
// Multiple Hosts, Connectors and Profiles can be specified.
// The Selection option allows for choosing the behavior of this Group and the
// order connections will be used.
//
// The default Profile/Connector can be used to set the Profile and/or Connector
// used if one is not specified when a Group is added.
type Group struct {
	def, last *link
	l         []*link
	err       bool

	Selection selector
}
type static struct {
	c Connector
	h string
}
type linker interface {
	String() string
	connect(*Session) (net.Conn, error)
}

// Reset will set the 'last' value to nil and will make this Group behave as
// it was not used before.
func (g *Group) Reset() {
	g.last = nil
}

// Len implements the 'sort.Interface' interface.
func (g *Group) Len() int {
	return len(g.l)
}
func (g *Group) next() *link {
	if len(g.l) == 0 {
		return g.last
	}
	switch g.Selection {
	case GroupRandom:
		return g.l[util.FastRandN(len(g.l))]
	case GroupLastValid:
		if !g.err && g.last != nil {
			return g.last
		}
		fallthrough
	case GroupRoundRobin:
		if !g.err && g.last != nil {
			return g.last
		}
		if g.last == nil {
			return g.l[0]
		}
		var x bool
		for i := range g.l {
			if g.l[i] == g.last {
				x = true
				continue
			} else if x {
				return g.l[i]
			}
		}
		g.last = g.l[0]
	default:
	}
	return g.last
}

// Swap implements the 'sort.Interface' interface.
func (g *Group) Swap(i, j int) {
	g.l[i], g.l[j] = g.l[j], g.l[i]
}

// UnsetDefault will remove the default values set for this Group.
func (g *Group) UnsetDefault() {
	g.def = nil
}
func (l *link) String() string {
	v, ok := l.c.(stringer)
	if !ok {
		return ""
	}
	return v.String()
}
func (g *Group) String() string {
	if g.last != nil {
		return g.last.String()
	}
	if g.def != nil {
		return g.def.String()
	}
	return ""
}
func (s static) String() string {
	v, ok := s.c.(stringer)
	if !ok {
		return ""
	}
	return v.String()
}
func (g *Group) profile() Profile {
	if g.last == nil {
		return g.def.p
	}
	return g.last.p
}

// Less implements the 'sort.Interface' interface.
func (g *Group) Less(i, j int) bool {
	if g.l[i].w == g.l[j].w {
		return g.l[i].h < g.l[j].h
	}
	return g.l[i].w > g.l[j].w
}

// NewGroup will return a new Group struct with the specified defaults
// set if not-empty/nil.
//
// This is the same as 'new(Group).Default(h, p, c)'.
func NewGroup(c Connector, p Profile) *Group {
	g := new(Group)
	g.Default(c, p)
	return g
}

// Default sets the default values for this Group.
//
// These values can be used in-place of empty options for each additional section.
// Nil values DO NOT clear the defaults already set. Use the 'UnsetDefault' function
// to reset the defaults.
func (g *Group) Default(c Connector, p Profile) {
	if g.def == nil {
		g.def = new(link)
	}
	if p != nil {
		g.def.p = p
	}
	if c != nil {
		g.def.c = c
	} else if p != nil {
		if v, ok := p.(hinter); ok && v != nil {
			g.def.c = v.Connector()
		}
	}
}
func (l *link) connect(s *Session) (net.Conn, error) {
	c, err := l.c.Connect(l.h)
	if err != nil {
		return nil, err
	}
	if s.host = l.h; l.p != nil {
		s.w, s.t = l.p.Wrapper(), l.p.Transform()
	}
	return c, nil
}
func (s static) connect(_ *Session) (net.Conn, error) {
	return s.c.Connect(s.h)
}
func (g *Group) connect(s *Session) (net.Conn, error) {
	if g.last, g.err = g.next(), false; g.last == nil {
		if g.def == nil {
			return nil, ErrNoConnector
		}
		return g.def.connect(s)
	}
	c, err := g.last.connect(s)
	if err != nil {
		g.err = true
		return nil, err
	}
	return c, nil
}

// Must will add the specified host, Profile and Connector to the Group to be used
// for Session connections.
//
// If the supplied Profile and/or Connector are nil, the default Profile and/or
// Connector in this Group will be used.
//
// If the Connector supplied and default Connector are nil, the Connector will
// try to be exposed via the supplied Profile (if it exposes the 'hinter' interface).
//
// If no Connector is found, 'ErrNoConnector' this function will panic.
func (g *Group) Must(h string, c Connector, p Profile) {
	if err := g.AddWeighted(0, h, c, p); err != nil {
		panic(err)
	}
}

// Add will add the specified host, Profile and Connector to the Group to be used
// for Session connections.
//
// If the supplied Profile and/or Connector are nil, the default Profile and/or
// Connector in this Group will be used.
//
// If the Connector supplied and default Connector are nil, the Connector will
// try to be exposed via the supplied Profile (if it exposes the 'hinter' interface).
//
// If no Connector is found, 'ErrNoConnector' will be returned.
func (g *Group) Add(h string, c Connector, p Profile) error {
	return g.AddWeighted(0, h, c, p)
}

// MustWeighted will add the specified host, Profile and Connector to the Group to be used
// for Session connections. The initial integer specified will allow for specifying a
// weight to determine when to use this group entry. The weight will be rounded to
// [0 - 100] (inclusive).
//
// If the supplied Profile and/or Connector are nil, the default Profile and/or
// Connector in this Group will be used.
//
// If the Connector supplied and default Connector are nil, the Connector will
// try to be exposed via the supplied Profile (if it exposes the 'hinter' interface).
//
// If no Connector is found, 'ErrNoConnector' will panic.
func (g *Group) MustWeighted(w int, h string, c Connector, p Profile) {
	if err := g.AddWeighted(w, h, c, p); err != nil {
		panic(err)
	}
}

// AddWeighted will add the specified host, Profile and Connector to the Group to be used
// for Session connections. The initial integer specified will allow for specifying a
// weight to determine when to use this group entry. The weight will be rounded to
// [0 - 100] (inclusive).
//
// If the supplied Profile and/or Connector are nil, the default Profile and/or
// Connector in this Group will be used.
//
// If the Connector supplied and default Connector are nil, the Connector will
// try to be exposed via the supplied Profile (if it exposes the 'hinter' interface).
//
// If no Connector is found, 'ErrNoConnector' will be returned.
func (g *Group) AddWeighted(w int, h string, c Connector, p Profile) error {
	if p == nil {
		p = g.def.p
	}
	if c == nil {
		c = g.def.c
	}
	if p != nil {
		if v, ok := p.(hinter); ok {
			if c == nil {
				c = v.Connector()
			}
			if len(h) == 0 {
				h = v.Host()
			}
		}
	}
	if c == nil {
		return ErrNoConnector
	}
	if len(h) == 0 {
		return ErrNoHost
	}
	var x uint8
	if w >= 100 {
		x = 100
	} else if w < 0 {
		x = 0
	} else {
		x = uint8(w)
	}
	if g.l = append(g.l, &link{w: x, h: h, p: p, c: c}); w != 0 {
		sort.Sort(g)
	}
	return nil
}
