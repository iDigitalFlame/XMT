#!/usr/bin/python3

from glob import glob


#
# 1. Read in file data
# 2. Build package list.
# 3. Find exported functions (find "func _")
# 4. Map them
#

files = dict()  # file name = {exports, data}
packages = list()  # package names


def add_file(p):
    with open(p, "r") as f:
        b = f.read()
    if len(b) == 0 or "package" not in b:
        raise ValueError(f'invalid Go file "{p}"')
    v = None
    for i in b.split("\n"):
        if i.startswith("package "):
            v = i[8:].strip().lower()
    if v is None or len(v) == 0:
        raise ValueError(f'invalid Go file "{p}"')
    if v not in packages:
        packages.append(v)
    if p in files:
        raise ValueError(f'duplicate Go file "{p}"')
    e = list()
    for i in b.split("\n"):
        if not i.startswith("func "):
            continue
        if i[5] != "(":
            if not i[5].isupper():
                continue
            n = i.find("(", 5)
            if n == -1:
                continue
            e.append(i[5:n].strip())
            continue
        n = i.find(")", 5)
        if n == -1:
            continue
        x = i.find("(", n + 1)
        z = i[n + 1 : x].strip()
        if not z[0].isupper():
            continue
        e.append(z)
        print(z)
    files[p] = (list(), v, b)


for i in glob("**/**.go", recursive=True):
    if i.startswith("tests/") or i.startswith("unit_tests/") or "/" not in i:
        continue
    add_file(i)
