#!/usr/bin/python3
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
# strip_bin.py <in> <out>
#  Strips named symbols and obvious strings from Golang binaries.
#  MUST have compiled with "-trimpath -ldflags='-w -s'" for this to work 100%.
#
# This file is superseded via the "strip_binary" function in ThunderStorm's
# "strip_binary" function. (I might backport it here).
#
# Keep this in check with JetStream's crypt.py file.

from os import argv
from sys import stderr
from re import compile
from random import choice
from string import ascii_letters


PACKAGES = [
    b"\x00bufio",
    b"\x00bytes",
    b"\x00compress",
    b"\x00container",
    b"\x00context",
    b"\x00crypto",
    b"\x00debug",
    b"\x00encoding",
    b"\x00errors",
    b"\x00expvar",
    b"\x00flag",
    b"\x00fmt",
    b"\x00go",
    b"\x00hash",
    b"\x00image",
    b"\x00index",
    b"\x00internal",
    b"\x00io",
    b"\x00log",
    b"\x00math",
    b"\x00mime",
    b"\x00net",
    b"\x00os",
    b"\x00path",
    b"\x00reflect",
    b"\x00regexp",
    b"\x00runtime",
    b"\x00sort",
    b"\x00strconv",
    b"\x00strings",
    b"\x00sync",
    b"\x00syscall",
    b"\x00text",
    b"\x00time",
    b"\x00unicode",
    b"\x00unsafe",
]

CRUMB_ID = b' Go build ID: "'
CRUMB_VER = b"go1."
CRUMB_INF = b"\xFF Go buildinf:"
CRUMB_TAIL = b"/src/runtime/runtime.go\x00\x00"
CRUMB_DEPS = b"command-line-arguments"
CRUMB_FILE = b".go\x00"
CRUMB_STACK = [
    b"\x00github.com/",
    b"github.com/",
    b"\x00type..eq.",
    b"\x00type..hash.",
    b"\x00type:.eq.",
    b"\x00type:.hash.",
    b"\x00go.buildid",
    b"vendor/",
    b"\x00vendor/",
    b"struct {",
    b"map[",
    b"*map.",
    b"*func(",
    b"func(",
    b"\x00main.",
    b'asn1:"',
]
CRUMB_CLEANUP = [
    b"crypto/internal/",
    b"internal/syscall/",
    b"Mingw-w64 runtime failure:",
]
CRUMB_HEADERS = [
    b"\x00\x00\x00\x00\x05bufio",
    b"\x00\x00\x00\x00\x05bytes",
    b"\x00\x00\x00\x00\x06crypto",
    b"\x00\x00\x00\x00\x06crypto",
    b"\x00\x00\x00\x00\x06error",
    b"\x00\x00\x00\x00\x0Dcrypto/",
    b"\x00\x00\x00\x00\x0Dcompress/",
    b"\x00\x00\x00\x00\x0Ecompress/",
    b"\x00\x00\x00\x00\x0Econtainer/",
]

TABLE_IMPORTS = compile(
    b"([\x01-\x05]{0,1})([\x01-\x50]{1})([a-zA-Z0-9/\\-\\*]{2,50})\x00"
)
TABLE_STRINGS = compile(
    b"([\x01-\x50]{1})([a-zA-Z0-9/\\-\\*]{2,50})\\.([a-zA-Z0-9\\./\\[\\]\\]\\*]{4,50})([\x00-\x01]{1})"
)


def _is_valid(c):
    return (
        (c >= 0x41 and c <= 0x5A)
        or (c >= 0x61 and c <= 0x7A)
        or (c >= 0x30 and c <= 0x39)
        or (c >= 0x2C and c <= 0x2F)
        or c == 0x7B
        or c == 0x7D
        or c == 0x5B
        or c == 0x5D
        or c == 0x5F
        or c == 0x3B
        or c == 0x20
        or c == 0x28
        or c == 0x29
        or c == 0x2A
    )


def _is_ext(b, i):
    # Ignore anything that isn't "fmt.fmt"
    return (
        b[i - 5] == 0x2E and b[i - 4] != 0x66 and b[i - 3] != 0x6D and b[i - 2] != 0x74
    )


def _find_header(b):
    for i in CRUMB_HEADERS:
        x = b.find(i)
        if x > 0:
            return x
    return -1


def _mask_cleanup(b):
    x = 0
    while True:
        m = TABLE_STRINGS.search(b, x)
        if m is None:
            break
        if _is_ext(b, m.end()):
            x = m.end()
            continue
        print(b[m.start() : m.end() - 1])
        for x in range(m.start() + 2, m.end() - 1):
            b[x] = ord(choice(ascii_letters))
        x = m.end()
    x = 0
    for i in CRUMB_CLEANUP:
        p, x = 0, 0
        if i[0] == 0:
            p += 1
        while x < len(b):
            x = b.find(i)
            if x <= 0:
                break
            _fill_non_zero(b, x + p)
        del p, x


def _mask_build_id(b):
    x = b.find(CRUMB_ID)
    if x <= 0:
        return
    for i in range(x + 1, x + 128):
        if b[i] == 0x22 and b[i + 1] < 0x31:
            b[i] = 0
            break
        b[i] = 0
    del x


def _mask_build_inf(b):
    x = b.find(CRUMB_INF)
    if x > 0:
        for i in range(x + 2, x + 14):
            b[i] = 0
    x = 0
    while x < len(b):
        x = b.find(CRUMB_FILE, x + 1)
        if x <= 0 or b[x - 1] == 0:
            break
        s = x
        while s > x - 128:
            if b[s] == 0:
                break
            s -= 1
        if x - s < 64 and x - s > 0:
            for i in range(s, x):
                b[i] = 0
        del s
    del x


def _mask_tail(b, log):
    for i in CRUMB_STACK:
        p, x = 0, 0
        if i[0] == 0:
            p += 1
        while x < len(b):
            x = b.find(i)
            if x <= 0:
                break
            _fill_non_zero(b, x + p, real=True)
        del p, x
    if callable(log):
        log("Removing unused package names..")
    for i in range(0, len(PACKAGES)):
        x = 0
        if callable(log) and i % 10 == 0:
            log(f"Checking package {i+1} out of {len(PACKAGES)}..")
        while x < len(b):
            x = b.find(PACKAGES[i], x)
            if x <= 0:
                break
            if b[x + 3] == 0:
                x += 2
                continue
            _fill_non_zero(b, x + 1)
        del x
    if callable(log):
        log("Looking for path values..")
    x = b.find(CRUMB_TAIL)
    if x <= 0:
        return
    while x > 0:
        if b[x] == 0 and b[x - 1] != 0 and b[x - 1] < 0x21 and b[x - 2] != 0:
            if callable(log):
                log(f"Found start of paths at 0x{x:X}")
            break
        x -= 1
    if x <= 0:
        return
    c = 0
    while x < len(b):
        if b[x] == 0:
            c += 1
            x += 1
            continue
        if c > 3:
            break
        x, c = _fill_non_zero(b, x), 0
    del x, c


def _mask_tables(b, log):
    for _ in range(0, 2):
        _mask_tables_inner(b, log)


def _mask_deps(b, start, log):
    x = start
    for r in range(0, 2):
        if callable(log):
            log(f"Searching for command line args (round {r+1})..")
        x = b.find(CRUMB_DEPS, start)
        if x <= 0:
            continue
        x -= 5
        while x < len(b):
            if b[x] < 0x21 and b[x + 1] > 0x7E:
                break
            if b[x] > 0x21:
                b[x] = 0
            x += 1
    del x


def _mask_tables_inner(b, log):
    x = -1
    for i in CRUMB_HEADERS:
        x = b.find(i)
        if x > 0:
            break
    if x <= 0:
        return
    c = 0
    while True:
        s, e = _find_next_vtr(b, x)
        if s == -1:
            break
        if (e - s) > 3:
            for i in range(s, e):
                # NOTE(dij): These MUST be random chars or the program will CRASH!!
                b[i] = ord(choice(ascii_letters))
        x = e
        c += 1
    del x
    if callable(log):
        log(f"Masked {c} stack trace strings!")
    del c


def _fill_non_zero(b, start, max=256, real=False):
    for i in range(start, start + max):
        if b[i] == 0 or (real and b[i] < 32):
            return i
        b[i] = 0
    return start + max


def _mask_version(b, start, root=None, path=None):
    if root is not None and len(root) > 0:
        m = root.encode("UTF-8")
        x = start + 1
        while x > start:
            x = b.find(m, x)
            if x == -1:
                break
            _fill_non_zero(b, x)
        del x
    if path is not None and len(path) > 0:
        m = path.encode("UTF-8")
        x = start + 1
        while x > start:
            x = b.find(m, x)
            if x == -1:
                break
            _fill_non_zero(b, x)
        del x
    x = b.find(CRUMB_VER, start)
    while x > start and x < len(b):
        b[x], b[x + 1], b[x + 2] = 0, 0, 0
        x += 3
        for i in range(0, 16):
            x += len(CRUMB_VER) + i
            if b[x] < 0x2E or b[x] < 0x39:
                break
            b[x] = 0
        x = b.find(CRUMB_VER)
    del x


def _find_next_vtr(b, start, zeros=32, max_len=128):
    # Golang string identifiers are usually BB<str>.
    # Two bytes, the first one idk what it means, it's usually 1|0 (but not always)
    # and then a byte that specifies how long the following string is.
    #
    # We use this to our advantage by reading this identifier and discounting anything
    # that A). doesn't fit Go name conventions and anything that doesn't equal the
    # supplied length. This allows us to scroll through the virtual string table
    # (vtr).
    i, z, c, s = 0, 0, 0, 0
    for i in range(start, start + max_len):
        if z > zeros:
            break
        if s > start:
            if c > 0 and (i - s) > c:
                s, c = 0, 0
                continue
            if b[i] == 0 or not _is_valid(b[i]):
                if (b[i] == 0x3A and b[i + 1] == 0x22) or (
                    b[i] == 0x22 and b[i - 1] == 0x3A
                ):  # Find usage of :"
                    continue
                if b[i] == 0x22:  # Scroll back to find the missing "
                    q = i - 1
                    while q > start and q > q - 64:
                        if b[q] == 0x22:
                            break
                        q -= 1
                    if b[q] == 0x22:
                        continue
                    del q
                if (
                    b[i] == 0x3C
                    and b[i + 1] == 0x2D
                    and (b[i - 1] == 0x6E or b[i + 2] == 0x63)
                ):  # Check for <-chan or chan<-
                    continue
                if i - s != c:
                    s, c = 0, 0
                    continue
                return i - (i - s), i
            continue
        if b[i] == 0:
            z += 1
            continue
        else:
            z = 0
        if s == 0 and b[i] >= 0x30 and b[i] <= 0x39:
            # No valid identifiers start with a number.
            continue
        if s == 0 and _is_valid(b[i]) and b[i - 1] != 0:
            s, c = i, b[i - 1]
    del z, c, s
    return -1, i


def strip_binary(file, out, log=None, root=None, path=None):
    with open(file, "rb") as f:
        b = bytearray(f.read())
    x = _find_header(b)
    if x <= 0:
        if callable(log):
            log("Could not find header (are you using Garble?)")
        return
    _mask_tables(b, log)
    _mask_tail(b, log)
    _mask_deps(b, x, log)
    if callable(log):
        log("Removing version info..")
    _mask_version(b, x, root, path)
    _mask_build_id(b)
    _mask_build_inf(b)
    if callable(log):
        log("Cleaning up..")
    _mask_cleanup(b)
    del x
    with open(out, "wb") as f:
        f.write(b)
    del b


if __name__ == "__main__":
    if len(argv) != 3:
        print(f"{argv[0]} <file> <out>", file=stderr)
        exit(1)

    strip_binary(argv[1], argv[2], print)
