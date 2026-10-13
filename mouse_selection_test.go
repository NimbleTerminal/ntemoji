package ntemoji

import (
	"fmt"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	zone "github.com/lrstanley/bubblezone/v2"
)

func TestWrappedCategoryMouseSelection(t *testing.T) {
	for _, width := range []int{12, 16, 20, 24, 28, 80} {
		t.Run(fmt.Sprint(width), func(t *testing.T) {
			zm := zone.New()
			defer zm.Close()
			picker := New()
			picker.SetZoneManager(zm)
			updated, _ := picker.Update(tea.WindowSizeMsg{Width: width, Height: 50})
			picker = updated.(Model)
			block := scanPickerZone(t, picker, zm, ZoneCategories)
			lines := strings.Split(zm.Scan(picker.View().Content), "\n")
			for i, palette := range picker.Palettes {
				found := false
				for y := block.StartY; y <= block.EndY; y++ {
					before, _, ok := strings.Cut(lines[y], palette.Icon)
					if !ok {
						continue
					}
					found = true
					mouse := tea.Mouse{X: lipgloss.Width(before), Y: y, Button: tea.MouseLeft}
					for _, msg := range []tea.Msg{tea.MouseClickMsg(mouse), tea.MouseReleaseMsg(mouse)} {
						updated, _ = picker.Update(msg)
						picker = updated.(Model)
						if picker.ActiveTab != i {
							t.Fatalf("%T on %s selected tab %d, want %d", msg, palette.Name, picker.ActiveTab, i)
						}
					}
				}
				if !found {
					t.Fatalf("category %s was not rendered", palette.Name)
				}
			}
		})
	}
}

func scanPickerZone(t *testing.T, picker Model, zm *zone.Manager, id string) *zone.ZoneInfo {
	t.Helper()
	zm.Scan(picker.View().Content)
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if z := zm.Get(id); z != nil {
			return z
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("missing zone %s", id)
	return nil
}

func TestScrolledGridMouseSequenceKeepsViewport(t *testing.T) {
	for _, hover := range []bool{false, true} {
		zm := zone.New()
		picker := New()
		picker.SetZoneManager(zm)
		picker.GridCols, picker.GridRows = 4, 2
		picker.GridCursorY, picker.Focus = 3, FocusGrid
		grid := scanPickerZone(t, picker, zm, ZoneGrid)
		want := picker.Palettes[0].Emojis[8]
		mouse := tea.Mouse{X: grid.StartX + 1, Y: grid.StartY, Button: tea.MouseLeft}
		messages := []tea.Msg{tea.MouseClickMsg(mouse), tea.MouseReleaseMsg(mouse)}
		if hover {
			messages = append([]tea.Msg{tea.MouseMotionMsg(mouse), tea.MouseMotionMsg(mouse)}, messages...)
		}
		for _, msg := range messages {
			updated, cmd := picker.Update(msg)
			picker = updated.(Model)
			if picker.SelectedEmoji() != want {
				t.Fatalf("hover=%v, %T: cursor selected %q, want %q", hover, msg, picker.SelectedEmoji(), want)
			}
			// Repaint between events as the real application does. The emoji
			// beneath the pointer must remain in the first visible row.
			view := zm.Scan(picker.View().Content)
			if !strings.Contains(strings.Split(view, "\n")[grid.StartY], want) {
				t.Fatalf("%T scrolled the emoji away from the pointer", msg)
			}
			if _, release := msg.(tea.MouseReleaseMsg); release {
				if cmd == nil || cmd().(EmojiChangedMsg).Emoji != want {
					t.Fatal("release did not select the emoji under the pointer")
				}
			} else if cmd != nil {
				t.Fatalf("%T prematurely selected an emoji", msg)
			}
		}
		// Moving past the top still scrolls upward, and moving back within
		// the viewport must not move the viewport again.
		updated, _ := picker.Update(tea.KeyPressMsg{Code: tea.KeyUp})
		picker = updated.(Model)
		if picker.GridCursorY != 1 || picker.gridScrollOffset() != 1 {
			t.Fatal("keyboard navigation did not reveal the previous row")
		}
		updated, _ = picker.Update(tea.KeyPressMsg{Code: tea.KeyDown})
		picker = updated.(Model)
		if picker.GridCursorY != 2 || picker.gridScrollOffset() != 1 {
			t.Fatal("keyboard navigation moved the viewport unnecessarily")
		}
		zm.Close()
	}
}

func TestWrappedPresetMouseSelection(t *testing.T) {
	for _, cols := range []int{1, 2, 3, 8} {
		zm := zone.New()
		defer zm.Close()
		presets := []string{"😀", "👍", "🔥", "✅", "❌", "🚨", "A"}
		picker := New(WithPresets(presets))
		picker.SetZoneManager(zm)
		picker.GridCols = cols
		block := scanPickerZone(t, picker, zm, ZonePresets)
		for i, want := range presets {
			x, y := block.StartX+(i%cols)*4, block.StartY+i/cols
			view := zm.Scan(picker.View().Content)
			line := strings.Split(view, "\n")[y]
			// Check the actual glyph position independently of hit-testing.
			before, _, found := strings.Cut(line, want)
			if !found || lipgloss.Width(before) < x || lipgloss.Width(before) >= x+4 {
				t.Fatalf("cols=%d: preset %q is outside its expected cell", cols, want)
			}
			mouse := tea.Mouse{X: x + 1, Y: y, Button: tea.MouseLeft}
			for _, msg := range []tea.Msg{tea.MouseMotionMsg(mouse), tea.MouseClickMsg(mouse), tea.MouseReleaseMsg(mouse)} {
				updated, cmd := picker.Update(msg)
				picker = updated.(Model)
				if picker.PresetCursor != i {
					t.Fatalf("cols=%d, %T: cursor=%d, want %d", cols, msg, picker.PresetCursor, i)
				}
				if _, release := msg.(tea.MouseReleaseMsg); release {
					if cmd == nil || cmd().(EmojiChangedMsg).Emoji != want {
						t.Fatalf("cols=%d: release did not select %q", cols, want)
					}
				}
			}
		}
		if len(presets)%cols != 0 {
			x, y := block.StartX+(len(presets)%cols)*4, block.StartY+len(presets)/cols
			_, cmd := picker.Update(tea.MouseReleaseMsg{X: x, Y: y, Button: tea.MouseLeft})
			if cmd != nil {
				t.Fatal("blank cell in final preset row selected an emoji")
			}
		}
	}
}
