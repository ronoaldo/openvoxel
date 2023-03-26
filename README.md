# openvoxel

> This is a work in progress project.

**openvoxel** is an [open source](https://en.wikipedia.org/wiki/Open_source)
[voxel](https://en.wikipedia.org/wiki/Voxel) [game
engine](https://en.wikipedia.org/wiki/Game_engine).

Experimental demo available here:

* [**Webgl** version](https://ronoaldo.github.io/openvoxel/helloworld_js_wasm.html)
* [**Windows (64 bits)** version](https://ronoaldo.github.io/openvoxel/helloworld_windows_amd64.exe)
* [**Linux (64 bits)** version](https://ronoaldo.github.io/openvoxel/helloworld_linux_amd64)

## Development Setup

To start developing, you can use the helper scripts in the `scripts/` folder.
You must have already a working `Go` installation, we tested on Go 1.18 and
newer and be using either Debian/Ubuntu or a debian-based docker container.

After checking out the repository, you can then execute:

    export OPENVOXEL_ARCHS=amd64
    ./scripts/cross-setup.sh

This will install all the OpenGL dependencies for you.  To get started testing,
use the `go build` or `go run` commands, like:

    cd exp/cmd/helloworld
    go run main.go

To speed up the testing cycle, run `go install` once so you can benefit from
cached packages built with CGO:

    cd exp/cmd/helloworld
    go install
