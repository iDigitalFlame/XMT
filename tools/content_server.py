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

from json import loads
from threading import Lock
from sys import argv, stderr, exit
from socketserver import TCPServer
from watchdog.observers import Observer
from http.server import SimpleHTTPRequestHandler
from os.path import expanduser, expandvars, isfile
from watchdog.events import FileSystemEventHandler, FileModifiedEvent


class Server(FileSystemEventHandler):
    __slots__ = ("_dir", "_entries", "_lock", "_file", "_obs")

    def __init__(self, file, dir):
        self._dir = dir
        self._lock = Lock()
        self._file = expanduser(expandvars(file))
        self._reload()
        self._obs = Observer()
        self._obs.schedule(self, self._file)
        self._obs.start()

    def stop(self):
        if self._obs is None:
            return
        self._obs.stop()
        self._obs.unschedule_all()
        self._obs = None

    def _reload(self):
        print(f'Reading config file "{self._file}"..')
        with open(self._file) as f:
            e = loads(f.read())
        if not isinstance(e, dict):
            raise ValueError(f'file "{self._file}" does not contain a JSON dict')
        for k, v in e.items():
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
        self._entries = None
        self._entries = e
        print(f"Reading config done, {len(self._entries)} loaded.")

    def start(self, addr, port):
        with TCPServer((addr, port), self._request) as h:
            h.serve_forever()

    def on_modified(self, event):
        if not isinstance(event, FileModifiedEvent):
            return
        self._lock.acquire(True)
        try:
            self._reload()
        except Exception as err:
            print(f"Failed to reload config: {err}!", file=stderr)
        finally:
            self._lock.release()

    def _request(self, req, address, server):
        self._lock.acquire(True)
        r = _WebRequest(self._entries, self._dir, req, address, server)
        self._lock.release()
        return r


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
