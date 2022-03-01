//go:build !stdrand
// +build !stdrand

package util

type random struct{}

func getRandom() *random {
	return new(random)
}
func (r random) Int() int {
	return int(abs64(r.Uint64()))
}
func abs32(v uint32) uint32 {
	return v &^ (1 << 31)
}
func abs64(v uint64) uint64 {
	return v &^ (1 << 63)
}
func (random) Int31() int32 {
	return int32(abs32(FastRand()))
}
func (r random) Int63() int64 {
	return int64(abs64(r.Uint64()))
}
func (random) Uint32() uint32 {
	return FastRand()
}
func (random) Uint64() uint64 {
	return uint64(FastRand())<<32 | uint64(FastRand())
}
func (r random) Intn(n int) int {
	return int(int(abs64(r.Uint64())) % n)
}
func (random) Int31n(n int32) int32 {
	return int32(int32(abs32(FastRand())) & n)
}
func (r random) Int63n(n int64) int64 {
	return int64(int64(abs64(r.Uint64())) % n)
}
func (random) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = byte(FastRand() % uint32(256))
	}
	return len(p), nil
}
