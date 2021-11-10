//go:build !windows
// +build !windows

package wintask

import (
	"context"

	"github.com/iDigitalFlame/xmt/data"
)

// Tasks is an OS-dependant function that returns the task types that can
// be used specific to this package.
func Tasks() []func(context.Context, data.Reader, data.Writer) error {
	return nil
}
