# ntemoji — Emoji picker and swatch keyboard overlay for Bubble Tea

<p>
    <a href="https://github.com/NimbleTerminal/ntemoji/tags"><img src="https://img.shields.io/github/tag/NimbleTerminal/ntemoji.svg" alt="Latest Release"></a>
    <a href="https://pkg.go.dev/nimbleterminal.dev/ntemoji"><img src="https://pkg.go.dev/badge/nimbleterminal.dev/ntemoji.svg" alt="GoDoc"></a>
</p>

`ntemoji` is a [Bubble Tea](https://github.com/charmbracelet/bubbletea) widget that provides a keyboard- and mouse-enabled emoji picker overlay keypad for terminal user interfaces (TUIs). 

It features curated categories of TUI-safe, highly compatible emojis, support for custom palettes, search/filter capabilities, preset shortcuts, and built-in overlay positioning with [`bubblezone`](https://github.com/lrstanley/bubblezone) mouse integration.

[Try the emoji picker and swatch grid online.](https://nimbleterminal.github.io/ntemoji)

This module was inspired by the [@madicen's `bubble-color-picker` library](https://github.com/madicen/bubble-color-picker).

## Installation

Requires Go 1.26.8 or newer.

```sh
go get nimbleterminal.dev/ntemoji
```

The repository is now [NimbleTerminal/ntemoji](https://github.com/NimbleTerminal/ntemoji).
Existing users should change their imports to `nimbleterminal.dev/ntemoji` and run
`go mod tidy` after installing a release that uses the new module path.

## Quickstart

```go
package main

import (
	"fmt"
	"os"

	"nimbleterminal.dev/ntemoji"
	tea "charm.land/bubbletea/v2"
	zone "github.com/lrstanley/bubblezone/v2"
)

type model struct {
	picker ntemoji.Model
	zm     *zone.Manager
	chosen string
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case ntemoji.EmojiChangedMsg:
		m.chosen = msg.Emoji
		return m, tea.Quit
	case ntemoji.EmojiCanceledMsg:
		return m, tea.Quit
	}

	updated, cmd := m.picker.Update(msg)
	m.picker = updated.(ntemoji.Model)
	return m, cmd
}

func (m model) View() tea.View {
	v := tea.NewView(m.zm.Scan(m.picker.View().Content))
	v.AltScreen = true
	v.MouseMode = tea.MouseModeAllMotion
	return v
}

func main() {
	zm := zone.New()
	picker := ntemoji.New(
		ntemoji.WithInitialEmoji("🚀"),
		ntemoji.WithPresets([]string{"😀", "👍", "🔥", "✅", "❌"}),
		ntemoji.WithShowSearch(true),
	)
	picker.SetZoneManager(zm)

	p := tea.NewProgram(model{picker: picker, zm: zm})
	result, err := p.Run()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if chosen := result.(model).chosen; chosen != "" {
		fmt.Printf("Selected emoji: %s\n", chosen)
	}
}
```

## SwatchPicker (Popup Modal overlay)

The `SwatchPicker` is a small UI component representing the selected emoji with a dropdown arrow (e.g. `😀 ▼`). Clicking it opens the full emoji keyboard popup centered on the button.

1. **Create the swatch**: `swatch := ntemoji.NewSwatchPicker("🚀", "Trigger")`
2. **Setup bubblezone**: `swatch.SetZoneManager(zm)`
3. **Set bounds & render**: call `swatch.SetBounds(row, col, width, height)` and wrap with overlay rendering:
   ```go
   mainView = swatch.ViewWithOverlay(mainView, width, height)
   return zm.Scan(mainView)
   ```
4. **Update**: forward `tea.WindowSizeMsg` to every swatch, including closed ones, and key and mouse events to the open swatch. `WindowSizeMsg` determines picker size; `ViewWithOverlay` records the actual drawn position for mouse hit-testing.
5. **Keyboard opening**: call `swatch.OpenPicker()` and return its command from your update handler. `swatch.Open()` reports whether the modal is open.

## Keyboard & Mouse Interactions

- **Tab / Shift+Tab**: Cycle focus among the **Presets row**, **Search bar**, **Category selectors**, and the **Emoji grid**.
- **Search**: Type or paste text while the search field is focused. Left/right edit the cursor; long queries scroll within the field.
- **Arrows / hjkl**: Navigate selections within the active presets or the emoji grid. Changing focus to category selection and using ←/→ instantly switches category tabs.
- **Enter**: Confirms the selected emoji and fires an `EmojiChangedMsg`.
- **Escape**: Cancels selection and fires an `EmojiCanceledMsg`.
- **Mouse clicks**: Clicking any category tab switches categories. Hovering over presets/grid items highlights them, and releasing the left-click selects the emoji.

## Configuration Options

Construct the emoji picker with **`New(...Option)`**:

| Option | Purpose |
|--------|---------|
| `WithInitialEmoji(emoji)` | Sets the starting emoji (auto-focuses it in grid/presets). |
| `WithPresets([]string)` | Shows a row of quick-select preset emojis above search/categories. |
| `WithPalettes([]Palette)` | Supplies custom emoji lists and category tabs. |
| `WithTheme(Theme)` | Customizes colors and styles throughout the picker. |
| `WithStyle(lipgloss.Style)` | Configures custom outer frame border and padding. |
| `WithAutoDismiss(bool)` | When true, selection includes `Dismiss: true` to auto-hide the modal. |
| `WithShowSearch(bool)` | Toggles the search input box for matching keywords. |

## Emoji names and search keywords

```go
ntemoji.EmojiName("👍") // "thumbs up"
ntemoji.Tags("👍")      // []string{"+1", "good", "hand", "like", "thumb", "up", "yes"}
```

`EmojiName` returns the English Unicode CLDR short name (TTS name) for supported
emoji, sequences, and annotated symbols, including every built-in emoji constant
and default palette entry. Unknown values are returned unchanged; an empty input
returns an empty string. Applications with custom palettes can provide their own
labels before falling back to this lookup. Names are display text, not stable IDs.

`Tags` returns a fresh slice of CLDR search keywords, or the descriptive name as its
only entry when there are no extra keywords. Unknown values return `nil`.
Changing the returned slice cannot affect another picker or its search results.
Search matches both descriptive names and keywords, ignoring case.

Metadata is generated from pinned **CLDR 48.2** English annotations and derived
annotations. Skin tones, flags, and joined sequences are included. Presentation
selectors are ignored for lookup, so `❤` and `❤️` share a name and keywords.
The default palettes remain curated for terminal use; broader metadata coverage
does not imply that every supported sequence renders reliably in every terminal.

Run `task generate-cldr` to reproduce the tables offline from checked-in sources
(or run `go generate ./...` directly).
See [source versions, checksums, and update instructions](internal/cldr/README.md).

Use `EmojiName` when supplying accessibility labels in your application. These
lookups provide metadata; applications still need to expose those labels through
their terminal or browser accessibility integration.

**API migration:** replace `EmojiTags[emoji]` with `Tags(emoji)`. The exported
mutable `EmojiTags` map has been removed; global keyword mutation is no longer
supported. Search keywords now come from CLDR rather than the previous custom
synonyms, so keyword contents and search results can differ.

## Emoticon conversion

```go
ntemoji.EmoticonToEmoji(":-)") // "🙂"
ntemoji.EmojiToEmoticon("🙂")  // ":)"
ntemoji.EmoticonToEmoji("<3")  // "❤️"
```

These functions convert one complete value, matching exactly without trimming
whitespace or replacing text inside sentences. Unknown inputs (including empty
strings) pass through unchanged. Multiple emoticon spellings can map to the same
emoji; reverse conversion returns the preferred spelling below.

| Preferred emoticon | Emoji | Other accepted spellings |
|--------------------|-------|--------------------------|
| `:)` | 🙂 | `:-)` |
| `:(` | 🙁 | `:-(` |
| `:D` | 😃 | `:-D`, `:d`, `:-d` |
| `;)` | 😉 | `;-)` |
| `XD` | 😆 | `xD` |
| `:P` | 😛 | `:-P`, `:p`, `:-p` |
| `:O` | 😮 | `:-O`, `:o`, `:-o` |
| `:/` | 😕 | `:-/`, `:\`, `:-\` |
| `:\|` | 😐 | `:-\|` |
| `:'(` | 😢 | `:'-(` |
| `>:(` | 😠 | `>:-(` |
| `:*` | 😘 | `:-*` |
| `B)` | 😎 | `B-)`, `8)`, `8-)` |
| `<3` | ❤️ | — |
| `</3` | 💔 | — |

Every conversion result has an `EmojiName` label. Conversion does not change the
default palettes or automatically rewrite search input.

## TUI Emoji Compatibility & Box-Sizing

Emojis are generally 2 columns wide in terminal user interfaces. However, some emojis (such as those containing the variation selector `\ufe0f`, e.g., `⚠️` or `🌧️`) report a width of **2** columns under standard Go string-width libraries (like Lip Gloss / `uniseg`), but render as only **1** column wide on standard terminal emulators.

When designing custom palettes or presets for terminal widgets like `ntemoji`, this discrepancy causes right-border layout alignment shifts (as the terminal renders the row narrower than Lip Gloss's box-sizing layout calculations).

To reduce layout differences across terminals:
- **Only use standard 2-column emojis** where `lipgloss.Width` and terminal rendering widths agree.
- **Avoid variation selectors** (`\ufe0f`) in custom emoji palettes.
- Default palettes are curated for consistent widths; rendering still depends on the terminal and font.

## Demos

Built-in demo examples compile out-of-the-box:

```sh
# Run full-screen demo
task build-ex-simple
./bin/simple

# Run multi-swatch grid overlay demo
task build-ex-swatch
./bin/swatch
```

The [online playground](https://nimbleterminal.github.io/ntemoji/) includes both
examples, keyboard help, source links, a restart control, and Booba terminal
action buttons that change when a picker opens or closes. Switching examples
starts a fresh session.

To build and serve it locally:

```sh
task serve-wasm-site
# Open http://localhost:8000/ntemoji/
```

Native builds use upstream Bubble Tea v2 from `go.mod`. Browser builds opt into
`wasm.work`, which pins the fork needed for WASM support. The Taskfile and CI
select this workspace automatically for both WASM builds. To build the full site directly:

```sh
sh scripts/build-wasm-site.sh
```

Applications importing `ntemoji` that target WASM must configure the replacement
from `wasm.work` in their own module or workspace; Go does not inherit it.

The `booba-assets` tool still uses its upstream `github.com/NimbleMarkets/go-booba`
module path. The examples use `internal/demorun` to select the native or WASM
launcher and publish their current controls through Booba terminal actions. The tracked page source lives in `site/`.
The shared build script compiles both examples, stages the runtime, and copies the page into `web/`.
Generated browser assets are ignored by Git; edit `site/`, not `web/`.

Run the same validation as GitHub Actions with `task ci`: race tests, vet, native
example builds, and the complete WASM demo site build.

## Vibe coded

This comment was inserted by a human.  Except it wasn't, but then it was reviewed and I wrote this.

## License

[MIT License](./LICENSE.txt) — Copyright (c) 2026 [Neomantra Corp](https://www.neomantra.com).

Unicode CLDR data and generated metadata tables are separately covered by
[Unicode License v3](third_party/unicode/LICENSE.txt). Include the
[Unicode notice](third_party/unicode/NOTICE.txt) and license when distributing
those tables or applications embedding them. The browser demo build includes both.

----
Made with :heart: and :fire: by the team behind [Nimble Terminal](https://nimbleterminal.dev).

## Descriptive views and accessibility state

For background and assistive-technology resources, see Perkins School for the
Blind’s [Emojis and Accessibility](https://www.perkins.org/resource/emojis-and-accessibility/),
which links to guidance on using emoji with visual impairments, VoiceOver, and NVDA.

The preview follows the focused emoji or category, independently of the confirmed
selection. Names come from CLDR data (English by default). Enable a labeled result list:

```go
picker := ntemoji.New(
    ntemoji.WithListMode(true),
    ntemoji.WithShowSearch(true),
)
```

List mode shows one result per row with its name and a `>` focus marker. Categories
use text labels (`>` for focus, `*` for the active category when another section
has focus). Presets stay compact; their focused item has a `>` marker and its name
in the preview. Up/down or j/k move through results; left/right or h/l move through
categories and presets. Tab and Shift+Tab switch sections. Enter confirms.
Long labels are clipped to the available width; state snapshots retain full names.

Read `picker.AccessibilityState()` after updating the model. It includes focus,
the focused name and glyph, one-based position and section count, category, query,
result count, and the initial or last confirmed selection. An absent item has
position zero. Search has no item position or section count; use `ResultCount`.
During cross-category search, an emoji's category is its first matching palette.
Unknown emoji names fall back to the original glyph.

```go
before := picker.AccessibilityState()
updated, cmd := picker.Update(msg)
picker = updated.(ntemoji.Model)
after := picker.AccessibilityState()
if after != before {
    // Feed relevant changes to your host's status or accessibility adapter.
    // after.FocusedName describes navigation; after.SelectedName describes value.
}
// Return cmd from your host's Update so picker messages are delivered.
```

`swatch.AccessibilityState()` includes the label passed to `NewSwatchPicker`, the
current emoji and name, focus and expanded flags, and the picker snapshot while
open (zero-valued while closed).

Try the list and snapshot integration in the simple example:

```sh
NTEMOJI_LIST=1 go run ./examples/simple
```

The example also publishes focused names in its browser action status. These APIs
supply information for host adapters; they do not create semantic browser controls
or send screen-reader announcements. VoiceOver/NVDA integration and testing remain
separate work. The default view remains the compact emoji grid.

## Multilingual emoji metadata

Bundled metadata languages are English (`en`), German (`de`), Spanish (`es`), French
(`fr`), Japanese (`ja`), and Simplified Chinese (`zh`). `Locales()` returns a fresh,
sorted list. All are extracted from the same pinned CLDR 48.2 release, offline via
`task generate-cldr`.

```go
ntemoji.EmojiNameFor("🚀", "es")           // "cohete"
ntemoji.TagsFor("🚀", "fr")               // independent French keyword slice
ntemoji.MatchEmojiFor("🚀", "fusée", "fr") // true
picker := ntemoji.New(ntemoji.WithLocale("ja"), ntemoji.WithListMode(true))
```

`EmojiName` and `Tags` remain English. The locale-aware variants fall back to
English for missing entries or unsupported languages, then preserve unknown
emoji inputs (keywords return nil). Search accepts both localized and English
names/keywords, without accent folding or transliteration.

Locale matching accepts case-insensitive base languages and regional variants,
with hyphens or underscores (`es-MX`, `fr_CA`). Explicit unsupported scripts fall
back to English. `zh-TW`, `zh-HK`, `zh-MO`, and `zh-Hant` fall back to English;
`zh-Hans` selects Simplified Chinese. This is a limited bundle matcher, not a full
CLDR locale resolver.

`WithLocale` controls emoji names in previews, list rows, accessibility snapshots,
and search. Set it through `SwatchPicker.SetPickerOptions` for swatch names and
its open picker. UI instructions and palette/category names are not translated.
The bundled tables increase binary size; there are no runtime downloads.

```sh
NTEMOJI_LIST=1 NTEMOJI_LOCALE=es go run ./examples/simple
```

## Converting emoji in text

`Demoji` replaces emoji with descriptive names; plain names are the default.
`Remoji` defaults to `:name:` and can also match plain names with empty delimiters. Configure
both delimiters with `WithDelimiters` and the language with `WithTextLocale`:

```go
plain, err := ntemoji.DemojiString("Go 🚀") // "Go rocket"
encoded, err := ntemoji.DemojiString("Go 🚀", ntemoji.WithDelimiters("[", "]"))
// encoded: "Go [rocket]"
decoded, err := ntemoji.RemojiString(encoded, ntemoji.WithDelimiters("[", "]"))
// decoded: "Go 🚀"
spanish, err := ntemoji.DemojiString("🚀", ntemoji.WithTextLocale("es")) // "cohete"
```

All helpers return errors. `DemojiBytes` and `RemojiBytes` accept `[]byte` and
return `([]byte, error)` without modifying or aliasing their input. Stream
functions use destination-first argument order:

```go
err := ntemoji.Demoji(writer, reader, ntemoji.WithTextLocale("fr"))
err = ntemoji.Remoji(writer, reader, ntemoji.WithDelimiters("[", "]"))
```

Streams are processed incrementally and are not closed. Errors can leave partial
output. `Demoji` buffers at most a 1 MiB grapheme (larger clusters return an error);
`Remoji` buffers its metadata index and longest possible token, not the entire
input. Both propagate reader/writer failures, including short writes.

Plain output inserts spaces between names and adjacent words or emoji names.
Delimited output preserves surrounding spacing. Unknown emoji sequences, ordinary
punctuation, and invalid UTF-8 bytes pass through unchanged. Text-presentation
selectors are respected; text-default symbols such as `©` and `❤` require VS16
(emoji presentation) to be converted.

`Remoji` matches exact, case-sensitive CLDR names, including spaces and punctuation
inside the name. It uses a trie and scans left to right, taking the longest
complete matching name at each position. Unknown or ambiguous names remain
unchanged; an ambiguous longest name also prevents replacement of a shorter
prefix. Keywords and emoticons are not decoded. Choose the same locale for both
directions. Reverse output uses canonical emoji presentation.

Opt into plain names by supplying empty delimiters:

```go
text, err := ntemoji.RemojiString("fire engine rocket", ntemoji.WithDelimiters("", ""))
// "🚒 🚀", rather than matching "fire" first
```

Each missing delimiter requires a boundary against Unicode letters, numbers,
combining marks, connector punctuation (including `_`), and joiners. Thus
`rocketship` remains unchanged. One-sided delimiters are also supported. Matching
remains case-sensitive and does not perform linguistic word segmentation: an
isolated Chinese name can match, but the same name embedded in a continuous run
of Chinese letters does not. Plain mode can replace ordinary prose and cannot
reliably undo plain Demoji output. The deprecated `ErrEmptyDelimiter` is retained
for compatibility but is no longer returned.

This is text replacement, not a lossless serialization: literal `:rocket:` can
be converted, delimiters are not escaped, some names are ambiguous, and CLDR names
can change with data releases. Plain descriptive output is intended for readable
text; use explicit delimiters when reverse conversion is desired. Actual spoken
punctuation depends on the screen reader and its settings.
