//go:build !windows

package man

import "github.com/iDigitalFlame/xmt/device"

func (objSync) check(_ string) (bool, error) {
	return false, device.ErrNoWindows
}
func (objSync) create(_ string) (listener, error) {
	return nil, device.ErrNoWindows
}
