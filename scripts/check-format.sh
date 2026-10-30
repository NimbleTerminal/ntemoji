#!/bin/sh
set -eu
cd "$(dirname "$0")/.."
# Include new source files, skipping deleted paths and ignored local artifacts.
unformatted=$(git ls-files --cached --others --exclude-standard -z -- '*.go' | xargs -0 sh -ec '
    for file do
        if [ -f "$file" ]; then
            gofmt -l "$file"
        fi
    done
' sh)
if [ -n "$unformatted" ]; then
    printf 'Run gofmt on these files:\n%s\n' "$unformatted" >&2
    exit 1
fi
