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

from io import BytesIO
from shlex import split
from json import dumps, loads
from secrets import token_bytes
from struct import pack, unpack
from traceback import format_exc
from base64 import b64decode, b64encode
from sys import argv, exit, stdin, stderr, stdout
from os.path import isfile, expanduser, expandvars
from argparse import ArgumentParser, BooleanOptionalAction

HELP_TEXT = """XMT man.Sentinel Builder v1 Release

Builds or reads a Sentinel file based on the supplied arguments.
Files can be converted from and to JSON.

NOTE: JSON Files are NOT supported by XMT directly. They are only
to be used for generation.

Usage: {binary} <options>

BASIC ARGUMENTS:
  -h                              Show this help message and exit.
  --help

INPUT/OUTPUT ARGUMENTS:
  -f                <file>        Input/Output file path. Use '-' for
  --file                            stdin/stdout.
  -S                              Force saving the file. Disables JSON
  --save                            and Print. Used mainly for updating the
                                    Filter settings, which may not be
                                    automatically detected.
  -k                              Provide a key string value to encrypt or
  --key                             decrypt the Sentinel output with XOR CFB.
                                    Only valid when reading/writing from a
                                    binary file.
  -K                              Provide a base64 encoded key string value
  --key-b64                         to encrypt or decrypt the Sentinel output
                                    with XOR CFB. Only valid when reading/writing
                                    from a binary file.
  -y                              Provide a path to a file that contains the
  --key-file                        binary data for the key to be used to encrypt
                                    or decrypt the Sentinel output with XOR CFB.
                                    Only valid when reading/writing from a binary
                                    file.
  -j                              Output in JSON format. Omit for raw
  --json                            binary. (Or base64 when output to
                                    stdout.)
  -I                              Accept stdin input as commands. Each
  --stdin                           line from stdin will be treated as a
                                    'append' line to the supplied config.
                                    Input and Output are ignored and are
                                    only set via the command line.
                                    This option disables using stdin for
                                    Sentinel data.

OPERATION ARGUMENTS:
  -p
  --print                         List values contained in the file
                                    input. Fails if no input is found or
                                    invalid. Output format can be modified
                                    using -j/-p.

SENTINEL ARGUMENTS:
  -d                <path>        Supply a path for a file to be used as a DLL
  --dll                             Sentinel entry. The path is not expanded
                                    until the Sentinel is ran.
  --s               <path>        Supply a path for a file to be used as an
  --asm                             Assembly Sentinel entry. The path is not
                                    expanded until the Sentinel is ran.
  -z                <path>        Supply a path for a Assembly or DLL file to
  --zombie                          be used as a Zombie (Hallowed) Sentinel entry.
                                    DLLs will be converted to Assembly by the
                                    Sentinel (if enabled). This requires at
                                    least ONE fake command to be added with '-F'
                                    or '--fake'.
  -c                <command>     Supply a command to be used as a Sentinel entry.
  --command                         Any environment variables will not be expanded
                                    until the Sentinel is ran.
  -u                <url>         Supply a URL to be used as a Sentinel entry.
  --url                             The downloaded target will be executed depending
  --download                        on the resulting 'Content-Type' header.
                                    The '-A' or '--agent' value can be specified
                                    to change the 'User-Agent' header to be used
                                    when downloading.
  -A                <user-agent>  Sets or adds a 'User-Agent' string that can be
  --agent                           used when downloading a URL path. This argument
                                    may be used multiple times to add more User-Agents.
                                    When multiple are present, one is selected at
                                    random. Supports the Text matcher verbs in the
                                    'text' package. See the 'ADDITIONAL RESOURCES'
                                    section for more info.
  -F                <fake-cmd>    Sets or adds the 'Fake' commands line args used
  --fake                            when a Zombie process is started. The first
                                    argument (the target binary) MUST exist. This
                                    argument may be used multiple times to add more
                                    command lines. When multiple are present, one
                                    is selected at random.

FILTER ARGUMENTS:
  -n                <pid>         Specify the PID to use for the Parent Filter.
  --pid                             Takes priority over all other options.
  -i                <name1,nameX> Specify a (comma|space) seperated list of process
  --include                         names to INCLUDE in the Filter search process.
                                    This may be used more than one time in the command.
  -x                <name1,nameX> Specify a (comma|space) seperated list of process
  --exclude                         names to EXCLUDE from the Filter search process.
                                    This may be used more than one time in the command.
  -v                              Enable the search to only ALLOW FOREGROUND or
  --desktop                         DESKTOP processes to be used. Default is don't
                                    care. Takes priority over any disable arguments.
  -V                              Enable the search to only ALLOW BACKGROUND or
  --no-desktop                      SERVICE processes to be used. Default is don't
                                    care.
  -e                              Enable the search to only ALLOW ELEVATED processes
  --admin                           to be used. Default is don't care. Takes priority
  --elevated                        over any disable arguments.
  -E                              Enable the search to only ALLOW NON-ELEVATED
  --no-admin                        processes to be used. Default is don't care.
  --no-elevated
  -r                              Enable the Filter to fallback if no suitable
  --fallback                        processes were found during the first run and
                                    run again with less restrictive settings.
  -R                              Disable the Filter's ability to fallback if no
  --no-fallback                     suitable processes were found during the first
                                    run.

ADDITIONAL RESOURCES:
  Text Matcher Guide
    https://pkg.go.dev/github.com/iDigitalFlame/xmt@v0.3.3/util/text#Matcher
"""


def _copy(d, s, p=0):
    n, i = 0, p
    for x in range(0, len(s)):
        if i >= len(d):
            break
        d[i] = s[x]
        i += 1
        n += 1
    return n


def _join_split(r, v):
    if "," not in v:
        s = v.strip()
        if len(s) == 0:
            return
        return r.append(s)
    for i in v.split(","):
        if len(i) == 0:
            continue
        r.append(i.strip())


def _join(a, split=False):
    if not isinstance(a, list) or len(a) == 0:
        return None
    r = list()
    for i in a:
        if isinstance(i, str):
            if split:
                _join_split(r, i)
                continue
            v = i.strip()
            if len(v) == 0:
                continue
            r.append(v)
            del v
            continue
        if not isinstance(i, list):
            continue
        for x in i:
            if split:
                _join_split(r, x)
                continue
            v = x.strip()
            if len(v) == 0:
                continue
            r.append(v)
            del v
    return r


def _read_file_input(v, k):
    if v.strip() == "-" and not stdin.isatty():
        if hasattr(stdin, "buffer"):
            b = stdin.buffer.read()
        else:
            b = stdin.read()
        stdin.close()
    else:
        p = expandvars(expandvars(v))
        if not isfile(p):
            return Sentinel(), False
        with open(p, "rb") as f:
            b = f.read()
        del p
    if len(b) == 0:
        raise ValueError("input: empty input data")
    return Sentinel(raw=b, key=k), True


def _nes(s, min=0, max=-1):
    if max > min:
        return isinstance(s, str) and len(s) < max and len(s) > min
    return isinstance(s, str) and len(s) > min


def _xor(dst, src, key, o=0):
    n = len(key)
    if n > len(src):
        n = len(src)
    for i in range(0, n):
        dst[i + o] = src[i] ^ key[i]
    return n


def _write_out(s, key, v, pretty, json):
    f = stdout
    if _nes(v) and v != "-":
        if not pretty and not json:
            f = open(v, "wb")
        else:
            f = open(v, "w")
    try:
        if pretty or json:
            return print(
                dumps(s.to_json(), sort_keys=False, indent=(4 if pretty else None)),
                file=f,
            )
        b = s.save(None, key)
        if f == stdout and not f.isatty():
            return f.buffer.write(b)
        if f.mode == "wb":
            return f.write(b)
        f.write(b64encode(b).decode("UTF-8"))
        del b
    finally:
        if f == stdout:
            print(end="")
        else:
            f.close()
        del f


class CTRXor(object):
    __slots__ = ("ctr", "key", "used", "out", "total")

    def __init__(self, iv, key):
        if len(iv) != len(key):
            raise ValueError("key and iv lengths must be equal")
        self.key = key
        self.used = 0
        self.total = 0
        self.ctr = bytearray(len(key))
        k = 512
        if k < len(key):
            k = len(key)
        self.out = bytearray(k)
        del k
        _copy(self.ctr, iv)

    def refill(self):
        r = self.total - self.used
        _copy(self.out, self.out[self.used :])
        self.total = len(self.out)
        n = len(self.key)
        while r <= (self.total - n):
            _xor(self.out, self.ctr, self.key, r)
            r += n
            for x in range(len(self.ctr) - 1, -1, -1):
                v = self.ctr[x] + 1
                if v >= 0x100:
                    self.ctr[x] = v - 0x100
                else:
                    self.ctr[x] = v
                del v
                if self.ctr[x] != 0:
                    break
        self.total, self.used = r, 0
        del n

    def xor(self, dst, src):
        if len(dst) < len(src):
            raise ValueError("output smaller than input")
        x, y = 0, 0
        while x < len(src):
            if self.used >= self.total - len(self.out):
                self.refill()
            n = _xor(dst, src[x:], self.out[self.used :], y)
            x += n
            y += n
            self.used += n
        del x, y


class Reader(object):
    __slots__ = ("r",)

    def __init__(self, r):
        self.r = r

    def read_str(self):
        return self.read_bytes().decode("UTF-8")

    def read_bool(self):
        return self.read_uint8() == 1

    def read_bytes(self):
        t = self.read_uint8()
        if t == 0:
            return bytearray(0)
        n = 0
        if t == 1:
            n = self.read_uint8()
        elif t == 3:
            n = self.read_uint16()
        elif t == 5:
            n = self.read_uint32()
        elif t == 7:
            n = self.read_uint64()
        else:
            raise ValueError("read_bytes: invalid buffer type")
        if n < 0:
            raise ValueError("read_bytes: invalid buffer size")
        b = self.r.read(n)
        del t, n
        return b

    def read_uint8(self):
        return unpack(">B", self.r.read(1))[0]

    def read_uint16(self):
        return unpack(">H", self.r.read(2))[0]

    def read_uint32(self):
        return unpack(">I", self.r.read(4))[0]

    def read_uint64(self):
        return unpack(">Q", self.r.read(8))[0]

    def read_str_list(self):
        t = self.read_uint8()
        if t == 0:
            return list()
        n = 0
        if t == 1:
            n = self.read_uint8()
        elif t == 3:
            n = self.read_uint16()
        elif t == 5:
            n = self.read_uint32()
        elif t == 7:
            n = self.read_uint64()
        else:
            raise ValueError("invalid buffer type")
        if n < 0:
            raise ValueError("invalid list size")
        r = list()
        for _ in range(0, n):
            r.append(self.read_str())
        del t, n
        return r


class Writer(object):
    __slots__ = ("w",)

    def __init__(self, w):
        self.w = w

    def write_str(self, v):
        if v is None:
            return self.write_bytes(None)
        if not isinstance(v, str):
            raise ValueError("write: not a string")
        self.write_bytes(v.encode("UTF-8"))

    def write_bool(self, v):
        self.write_uint8(1 if v else 0)

    def write_bytes(self, v):
        if v is None or len(v) == 0:
            return self.write_uint8(0)
        if not isinstance(v, (bytes, bytearray)):
            raise ValueError("write: not a bytes type")
        n = len(v)
        if n < 0xFF:
            self.write_uint8(1)
            self.write_uint8(n)
        elif n < 0xFFFF:
            self.write_uint8(3)
            self.write_uint16(n)
        elif n < 0xFFFFFFFF:
            self.write_uint8(5)
            self.write_uint32(n)
        else:
            self.write_uint8(7)
            self.write_uint64(n)
        self.w.write(v)
        del n

    def write_uint8(self, v):
        if not isinstance(v, int):
            raise ValueError("write: not a number")
        self.w.write(pack(">B", v))

    def write_uint16(self, v):
        if not isinstance(v, int):
            raise ValueError("write: not a number")
        self.w.write(pack(">H", v))

    def write_uint32(self, v):
        if not isinstance(v, int):
            raise ValueError("write: not a number")
        self.w.write(pack(">I", v))

    def write_uint64(self, v):
        if not isinstance(v, int):
            raise ValueError("write: not a number")
        self.w.write(pack(">Q", v))

    def write_str_list(self, v):
        if v is None or len(v) == 0:
            return self.write_uint8(0)
        if not isinstance(v, list):
            raise ValueError("not a list")
        n = len(v)
        if n < 0xFF:
            self.write_uint8(1)
            self.write_uint8(n)
        elif n < 0xFFFF:
            self.write_uint8(3)
            self.write_uint16(n)
        elif n < 0xFFFFFFFF:
            self.write_uint8(5)
            self.write_uint32(n)
        else:
            self.write_uint8(7)
            self.write_uint64(n)
        del n
        for i in v:
            self.write_str(i)


class ReadCTR(object):
    __slots__ = ("r", "ctr")

    def __init__(self, ctr, r):
        self.r = r
        self.ctr = ctr

    def read(self, n):
        r = self.r.read(n)
        if r is None:
            return None
        b = bytearray(len(r))
        self.ctr.xor(b, r)
        del r
        return b


class WriteCTR(object):
    __slots__ = ("w", "ctr")

    def __init__(self, ctr, w):
        self.w = w
        self.ctr = ctr

    def write(self, b):
        if not isinstance(b, (bytes, bytearray)):
            raise ValueError("write: not a bytes type")
        r = bytearray(len(b))
        self.ctr.xor(r, b)
        self.w.write(r)
        del r


class Filter(object):
    __slots__ = ("pid", "session", "exclude", "include", "elevated", "fallback")

    def __init__(self, json=None):
        self.pid = 0
        self.session = None
        self.exclude = None
        self.include = None
        self.elevated = None
        self.fallback = False
        if not isinstance(json, dict):
            return
        self.from_json(json)

    def to_json(self):
        r = dict()
        if isinstance(self.session, bool):
            r["session"] = self.session
        if isinstance(self.elevated, bool):
            r["elevated"] = self.elevated
        if isinstance(self.fallback, bool):
            r["fallback"] = self.fallback
        if isinstance(self.pid, int) and self.pid > 0:
            r["pid"] = self.pid
        if isinstance(self.exclude, list) and len(self.exclude) > 0:
            r["exclude"] = self.exclude
        if isinstance(self.include, list) and len(self.include) > 0:
            r["include"] = self.include
        return r

    def read(self, r):
        if not r.read_bool():
            return
        self.pid = r.read_uint32()
        self.fallback = r.read_bool()
        b = r.read_uint8()
        if b == 0:
            self.session = None
        elif b == 1:
            self.session = False
        elif b == 2:
            self.session = True
        b = r.read_uint8()
        if b == 0:
            self.elevated = None
        elif b == 1:
            self.elevated = False
        elif b == 2:
            self.elevated = True
        self.exclude = r.read_str_list()
        self.include = r.read_str_list()

    def is_empty(self):
        if isinstance(self.session, bool):
            return False
        if isinstance(self.elevated, bool):
            return False
        if isinstance(self.pid, int) and self.pid > 0:
            return False
        if isinstance(self.exclude, list) and len(self.exclude) > 0:
            return False
        if isinstance(self.include, list) and len(self.include) > 0:
            return False
        return True

    def write(self, w):
        if self is None or self.is_empty():
            return w.write_bool(False)
        w.write_bool(True)
        w.write_uint32(self.pid)
        w.write_bool(self.fallback)
        if self.session is True:
            w.write_uint8(2)
        elif self.session is False:
            w.write_uint8(1)
        else:
            w.write_uint8(0)
        if self.elevated is True:
            w.write_uint8(2)
        elif self.elevated is False:
            w.write_uint8(1)
        else:
            w.write_uint8(0)
        w.write_str_list(self.exclude)
        w.write_str_list(self.include)

    def from_json(self, d):
        if not isinstance(d, dict):
            raise ValueError("from_json: value provided was not a dict")
        if "session" in d and isinstance(d["session"], bool):
            self.session = d["session"]
        if "elevated" in d and isinstance(d["elevated"], bool):
            self.elevated = d["elevated"]
        if "fallback" in d and isinstance(d["fallback"], bool):
            self.fallback = d["fallback"]
        if "pid" in d and isinstance(d["pid"], int) and d["pid"] > 0:
            self.pid = d["pid"]
        if "exclude" in d and isinstance(d["exclude"], list) and len(d["exclude"]) > 0:
            self.exclude = d["exclude"]
            for i in self.exclude:
                if isinstance(i, str) and len(i) > 0:
                    continue
                raise ValueError('from_json: empty or non-string value in "exclude"')
        if "include" in d and isinstance(d["include"], list) and len(d["include"]) > 0:
            self.include = d["include"]
            for i in self.include:
                if isinstance(i, str) and len(i) > 0:
                    continue
                raise ValueError('from_json: empty or non-string value in "include"')


class Sentinel(object):
    __slots__ = ("paths", "filter")

    def __init__(self, raw=None, file=None, key=None, json=None):
        self.paths = None
        self.filter = Filter()
        if _nes(file):
            return self.load(file, key)
        if isinstance(raw, (str, bytes, bytearray)):
            return self.from_raw(raw, key)
        if not isinstance(json, dict):
            return
        self.from_json(json)

    def to_json(self):
        return {
            "filter": self.filter.to_json(),
            "paths": [i.to_json() for i in self.paths],
        }

    def read(self, r):
        self.filter.read(r)
        n = r.read_uint16()
        if n < 0:
            raise ValueError("invalid entry size")
        self.paths = list()
        for _ in range(0, n):
            self.paths.append(SentinelPath(reader=r))
        del n

    def write(self, w):
        self.filter.write(w)
        if not isinstance(self.paths, list) or len(self.paths) == 0:
            return w.write_uint16(0)
        w.write_uint16(len(self.paths))
        for x in range(0, min(len(self.paths), 0xFFFF)):
            self.paths[x].write(w)

    def from_json(self, j):
        if not isinstance(j, dict):
            raise ValueError("from_json: value provided was not a dict")
        if "filter" in j:
            self.filter.from_json(j["filter"])
        if "paths" not in j or not isinstance(j["paths"], list) or len(j["paths"]) == 0:
            return
        self.paths = list()
        for i in j["paths"]:
            self.paths.append(SentinelPath.from_json(i))

    def add_dll(self, path):
        if not _nes(path):
            raise ValueError('add_dll: "path" must be a non-empty string')
        if self.paths is None:
            self.paths = list()
        self.paths.append(SentinelPath(type=SentinelPath.DLL, path=path))

    def add_asm(self, path):
        if not _nes(path):
            raise ValueError('add_asm: "path" must be a non-empty string')
        if self.paths is None:
            self.paths = list()
        self.paths.append(SentinelPath(type=SentinelPath.ASM, path=path))

    def add_execute(self, cmd):
        if not _nes(cmd):
            raise ValueError('add_execute: "path" must be a non-empty string')
        if self.paths is None:
            self.paths = list()
        self.paths.append(SentinelPath(type=SentinelPath.EXECUTE, path=cmd))

    def save(self, path, key=None):
        k = key
        if _nes(key):
            k = key.encode("UTF-8")
        elif key is not None and not isinstance(key, (bytes, bytearray)):
            raise ValueError("save: key must be a string or bytes type")
        b = BytesIO()
        if k is not None:
            i = token_bytes(len(k))
            b.write(i)
            o = WriteCTR(CTRXor(i, k), b)
            del i
        else:
            o = b
        w = Writer(o)
        del o
        self.write(w)
        del w
        r = b.getvalue()
        b.close()
        del b
        if not _nes(path):
            return r
        with open(expanduser(expandvars(path)), "wb") as f:
            f.write(r)
        del r

    def add_zombie(self, path, fakes):
        if not _nes(path):
            raise ValueError('add_zombie: "path" must be a non-empty string')
        if not isinstance(fakes, (str, list)) or len(fakes) == 0:
            raise ValueError(
                'add_zombie: "fakes" must be a non-empty string or string list'
            )
        if self.paths is None:
            self.paths = list()
        if isinstance(fakes, str):
            fakes = [fakes]
        self.paths.append(
            SentinelPath(type=SentinelPath.ZOMBIE, path=path, extra=fakes)
        )

    def from_raw(self, data, key=None):
        if isinstance(data, str) and len(data) > 0:
            if data[0] == "{" and data[-1].strip() == "}":
                return self.from_json(loads(data))
            return self.load(None, key=key, buf=b64decode(data, validate=True))
        if isinstance(data, (bytes, bytearray)) and len(data) > 0:
            if data[0] == 91 and data.decode("UTF-8", "ignore").strip()[-1] == "]":
                return self.from_json(loads(data.decode("UTF-8")))
            return self.load(None, key=key, buf=data)
        raise ValueError("from_raw: a bytes or string type is required")

    def load(self, path, key=None, buf=None):
        k = key
        if _nes(key):
            k = key.encode("UTF-8")
        elif key is not None and not isinstance(key, (bytes, bytearray)):
            raise ValueError("load: key must be a string or bytes type")
        if isinstance(buf, (bytes, bytearray)):
            b = BytesIO(buf)
        elif isinstance(buf, BytesIO):
            b = buf
        else:
            b = open(expanduser(expandvars(path)), "rb")
        if k is not None:
            i = b.read(len(k))
            o = ReadCTR(CTRXor(i, k), b)
            del i
        else:
            o = b
        r = Reader(o)
        del o
        try:
            self.read(r)
        finally:
            b.close()
            del b
        del r

    def add_download(self, url, agents=None):
        if not _nes(url):
            raise ValueError('add_download: "url" must be a non-empty string')
        if agents is not None and not isinstance(agents, (str, list)):
            raise ValueError('add_download: "agents" must be a string or string list')
        if self.paths is None:
            self.paths = list()
        if isinstance(agents, str):
            agents = [agents]
        self.paths.append(
            SentinelPath(type=SentinelPath.DOWNLOAD, path=url, extra=agents)
        )


class SentinelPath(object):
    EXECUTE = 0
    DLL = 1
    ASM = 2
    DOWNLOAD = 3
    ZOMBIE = 4

    __slots__ = ("type", "path", "extra")

    def __init__(self, reader=None, type=None, path=None, extra=None):
        self.type = type
        self.path = path
        self.extra = extra
        if not isinstance(reader, Reader):
            return
        self.read(reader)

    def valid(self):
        if self.type > SentinelPath.ZOMBIE or not self.path:
            return False
        if self.type > SentinelPath.DOWNLOAD and not self.extra:
            return False
        return True

    @staticmethod
    def from_json(d):
        if not isinstance(d, dict):
            raise ValueError("from_json: value provided was not a dict")
        if "type" not in d or "path" not in d:
            raise ValueError("from_json: invalid JSON data")
        t = d["type"]
        if not _nes(t):
            raise ValueError('from_json: invalid "type" value')
        v = t.lower()
        del t
        p = d["path"]
        if not _nes(p):
            raise ValueError('from_json: invalid "path" value')
        s = SentinelPath()
        s.path = p
        del p
        if v == "execute":
            s.type = SentinelPath.EXECUTE
        elif v == "dll":
            s.type = SentinelPath.DLL
        elif v == "asm":
            s.type = SentinelPath.ASM
        elif v == "download":
            s.type = SentinelPath.DOWNLOAD
        elif v == "zombie":
            s.type = SentinelPath.ZOMBIE
        else:
            raise ValueError('from_json: unknown "type" value')
        del v
        if s.type < SentinelPath.DOWNLOAD:
            return s
        if "extra" not in d:
            if s.type == SentinelPath.DOWNLOAD:
                return s
            raise ValueError('from_json: missing "extra" value')
        e = d["extra"]
        if not isinstance(e, list) or len(e) == 0:
            if s.type == SentinelPath.DOWNLOAD:
                return s
            raise ValueError('from_json: invalid "extra" value')
        for i in e:
            if _nes(i):
                continue
            raise ValueError('from_json: invalid "extra" sub-value')
        s.extra = e
        del e
        return s

    def to_json(self):
        if self.type > SentinelPath.ZOMBIE:
            raise ValueError("to_json: invalid path type")
        if self.type < SentinelPath.DOWNLOAD or (
            not isinstance(self.extra, list) and not self.extra
        ):
            return {"type": self.typename(), "path": self.path}
        return {"type": self.typename(), "path": self.path, "extra": self.extra}

    def read(self, r):
        self.type = r.read_uint8()
        self.path = r.read_str()
        if self.type < SentinelPath.DOWNLOAD:
            return
        self.extra = r.read_str_list()

    def __str__(self):
        if self.type == SentinelPath.EXECUTE:
            return f"Execute: {self.path}"
        if self.type == SentinelPath.DLL:
            return f"DLL: {self.path}"
        if self.type == SentinelPath.ASM:
            return f"ASM: {self.path}"
        if self.type == SentinelPath.DOWNLOAD:
            if isinstance(self.extra, list) and len(self.extra) > 0:
                return f'Download: {self.path} (Agents: {", ".join(self.extra)})'
            return f"Download: {self.path}"
        if self.type == SentinelPath.ZOMBIE:
            if isinstance(self.extra, list) and len(self.extra) > 0:
                return f'Zombie: {self.path} (Fakes: {", ".join(self.extra)})'
            return f"Zombie: {self.path}"
        return "Unknown"

    def write(self, w):
        w.write_uint8(self.type)
        w.write_str(self.path)
        if self.type < SentinelPath.DOWNLOAD:
            return
        w.write_str_list(self.extra)

    def typename(self):
        if self.type == SentinelPath.EXECUTE:
            return "execute"
        if self.type == SentinelPath.DLL:
            return "dll"
        if self.type == SentinelPath.ASM:
            return "asm"
        if self.type == SentinelPath.DOWNLOAD:
            return "download"
        if self.type == SentinelPath.ZOMBIE:
            return "zombie"
        return "invalid"


class _Builder(ArgumentParser):
    def __init__(self):
        ArgumentParser.__init__(self, description="XMT man.Sentinel Tool")
        self.add_argument("-j", "--json", dest="json", action="store_true")
        self.add_argument("-p", "--print", dest="print", action="store_true")
        self.add_argument("-I", "--stdin", dest="stdin", action="store_true")
        self.add_argument("-f", "--file", type=str, dest="file")
        self.add_argument("-k", "--key", type=str, dest="key")
        self.add_argument("-y", "--key-file", type=str, dest="key_file")
        self.add_argument("-K", "--key-b64", type=str, dest="key_base64")
        self.add_argument("-d", "--dll", dest="dll", action="store_true")
        self.add_argument("-s", "--asm", dest="asm", action="store_true")
        self.add_argument("-S", "--save", dest="save", action="store_true")
        self.add_argument("-z", "--zombie", dest="zombie", action="store_true")
        self.add_argument("-c", "--command", dest="command", action="store_true")
        self.add_argument(
            "-u", "--url", "--download", dest="download", action="store_true"
        )
        self.add_argument(nargs="*", type=str, dest="path")
        self.add_argument(
            "-A",
            "-F",
            "--fake",
            "--agent",
            nargs="*",
            type=str,
            dest="extra",
            action="append",
        )
        self.add_argument("-n", "--pid", type=int, dest="pid")
        self.add_argument("-V", dest="no_desktop", action="store_false")
        self.add_argument(
            "-v", "--desktop", dest="desktop", action=BooleanOptionalAction
        )
        self.add_argument("-R", dest="no_fallback", action="store_false")
        self.add_argument(
            "-r", "--fallback", dest="fallback", action=BooleanOptionalAction
        )
        self.add_argument("-E", dest="no_admin", action="store_false")
        self.add_argument(
            "-e",
            "--admin",
            "--elevated",
            dest="admin",
            action=BooleanOptionalAction,
        )
        self.add_argument(
            "-x",
            "--exclude",
            nargs="*",
            type=str,
            dest="exclude",
            action="append",
        )
        self.add_argument(
            "-i",
            "--include",
            nargs="*",
            type=str,
            dest="include",
            action="append",
        )

    def run(self):
        a, k = self.parse_args(), None
        if _nes(a.key_file):
            with open(expanduser(expandvars(a.key_file)), "rb") as f:
                k = f.read()
        elif _nes(a.key_base64):
            k = b64decode(a.key_base64, validate=True)
        elif _nes(a.key):
            k = a.key.encode("UTF-8")
        if a.file:
            s, z = _read_file_input(a.file, k)
        else:
            s, z = Sentinel(), False
        if a.stdin and a.file != "-":
            if stdin.isatty():
                raise ValueError("stdin: no input found")
            if hasattr(stdin, "buffer"):
                b = stdin.buffer.read().decode("UTF-8")
            else:
                b = stdin.read()
            stdin.close()
            for v in b.split("\n"):
                _Builder.build(s, super(__class__, self).parse_args(split(v)))
        elif (
            isinstance(a.path, list)
            and len(a.path) > 0
            and (not a.print or (a.print and not a.file))
        ):
            _Builder.build(s, a)
        elif not z:
            raise ValueError("no paths added to an empty Sentinel")
        elif not a.save and (a.print or a.json):
            a.file = None
        elif a.save:
            _Builder._parse_filter(s.filter, a)
        if not isinstance(s.paths, list) or len(s.paths) == 0:
            return
        if a.save:
            a.print, a.json = False, False
        _write_out(s, k, a.file, a.print, a.json)
        del s, z, a, k

    @staticmethod
    def build(s, a):
        _Builder._parse_filter(s.filter, a)
        if not isinstance(a.path, list) or len(a.path) == 0:
            return
        if a.command:
            s.add_execute(" ".join(a.path))
        elif a.dll:
            s.add_dll(" ".join(a.path))
        elif a.asm:
            s.add_asm(" ".join(a.path))
        elif a.download:
            s.add_download(" ".join(a.path), agents=_join(a.extra))
        elif a.zombie:
            s.add_zombie(" ".join(a.path), _join(a.extra))
        else:
            s.add_execute(" ".join(a.path))
        _Builder._parse_filter(s.filter, a)

    def parse_args(self):
        if len(argv) <= 1:
            return self.print_help()
        return super(__class__, self).parse_args()

    @staticmethod
    def _parse_filter(f, a):
        if isinstance(a.pid, int):
            if a.pid <= 0:
                f.pid = None
            else:
                f.pid = a.pid
        if a.exclude is not None and len(a.exclude) > 0:
            f.exclude = _join(a.exclude, True)
        if a.include is not None and len(a.include) > 0:
            f.include = _join(a.include, True)
        if not a.no_admin or a.admin is not None:
            f.elevated = (a.admin is None and a.no_admin) or (
                a.admin is True and a.no_admin
            )
        if not a.no_desktop or a.desktop is not None:
            f.session = (a.desktop is None and a.no_desktop) or (
                a.desktop is True and a.no_desktop
            )
        if not a.no_fallback or a.fallback is not None:
            f.fallback = (a.fallback is None and a.no_fallback) or (
                a.fallback is True and a.no_fallback
            )

    def print_help(self, file=None):
        print(HELP_TEXT.format(binary=argv[0]), file=file)
        exit(2)


if __name__ == "__main__":
    try:
        _Builder().run()
    except Exception as err:
        print(f"Error: {err}\n{format_exc(3)}", file=stderr)
        exit(1)
