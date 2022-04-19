package c2

import (
	"sync/atomic"

	"github.com/iDigitalFlame/xmt/util/bugtrack"
)

const (
	stateCanRecv uint32 = 1 << iota
	stateReady
	stateClosed
	stateClosing
	stateShutdown
	stateSendClose
	stateRecvClose
	stateWakeClose
	stateChannel
	stateChannelValue
	stateChannelUpdated
	stateChannelProxy
	stateSeen
	stateMoving
	stateReplacing
)

type state uint32

func (s *state) Tag() bool {
	if !s.Seen() {
		return false
	}
	s.Unset(stateSeen)
	return true
}
func (s state) Seen() bool {
	return atomic.LoadUint32((*uint32)(&s))&stateSeen != 0
}
func (s state) Ready() bool {
	if s.Closed() {
		return false
	}
	return atomic.LoadUint32((*uint32)(&s))&stateReady != 0
}
func (s state) Last() uint16 {
	return uint16(atomic.LoadUint32((*uint32)(&s)) >> 16)
}
func (s state) Moving() bool {
	return atomic.LoadUint32((*uint32)(&s))&stateMoving != 0
}
func (s state) Closed() bool {
	return atomic.LoadUint32((*uint32)(&s))&stateClosed != 0
}
func (s *state) Set(v uint32) {
	atomic.StoreUint32((*uint32)(s), atomic.LoadUint32((*uint32)(s))|v)
}
func (s state) CanRecv() bool {
	if s.Closed() || s.RecvClosed() {
		return false
	}
	return atomic.LoadUint32((*uint32)(&s))&stateCanRecv != 0
}
func (s state) Closing() bool {
	if s.Closed() {
		return true
	}
	return atomic.LoadUint32((*uint32)(&s))&stateClosing != 0
}
func (s state) Channel() bool {
	return atomic.LoadUint32((*uint32)(&s))&stateChannel != 0
}
func (s state) Shutdown() bool {
	if s.Closed() {
		return true
	}
	return atomic.LoadUint32((*uint32)(&s))&stateShutdown != 0
}
func (s state) Replacing() bool {
	return atomic.LoadUint32((*uint32)(&s))&stateReplacing != 0
}
func (s *state) Unset(v uint32) {
	d := atomic.LoadUint32((*uint32)(s)) &^ v
	atomic.StoreUint32((*uint32)(s), d)
}
func (s state) RecvClosed() bool {
	if s.Closed() {
		return true
	}
	return atomic.LoadUint32((*uint32)(&s))&stateRecvClose != 0
}
func (s state) SendClosed() bool {
	if s.Closed() {
		return true
	}
	return atomic.LoadUint32((*uint32)(&s))&stateSendClose != 0
}
func (s state) WakeClosed() bool {
	if s.Closed() {
		return true
	}
	return atomic.LoadUint32((*uint32)(&s))&stateWakeClose != 0
}
func (s *state) SetLast(v uint16) {
	atomic.StoreUint32((*uint32)(s), (uint32(v)<<16)|uint32(uint16(atomic.LoadUint32((*uint32)(s)))))
}
func (s state) ChannelValue() bool {
	return atomic.LoadUint32((*uint32)(&s))&stateChannelValue != 0
}
func (s state) ChannelProxy() bool {
	return atomic.LoadUint32((*uint32)(&s))&stateChannelProxy != 0
}
func (s state) ChannelUpdated() bool {
	return atomic.LoadUint32((*uint32)(&s))&stateChannelUpdated != 0
}
func (s *state) ChannelCanStop() bool {
	if s.Closed() || !s.Channel() {
		return true
	}
	if s.ChannelUpdated() {
		s.Unset(stateChannelUpdated)
		return !s.ChannelValue()
	}
	return !s.Channel()
}
func (s state) ChannelCanStart() bool {
	if s.Closed() {
		return false
	}
	if s.Channel() {
		return true
	}
	return s.ChannelValue()
}
func (s *state) SetChannel(e bool) bool {
	if e {
		if s.ChannelValue() {
			if bugtrack.Enabled {
				bugtrack.Track("c2.state.SetChannel(): e=%t, s.ChannelValue()=true, setting channel is NOP since we are in a channel.", e)
			}
			return false
		}
		s.Set(stateChannelValue)
	} else {
		if (!s.Channel() || !s.ChannelProxy()) && !s.ChannelValue() {
			if bugtrack.Enabled {
				bugtrack.Track(
					"c2.state.SetChannel(): e=%t, s.Channel()=%t, s.ChannelProxy()=%t, s.ChannelValue()=%t, canceling channel is NOP since we are not in a channel.",
					e, s.Channel(), s.ChannelProxy(), s.ChannelValue(),
				)
			}
			return false
		}
		s.Unset(stateChannelValue)
	}
	s.Set(stateChannelUpdated)
	return true
}
