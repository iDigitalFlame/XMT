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


def add(index, text, pkg=None, ident=None):
    ERRORS[index] = ErrorValue(text, pkg, ident)


add(0x00, "invalid/unknown error")
add(0x01, "unspecified error")
add(0x10, "empty or invalid Guardian name", "man")
add(0x11, "no paths found", "man", "ErrNoEndpoints")
add(0x12, "invalid path/name", "man")
add(0x13, "invalid link type", "man")
add(0x14, "update without a Service status handle", "device/winapi/svc")
add(0x15, "unexpected key size", "device/winapi/registry")
add(0x16, "unexpected key type", "device/winapi/registry")
add(0x17, "invalid env value", "device/winapi")
add(0x18, "cannot load DLL function", "device/winapi")
add(0x19, "base is not a valid DOS header", "device/winapi")
add(0x1A, "offset base is not a valid NT header", "device/winapi")
add(0x1B, "header does not represent a DLL", "device/winapi")
add(0x1C, "header has an invalid first entry point", "device/winapi")
add(0x1D, "cannot find data section", "device/winapi")
add(0x1E, "invalid address value", "device")
add(0x1F, "quit", "device", "ErrQuit")
add(0x20, "only supported on Windows devices", "device", "ErrNoWindows")
add(0x21, "only supported on *nix devices", "device", "ErrNoNix")
add(0x22, "cannot dump self", "device")
add(0x23, "buffer limit reached", "data", "ErrLimit")
add(0x24, "invalid buffer type", "data", "ErrInvalidType")
add(0x25, "invalid index", "data", "ErrInvalidIndex")
add(0x26, "buffer is too large", "data", "ErrTooLarge")
add(0x27, "invalid whence", "data")
add(0x28, "block size must be a power of two between 16 and 128", "data/crypto")
add(0x29, "block size must equal IV size", "data/crypto")
add(0x2A, "malformed Tag", "com", "ErrMalformedTag")
add(0x2B, "tags list is too large", "com", "ErrTagsTooLarge")
add(0x2C, "packet ID does not match the supplied ID", "com")
add(0x2D, "invalid or missing TLS certificates", "com", "ErrInvalidTLSConfig")
add(0x2E, "invalid permissions", "com/pipe")
add(0x2F, "invalid permissions size", "com/pipe")
add(0x30, "empty host field", "com/wc2")
add(0x31, "invalid port specified", "com/wc2")
add(0x32, "invalid HTTP response", "com/wc2")
add(0x33, "body is not writable", "com/wc2")
add(0x34, "could not get underlying net.Conn", "com/wc2")
add(0x35, "not a file", "com/wc2")
add(0x36, "not a directory", "com/wc2")
add(0x37, "stdout already set", "cmd")
add(0x38, "stderr already set", "cmd")
add(0x39, "stdin already set", "cmd")
add(0x3A, "process has not started", "cmd", "ErrNotStarted")
add(0x3B, "process arguments are empty", "cmd", "ErrEmptyCommand")
add(0x3C, "process still running", "cmd", "ErrStillRunning")
add(0x3D, "process already started", "cmd", "ErrAlreadyStarted")
add(0x3E, "could not find a suitable process", "cmd/filter", "ErrNoProcessFound")
add(0x3F, "empty or nil Host", "c2", "ErrNoHost")
add(0x40, "other side did not come up", "c2", "ErrNoConn")
add(0x41, "empty or nil Profile", "c2/cfg", "ErrInvalidProfile")
add(0x42, "first Packet is invalid", "c2")
add(0x43, "empty or invalid pipe name", "c2")
add(0x45, "unexpected OK value", "c2")
add(0x46, "empty or nil Packet", "c2", "ErrMalformedPacket")
add(0x47, "not a Listener", "c2/cfg", "ErrNotAListener")
add(0x48, "not a Connector", "c2/cfg", "ErrNotAConnector")
add(0x49, "unable to listen", "c2")
add(0x4A, "empty Listener name", "c2")
add(0x4B, "listener already exists", "c2")
add(0x4C, "send buffer is full", "c2", "ErrFullBuffer")
add(
    0x4D,
    "frag/multi total is zero on a frag/multi packet",
    "c2",
    "ErrInvalidPacketCount",
)
add(0x4E, "must be a client session", "c2")
add(0x4F, "migration in progress", "c2")
add(0x50, "cannot marshal Profile", "c2")
add(0x51, "cannot marshal Proxy data", "c2")
add(0x52, "packet ID does not match the supplied ID", "c2")
add(0x53, "proxy support disabled", "c2")
add(0x54, "cannot marshal Proxy Profile", "c2")
add(0x55, "only a single Proxy per session can be active", "c2")
add(0x56, "frag/multi count is larger than 0xFFFF", "c2", "ErrTooManyPackets")
add(0x57, "received Packet that does not match our own device ID", "c2")
add(0x58, "no Job created for client Session", "c2", "ErrNoTask")
add(0x59, "empty or nil Job", "c2")
add(0x5A, "cannot assign a Job ID", "c2")
add(0x5B, "job already registered", "c2")
add(0x5C, "empty or nil Tasklet", "c2")
add(0x5D, "setting is invalid", "c2/cfg", "ErrInvalidSetting")
add(0x5E, "cannot add multiple transforms", "c2/cfg", "ErrMultipleTransforms")
add(0x5F, "cannot add multiple connections", "c2/cfg", "ErrMultipleConnections")
add(0x60, "binary source not available", "c2/cfg")
add(0x61, 'missing "type" string', "c2/cfg")
add(0x62, "key not found", "c2/cfg")
add(0x63, "mapping ID is invalid", "c2/task")
add(0x64, "mapping ID is already exists", "c2/task")
add(0x65, "empty host field", "c2/task")
add(0x66, "invalid port specified", "c2/task")
add(0x67, "invalid HTTP response", "c2/task")
add(0x68, "invalid operation", "c2/task")
add(0x69, "invalid Packet", "c2/task")
add(0x6A, "empty or nil Tasklet", "c2/task")
add(0x6B, "script is empty", "c2/task")
add(0x6C, "empty key name", "c2/task")
add(0x6D, "empty value name", "c2/task")
add(0x6E, "arguments cannot be nil or empty", "c2/wrapper")
add(0xFE, "invalid Task mapping", "c2")
add(0x6F, "cannot find function", "device/winapi")
add(0x6F, "function is a forward", "device/winapi")
add(0x70, "invalid StartHour value", "c2")
add(0x71, "invalid StartMin value", "c2")
add(0x72, "invalid EndHour value", "c2")
add(0x73, "invalid EndMin value", "c2")


if __name__ == "__main__":
    if len(argv) == 2:
        print(find_error(argv[1]))
    else:
        print(f"{argv[0]} <error_code>")
