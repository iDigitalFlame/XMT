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

// Package bugtrack enables the bug tracking system, which is comprised of a
// global logger that will write to Standard Error and on the filesystem in a
// temporary directory, "$TEMP" in *nix and "%TEMP%" on Windows, that is named
// "bugtrack-<PID>.log".
//
// To enable bug tracking, use the "bugs" build tag.
//
package bugtrack
