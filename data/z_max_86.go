//go:build 386 || arm || mips || mipsle

package data

// MaxSlice is the max slice value used when creating slices to prevent OOM
// issues. XMT will refuse to  make a slice any larger than this and will return
// 'ErrToLarge'
const MaxSlice = 1_952_257_862 // round((2<<30)/1.1)
