//go:build !crypt

package com

func (udpErr) Error() string {
	return "deadline exceeded"
}
