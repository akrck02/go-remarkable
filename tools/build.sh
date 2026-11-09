#!/bin/sh
# Builds a program for the tablet: tools/build.sh ./cmd/hello leaves out/hello/, ready to install.
# Go cross-compiles by itself, so all this needs is Go.
set -eu

if [ $# -ne 1 ]; then
    echo "usage: tools/build.sh ./cmd/<program>" >&2
    exit 2
fi

root=$(cd "$(dirname "$0")/.." && pwd)
source=${1%/}
name=$(basename "$source")
manifest=$source/external.manifest.json

cd "$root"
if [ ! -f "$manifest" ]; then
    echo "no $manifest: AppLoad needs one to list the program" >&2
    exit 1
fi

mkdir -p "out/$name"
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o "out/$name/$name" "$source"
cp "$manifest" "out/$name/"
echo "out/$name"
