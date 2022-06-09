//go:build windows || !implant

// Copyright (C) 2020 - 2022 iDigitalFlame
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
	taskWindowShow
	taskWindowClose
	taskWindowMessage
	taskWindowMove
	taskWindowFocus
	taskWindowType
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
	taskTrollWTF
)
