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

func TestSwatchFitsTerminalWithCustomFrame(t *testing.T) {
	for _, custom := range []bool{false, true} {
		s := NewSwatchPicker("😀", "test")
		opts := []Option{WithPresets([]string{"😀"}), WithShowSearch(true)}
		if custom {
			opts = append(opts, WithStyle(lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(2, 4)))
		}
		s.SetPickerOptions(opts...)
		s.SetBounds(0, 0, 4, 1)
		s, _ = s.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
		s, _ = s.Update(tea.MouseClickMsg{X: 1, Y: 0, Button: tea.MouseLeft})
		for _, size := range []tea.WindowSizeMsg{{Width: 80, Height: 24}, {Width: 60, Height: 23}, {Width: 100, Height: 30}} {
			s, _ = s.Update(size)
			w, h := s.picker.ViewSize()
			if w > size.Width || h > size.Height {
				t.Fatalf("custom=%v: %dx%d exceeds %dx%d", custom, w, h, size.Width, size.Height)
			}
			if s.lastModalW != w || s.lastOverlayHeight != h {
				t.Fatal("stale modal bounds")
			}
		}
	}
}

func TestOverlayZoneBounds(t *testing.T) {
	zm := zone.New()
	defer zm.Close()
	marked := zm.Mark("outer", "a"+zm.Mark("inner", "bc")+"d")
	got := OverlayView("..........", marked, 10, 1, 0, 3)
	if strings.Count(got, "z") != strings.Count(marked, "z") {
		t.Fatalf("zone markers replayed: %q", got)
	}
	zm.Scan(got)
	deadline := time.Now().Add(time.Second)
	for zm.Get("outer") == nil && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	for id, want := range map[string][2]int{"outer": {3, 6}, "inner": {4, 5}} {
		z := zm.Get(id)
		if z == nil || z.StartX != want[0] || z.EndX != want[1] {
			t.Fatalf("%s bounds: %#v, want %v", id, z, want)
		}
	}
}

func TestResizeMatchesRenderedFit(t *testing.T) {
	for _, style := range []lipgloss.Style{lipgloss.NewStyle().Padding(2, 4), lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(1, 2), lipgloss.NewStyle().Width(60).Height(26), lipgloss.NewStyle().MaxWidth(30).MaxHeight(24)} {
		for _, size := range []tea.WindowSizeMsg{{Width: 20, Height: 24}, {Width: 40, Height: 24}, {Width: 80, Height: 30}} {
			picker := New(WithStyle(style), WithPresets([]string{"😀", "👍", "🔥", "✅", "❌", "🚨"}), WithShowSearch(true))
			updated, _ := picker.Update(size)
			got := updated.(Model)
			want := picker
			want.GridCols, want.GridRows = 12, 1
			for want.GridCols > 1 {
				w, _ := want.ViewSize()
				if w <= size.Width {
					break
				}
				want.GridCols--
			}
			want.GridRows = 8
			for want.GridRows > 1 {
				_, h := want.ViewSize()
				if h <= size.Height {
					break
				}
				want.GridRows--
			}
			if got.GridCols != want.GridCols || got.GridRows != want.GridRows {
				t.Fatalf("%v: got %dx%d grid, want %dx%d", size, got.GridCols, got.GridRows, want.GridCols, want.GridRows)
			}
		}
	}
}

func TestSwatchHitBoundsFollowRenderedOrigin(t *testing.T) {
	zm := zone.New()
	defer zm.Close()
	s := NewSwatchPicker("😀", "test")
	s.SetZoneManager(zm)
	s.SetBounds(35, 110, 4, 1)
	s, _ = s.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	s.OpenPicker()
	oldLeft := s.lastOverlayLeft
	zm.Scan(s.ViewWithOverlay(strings.Repeat(strings.Repeat(" ", 80)+"\n", 29), 80, 30))
	if s.lastOverlayLeft == oldLeft {
		t.Fatal("render did not update hit origin")
	}
	deadline := time.Now().Add(time.Second)
	for zm.Get(ZoneGrid) == nil && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	grid := zm.Get(ZoneGrid)
	if grid == nil {
		t.Fatal("missing rendered grid zone")
	}
	s, cmd := s.Update(tea.MouseClickMsg{X: grid.StartX + 1, Y: grid.StartY, Button: tea.MouseLeft})
	if !s.Open() || cmd != nil {
		t.Fatal("click on visible modal canceled it")
	}
	_, cmd = s.Update(tea.MouseReleaseMsg{X: grid.StartX + 1, Y: grid.StartY, Button: tea.MouseLeft})
	if cmd == nil {
		t.Fatal("visible emoji click did not select")
	}
	if _, ok := cmd().(EmojiChangedMsg); !ok {
		t.Fatal("expected emoji selection")
	}
}
