#!/bin/bash
set -e
set -o pipefail
#set -x

# go_build is a helper function to execute the proper Go command passing all the
# required build flags.
go_build() {
    export PROG=$1 OS=$2 ARCH=$3
    export OUT="$(basename ${PROG})_${OS}_${ARCH}"
    echo "** Building ${PROG} for ${OS}/${ARCH} into ${OUT}... **"

    export GOOS=$OS GOARCH=$ARCH CC= CXX=
    case $OS in
        windows)
            export OUT="${OUT}.exe"
            case $ARCH in
                amd64) export CC=x86_64-w64-mingw32-gcc CXX=x86_64-w64-mingw32-g++ ;;
                386)   export CC=i686-w64-mingw32-gcc CXX=i686-w64-mingw32-g++ ;;
            esac
        ;;
        linux)
            case $ARCH in
                386)   export CC=i686-linux-gnu-gcc    CXX=i686-linux-gnu-g++ ;;
                arm64) export CC=aarch64-linux-gnu-gcc CXX=aarch64-linux-gnu-g++ ;;
                arm)   export CC=arm-linux-gnueabi-gcc CXX=arm-linux-gnueabi-gcc ;;
            esac
        ;;
        js)
            cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" build
            cp "${PROG}/index.html" build/${OUT}.html
            export OUT="${OUT}.wasm" 
        ;;
    esac
    env CGO_ENABLED=1 CC=$CC CXX=$CXX GOOS=$GOOS GOARCH=$GOARCH \
        go build ${GO_BUILD_FLAGS} -o build/${OUT} ./${PROG} &&\
	echo "** Binary created at ${OUT} **" ||\
	echo "** Build failed **"
}

# Main
_NAME=$(readlink -f $0)
_DIR=$(dirname $_NAME)
cd $_DIR/..

echo "Building into $PWD/build ..."
mkdir -p build
rm -rvf build/*
git status 2>&1 >/dev/null || git config --global --add safe.directory "$PWD"

if [ x"$DEBUG" = x"true" ] ; then 
    echo "Debug enabled to check 'go build' flags..."
	export GO_BUILD_FLAGS='-x -v'
	set -x
fi

if [ x"$1" = x"--ci" ]; then
    echo "Setting up CI environment"
    ./scripts/cross-setup.sh
    shift
fi

if [ x"$1" != x"" ]; then
    # Build a specific OS/ARCH pair
    go_build exp/cmd/helloworld $1 $2
else
    # Build all supported OS/ARCH pairs
    go_build exp/cmd/helloworld windows 386
    go_build exp/cmd/helloworld windows amd64

    go_build exp/cmd/helloworld linux 386
    go_build exp/cmd/helloworld linux amd64
    go_build exp/cmd/helloworld linux arm64

    go_build exp/cmd/helloworld js wasm
fi