#!/usr/bin/python3
# Copyright (C) 2021 - 2022 iDigitalFlame
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
# strip_bin.py <in> <out>
#  Strips named symbols and obvious strings from Golang binaries.
#  MUST have compiled with "-trimpath -ldflags='-w -s'" for this to work 100%.
#

from random import choice
from string import ascii_letters
from sys import argv, stderr, exit

SIG_GO_ID = b' Go build ID: "'
SIG_FOOTER = b"\x00\x00\x00\x00"
SIG_MAKERS = b"\x00\x00internal/"
SIG_BUILD_ID = b"\x00\x00\x00\x00go.buildid\x00"
SIG_VERSION_A = b"\x03\x00\x00\x00\x00\x00\x00\x00go"
SIG_COMMAND_ARGS = b"path\x09command-line-arguments\x0A"
SIG_VARS_START = b"\x50\x00\x00\x00\x00\x00\x00\x06crypto"
SIG_SCRAP = b"\x00\x08\x08\x08\xFF\x00\x00\x00\x00\x03fmt\x00\x00\x00\x00"
# SIG_VARS_END = b"\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x20\x2B\x40"
SIG_VARS_END = (
    b"\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x26\x40\x00"
)
SIG_MAP_START = b"\x40\x00\x00\x00\x00\x00\x20\x00\x00\x00\x00\x00\x00\x00\x00"
SIG_MAP_END = (
    b"\00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00"
    b"\x00\x00\x00\x00\x00\x08\x00\x00\x00\x00\x00\x00\x00\x08\x00\x00\x00\x00"
)


def is_abc(c):
    if c >= ord("A") and c <= ord("Z"):
        return True
    return c >= ord("a") and c <= ord("z")


def rem_go_id(data):
    s = data.find(SIG_GO_ID)
    if s < 0:
        raise ValueError("Cannot find 'Go build ID' signature!")
    e = data.find(34, s + 15)
    if e < s or e - s < 20:
        raise ValueError("Invalid 'Go build ID' signature!")
    print(
        f'Filling build ID: "{data[s : e + 1].decode("UTF-8", "ignore")}" at {s:X}:{e:X}!'
    )
    for x in range(s, e + 1):
        data[x] = ord(choice(ascii_letters))
    return e + 1


def print_bytes(data):
    return f'[{" ".join([hex(v) for v in data])}]'


def rem_scrap(data, index=0):
    s = data.find(SIG_SCRAP, index)
    if s < index:
        raise ValueError("Could not find scrap signature!")
    i = s + 10
    while True:
        if data[i] == 64 and data[i + 1] == 0:
            break
        if is_abc(data[i]):
            data[i] = ord(choice(ascii_letters))
        i += 1
    print(f"Found scrap offsets at {s:X}:{i:X}, filling..")


def strip_binary(input, out):
    with open(input, "rb") as f:
        b = bytearray(f.read())
    i = rem_go_id(b)
    rem_command_args(b, i)
    rem_scrap(b, i)
    rem_version(b, i)
    n = dict()
    v = dict()
    map_swap(b, SIG_MAP_START, SIG_MAP_END, i, 15, v, n)
    map_swap(b, SIG_VARS_START, SIG_VARS_END, i, 8, v, n)
    fill(b, i, SIG_BUILD_ID, SIG_FOOTER, ord(choice(ascii_letters)), 4, "build")
    try:
        fill(b, i, SIG_MAKERS, SIG_FOOTER, ord(choice(ascii_letters)), 2, "markers")
    except ValueError as err:
        raise ValueError(f'{err}: Did you compile with "-trimpath?')
    with open(out, "wb") as f:
        f.write(b)


def rem_version(data, index=0):
    s = data.find(SIG_VERSION_A, index)
    if s < index:
        raise ValueError("Could not find version signature!")
    print(
        f'Found version "{data[s+8:s+17].decode("UTF-8", "ignore")}" offsets at {s+8:X}:{s+17:X}, randomizing..'
    )
    for x in range(s + 8, s + 17):
        data[x] = ord(choice(ascii_letters))


def rem_command_args(data, index=0):
    s = data.find(SIG_COMMAND_ARGS, index)
    if s < index:
        raise ValueError("Could not find command args signature!")
    while True:
        i = data.find(b"\x0A", s + 1)
        print(f"Found command-line offsets at {s:X}:{i:X}, filling..")
        for x in range(s + 1, i):
            if not is_abc(data[x]):
                continue
            data[x] = ord(choice(ascii_letters))
        if data[i + 1] != 0x6D and data[i + 1] != 0x64:
            break
        s = i


def fill(data, index, start, end, fill, skip=None, name="data"):
    s = data.find(start, index)
    if s < index:
        raise ValueError(
            f'Could not find start value "{print_bytes(start)}" after offset {index:2X}!'
        )
    e = data.find(end, s + len(start))
    if e < s:
        raise ValueError(
            f'Could not find end value "{print_bytes(end)}" after offset {s+len(start):2X}!'
        )
    print(f"Found {name} offsets at {s:X}:{e:X}, filling with {fill}..")
    if not isinstance(skip, int):
        skip = len(start)
    for x in range(s + skip, e):
        if not is_abc(data[x]):
            continue
        data[x] = fill
    return e + len(end)


def map_swap(data, start, end, index=0, skip=None, vars=dict(), names=dict()):
    s = data.find(start, index)
    if s < index:
        raise ValueError(f'Could not find map start value "{print_bytes(start)}"!')
    e = data.find(end, s + len(start))
    if e < s:
        raise ValueError(f'Could not find map end value "{print_bytes(end)}"!')
    z = 0
    if not isinstance(skip, int):
        skip = len(start)
    while True:
        while not is_abc(data[s]):
            s += 1
            if s > e:
                return
        z = s + 1
        while is_abc(data[z]):
            z += 1
            if z >= e:
                break
        if z - s >= 3:
            v = data[s:z].decode("UTF-8")
            if v not in vars:
                d = ""
                while True:
                    d = "".join(choice(ascii_letters) for i in range(z - s))
                    if d not in names:
                        names[d] = True
                        break
                vars[v] = d
                print(f'Mapped "{v}" => "{d}".')
            h = vars[v]
            i = 0
            for x in range(s, z):
                data[x] = ord(h[i])
                i += 1
        s = z + 1


if __name__ == "__main__":
    if len(argv) != 3:
        print(f"{argv[0]} <file> <out>", file=stderr)
        exit(1)

    strip_binary(argv[1], argv[2])
