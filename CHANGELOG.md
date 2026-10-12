# `ntemoji` CHANGELOG

## Unreleased

 * Move the repository to `NimbleTerminal/ntemoji` and the Go module to `nimbleterminal.dev/ntemoji`.
 * Update documentation, demo links, and the quickstart for Bubble Tea v2.
 * Refresh UI dependencies and go-booba assets; require Go 1.26.8 or newer.
 * Isolate the Bubble Tea WASM fork in an opt-in `wasm.work`, pinned to the revision used by go-booba v0.9.2; native builds use upstream Bubble Tea.
 * Add native and WASM build checks for pushes and pull requests.

## `v0.1.1` (2026-05-25)

 * fix: resolve row width layout shifts, overlay ANSI bugs, and swatch hit-bounds misalignment

## `v0.1.0` (2026-05-24)

 * Initial release
