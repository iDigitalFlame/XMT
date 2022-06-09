//go:build implant

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

package cout

import "github.com/PurpleSec/logx"

// Enabled is a compile time constant that can be used to disable/enable the logx Logger and prevent any
// un-needed fmt calls as the client does not /naturally/ need to produce output. Only needed for debug
// purposes
const Enabled = false

var log = new(Log)

// Log is an interface for any type of struct that supports standard Logging functions.
type Log struct{}

// New creates a Log instance from a logx Logger.
func New(logx.Log) *Log {
	return log
}

// Set updates the internal logger. This function is a NOP if the logger is nil or logging is not
// enabled via the 'client' build tag.
func (Log) Set(_ logx.Log) {}

// Info writes a informational message to the logger.
// The function arguments are similar to fmt.Sprintf and fmt.Printf. The first argument is
// a string that can contain formatting characters. The second argument is a vardict of
// interfaces that can be omitted or used in the supplied format string.
func (Log) Info(_ string, _ ...any) {}

// Error writes a error message to the logger.
// The function arguments are similar to fmt.Sprintf and fmt.Printf. The first argument is
// a string that can contain formatting characters. The second argument is a vardict of
// interfaces that can be omitted or used in the supplied format string.
func (Log) Error(_ string, _ ...any) {}

// Trace writes a tracing message to the logger.
// The function arguments are similar to fmt.Sprintf and fmt.Printf. The first argument is
// a string that can contain formatting characters. The second argument is a vardict of
// interfaces that can be omitted or used in the supplied format string.
func (Log) Trace(_ string, _ ...any) {}

// Debug writes a debugging message to the logger.
// The function arguments are similar to fmt.Sprintf and fmt.Printf. The first argument is
// a string that can contain formatting characters. The second argument is a vardict of
// interfaces that can be omitted or used in the supplied format string.
func (Log) Debug(_ string, _ ...any) {}

// Warning writes a warning message to the logger.
// The function arguments are similar to fmt.Sprintf and fmt.Printf. The first argument is
// a string that can contain formatting characters. The second argument is a vardict of
// interfaces that can be omitted or used in the supplied format string.
func (Log) Warning(_ string, _ ...any) {}
