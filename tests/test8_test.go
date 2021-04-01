package main

import (
	"testing"

	"github.com/iDigitalFlame/xmt/device"
)

func BenchmarkIDEqualHash(b *testing.B) {
	var (
		id2 = device.UUID
		id1 device.ID
	)
	for n := 0; n < b.N; n++ {
		_ = id2.Equal(id1)
	}
}

func BenchmarkIDEqualSign(b *testing.B) {
	var (
		id2 = device.UUID
		id1 device.ID
	)
	for n := 0; n < b.N; n++ {
		_ = id2 == id1
	}
}
func BenchmarkIDMapHash(b *testing.B) {
	var (
		id2 = device.UUID
		m   = make(map[uint32]string)
	)

	m[id2.Hash()] = "hello!"

	for n := 0; n < b.N; n++ {
		_, _ = m[id2.Hash()]
	}
}

func BenchmarkIDMapSign(b *testing.B) {
	var (
		id2 = device.UUID
		m   = make(map[device.ID]string)
	)

	m[id2] = "hello!"

	for n := 0; n < b.N; n++ {
		_, _ = m[id2]
	}
}

/*
func BenchmarkErrorf(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_ = fmt.Errorf("this error is nil")
	}
}
func BenchmarkErrorNew(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_ = errors.New("this error is nil")
	}
}

func BenchmarkXErrNew(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_ = xerr.New("this error is nil")
	}
}

func BenchmarkErrorfWrap(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_ = fmt.Errorf("this error is nil: %w", io.EOF)
	}
}
func BenchmarkErrWrap(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_ = xerr.Wrap("this error is nil", io.EOF)
	}
}

func BenchmarkSprintf(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_ = fmt.Sprintf("string value %x!", n)
	}
}
func BenchmarkStringPlus(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_ = "string value " + strconv.FormatInt(int64(n), 16) + "!"
	}
}

func BenchmarkStringMatch(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_ = strings.ToLower("DERPMASTER69!") == strings.ToLower("DERpmASTER69!")
	}
}

func BenchmarkFastStringMatch(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_ = util.FastUTF8Match("DERPMASTER69!", "DERpmASTER69!")
	}
}

func BenchmarkStringMatchBad(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_ = strings.ToLower("DERPMASTER69!") == strings.ToLower("DERpmASTER68!")
	}
}

func BenchmarkFastStringMatchBad(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_ = util.FastUTF8Match("DERPMASTER69!", "DERpmASTER68!")
	}
}

func BenchmarkStringMatchBad2(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_ = strings.ToLower("DERPMASTER69!") == strings.ToLower("mASTER68!")
	}
}

func BenchmarkFastStringMatchBad2(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_ = util.FastUTF8Match("DERPMASTER69!", "mASTER68!")
	}
}
*/
