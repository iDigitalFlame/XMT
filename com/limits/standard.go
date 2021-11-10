//go:build !s_tiny || !s_small || !s_medium || !s_nofrag
// +build !s_tiny !s_small !s_medium !s_nofrag

package limits

// Frag is the max size used to fragment packets into.
// Any packet over this byte size will be fragmented.
const Frag = 33554432

// Buffer is the default size of buffers for stateless connections.
// This number affects how UDP/IP connections perform and how large their backlog
// will be before blocking. (Server side only).
const Buffer = 65536

// Packets determines how many Packets may be processed by the Session thread before
// waiting another wait cycle. If this is set to anything less than one, only a single
// Packet will be processed at a time.
const Packets = 256
