//go:build nosweep

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
	"context"
	"time"
)

// MemorySweep enables the GC memory sweeper, which keeps the process memory
// clean to prevent any crashes while in DLL format or injected. This function
// only needs to be called once and will return immediately.
//
// The context is recommended to prevent any leaking Goroutines from being left
// behind.
//
// Defaults to a time of one minute.
func MemorySweep(_ context.Context) {}

// MemorySweepEx enables the GC memory sweeper, which keeps the process memory
// clean to prevent any crashes while in DLL format or injected. This function
// only needs to be called once and will return immediately.
//
// The context is recommended to prevent any leaking Goroutines from being left
// behind.
//
// Allows for specification of the time span between sweeps.
func MemorySweepEx(_ context.Context, _ time.Duration) {}
