#!/bin/sh
# ===========================================================================
# File: build.sh
# Description: usage: ./build [outdir]
# ===========================================================================
set -e

mkdir_output() {
    if [ -z "$1" ]; then
        mkdir -p build
        OUTPUT_DIR=$(cd build > /dev/null && pwd)
    else
        OUTPUT_DIR="$1"
    fi
    echo "$OUTPUT_DIR"
}

OUTPUT_DIR=$(mkdir_output "$1")
GOOS=$(go env GOOS)
GOARCH=$(go env GOARCH)

flags="-X 'main.version=${VERSION}'
-X 'main.commitID=$(git rev-parse --short HEAD)'
-X 'main.buildDate=$(date -u +"%Y-%m-%dT%H:%M:%SZ")'"

for exec in kcheck makecheck; do
  OUTPUT_BINARY=$OUTPUT_DIR/$exec"_"$GOOS"_"$GOARCH
  # -ldflags="-w -s" means omit DWARF symbol table and the symbol table and debug information
  go build -p=16 --tags "dev" -ldflags "-w -s $flags" -o ${OUTPUT_BINARY} ./cmd/kcheck
done


