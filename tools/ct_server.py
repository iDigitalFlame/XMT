#!/usr/bin/python
# Copyright (C) 2020 - 2022 iDigitalFlame
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

from json import loads
from sys import argv, stderr, exit
from socketserver import TCPServer
from http.server import SimpleHTTPRequestHandler
from os.path import expanduser, expandvars, isfile


class Server(object):
    __slots__ = ("_dir", "_entries")

    def __init__(self, file, dir):
        self._dir = dir
        with open(expanduser(expandvars(file))) as f:
            self._entries = loads(f.read())
        if not isinstance(self._entries, dict):
            raise ValueError(f'file "{file}" does not contain a JSON dict')
        for k, v in self._entries.items():
            if not isinstance(k, str) or len(k) == 0:
                raise ValueError("invalid JSON key")
            if not isinstance(v, dict):
                raise ValueError(f'invalid JSON value for "{k}"')
            if "file" not in v or "type" not in v:
                raise ValueError(f'missing values for JSON value "{k}"')
            if not isinstance(v["file"], str) or not isinstance(v["type"], str):
                raise ValueError(f'invalid value types for JSON value "{k}"')
            p = expandvars(expanduser(v["file"]))
            if not isfile(p):
                raise ValueError(f'path "{p}" for JSON value "{k}" does not exist')
            v["file"] = p
            del p

    def start(self, addr, port):
        with TCPServer((addr, port), self._request) as h:
            h.serve_forever()

    def _request(self, req, address, server):
        return _WebRequest(self._entries, self._dir, req, address, server)


class _WebRequest(SimpleHTTPRequestHandler):
    def __init__(self, entries, base, req, address, server):
        self._entries = entries
        SimpleHTTPRequestHandler.__init__(
            self, request=req, client_address=address, server=server, directory=base
        )

    def do_GET(self):
        e = self._entries.get(self.path.lower())
        if not isinstance(e, dict):
            return super(__class__, self).do_GET()
        try:
            with open(e["file"], "rb") as f:
                b = f.read()
        except OSError as err:
            print(f"Server read error {self.path}: {err}", file=stderr)
            return self.send_error(500, "Server error")
        else:
            self.send_response(200)
            self.send_header("Content-type", e["type"])
        finally:
            del e
        self.send_header("Content-Length", len(b))
        self.end_headers()
        self.wfile.write(b)


if __name__ == "__main__":
    if len(argv) < 5:
        print(f"{argv[0]} <config> <dir> <addr> <port>", file=stderr)
        exit(1)
    try:
        Server(argv[1], argv[2]).start(argv[3], int(argv[4]))
    except KeyboardInterrupt:
        print("Interrupted!", file=stderr)
        exit(1)
    except Exception as err:
        print(f"Error: {err}!", file=stderr)
        exit(1)
