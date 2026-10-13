# `ntemoji` CHANGELOG

## Unreleased

 * Preserve overlay zone markers and leading zero-width text without replaying them as styles.
 * Size swatch modals against the available terminal using measured layout, including custom frames.
 * Keep cursor graphemes visible in narrow searches and render text styles per run.

 * Route search editing keys before navigation and ignore key releases.
 * Keep search text within its cell width, support single-line paste, and edit whole graphemes.
 * Measure and overlay combined emoji graphemes consistently with Lip Gloss.

 * Use uniform emoji grid cells with spaced rows, accurate picker dimensions, and matching mouse targets when scrolled.

 * Add a multi-demo browser playground with example navigation, restart controls, keyboard help, and loading/error states.
 * Add context-sensitive Booba terminal actions and the explicit browser bridge.
 * Preserve the selected emoji when resizing the terminal.
 * Let searches contain `q` without exiting the example.

 * Move the repository to `NimbleTerminal/ntemoji` and the Go module to `nimbleterminal.dev/ntemoji`.
 * Update documentation, demo links, and the quickstart for Bubble Tea v2.
 * Refresh UI dependencies and go-booba assets; require Go 1.26.8 or newer.
 * Isolate the Bubble Tea WASM fork in an opt-in `wasm.work`, pinned to the revision used by go-booba v0.9.2; native builds use upstream Bubble Tea.
 * Add native and WASM build checks for pushes and pull requests.

## `v0.1.1` (2026-05-25)

 * fix: resolve row width layout shifts, overlay ANSI bugs, and swatch hit-bounds misalignment

## `v0.1.0` (2026-05-24)

 * Initial release
