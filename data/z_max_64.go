//go:build !386 && !arm && !mips && !mipsle

package data

// MaxSlice is the max slice value used when creating slices to prevent OOM
// issues. XMT will refuse to  make a slice any larger than this and will return
// 'ErrToLarge'
const MaxSlice = 4_398_046_511_104 // (2 << 41)
