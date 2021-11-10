//go:build stdrand
// +build stdrand

package util

type random struct {
	*rand.Rand
}

func getRandom() *random {
	return &random{Rand: rand.New(rand.NewSource(cputicks()))}
}
