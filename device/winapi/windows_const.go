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

package winapi

const (
	// SwHide hides the window and activates another window.
	SwHide uint8 = iota
	// SwNormal activates and displays a window. If the window is minimized or
	// maximized, the system restores it to its original size and position. An
	// application should specify this flag when displaying the window for the
	// first time.
	SwNormal
	// SwMinimized activates the window and displays it as a minimized window.
	SwMinimized
	// SwMaximize activates the window and displays it as a maximized window.
	SwMaximize
	// SwNoActive displays a window in its most recent size and position. This
	// value is similar to SwNormal, except that the window is not activated.
	SwNoActive
	// SwShow activates the window and displays it in its current size and
	// position.
	SwShow
	// SwMinimize minimizes the specified window and activates the next top-level
	// window in the Z order.
	SwMinimize
	// SwMinimizeNoActive displays the window as a minimized window. This value
	// is similar to SwMinimizeNoActive, except the window is not activated.
	SwMinimizeNoActive
	// SwShowNoActive displays the window in its current size and position.
	// This value is similar to SwShow, except that the window is not activated.
	SwShowNoActive
	// SwRestore activates and displays the window. If the window is minimized
	// or maximized, the system restores it to its original size and position.
	// An application should specify this flag when restoring a minimized window.
	SwRestore
	// SwDefault sets the show state based on the SW_ value specified in the
	// STARTUPINFO structure passed to the CreateProcess function by the program
	// that started the application.
	SwDefault
	// SwMinimizeForce minimizes a window, even if the thread that owns the
	// window is not responding. This flag should only be used when minimizing
	// windows from a different thread.
	SwMinimizeForce
)
