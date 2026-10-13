package ntemoji

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	zone "github.com/lrstanley/bubblezone/v2"
)

func TestAccessibilityFocusAndSelection(t *testing.T) {
	m := New(WithInitialEmoji("😀"), WithPresets([]string{"😀", "👍"}), WithShowSearch(true))
	updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	m = updated.(Model)
	s := m.AccessibilityState()
	if s.FocusedEmoji != "👍" || s.FocusedName != "thumbs up" || s.Position != 2 || s.Count != 2 || s.SelectedEmoji != "😀" {
		t.Fatalf("preset state: %+v", s)
	}
	if !strings.Contains(ansi.Strip(m.View().Content), "thumbs up") {
		t.Fatal("preview does not follow preset focus")
	}
	updated, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = updated.(Model)
	if cmd == nil || cmd().(EmojiChangedMsg).Emoji != "👍" || m.AccessibilityState().SelectedEmoji != "👍" {
		t.Fatal("confirmation not reflected")
	}
	m.Focus = FocusSearch
	m.SearchInput.SetValue("rocket")
	s = m.AccessibilityState()
	if s.Query != "rocket" || s.ResultCount == 0 || s.FocusedEmoji != "" || s.FocusedName != "Search" || s.Position != 0 {
		t.Fatalf("search: %+v", s)
	}
	m.Focus = FocusGrid
	s = m.AccessibilityState()
	if s.FocusedName != EmojiName(m.SelectedEmoji()) || s.Category == "" {
		t.Fatalf("search result: %+v", s)
	}
	m.SearchInput.SetValue("no-match-1234")
	s = m.AccessibilityState()
	if s.ResultCount != 0 || s.Position != 0 || s.FocusedEmoji != "" {
		t.Fatalf("empty: %+v", s)
	}
	m.Focus = FocusCategories
	s = m.AccessibilityState()
	if s.FocusedName != m.Palettes[m.ActiveTab].Name || s.Count != len(m.Palettes) {
		t.Fatalf("category: %+v", s)
	}
}

func TestListNavigationResizeAndMouse(t *testing.T) {
	zm := zone.New()
	defer zm.Close()
	emojis := []string{"😀", "👍", "🚀", "🔥", "🎉", "❤️"}
	m := New(WithListMode(true), WithPalettes([]Palette{{Name: "Examples", Emojis: emojis}}), WithInitialEmoji("👍"))
	m.SetZoneManager(zm)
	m.GridRows = 2
	for i := 0; i < 3; i++ {
		next, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
		m = next.(Model)
	}
	if m.SelectedEmoji() != "🎉" || m.AccessibilityState().Position != 5 {
		t.Fatalf("navigation: %+v", m.AccessibilityState())
	}
	if !strings.Contains(ansi.Strip(m.View().Content), "> 🎉 party popper") {
		t.Fatal("missing labeled focus marker")
	}
	z := scanPickerZone(t, m, zm, ZoneGrid)
	next, cmd := m.Update(tea.MouseReleaseMsg{X: z.StartX + 15, Y: z.StartY, Button: tea.MouseLeft})
	m = next.(Model)
	if cmd == nil || cmd().(EmojiChangedMsg).Emoji != "🔥" {
		t.Fatalf("scrolled row click selected %q", m.SelectedEmoji())
	}
	for _, width := range []int{30, 80, 20, 40} {
		next, _ = m.Update(tea.WindowSizeMsg{Width: width, Height: 40})
		m = next.(Model)
		if m.SelectedEmoji() != "🔥" {
			t.Fatal("resize changed focused emoji")
		}
		w, h := m.ViewSize()
		if w > width || h > 40 {
			t.Fatalf("view %dx%d exceeds %dx40", w, h, width)
		}
	}
}

func TestListCategoryZones(t *testing.T) {
	for _, width := range []int{20, 30, 80} {
		t.Run(fmt.Sprint(width), func(t *testing.T) {
			zm := zone.New()
			defer zm.Close()
			m := New(WithListMode(true))
			m.SetZoneManager(zm)
			next, _ := m.Update(tea.WindowSizeMsg{Width: width, Height: 60})
			m = next.(Model)
			for i := range m.Palettes {
				z := scanPickerZone(t, m, zm, fmt.Sprintf("picker-tab-%d", i))
				next, _ = m.Update(tea.MouseClickMsg{X: z.StartX, Y: z.StartY, Button: tea.MouseLeft})
				m = next.(Model)
				if m.ActiveTab != i {
					t.Fatalf("tab %d selected %d", i, m.ActiveTab)
				}
			}
			if lipgloss.Width(m.View().Content) > width {
				t.Fatal("category tabs exceeded width")
			}
		})
	}
}

func TestSwatchAccessibilityState(t *testing.T) {
	s := NewSwatchPicker("🚀", "Launch status")
	s.SetFocused(true)
	state := s.AccessibilityState()
	if state.Label != "Launch status" || state.Name != "rocket" || !state.Focused || state.Expanded || state.Picker != (AccessibilityState{}) {
		t.Fatalf("closed: %+v", state)
	}
	s.SetPickerOptions(WithListMode(true))
	s.OpenPicker()
	state = s.AccessibilityState()
	if !state.Expanded || state.Picker.FocusedName != "rocket" {
		t.Fatalf("open: %+v", state)
	}
	s, _ = s.Update(EmojiChangedMsg{Emoji: "👍"})
	state = s.AccessibilityState()
	if state.Expanded || state.Name != "thumbs up" || state.Picker != (AccessibilityState{}) {
		t.Fatalf("selected: %+v", state)
	}
}

func ExampleModel_AccessibilityState() {
	picker := New(WithListMode(true), WithInitialEmoji("🚀"))
	state := picker.AccessibilityState()
	// Read again after Update; compare snapshots before announcing to the host.
	fmt.Println(state.FocusedName)
	fmt.Println(state.SelectedName)
	// Output:
	// rocket
	// rocket
}

func TestFocusedPreviewDoesNotChangeSelection(t *testing.T) {
	for _, list := range []bool{false, true} {
		m := New(WithListMode(list), WithShowSearch(true), WithInitialEmoji("😀"), WithPalettes([]Palette{{Name: "Test", Emojis: []string{"😀", "👍"}}}))
		key := tea.KeyRight
		if list {
			key = tea.KeyDown
		}
		next, _ := m.Update(tea.KeyPressMsg{Code: key})
		m = next.(Model)
		before := m.AccessibilityState()
		view := ansi.Strip(m.View().Content)
		if !strings.Contains(view, "thumbs up") || before.FocusedName != "thumbs up" || before.SelectedEmoji != "😀" {
			t.Fatalf("list=%v: %+v\n%s", list, before, view)
		}
		if m.AccessibilityState() != before {
			t.Fatal("rendering mutated state")
		}
	}
}

func TestListLabelsKeepFullStateWhenClipped(t *testing.T) {
	m := New(WithListMode(true), WithInitialEmoji("👩🏽‍🚀"), WithPalettes([]Palette{{Name: "Astronauts", Emojis: []string{"👩🏽‍🚀"}}}))
	m.GridCols = 3
	m.GridRows = 1
	line := ansi.Strip(m.labeledResults(12, 1))
	if lipgloss.Width(line) != 12 || !strings.Contains(line, "👩🏽‍🚀") || !strings.HasSuffix(line, "…") {
		t.Fatalf("clipped row: %q", line)
	}
	if m.AccessibilityState().FocusedName != "woman astronaut: medium skin tone" {
		t.Fatalf("truncated state: %+v", m.AccessibilityState())
	}
}
