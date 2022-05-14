//go:build !implant

package device

import "github.com/iDigitalFlame/xmt/util"

// String returns a string representation of the OSType.
func (o OSType) String() string {
	switch o {
	case Windows:
		return "Windows"
	case Linux:
		return "Linux"
	case Unix:
		return "Unix/BSD"
	case Mac:
		return "MacOS"
	case IOS:
		return "iOS"
	case Android:
		return "Android"
	case Plan9:
		return "Plan9"
	case Unsupported:
		return "Unsupported"
	}
	return "Unknown"
}
func (d device) String() string {
	b := builders.Get().(*util.Builder)
	b.Grow(30)
	b.WriteString(d.Name + "(" + d.Mac.String() + "): [")
	for i := range d.Address {
		if i > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(d.Address[i].String())
	}
	b.WriteByte(']')
	r := b.Output()
	builders.Put(b)
	return r
}

// String returns a simple string representation of the Machine instance.
func (m Machine) String() string {
	var e string
	if m.IsElevated() {
		e = "*"
	}
	return "[" + m.ID.String() + "] " + m.Hostname + " (" + m.Version + ") " + e + m.User
}
