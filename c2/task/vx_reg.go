//go:build windows || !implant

package task

const (
	regOpLs  uint8 = 0
	regOpGet uint8 = iota
	regOpMake
	regOpDeleteKey
	regOpDelete
	regOpSet
	regOpSetString
	regOpSetDword
	regOpSetQword
	regOpSetBytes
	regOpSetExpandString
	regOpSetStringList
)
