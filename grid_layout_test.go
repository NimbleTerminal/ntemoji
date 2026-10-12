package ntemoji

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	zone "github.com/lrstanley/bubblezone/v2"
)

func TestGridUniformSpacing(t *testing.T) {
	picker := New(WithPalettes([]Palette{{Name: "Mixed", Icon: "😀", Emojis: []string{"😀", "A", "🔥", "B", "✅"}}}))
	picker.GridCols, picker.GridRows = 4, 2
	view := picker.View().Content
	lines := strings.Split(view, "\n")
	first, second := -1, -1
	for i, line := range lines {
		if strings.Contains(line, "🔥") {
			first = i
			start := strings.Index(line, "😀")
			fire := strings.Index(line, "🔥")
			if got := lipgloss.Width(line[start:fire]); got != 8 {
				t.Fatalf("emoji cells are %d columns apart, want 8", got)
			}
		}
		if strings.Contains(line, "✅") {
			second = i
		}
	}
	if first < 0 || second-first != 2 {
		t.Fatalf("grid rows at %d and %d, want two-line spacing", first, second)
	}
	width, height := picker.ViewSize()
	if width != lipgloss.Width(view) || height != lipgloss.Height(view) {
		t.Fatalf("ViewSize does not match rendered dimensions")
	}
}

func TestGridMouseWithSpacingAndScroll(t *testing.T) {
	zm := zone.New()
	defer zm.Close()
	picker := New()
	picker.SetZoneManager(zm)
	picker.GridCols, picker.GridRows = 4, 2
	picker.Focus = FocusGrid
	picker.GridCursorY = 3 // visible rows are 2 and 3
	zm.Scan(picker.View().Content)
	grid := zm.Get(ZoneGrid)
	deadline := time.Now().Add(time.Second)
	for grid == nil && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
		grid = zm.Get(ZoneGrid)
	}
	if grid == nil {
		t.Fatal("missing grid zone")
	}
	emojis := picker.filteredEmojis()
	for _, tc := range []struct {
		name string
		x, y int
		want int
	}{
		{"first visible row", 1, 0, 8},
		{"second visible row", 5, 2, 13},
		{"separator", 1, 1, -1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			updated, cmd := picker.Update(tea.MouseReleaseMsg{X: grid.StartX + tc.x, Y: grid.StartY + tc.y, Button: tea.MouseLeft})
			if tc.want < 0 {
				if cmd != nil || updated.(Model).GridCursorY != picker.GridCursorY {
					t.Fatal("separator selected an emoji")
				}
				return
			}
			if cmd == nil {
				t.Fatal("click did not select an emoji")
			}
			if got := cmd().(EmojiChangedMsg).Emoji; got != emojis[tc.want] {
				t.Fatalf("clicked %q, want %q", got, emojis[tc.want])
			}
		})
	}
}

func TestPickerFitsHeightAndKeepsGridSize(t *testing.T) {
	picker := New(WithPresets([]string{"😀"}), WithShowSearch(true))
	updated, _ := picker.Update(tea.WindowSizeMsg{Width: 56, Height: 27})
	picker = updated.(Model)
	w, h := picker.ViewSize()
	if w > 56 || h > 27 {
		t.Fatalf("picker size %dx%d exceeds available 56x27", w, h)
	}
	updated, _ = picker.Update(tea.WindowSizeMsg{Width: w, Height: h})
	if got := updated.(Model); got.GridCols != picker.GridCols || got.GridRows != picker.GridRows {
		t.Fatal("resizing to rendered dimensions changed grid dimensions")
	}
}
