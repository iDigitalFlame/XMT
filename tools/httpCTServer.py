#!/usr/bin/python

from json import loads
from sys import argv, stderr, exit
from socketserver import TCPServer
from http.server import BaseHTTPRequestHandler
from os.path import expanduser, expandvars, isfile


class Server(object):
    __slots__ = ("_entries",)

    def __init__(self, file):
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
        return _WebRequest(self._entries, req, address, server)


class _WebRequest(BaseHTTPRequestHandler):
    def __init__(self, entries, req, address, server):
        self._entries = entries
        BaseHTTPRequestHandler.__init__(self, req, address, server)

    def do_GET(self):
        e = self._entries.get(self.path.lower())
        if not isinstance(e, dict):
            return self.send_error(404, "File not found")
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
    if len(argv) < 4:
        print(f"{argv[0]} <config> <addr> <port>", file=stderr)
        exit(1)
    try:
        Server(argv[1]).start(argv[2], int(argv[3]))
    except KeyboardInterrupt:
        print("Interrupted!", file=stderr)
        exit(1)
    except Exception as err:
        print(f"Error: {err}!", file=stderr)
        exit(1)
