# `ntemoji` CHANGELOG

## Unreleased

## `v0.3.0` (2026-11-04)

This release hardens terminal display of untrusted text and adds a dedicated
emoji font comparison demo with Safari-compatible Noto rendering.

Compatibility: Demoji and Remoji now reject opening and closing delimiters whose
combined length exceeds 128 UTF-8 bytes, returning `ErrDelimiterTooLong` before
reading input or writing output. TextInput retains at most 4,096 UTF-8 bytes,
truncating excess input at a rune boundary after sanitization. Widget content and
accessibility snapshots strip terminal and bidi controls; selection values remain
raw application data.

Raw conversion output remains the default for the library and CLI pipelines.
Use `--terminal-safe`, `TerminalSafe`, or `NewTerminalSafeWriter` when displaying
untrusted conversion output in a terminal. The streaming writer must be closed
to flush trailing incomplete UTF-8; it does not close the underlying writer.

 * Add an optional ntfonts Noto Color Emoji comparison as a dedicated browser demo, with localized sequence specimens, Unicode inspection, lazy font loading, and automatic CBDT/sbix selection for Safari using ntfonts v0.3.1.

 * Harden terminal display with streaming control escaping, a `--terminal-safe` CLI option, safe diagnostics and widget content, bounded delimiters/search input, and security regression/fuzz coverage.

 * Migrate example imports and the asset tool to `nimbleterminal.dev/booba` v0.12.1, and update Go dependencies.

## `v0.2.0` (2026-10-30)

This release adds multilingual emoji conversion, the `demoji` and `remoji` CLI
tools, browser conversion demos, and picker accessibility and layout improvements.

Migration: use the module path `nimbleterminal.dev/ntemoji` and replace
`EmojiTags[emoji]` with `Tags(emoji)`. Go 1.26.8 or newer is required. Repository
examples now run from their own module: `go -C examples run ./simple`.

 * Make generated-data freshness checks insensitive to equivalent Unicode string quoting across Go versions, and test Go 1.26.8 and 1.27 in CI.

 * Split demos and Booba tooling into a nested examples module and update Go dependencies and the browser Bubble Tea fork.

 * Include named hair and skin-tone emoji components in Demoji/Remoji conversion, including `red hair` → 🦰.

 * Restore emoji presentation for desktop and gear, emit completed Demoji control graphemes without waiting for another read, and preserve background mouse-zone endpoints under overlays.
 * Reuse picker search results within rendering and resize measurements; cache Remoji indexes by normalized locale and delimiters with bounded retention.
 * Pin English CLDR sources in the shared locale manifest and explicitly reject unsupported regional/script tables.
 * Correct the swatch example binary name to `ntemoji-swatch`.

 * Publish `demoji` and `remoji` as a Homebrew cask, `nimbleterminal/tap/ntemoji`, with release archives for macOS, Linux, and Windows, and add `--version` to both tools.

 * Allow empty Remoji delimiters for greedy, longest-name trie matching with Unicode word boundaries, while retaining colon-delimited defaults.

 * Add streaming Demoji/Remoji CLI tools and dedicated browser conversion pages with language selection, custom delimiters, localized examples, and copyable multiline output, backed by the Go library in a WASM worker.

 * Add streaming `Demoji`/`Remoji` and String/Bytes helpers, configurable delimiters and locale, plain descriptive output by default, and canonical emoji sequence recognition from pinned Unicode Emoji 17.0 data.

 * Extract pinned CLDR metadata for German, Spanish, French, Japanese, and Simplified Chinese. Add `Locales`, `EmojiNameFor`, `TagsFor`, `MatchEmojiFor`, and `WithLocale` with English fallback, localized picker labels, search, and accessibility state.

 * Add a focus-following descriptive preview, optional labeled result lists with text category tabs and visible focus markers (`WithListMode`), and comparable accessibility snapshots for pickers and swatches.

 * Generate English names and search keywords from a checksum-pinned CLDR 48.2 snapshot, including derived sequences. Normalize emoji presentation selectors during lookup. Include Unicode License v3 and attribution in source and browser demo distributions. Search now uses CLDR keywords instead of the previous custom synonyms.

 * Add `EmoticonToEmoji` and `EmojiToEmoticon` for common ASCII emoticons, with preferred reverse spellings, unchanged unknown inputs, and descriptive labels for every conversion result.

 * Replace the exported mutable `EmojiTags` map with `Tags(emoji)`, which returns an independent keyword slice, and `EmojiName(emoji)` for descriptive English labels. Cover every built-in emoji and include names in search. This is a breaking API change; replace `EmojiTags[emoji]` with `Tags(emoji)`.

 * Keep wrapped category tabs clickable and clip overlays, graphemes, and mouse zones to the supplied viewport.

 * Keep the scrolled grid stable during mouse selection and align wrapped preset rows with their mouse targets.

 * Keep modal mouse bounds aligned with the actual rendered position when view and window sizes differ.
 * Reserve narrow input space for end cursors and let demos finish normally after action errors.
 * Share grid dimension defaults and reduce unconstrained resize fitting to two measurements.

 * Add `SwatchPicker.OpenPicker` for keyboard activation without dropping the first mouse selection.
 * Keep swatch rendering free of layout mutations and forward terminal resize events to every example swatch.
 * Correct gear and desktop constants and gear/wrench search tags.
 * Check Go formatting in CI, document the package, and remove unused helpers and obsolete VHS tasks.

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
