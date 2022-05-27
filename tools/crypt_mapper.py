#!/usr/bin/python3

from glob import glob
from json import dumps
from os.path import join
from sys import argv, exit


def _make_tags(v):
    s = v.strip()
    if s == "crypt":
        return None
    e = s.split(" ")
    if len(e) == 0:
        return None
    if len(e) == 1 and e[0] == "crypt":
        return None
    r = list()
    for i in e:
        if i == "||" or i == "&&" or i == "crypt":
            continue
        r.append(i.replace("(", "").replace(")", ""))
    r.sort()
    return r


def _merge_tags(one, two):
    if one is None or two is None:
        return None
    if len(one) == 0 or len(two) == 0:
        return None
    r = one.copy()
    for i in two:
        if i[0] == "!":  # We have a negate
            if i[1:] in r:  # We have a negative but theres a positive in the list
                r.remove(i[1:])  # Remove the positive
                continue  # Skip the negative
            if i in r:
                continue
            r.append(i)  # Add it since we don't care.
            continue
        # It's a positive
        if f"!{i}" in r:  # Is a negative of that in the current?
            r.remove(f"!{i}")  # Remove it
            continue  # Skip the positive, so we can have both
        if i in r:
            continue
        r.append(i)
    if len(r) == 0:
        return None
    return r


class CryptMapper(object):
    def __init__(self):
        self.text = dict()
        self._count = list()

    def _add(self, value, tags):
        t = _make_tags(tags)
        if "\\n" in value:
            value = value.replace("\\n", "\n")
        if value in self.text:
            n = self.text[value]
            n[1] = _merge_tags(n[1], t)
            return n[0]
        i = str(len(self._count))
        self._count.append((i))
        self.text[value] = [i, t]
        return i

    def _index(self, p, d, tags):
        if "crypt.Get(" not in d:
            return None
        s = d.find("crypt.Get(")
        if s <= 0:
            raise ValueError(f'{p}: Invalid crypt entry "{d.strip()}"')
        e = d.find(")", s + 1)
        if e <= s:
            raise ValueError(f'{p}: Invalid crypt entry "{d.strip()}"')
        c = d.find(" // ")
        if c <= e:
            raise ValueError(f'{p}: Crypt entry lacks comment entry "{d.strip()}"')
        return (
            d[:s]
            + "crypt.Get("
            + self._add(d[c + 3 :].strip(), tags)
            + ")"
            + d[e + 1 :]
        )

    def start(self, out, path=""):
        m = list()
        for i in glob(join(path, "**/**.go"), recursive=True):
            if i.startswith("unit_tests") or i.endswith(".test"):
                continue
            with open(i) as f:
                d = f.read()
            if len(d) == 0:
                continue
            if "package main\n" in d:
                continue
            if 'xmt/util/crypt"' not in d:
                continue
            if "crypt.Get(" not in d:
                continue
            m.append((i, d, d.split("\n")))
        if len(m) == 0:
            return
        print(f"Examining {len(m)} files..")
        for i in m:
            r, t = False, None
            f = open(i[0], "w")
            for x in range(0, len(i[2])):
                if t is None and i[2][x].startswith("//go:build"):
                    t = i[2][x][10:].strip().lower()
                o = self._index(i[0], i[2][x], t)
                if o is not None:
                    r = o != i[2][x]
                    f.write(o)
                else:
                    f.write(i[2][x])
                if len(i[2][x]) == 0 and x + 1 >= len(i[2]):
                    continue
                f.write("\n")
                del o
            f.close()
            if not r:
                continue
            print(f'Reformatted "{i[0]}"')
        del m
        z = dict()
        for k, v in self.text.items():
            z[v[0]] = {"value": k}
            if v[1] is not None and len(v[1]) > 0:
                z[v[0]]["tags"] = v[1]
        with open(out, "w") as f:
            f.write(dumps(z, indent=4, sort_keys=False))
        del z


if __name__ == "__main__":
    if len(argv) < 2:
        print(f"{argv[0]} <json> [dir]")
        exit(1)

    d = ""
    if len(argv) == 3:
        d = argv[2]

    v = CryptMapper()
    v.start(argv[1], d)
    del d
    del v
