#!/usr/bin/bash

arch_386=( android freebsd linux netbsd openbsd plan9 windows )
arch_amd64=( android darwin dragonfly freebsd illumos ios linux netbsd openbsd plan9 solaris windows )
arch_arm=( android freebsd linux netbsd openbsd plan9 windows )
arch_arm64=( android darwin freebsd ios linux netbsd openbsd windows )
arch_mips=( linux )
arch_mips64=( linux openbsd )
arch_mips64le=( linux )
arch_mipsle=( linux )
arch_ppc64=( aix linux )
arch_ppc64le=( linux )
arch_riscv64=( linux )
arch_s390x=( linux )
arch_wasm=( js )

build_tags=(
              bugs implant crypt stdrand nojson noprotect nosweep no6 tiny small medium nofrag regexp nopanic noservice
              bugs,implant
              bugs,implant,crypt
              bugs,implant,crypt,stdrand
              bugs,implant,crypt,stdrand,nojson
              bugs,implant,crypt,stdrand,nojson,noprotect
              bugs,implant,crypt,stdrand,nojson,noprotect,nosweep
              bugs,implant,crypt,stdrand,nojson,noprotect,nosweep,no6
              bugs,implant,crypt,stdrand,nojson,noprotect,nosweep,no6,tiny
              bugs,implant,crypt,stdrand,nojson,noprotect,nosweep,no6,small
              bugs,implant,crypt,stdrand,nojson,noprotect,nosweep,no6,medium
              bugs,implant,crypt,stdrand,nojson,noprotect,nosweep,no6,nofrag
              bugs,implant,crypt,stdrand,nojson,noprotect,nosweep,no6,nofrag,regexp
              implant,crypt,stdrand,nojson,noprotect,nosweep,no6,nofrag,regexp
              crypt,stdrand,nojson,noprotect,nosweep,no6,nofrag,regexp
              stdrand,nojson,noprotect,nosweep,no6,nofrag,regexp
              nojson,noprotect,nosweep,no6,nofrag,regexp
              noprotect,nosweep,no6,nofrag,regexp
              nosweep,no6,nofrag,regexp
              no6,nofrag,regexp
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
    env GOARCH=$1 GOOS=$2 go vet ./device
    env GOARCH=$1 GOOS=$2 go vet ./man
    env GOARCH=$1 GOOS=$2 go vet ./util
}
run_vet_all() {
    for os in ${arch_ppc64[@]}; do
        run_vet ppc64 $os
    done
    for os in ${arch_386[@]}; do
        run_vet 386 $os
    done
    for os in ${arch_arm[@]}; do
        run_vet arm $os
    done
    for os in ${arch_arm64[@]}; do
        run_vet arm64 $os
    done
    for os in ${arch_wasm[@]}; do
        run_vet wasm $os
    done
    for os in ${arch_mips[@]}; do
        run_vet mips $os
    done
    for os in ${arch_mips64[@]}; do
        run_vet mips64 $os
    done
    for os in ${arch_mips64le[@]}; do
        run_vet mips64le $os
    done
    for os in ${arch_mipsle[@]}; do
        run_vet mipsle $os
    done
    for os in ${arch_ppc64le[@]}; do
        run_vet ppc64le $os
    done
    for os in ${arch_riscv64[@]}; do
        run_vet riscv64 $os
    done
    for os in ${arch_s390x[@]}; do
        run_vet s390x $os
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
    for os in ${arch_ppc64[@]}; do
        run_staticcheck ppc64 $os
    done
    for os in ${arch_386[@]}; do
        run_staticcheck 386 $os
    done
    for os in ${arch_arm[@]}; do
        run_staticcheck arm $os
    done
    for os in ${arch_arm64[@]}; do
        run_staticcheck arm64 $os
    done
    for os in ${arch_wasm[@]}; do
        run_staticcheck wasm $os
    done
    for os in ${arch_mips[@]}; do
        run_staticcheck mips $os
    done
    for os in ${arch_mips64[@]}; do
        run_staticcheck mips64 $os
    done
    for os in ${arch_mips64le[@]}; do
        run_staticcheck mips64le $os
    done
    for os in ${arch_mipsle[@]}; do
        run_staticcheck mipsle $os
    done
    for os in ${arch_ppc64le[@]}; do
        run_staticcheck ppc64le $os
    done
    for os in ${arch_riscv64[@]}; do
        run_staticcheck riscv64 $os
    done
    for os in ${arch_s390x[@]}; do
        run_staticcheck s390x $os
    done
}

run_vet_all
run_staticcheck_all
