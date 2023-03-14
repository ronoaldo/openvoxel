#!/bin/bash
set -e
set -o pipefail
set -x

mkdir -p build/
WORK=$(mktemp -d)

SCRIPT="$1"
shift

docker run --rm \
    -v $PWD:/workspace \
    -v $WORK:/tmp/work \
    -e VERSION=${VERSION:-dev} \
    golang:1.20.2-bullseye /workspace/${SCRIPT} "$@" 2>&1 | tee $LOG

echo "Build log: $LOG"
echo "Build work directory: $WORK"

