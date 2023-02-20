//go:build plan9
// +build plan9

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

package limits

import (
	"os"
	"os/signal"
	"syscall"
)

// Reset will undo all the signals ignored by the 'Ignore' function.
func Reset() {
	signal.Reset(syscall.SIGHUP, syscall.SIGABRT)
}

// Ignore is a simple helper method that can be used to ignore signals
// that can be used to abort or generate a stack-trace.
//
// Used for anti-debugging measures.
func Ignore() {
	signal.Ignore(syscall.SIGHUP, syscall.SIGABRT)
}
func convertSignal(s uint32) os.Signal {
	return syscall.Note(s)
}
