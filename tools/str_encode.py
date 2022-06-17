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

from sys import argv, exit, stderr


def xor(key: bytes, data: bytes) -> bytearray:
    if len(data) == 0 or len(key) == 0:
        return bytearray()
    r = bytearray(len(data))
    for i in range(0, len(r)):
        r[i] = data[i] ^ key[i % len(key)]
    return r


if __name__ == "__main__":
    if len(argv) != 3:
        print("XMT String Encoder\nEncodes a string payload\n", file=stderr)
        print(f"{argv[0]} <key> <string>", file=stderr)
        exit(2)

    try:
        key = argv[1].encode("UTF-8")
    except UnicodeDecodeError as err:
        print(f'Error encoding key "{argv[1]}": {err}!', file=stderr)
        exit(1)

    try:
        data = argv[2].encode("UTF-8")
    except UnicodeDecodeError as err:
        print(f'Error encoding string "{argv[2]}": {err}!', file=stderr)
        exit(1)

    out = xor(key, data)
    del key, data

    print(
        "Paste this data into your Golang code, then call 'crypto.UnwrapString(key, value)'\n",
        file=stderr,
    )
    print("var value = []byte{", end="")
    for i in range(0, len(out)):
        if i == 0 or i % 20 == 0:
            if i % 20 == 0 and i != 0:
                print(",", end="")
            print("\n\t", end="")
        else:
            print(", ", end="")
        print(f"0x{('%X' % out[i]).zfill(2)}", end="")
    print(",\n}")
    del out
