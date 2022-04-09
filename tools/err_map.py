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


class ErrorType(object):
    def __init__(self, name, refs=None, messages=None):
        self.name = name
        self.refs = refs
        self.messages = messages


ERRORS = {
    0: ErrorType("unknown error"),
    1: ErrorType("script error"),
    2: ErrorType(
        "feature disabled",
        messages=[
            "json disabled",
            "proxy support disabled",
            "only a single Proxy per session can be active",
        ],
    ),
    3: ErrorType(
        "unexpected value",
        messages=[
            "could not get underlying net.Conn",
        ],
    ),
    4: ErrorType("migration in progress"),
    5: ErrorType(
        "precondition failed",
        [
            "c2.ErrNoTask",
            "cmd.ErrNotStarted",
        ],
        [
            "migration in progress",
            "must be a client session",
            "cannot be a client session",
            "process has not been started",
            "no Job created for client Session",
        ],
    ),
    7: ErrorType(
        "buffer is full",
        ["c2.ErrFullBuffer"],
        [
            "cannot assign a Job ID",
        ],
    ),
    8: ErrorType(
        "unable to complete action",
        [
            "c2.ErrNotAListener",
            "c2.ErrNotAConnector",
        ],
        [
            "not a listener",
            "not a connector",
            "unable to listen",
            "stdin already set",
            "stdout already set",
            "stderr already set",
            "migration in progress",
            "other side did not come up",
            "binary source not available",
        ],
    ),
    9: ErrorType(
        "empty or nil value",
        [
            "c2.ErrNoHost",
            "c2.ErrInvalidProfile",
            "c2.ErrMalformedPacket",
            "com.ErrInvalidTLSConfig",
            "filter.ErrNoProcessFound",
        ],
        [
            "empty DLL name",
            "empty host field",
            "empty or nil Host",
            "empty or nil Profile",
            "empty or nil Job",
            "empty or nil Tasklet",
            "empty or nil Packet",
            "empty Listener name",
            "empty Proxy name",
            "missing TLS certificates",
            "cannot find '.text' section",
            "arguments cannot be nil or empty",
            "could not find a suitable process",
        ],
    ),
    10: ErrorType("invalid name"),
    11: ErrorType("invalid path"),
    12: ErrorType(
        "invalid type",
        messages=[
            "cannot marshal Profile",
            "cannot marshal Proxy Profile",
            "body is not writable",
        ],
    ),
    13: ErrorType(
        "invalid value",
        ["com.ErrMalformedTag"],
        [
            "key not found",
            "malformed Tag",
            "invalid env value",
            "whence is invalid",
            "setting is invalid",
            "invalid permissions",
            "unexpected key type",
            'missing "type" string',
            "invalid port specified",
            "packet ID does not match the supplied ID",
            "received Packet that does not match our own ID",
        ],
    ),
    14: ErrorType(
        "invalid number",
        [
            "c2.ErrTooManyPackets",
            "c2.ErrInvalidPacketCount",
        ],
        [
            "mapping ID is invalid",
            "frag/multi count is larger than 0xFFFF",
            "frag/multi total is zero on a frag/multi packet",
        ],
    ),
    15: ErrorType(
        "invalid response",
        messages=[
            "invalid HTTP response",
            "first Packet is invalid",
        ],
    ),
    16: ErrorType(
        "invalid size",
        ["com.ErrTagsTooLarge"],
        [
            "unexpected key size",
            "tags list is too large",
            "invalid permission size",
            "block size must equal IV size",
            "maximum (255) Proxy limit reached",
            "block size must be between 16 and 128 and a power of two",
        ],
    ),
    19: ErrorType(
        "not found",
        ["c2.ErrNoEndpoints"],
        ["no paths found"],
    ),
    20: ErrorType(
        "not yet ready",
        ["cmd.ErrStillRunning"],
        [
            "process is still running",
        ],
    ),
    21: ErrorType(
        "already exists",
        ["cmd.ErrAlreadyStarted"],
        [
            "job ID is in use",
            "proxy already exists",
            "listener already exists",
            "mapping ID is already exists",
            "process has already been started",
        ],
    ),
    23: ErrorType(
        "invalid argument",
        [
            "cmd.ErrEmptyCommand",
            "cfg.ErrMultipleTransforms",
            "cfg.ErrMultipleConnections",
        ],
        [
            "process arguments are empty",
            "cannot add multiple transforms",
            "cannot add multiple connections",
        ],
    ),
    24: ErrorType("not a file"),
    25: ErrorType("not a directory"),
    250: ErrorType("windows only", ["device.ErrNoWindows"]),
    215: ErrorType("only supported on *nix devices", ["device.ErrNoNix"]),
}

for k, v in ERRORS.items():
    print(f"Error: 0x{k:X}: {v.name}")
    if not isinstance(v.messages, list) or len(v.messages) == 0:
        continue
    for n in v.messages:
        print(f" - {n}")
