package main

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"testing"

	"github.com/iDigitalFlame/xmt/util"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

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
