#!/usr/bin/python

from io import BytesIO
from json import loads
from os import environ
from subprocess import Popen
from secrets import token_bytes
from sys import stderr, exit, argv
from base64 import urlsafe_b64encode
from platform import system, architecture


def xor(key, data):
    if len(data) == 0 or len(key) == 0:
        return bytearray()
    r = bytearray(len(data))
    print(f"Encode {len(data)} : {len(key)}")
    for i in range(0, len(r)):
        r[i] = data[i] ^ key[i % len(key)]
    return r


def check_tags(args):
    for x in range(0, len(args)):
        if args[x] == "-tags":
            if "crypt" not in [k.strip() for k in args[x + 1].split(",")]:
                args[x + 1] = f"crypt,{args[x+1].strip()}"
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
    del o
    del a
    for x in range(0, len(args)):
        if args[x] == "-tags":
            for e in args[x + 1].split(","):
                v = e.strip()
                if v not in t:
                    t.append(v)
                del v
            break
    return t


def can_use_tag(tags, values):
    if not isinstance(tags, list) or not isinstance(values, list):
        return True
    if len(values) == 0:
        return True
    r = False
    for t in tags:
        for v in values:
            if t.lower() == v.lower():
                r = True
                break
            if v[0] == "!" and v[1:].lower() == t.lower():
                return False
    return r


class CryptWriter(BytesIO):
    def __init__(self, key=None):
        BytesIO.__init__(self)
        if isinstance(key, str) and len(key) > 0:
            self.key = key.encode("UTF-8")
        elif isinstance(key, bytes) or isinstance(key, bytearray):
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
            if not can_use_tag(tags, v.get("tags")):
                c[int(k)] = ""
            else:
                c[int(k)] = v["value"]
        for x in range(0, len(c)):
            self.add(c[x])
            if len(c[x]) == 0:
                continue
            print(f'+ {x:3} == "{c[x]}"')


if __name__ == "__main__":
    if len(argv) < 3:
        print(f"{argv[0]} <file> [go build args]", file=stderr)
        exit(2)

    w = CryptWriter()
    w.from_file(argv[1], get_env_tags(argv[2:]))

    a = [
        "go",
        "build",
        "-ldflags",
        f"-w -s -X "
        f"'github.com/iDigitalFlame/xmt/util/crypt.key={w.key_output()}'"
        f" -X 'github.com/iDigitalFlame/xmt/util/crypt.payload={w.output()}'",
    ]
    a.extend(check_tags(argv[2:]))

    Popen(a, env=environ)
    print()
