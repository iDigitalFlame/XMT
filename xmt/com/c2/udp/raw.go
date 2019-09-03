package udp

import (
	"net"
	"time"
)

var (
	// Raw is the UDP Raw connector.  This connector uses raw UDP
	// connections without any encoding or Transforms.
	Raw = &packetConnector{
		dial: &net.Dialer{
			Timeout: RawDefaultTimeout,
		},
		network: "udp",
	}

	// RawDefaultTimeout is the default timeout used for the Raw TCP connector.
	// The default is 5 seconds.
	RawDefaultTimeout = time.Duration(5) * time.Second
)
