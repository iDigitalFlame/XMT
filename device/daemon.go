package device

import (
	"context"

	"github.com/iDigitalFlame/xmt/util/xerr"
)

// ErrQuit is an error that can be returned from the DaemonFunction that
// will indicate a clean (non-error) break of the Daemon loop.
var ErrQuit = xerr.Sub("quit", 0x1F)

// DaemonFunc is a function type that can be used as a Daemon. This function
// should return nil to indicate a successful run or ErrQuit to break out of
// a 'DaemonTicker' loop.
//
// Any non-nil errors will be interpreted as exit code '1'.
type DaemonFunc func(context.Context) error
