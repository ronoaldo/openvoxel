#!/bin/sh

OPENVOXEL_ARCHS=${OPENVOXEL_ARCHS:-amd64 arm64}

if [ x"$(id -u)" != x"0" ] ; then
    echo "Should be executed as root."
    exit 1
fi

echo "Adding all architectures"
for arch in ${OPENVOXEL_ARCHS} ; do
    dpkg --add-architecture $arch
done
apt-get update

echo "Installing cross-build toolchain"
apt-get install gcc g++ pkg-config \
    gcc-i686-linux-gnu \
    g++-i686-linux-gnu \
    gcc-aarch64-linux-gnu \
    g++-aarch64-linux-gnu \
    mingw-w64 \
    -yq

echo "Installing required libraries"
export LIBS=""
for arch in ${OPENVOXEL_ARCHS} ; do
    export LIBS="${LIBS} libgl1-mesa-dev:${arch} libglfw3-dev:${arch} \
        libxxf86vm-dev:${arch} libxinerama-dev:${arch} \
        libxi-dev:${arch} libx11-dev:${arch} libxcursor-dev:${arch} \
        libxrandr-dev:${arch}"
done
apt-get install $LIBS -yq
