#!/bin/bash -e

echo "Making a Golang GOROOT clone to nuke stupid depdencies.."

GOROOT_LITE="/tmp/go-root-lite"

rsync -ar "${GOROOT}/" "$GOROOT_LITE"
rm ${GOROOT_LITE}/src/fmt/*.go
rm ${GOROOT_LITE}/src/unicode/*.go

echo "Downloaded patched source.."
git clone "https://github.com/iDigitalFlame/TinyPatchedGo" "${GOROOT_LITE}/patches"

echo 'Removing bloated "fmt" and "unicode"..'
mv ${GOROOT_LITE}/patches/fmt/*.go  ${GOROOT_LITE}/src/fmt/
mv ${GOROOT_LITE}/patches/unicode/*.go  ${GOROOT_LITE}/src/unicode/
export GOROOT=$GOROOT_LITE

echo "Vendering dependencies.."
rm -rf "vendor"
go mod tidy
go mod verify
go mod vendor

echo 'Removing bloated "idna", "norm" and "bidi"..'
rm -rf "vendor/golang.org/x/net/idna"
rm -rf "vendor/golang.org/x/text/unicode/bidi"
rm -rf "vendor/golang.org/x/text/unicode/norm"

echo 'Removing "idna" deps in "httpproxy"..'
sed -ie 's/return idna.Lookup.ToASCII(v)/return v, nil/g' "vendor/golang.org/x/net/http/httpproxy/proxy.go"
sed -ie 's/"golang.org\/x\/net\/idna"//g' "vendor/golang.org/x/net/http/httpproxy/proxy.go"
rm -f "vendor/golang.org/x/net/http/httpproxy/proxy.goe"

mkdir "vendor/golang.org/x/net/idna"
mkdir "vendor/golang.org/x/text/unicode/bidi"
mkdir "vendor/golang.org/x/text/unicode/norm"
printf "package idna\n\n" > "vendor/golang.org/x/net/idna/idna.go"
printf "package bidi\n\n" > "vendor/golang.org/x/text/unicode/bidi/bidi.go"
printf "package norm\n\n" > "vendor/golang.org/x/text/unicode/norm/norm.go"

echo "Done."
echo 'Make sure to run "export GOROOT=/tmp/go-root-lite" before starting any builds!'