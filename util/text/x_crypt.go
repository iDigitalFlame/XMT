//go:build crypt

package text

import "github.com/iDigitalFlame/xmt/util/crypt"

var alpha = crypt.Get(0) // abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789
