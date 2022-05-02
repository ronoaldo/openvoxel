#!/bin/sh

echo "Performing cross compilation to Windows 64 bits"

mkdir -p build

MING_GW=${PWD}/3rd_party/SDL2-2.0.22/x86_64-w64-mingw32

for program in exp/cmd/* ; do
    out="$(basename $program).exe"
    echo "Building $program into $out..."
    
    env CGO_ENABLED=1 \
    CC=x86_64-w64-mingw32-gcc \
    CXX=x86_64-w64-mingw32-g++ \
    CGO_LDFLAGS="-L${MING_GW}/lib -lmingw32 -lSDL2main -lSDL2" \
    CGO_CFLAGS="-I${MING_GW}/include -D_REENTRANT" \
    GOOS=windows GOARCH=amd64 \
    go build \
        -o build/$out \
        -x -v -i \
        $program/main.go
    cp ${MING_GW}/bin/SDL2.dll build/
done