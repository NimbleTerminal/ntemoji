#!/bin/sh
set -eu
cd "$(dirname "$0")/.."
# Include new source files, but exclude ignored local experiments and artifacts.
unformatted=$(git ls-files --cached --others --exclude-standard -z -- '*.go' | xargs -0 gofmt -l)
if [ -n "$unformatted" ]; then
    printf 'Run gofmt on these files:\n%s\n' "$unformatted" >&2
    exit 1
fi
