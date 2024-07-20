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

# crypt_builder.py
# Generates a crypt mapping string or build based off the build tags supplied.

from io import BytesIO
from json import loads
from os import environ
from subprocess import run
from secrets import token_bytes
from sys import stderr, exit, argv
from base64 import urlsafe_b64encode
from platform import system, architecture


def xor(key, data):
    if len(data) == 0 or len(key) == 0:
        return bytearray()
    r = bytearray(len(data))
    for i in range(0, len(r)):
        r[i] = data[i] ^ key[i % len(key)]
    return r


def check_tags(args):
    for x in range(0, len(args)):
        if args[x] == "-tags":
            if "crypt" not in [k.strip() for k in args[x + 1].split(",")]:
                args[x + 1] = f"crypt,{args[x + 1].strip()}"
            return args
    return args + ["-tags", "crypt"]


def get_env_tags(args):
    if "GOOS" in environ:
        o = environ["GOOS"]
    else:
        o = system().lower()
    if "GOARCH" in environ:
        a = environ["GOARCH"]
    else:
        v = architecture()
        if v[0] == "64bit":
            a = "amd64"
        elif v[0] == "32bit":
            a = "386"
        else:
            a = v[0].lower()
        del v
    t = [o, a]
    del o, a
    for x in range(0, len(args)):
        if args[x] == "-tags":
            for e in args[x + 1].split(","):
                v = e.strip()
                if v not in t:
                    t.append(v)
                del v
            break
    return t


def use_tag(current, tags):
    if not isinstance(tags, list) or not isinstance(current, list) or len(current) == 0:
        return True
    n = True
    for t in tags:
        if t[0] != "!":
            n = False
            break
    # Check negatives first
    for t in tags:
        if t[0] != "!":
            continue
        if t[1:] in current:
            return False
    # Check positives
    for t in tags:
        if t[0] == "!":
            continue
        if t in current:
            return True
    return n


class CryptWriter(BytesIO):
    __slots__ = ("key",)

    def __init__(self, key=None):
        BytesIO.__init__(self)
        if isinstance(key, str) and len(key) > 0:
            self.key = key.encode("UTF-8")
        elif isinstance(key, (bytes, bytearray)):
            self.key = key
        else:
            self.key = token_bytes(64)

    def add(self, v):
        self.write(v.encode("UTF-8"))
        self.write(bytearray(1))

    def output(self):
        return urlsafe_b64encode(xor(self.key, self.getvalue())).decode("UTF-8")

    def key_output(self):
        return urlsafe_b64encode(self.key).decode("UTF-8")

    def from_file(self, f, tags):
        with open(f, "r") as b:
            d = loads(b.read())
        if not isinstance(d, dict) or len(d) == 0:
            return
        c = [None] * len(d)
        for k, v in d.items():
            if not isinstance(v, dict) or "value" not in v:
                c[int(k)] = ""
                continue
            if not use_tag(tags, v.get("tags")):
                c[int(k)] = ""
            else:
                c[int(k)] = v["value"]
        print("Including strings in build...")
        for x in range(0, len(c)):
            self.add(c[x])
            if len(c[x]) == 0:
                continue
            s = c[x].replace("\n", "\\n")
            print(f'+ {x:3} == "{s}"')
            del s


if __name__ == "__main__":
    if len(argv) < 3:
        print(f"{argv[0]} [-no-build] <file> [go build args]", file=stderr)
        exit(2)

    n = False
    if argv[1] == "-no-build":
        if len(argv) < 4:
            print(f"{argv[0]} [-no-build] <file> [go build args]", file=stderr)
            exit(2)
        n = True
        argv.remove("-no-build")

    w = CryptWriter()
    w.from_file(argv[1], get_env_tags(argv[2:]))

    if n:
        print("Add this to ldflags:\n")
        print(
            f"-X 'github.com/iDigitalFlame/xmt/util/crypt.key={w.key_output()}'", end=""
        )
        print(f" -X 'github.com/iDigitalFlame/xmt/util/crypt.payload={w.output()}'")
        exit(0)

    a = [
        "go",
        "build",
        "-ldflags",
        f"-w -s -X "
        f"'github.com/iDigitalFlame/xmt/util/crypt.key={w.key_output()}'"
        f" -X 'github.com/iDigitalFlame/xmt/util/crypt.payload={w.output()}'",
    ]
    a.extend(check_tags(argv[2:]))
    run(a, env=environ, check=True, text=True, capture_output=True)
    print()
    del a
