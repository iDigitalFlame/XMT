//go:build stdrand

package util

import "math/rand"

type random struct {
	*rand.Rand
}

func getRandom() *random {
	return &random{Rand: rand.New(rand.NewSource(cputicks()))}
}
