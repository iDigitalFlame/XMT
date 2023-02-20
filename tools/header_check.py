#!/usr/bin/python
# Copyright (C) 2020 - 2023 iDigitalFlame
#
# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# any later version.
#
# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.
#
# You should have received a copy of the GNU General Public License
# along with this program.  If not, see <https://www.gnu.org/licenses/>.
#
# header_check.py <dir>
# Checks to see if all *.go files have the correct GNU license header.
#

from glob import glob
from os.path import join
from sys import argv, exit


def scan_dir(path):
    for i in glob(join(path, "**/**.go"), recursive=True):
        if "unit_tests" in i or "src/" in i:
            continue
        with open(i, "r") as f:
            b = f.read().split("\n")
        if len(b) < 8:
            raise ValueError(f"{i}: Empty file!")
        n = 0
        if b[0].startswith("//go:build"):
            if not b[1].startswith("// +build"):
                raise ValueError(f"{i}: No old build directive after new directive!")
            n = 3
            while b[n - 1].startswith("// +build"):
                n = n + 1
        if not b[n].startswith("// Copyright (C) 2020"):
            raise ValueError(f"{i}: Missing copyright!")
        if b[n + 1] != "//":
            raise ValueError(f"{i}: Missing black space after copyright!")
        if not b[n + 2].startswith("// This program is free software:"):
            raise ValueError(f"{i}: Missing freedom text!")
        del b


def scan_dir2(path):
    for i in glob(join(path, "**/**.go"), recursive=True):
        if "unit_tests" in i:
            continue
        with open(i, "r") as f:
            b = f.read().split("\n")
        if len(b) < 8:
            raise ValueError(f"{i}: Empty file!")
        n = 0
        if b[0].startswith("//go:build"):
            if len(b[1]) != 0:
                raise ValueError(f"{i}: No empty space after build directive!")
            n = 2
        if not b[n].startswith("// Copyright (C) 2020"):
            raise ValueError(f"{i}: Missing copyright!")
        if b[n + 1] != "//":
            raise ValueError(f"{i}: Missing black space after copyright!")
        if not b[n + 2].startswith("// This program is free software:"):
            raise ValueError(f"{i}: Missing freedom text!")
        del b


if __name__ == "__main__":
    if len(argv) < 2:
        print(f"{argv[0]} <dir>")
        exit(1)
    scan_dir(argv[1])
