//go:build !windows
// +build !windows

package man

import "github.com/iDigitalFlame/xmt/device/devtools"

func (objSync) check(_ string) (bool, error) {
	return false, devtools.ErrNoWindows
}
func (objSync) create(_ string) (listener, error) {
	return nil, devtools.ErrNoWindows
}
