package com

// Flag is a bitwise integer that repersents important
// information about the packet that its assigned to.
type Flag uint32

// Fragment is an integer that repersents the position of
// the holding packet in a stream of connected packets.  By default
// fragment is zero and is one indexed when a fragmented packet sequence is
// needed.
type Fragment uint32
