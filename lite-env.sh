#!/bin/bash -e

echo "Making a Golang GOROOT clone to nuke stupid depdencies.."

GOROOT_LITE="/tmp/go-root-lite"

rsync -ar "${GOROOT}/" "$GOROOT_LITE"
rm ${GOROOT_LITE}/src/fmt/*.go
rm ${GOROOT_LITE}/src/unicode/*.go

echo "Downloaded patched source.."
git clone "https://github.com/iDigitalFlame/TinyPatchedGo" "${GOROOT_LITE}/patches"

echo 'Removing bloated "fmt" and "unicode"..'
mv ${GOROOT_LITE}/patches/fmt/*.go  "${GOROOT_LITE}/src/fmt/"
mv ${GOROOT_LITE}/patches/unicode/*.go  "${GOROOT_LITE}/src/unicode/"

sed -ie 's/return envProxyFunc()(req.URL)/return req.URL, nil/g' "${GOROOT_LITE}/src/net/http/transport.go"
sed -ie 's/envProxyFuncValue = httpproxy.FromEnvironment().ProxyFunc()/envProxyFuncValue = nil/g' "${GOROOT_LITE}/src/net/http/transport.go"
sed -ie 's/"golang.org\/x\/net\/http\/httpproxy"//g' "${GOROOT_LITE}/src/net/http/transport.go"
rm "${GOROOT_LITE}/src/net/http/transport.goe"
sed -ie 's/if a, err := idna.ToASCII(host); err == nil {/if a := ""; len(a) > 0 {/g' "${GOROOT_LITE}/src/net/http/h2_bundle.go"
sed -ie 's/"golang.org\/x\/net\/idna"//g' "${GOROOT_LITE}/src/net/http/h2_bundle.go"
rm "${GOROOT_LITE}/src/net/http/h2_bundle.goe"
sed -ie 's/return idna.Lookup.ToASCII(v)/return v, nil/g' "${GOROOT_LITE}/src/net/http/request.go"
sed -ie 's/"golang.org\/x\/net\/idna"//g' "${GOROOT_LITE}/src/net/http/request.go"
rm "${GOROOT_LITE}/src/net/http/request.goe"
sed -ie 's/host, err = idna.ToASCII(host)/host, err = host, nil/g' "${GOROOT_LITE}/src/vendor/golang.org/x/net/http/httpguts/httplex.go"
sed -ie 's/"golang.org\/x\/net\/idna"//g' "${GOROOT_LITE}/src/vendor/golang.org/x/net/http/httpguts/httplex.go"
rm "${GOROOT_LITE}/src/vendor/golang.org/x/net/http/httpguts/httplex.goe"
export GOROOT=$GOROOT_LITE

echo "Vendering dependencies.."
rm -rf "vendor"
go mod tidy
go mod verify
go mod vendor

echo "Done."
echo 'Make sure to run "export GOROOT=/tmp/go-root-lite" before starting any builds!'
