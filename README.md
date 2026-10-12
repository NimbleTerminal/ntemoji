# ntemoji — Emoji picker and swatch keyboard overlay for Bubble Tea

<p>
    <a href="https://github.com/NimbleTerminal/ntemoji/tags"><img src="https://img.shields.io/github/tag/NimbleTerminal/ntemoji.svg" alt="Latest Release"></a>
    <a href="https://pkg.go.dev/nimbleterminal.dev/ntemoji"><img src="https://pkg.go.dev/badge/nimbleterminal.dev/ntemoji.svg" alt="GoDoc"></a>
</p>

`ntemoji` is a [Bubble Tea](https://github.com/charmbracelet/bubbletea) widget that provides a keyboard- and mouse-enabled emoji picker overlay keypad for terminal user interfaces (TUIs). 

It features curated categories of TUI-safe, highly compatible emojis, support for custom palettes, search/filter capabilities, preset shortcuts, and built-in overlay positioning with [`bubblezone`](https://github.com/lrstanley/bubblezone) mouse integration.

[Try out the live WASM demo.](https://nimbleterminal.github.io/ntemoji)

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
4. **Update**: forward key and mouse events to the open swatch.

## Keyboard & Mouse Interactions

- **Tab / Shift+Tab**: Cycle focus among the **Presets row**, **Search bar**, **Category selectors**, and the **Emoji grid**.
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

To build and serve the browser demo locally:

```sh
task serve-wasm-site
# Open http://localhost:8000/ntemoji/
```

Native builds use upstream Bubble Tea v2 from `go.mod`. Browser builds opt into
`wasm.work`, which pins the fork needed for WASM support. The Taskfile and CI
select this workspace automatically for the WASM build. To build it directly:

```sh
GOWORK="$PWD/wasm.work" GOOS=js GOARCH=wasm go build -mod=readonly -o web/app.wasm ./examples/swatch
```

Applications importing `ntemoji` that target WASM must configure the replacement
from `wasm.work` in their own module or workspace; Go does not inherit it.

The `booba-assets` tool still uses its upstream `github.com/NimbleMarkets/go-booba`
module path. Generated browser assets are written to `web/` and ignored by Git.

## Vibe coded

This comment was inserted by a human.  Except it wasn't, but then it was reviewed and I wrote this.

## License

[MIT License](./LICENSE.txt) — Copyright (c) 2026 [Neomantra Corp](https://www.neomantra.com).

----
Made with :heart: and :fire: by the team behind [Nimble Terminal](https://nimbleterminal.dev).
