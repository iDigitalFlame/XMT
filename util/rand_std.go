// +build stdrand

package util

import (
	// Import unsafe to use faster "cputicks" function instead of "time.Now().UnixNano()"
	_ "unsafe"
)

type random struct {
	*rand.Rand
}

func getRandom() *random {
	return &random{Rand: rand.New(rand.NewSource(cputicks()))}
}
