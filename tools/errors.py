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

from sys import argv


class ErrorValue(object):
    __slots__ = ("pkg", "text", "ident")

    def __init__(self, text, pkg=None, ident=None):
        self.pkg = pkg
        self.text = text
        self.ident = ident

    def __str__(self):
        if not isinstance(self.ident, str):
            return self.text
        if not isinstance(self.pkg, str):
            return f"{self.text} ({self.ident})"
        return f"{self.text} ({self.pkg}.{self.ident})"


ERRORS = [None] * 0xFF
ERRORS[0x00] = ErrorValue("invalid/unknown error")
ERRORS[0x01] = ErrorValue("unspecified error")
ERRORS[0x10] = ErrorValue("empty or invalid Guardian name", "man")
ERRORS[0x11] = ErrorValue("no paths found", "man", "ErrNoEndpoints")
ERRORS[0x12] = ErrorValue("invalid path/name", "man")
ERRORS[0x13] = ErrorValue("invalid link type", "man")
ERRORS[0x14] = ErrorValue("update without a service status handle", "device/winapi/svc")
ERRORS[0x15] = ErrorValue("unexpected key size", "device/winapi/registry")
ERRORS[0x16] = ErrorValue("unexpected key type", "device/winapi/registry")
ERRORS[0x17] = ErrorValue("invalid env value", "device/winapi")
ERRORS[0x18] = ErrorValue("cannot load DLL function", "device/winapi")
ERRORS[0x19] = ErrorValue("base is not a valid DOS header", "device/winapi")
ERRORS[0x1A] = ErrorValue("offset base is not a valid NT header", "device/winapi")
ERRORS[0x1B] = ErrorValue("header does not represent a DLL", "device/winapi")
ERRORS[0x1C] = ErrorValue("header has an invalid first entry point", "device/winapi")
ERRORS[0x1D] = ErrorValue("cannot find '.text' section", "device/evade")
ERRORS[0x1E] = ErrorValue("invalid address value", "device")
ERRORS[0x1F] = ErrorValue("quit", "device", "ErrQuit")
ERRORS[0x20] = ErrorValue("only supported on Windows devices", "device", "ErrNoWindows")
ERRORS[0x21] = ErrorValue("only supported on *nix devices", "device", "ErrNoNix")
ERRORS[0x22] = ErrorValue("cannot dump self", "device")
ERRORS[0x23] = ErrorValue("buffer limit reached", "data", "ErrLimit")
ERRORS[0x24] = ErrorValue("invalid buffer type", "data", "ErrInvalidType")
ERRORS[0x25] = ErrorValue("invalid index", "data", "ErrInvalidIndex")
ERRORS[0x26] = ErrorValue("buffer is too large", "data", "ErrTooLarge")
ERRORS[0x27] = ErrorValue("invalid whence", "data")
ERRORS[0x28] = ErrorValue(
    "block size must be a power of two between 16 and 128", "data/crypto"
)
ERRORS[0x29] = ErrorValue("block size must equal IV size", "data/crypto")
ERRORS[0x2A] = ErrorValue("malformed Tag", "com", "ErrMalformedTag")
ERRORS[0x2B] = ErrorValue("tags list is too large", "com", "ErrTagsTooLarge")
ERRORS[0x2C] = ErrorValue("packet ID does not match the supplied ID", "com")
ERRORS[0x2D] = ErrorValue(
    "invalid or missing TLS certificates", "com", "ErrInvalidTLSConfig"
)
ERRORS[0x2E] = ErrorValue("invalid permissions", "com/pipe")
ERRORS[0x2F] = ErrorValue("invalid permissions size", "com/pipe")
ERRORS[0x30] = ErrorValue("empty host field", "com/wc2")
ERRORS[0x31] = ErrorValue("invalid port specified", "com/wc2")
ERRORS[0x32] = ErrorValue("invalid HTTP response", "com/wc2")
ERRORS[0x33] = ErrorValue("body is not writable", "com/wc2")
ERRORS[0x34] = ErrorValue("could not get underlying net.Conn", "com/wc2")
ERRORS[0x35] = ErrorValue("not a file", "com/wc2")
ERRORS[0x36] = ErrorValue("not a directory", "com/wc2")
ERRORS[0x37] = ErrorValue("stdout already set", "cmd")
ERRORS[0x38] = ErrorValue("stderr already set", "cmd")
ERRORS[0x39] = ErrorValue("stdin already set", "cmd")
ERRORS[0x3A] = ErrorValue("process has not started", "cmd", "ErrNotStarted")
ERRORS[0x3B] = ErrorValue("process arguments are empty", "cmd", "ErrEmptyCommand")
ERRORS[0x3C] = ErrorValue("process still running", "cmd", "ErrStillRunning")
ERRORS[0x3D] = ErrorValue("process already started", "cmd", "ErrAlreadyStarted")
ERRORS[0x3E] = ErrorValue(
    "could not find a suitable process", "cmd/filter", "ErrNoProcessFound"
)
ERRORS[0x3F] = ErrorValue("empty or nil Host", "c2", "ErrNoHost")
ERRORS[0x40] = ErrorValue("other side did not come up", "c2", "ErrNoConn")
ERRORS[0x41] = ErrorValue("empty or nil Profile", "c2", "ErrInvalidProfile")
ERRORS[0x42] = ErrorValue("first Packet is invalid", "c2")
ERRORS[0x43] = ErrorValue("empty or invalid pipe name", "c2")
ERRORS[0x44] = ErrorValue("no Profile parser loaded", "c2")
ERRORS[0x45] = ErrorValue("unexpected OK value", "c2")
ERRORS[0x46] = ErrorValue("empty or nil Packet", "c2", "ErrMalformedPacket")
ERRORS[0x47] = ErrorValue("not a Listener", "c2", "ErrNotAListener")
ERRORS[0x48] = ErrorValue("not a Connector", "c2", "ErrNotAConnector")
ERRORS[0x49] = ErrorValue("unable to listen", "c2")
ERRORS[0x4A] = ErrorValue("empty Listener name", "c2")
ERRORS[0x4B] = ErrorValue("listener already exists", "c2")
ERRORS[0x4C] = ErrorValue("send buffer is full", "c2", "ErrFullBuffer")
ERRORS[0x4D] = ErrorValue(
    "frag/multi total is zero on a frag/multi packet", "c2", "ErrInvalidPacketCount"
)
ERRORS[0x4E] = ErrorValue("must be a client session", "c2")
ERRORS[0x4F] = ErrorValue("migration in progress", "c2")
ERRORS[0x50] = ErrorValue("cannot marshal Profile", "c2")
ERRORS[0x51] = ErrorValue("cannot marshal Proxy data", "c2")
ERRORS[0x52] = ErrorValue("packet ID does not match the supplied ID", "c2")

ERRORS[0x53] = ErrorValue("proxy support disabled", "c2")
ERRORS[0x54] = ErrorValue("cannot marshal Proxy Profile", "c2")
ERRORS[0x55] = ErrorValue("only a single Proxy per session can be active", "c2")
ERRORS[0x56] = ErrorValue(
    "frag/multi count is larger than 0xFFFF", "c2", "ErrTooManyPackets"
)
ERRORS[0x57] = ErrorValue("received Packet that does not match our own device ID", "c2")
ERRORS[0x58] = ErrorValue("no Job created for client Session", "c2", "ErrNoTask")
ERRORS[0x59] = ErrorValue("empty or nil Job", "c2")
ERRORS[0x5A] = ErrorValue("cannot assign a Job ID", "c2")
ERRORS[0x5B] = ErrorValue("job already registered", "c2")
ERRORS[0x5C] = ErrorValue("empty or nil Tasklet", "c2")
ERRORS[0x5D] = ErrorValue("setting is invalid", "c2/cfg", "ErrInvalidSetting")
ERRORS[0x5E] = ErrorValue(
    "cannot add multiple transforms", "c2/cfg", "ErrMultipleTransforms"
)
ERRORS[0x5F] = ErrorValue(
    "cannot add multiple connections", "c2/cfg", "ErrMultipleConnections"
)
ERRORS[0x60] = ErrorValue("binary source not available", "c2/cfg")
ERRORS[0x61] = ErrorValue('missing "type" string', "c2/cfg")
ERRORS[0x62] = ErrorValue("key not found", "c2/cfg")
ERRORS[0x63] = ErrorValue("mapping ID is invalid", "c2/task")
ERRORS[0x64] = ErrorValue("mapping ID is already exists", "c2/task")
ERRORS[0x65] = ErrorValue("empty host field", "c2/task")
ERRORS[0x66] = ErrorValue("invalid port specified", "c2/task")
ERRORS[0x67] = ErrorValue("invalid HTTP response", "c2/task")
ERRORS[0x68] = ErrorValue("invalid operation", "c2/task")
ERRORS[0x69] = ErrorValue("invalid Packet", "c2/task")
ERRORS[0x6A] = ErrorValue("empty or nil Tasklet", "c2/task")
ERRORS[0x6B] = ErrorValue("script is empty", "c2/task")
ERRORS[0x6C] = ErrorValue("empty key name", "c2/task")
ERRORS[0x6D] = ErrorValue("empty value name", "c2/task")
ERRORS[0x6E] = ErrorValue("arguments cannot be nil or empty", "c2/wrapper")
ERRORS[0xFE] = ErrorValue("invalid Task mapping", "c2")


def find_error(v):
    if isinstance(v, int):
        if v < 0 or v > 0xFF:
            return None
        return ERRORS[v]
    try:
        n = int(v, base=0)
        if n < 0 or n > 0xFF:
            return None
        return ERRORS[n]
    except ValueError:
        pass
    return None


if __name__ == "__main__":
    if len(argv) == 2:
        print(find_error(argv[1]))
    else:
        print(f"{argv[0]} <error_code>")
