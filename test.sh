#!/usr/bin/bash

build_tags=(
              bugs implant crypt stdrand nojson nosweep no6 tiny small medium large nofrag regexp nopanic noservice ews noproxy nokeyswap scripts
              bugs,implant
              bugs,implant,crypt
              bugs,implant,crypt,stdrand
              bugs,implant,crypt,stdrand,nojson
              bugs,implant,crypt,stdrand,nojson,nosweep
              bugs,implant,crypt,stdrand,nojson,nosweep,no6
              bugs,implant,crypt,stdrand,nojson,nosweep,no6,tiny
              bugs,implant,crypt,stdrand,nojson,nosweep,no6,small
              bugs,implant,crypt,stdrand,nojson,nosweep,no6,medium
              bugs,implant,crypt,stdrand,nojson,nosweep,no6,nofrag
              bugs,implant,crypt,stdrand,nojson,nosweep,no6,nofrag,regexp
              bugs,implant,crypt,stdrand,nojson,nosweep,no6,nofrag,regexp,ews
              bugs,implant,crypt,stdrand,nojson,nosweep,no6,nofrag,regexp,ews,noproxy
              implant,crypt,stdrand,nojson,nosweep,no6,nofrag,regexp,ews,noproxy
              crypt,stdrand,nojson,nosweep,no6,nofrag,regexp,ews,noproxy
              crypt,stdrand,nojson,nosweep,no6,nofrag,regexp,noproxy
              stdrand,nojson,nosweep,no6,nofrag,regexp,ews,noproxy
              nojson,nosweep,no6,nofrag,regexp,ews,noproxy
              nosweep,no6,nofrag,regexp,ews,noproxy
              nosweep,no6,nofrag,ews,noproxy
              nosweep,no6,nofrag,noproxy
              nosweep,no6,nofrag,ews
              no6,nofrag,regexp,ews
              no6,nofrag,regexp
              no6,nofrag,ews
              no6,nofrag
           )

go mod tidy
if [ $? -ne 0 ]; then
    printf "\x1b[1m\x1b[31m[!] Tidying modules failed!\x1b[0m\n"
    exit 1
fi

go mod verify
if [ $? -ne 0 ]; then
    printf "\x1b[1m\x1b[31m[!] Verifying modules failed!\x1b[0m\n"
    exit 1
fi

run_vet() {
    printf "\x1b[1m\x1b[36m[+] Vetting GOARCH=$1 GOOS=$2..\x1b[0m\n"
    env GOARCH=$1 GOOS=$2 go vet ./c2
    env GOARCH=$1 GOOS=$2 go vet ./cmd
    env GOARCH=$1 GOOS=$2 go vet ./com
    env GOARCH=$1 GOOS=$2 go vet ./data
    env GOARCH=$1 GOOS=$2 go vet ./device 2>&1 | grep -vE 'github.com/iDigitalFlame/xmt/device$|device/y_nix.go:84:24: possible misuse of unsafe.Pointer$'
    env GOARCH=$1 GOOS=$2 go vet ./man
    env GOARCH=$1 GOOS=$2 go vet ./util
}
run_vet_all() {
    for entry in $(go tool dist list); do
        run_vet "$(echo $entry | cut -d '/' -f 2)" "$(echo $entry | cut -d '/' -f 1)"
    done
}
run_staticcheck() {
    printf "\x1b[1m\x1b[33m[+] Static Check GOARCH=$1 GOOS=$2..\x1b[0m\n"
    env GOARCH=$1 GOOS=$2 staticcheck -checks all -f text ./... | grep -vE '^tests/|^unit_tests/' | grep -v '(ST1000)'
    for tags in ${build_tags[@]}; do
        printf "\x1b[1m\x1b[34m[+] Static Check GOARCH=$1 GOOS=$2 with tags \"${tags}\"..\x1b[0m\n"
        env GOARCH=$1 GOOS=$2 staticcheck -checks all -f text -tags $tags ./... | grep -vE '^tests/|^unit_tests/' | grep -vE '(ST1000)|(ST1003)'
    done
}
run_staticcheck_all() {
    for entry in $(go tool dist list); do
        run_staticcheck "$(echo $entry | cut -d '/' -f 2)" "$(echo $entry | cut -d '/' -f 1)"
    done
}

run_vet_all
run_staticcheck_all
