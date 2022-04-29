//go:build nofrag

// This tag "nofrag" can be used to enable/disable the Fragment system (by allowing a Zero frag value).
// While disabling fragments can show improvements in packet speed, it breaks stateless (ie: non-TCP/UNIX)
// connection buffers.

package limits

// Frag is the max size used to fragment packets into.
// Any packet over this byte size will be fragmented.
const Frag = -1

// Packets determines how many Packets may be processed by the Session thread before
// waiting another wait cycle. If this is set to anything less than one, only a single
// Packet will be processed at a time.
const Packets = 256
