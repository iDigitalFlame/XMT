//go:build !stdrand
// +build !stdrand

// Copyright (C) 2020 - 2023 iDigitalFlame
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.
//

package util

type random struct{}

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
	return int(abs64(r.Uint64())) % n
}
func (random) Int31n(n int32) int32 {
	return int32(abs32(FastRand())) & n
}
func (r random) Int63n(n int64) int64 {
	return int64(abs64(r.Uint64())) % n
}
func (random) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = byte(FastRandN(256))
	}
	return len(p), nil
}
