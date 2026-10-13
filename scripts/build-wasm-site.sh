#!/bin/sh
# Shared by Taskfile and GitHub Actions. Run from any directory.
set -eu
cd "$(dirname "$0")/.."
export GOWORK=off
mkdir -p web/demos
for demo in simple swatch; do
    GOWORK="$PWD/wasm.work" GOOS=js GOARCH=wasm go build -mod=readonly -o "web/demos/$demo.wasm" "./examples/$demo"
done
go tool booba-assets web/
# Always install the tracked showcase over the generated starter page.
cp site/index.html site/style.css site/demo.js web/
mkdir -p web/third_party/unicode
cp third_party/unicode/LICENSE.txt third_party/unicode/NOTICE.txt web/third_party/unicode/
