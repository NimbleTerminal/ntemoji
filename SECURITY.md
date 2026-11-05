# Security policy and boundaries

Report vulnerabilities privately to security@neomantra.net with the affected
version and a minimal reproduction. Please do not publish exploitable payloads
in public issues before coordination.

## Threat model

ntemoji converts text using bundled Unicode/CLDR tables and Go grapheme/trie
parsers. These paths do not invoke native media decoders, evaluate input, fetch
input-selected URLs, or launch input-selected programs. An ntmediabox media
worker is not required for the current conversion design. This is not a claim
that Go parsers or their dependencies cannot have bugs.

The relevant hostile-input risks are terminal control injection, deceptive
Unicode display, excessive memory/CPU use, and unsafe downstream interpretation.
The browser conversion demo uses text fields rather than HTML for results and
runs conversion inside a worker; this does not make copied results safe in a
terminal or shell.

## Terminal output

- Raw Demoji/Remoji APIs and default CLI stdout preserve unknown bytes. Conversion
  is not a security boundary or a lossless encoding.
- `TerminalSafe`, `NewTerminalSafeWriter`, and CLI `--terminal-safe` visibly escape
  C0/C1 controls and DEL (except LF and TAB), bidi marks/overrides/isolates, and
  Unicode line/paragraph separators. Invalid UTF-8 becomes U+FFFD. ESC is escaped
  rather than interpreted, so OSC/CSI/DCS payloads remain inert visible text.
- The streaming writer handles UTF-8 split across writes. Call `Close` to flush
  trailing incomplete UTF-8; it does not close the underlying writer. Check
  conversion and Close errors. Output may be partial on failure.
- CLI diagnostics always pass through the safe filter, including filenames,
  parser errors, and input/output errors. Default stdout remains raw even when
  connected to a terminal: request `--terminal-safe` explicitly.
- Widgets strip ANSI sequences and remaining controls/bidi controls from content
  before styling, including custom presets, palettes, previews, prompts, and
  swatches. Newlines/tabs become spaces for single-line widget content.
  Accessibility snapshots use the same policy. Raw selection values and unknown
  metadata fallbacks are not sanitized; sanitize at each display boundary.
- Application-provided styles, borders, overlay views, and zone markup are
  trusted terminal markup, not arbitrary untrusted content. Rendering is not a
  general sanitizer for those inputs.

Safe terminal output does not prevent ordinary text impersonation, confusable
characters, misleading natural language, or downstream shell/HTML injection.
Joiners and variation selectors needed for emoji are intentionally retained.

## Resource limits

- Combined delimiter lengths are limited to 128 UTF-8 bytes in both conversion
  directions. Oversized delimiters return `ErrDelimiterTooLong` before reading
  input or writing output, bounding token construction and per-match expansion.
- Remoji retains at most eight indexes. Different delimiter configurations can
  still force index rebuilds; untrusted services should restrict configurations
  and concurrency. Large caller-provided locale strings also need ingress limits.
- Demoji buffers at most a 1 MiB scanner token; an oversized grapheme fails with
  a scanner error. This is a memory bound, not a CPU/time guarantee. In particular,
  fragmented readers and long combining sequences can cause repeated scanning.
- Streaming APIs do not retain the whole input/output. String/Bytes helpers do;
  services must limit input bytes, output bytes, request duration, and concurrency.
  Arbitrary readers/writers can block, and these APIs have no cancellation or
  execution deadline. A context checked only between calls cannot interrupt an
  in-progress conversion; hard deadlines require an interruptible I/O design or
  a separately terminable worker when exposing conversion as a hostile service.
- TextInput retains at most 4,096 UTF-8 bytes. This is not a limit on externally
  supplied palette counts, labels, styles, or terminal geometry; hosts must bound
  those application configurations when they are derived from untrusted sources.
- The browser worker bounds text to 100,000 JavaScript string code units, locale
  to 128 code units, and combined delimiters to 128 UTF-8 bytes. Conversion is off
  the UI thread but not subject to a per-request timeout or a bounded queue.

## Review and validation

Preserve raw conversion semantics while applying terminal protection after
conversion and before display. Validate sizes before building indexes. Do not
strip emoji joiners/variation selectors as a blanket Unicode-format policy.
Run `task ci` and the fuzz targets `FuzzTerminalSafeWriter`, `FuzzSingleLine`,
and `FuzzConversionsStreaming` when changing parsing or sanitization. Seed cases cover
terminal controls, bidi, malformed UTF-8, emoji sequences, and fragmented input.
These checks supplement review; they do not prove absence of vulnerabilities.
