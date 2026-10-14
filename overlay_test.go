package ntemoji

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	zone "github.com/lrstanley/bubblezone/v2"
)

func TestOverlayClipsViewport(t *testing.T) {
	for _, tc := range []struct {
		name, bg, fg, want string
		w, h, top, left    int
	}{
		{"right", "abcdef", "XYZ", "abcXY", 5, 1, 0, 3},
		{"wide right", ".....", "a👩‍💻", "....a ", 6, 1, 0, 4},
		{"wide left", "......", "👩‍💻X", " X....", 6, 1, 0, -1},
		{"background wide edge", "ab🇺🇸", "", "ab ", 3, 1, 0, 0},
		{"top and bottom", "abc\ndef\nghi", "123\n456\n789\nXYZ", "456\n789", 3, 2, -1, 0},
		{"offscreen", "abcdef", "XYZ", "abc", 3, 1, 0, 5},
		{"zero width", "abc", "X", "", 0, 1, 0, 0},
		{"zero height", "abc", "X", "", 3, 0, 0, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := OverlayView(tc.bg, tc.fg, tc.w, tc.h, tc.top, tc.left)
			if ansi.Strip(got) != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestClippedOverlayZones(t *testing.T) {
	for _, offset := range []int{0, -2} {
		zm := zone.New()
		defer zm.Close()
		modal := zm.Mark("outer", "ab"+zm.Mark("inner", "cdef")+zm.Mark("hidden", "gh")+"\nabcdefgh\nabcdefgh")
		view := OverlayView("", modal, 4, 2, offset, offset)
		zm.Scan(view)
		deadline := time.Now().Add(time.Second)
		for zm.Get("outer") == nil && time.Now().Before(deadline) {
			time.Sleep(time.Millisecond)
		}
		outer := zm.Get("outer")
		wantBottom := 1
		if offset == -2 {
			wantBottom = 0
		}
		if outer == nil || outer.StartX != 0 || outer.EndX != 3 || outer.StartY != 0 || outer.EndY != wantBottom {
			t.Fatalf("offset %d: clipped outer bounds %+v", offset, outer)
		}
		if zm.Get("hidden") != nil {
			t.Fatal("fully clipped zone survived")
		}
		if offset == 0 {
			inner := zm.Get("inner")
			if inner == nil || inner.StartX != 2 || inner.EndX != 3 || inner.StartY != 0 || inner.EndY != 0 {
				t.Fatalf("clipped inner bounds %+v", inner)
			}
		} else if zm.Get("inner") != nil {
			t.Fatal("vertically clipped zone survived")
		}
	}
}

func TestSwatchOverlayClipsViewAndHitBounds(t *testing.T) {
	zm := zone.New()
	defer zm.Close()
	s := NewSwatchPicker("😀", "test")
	s.SetZoneManager(zm)
	s.Update(tea.WindowSizeMsg{Width: 80, Height: 30})
	s.OpenPicker()
	view := s.ViewWithOverlay("", 30, 14)
	if w, h := lipgloss.Size(view); w > 30 || h > 14 {
		t.Fatalf("clipped view is %dx%d, want at most 30x14", w, h)
	}
	if s.lastModalW != 30 || s.lastOverlayHeight != 14 {
		t.Fatalf("hit bounds are %dx%d", s.lastModalW, s.lastOverlayHeight)
	}
	zm.Scan(view)
	deadline := time.Now().Add(time.Second)
	for zm.Get(ZoneGrid) == nil && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	grid := zm.Get(ZoneGrid)
	if grid == nil || grid.EndX >= 30 || grid.EndY >= 14 {
		t.Fatalf("clipped grid zone: %+v", grid)
	}
	s, _ = s.Update(tea.MouseClickMsg{X: grid.StartX + 1, Y: grid.StartY, Button: tea.MouseLeft})
	_, cmd := s.Update(tea.MouseReleaseMsg{X: grid.StartX + 1, Y: grid.StartY, Button: tea.MouseLeft})
	if cmd == nil || cmd().(EmojiChangedMsg).Emoji != "😀" {
		t.Fatal("visible grid emoji lost its mouse target after clipping")
	}
}

func TestOverlayGraphemeWidths(t *testing.T) {
	for _, text := range []string{"👩‍💻", "👨‍👩‍👧‍👦", "👍🏽", "🇺🇸", "1️⃣", "e\u0301", "⚠️", "\x1b[31m👩‍💻\x1b[0m"} {
		w, h := ModalCellSize(text)
		if w != lipgloss.Width(text) || h != 1 {
			t.Errorf("%q: overlay %dx%d, Lip Gloss width %d", text, w, h, lipgloss.Width(text))
		}
		if got := renderOverlayCells(parseOverlayLine(text)); ansi.Strip(got) != ansi.Strip(text) {
			t.Errorf("round trip split grapheme %q: %q", text, got)
		}
	}
}

func TestOverlayCombinedEmojiBoundaries(t *testing.T) {
	for _, tc := range []struct {
		name, bg, fg string
		left         int
		want         string
	}{
		{"insert", "abcdef", "👩‍💻", 2, "ab👩‍💻ef"},
		{"overwrite first half", "a👩‍💻bcd", "X", 1, "aX bcd"},
		{"overwrite second half", "a👩‍💻bcd", "X", 2, "a Xbcd"},
		{"modifier", "abcdef", "👍🏽", 2, "ab👍🏽ef"},
		{"flag", "abcdef", "🇺🇸", 2, "ab🇺🇸ef"},
		{"combining accent", "abcdef", "e\u0301", 2, "abe\u0301def"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := OverlayView(tc.bg, tc.fg, 6, 1, 0, tc.left)
			if got != tc.want || lipgloss.Width(got) != 6 {
				t.Fatalf("got %q (width %d), want %q", got, lipgloss.Width(got), tc.want)
			}
		})
	}
}

func TestOverlayPreservesANSIWithGraphemes(t *testing.T) {
	got := OverlayView("abcdef", "\x1b[31m👩‍💻\x1b[0m", 6, 1, 0, 2)
	if ansi.Strip(got) != "ab👩‍💻ef" || !strings.Contains(got, "\x1b[31m👩‍💻\x1b[0m") {
		t.Fatalf("lost content or styling: %q", got)
	}
}

func TestOverlayPreservesZeroWidthText(t *testing.T) {
	for _, value := range []string{"\u200d", "\ufe0f", "\u200dA", "\ufe0fA"} {
		if got := renderOverlayCells(parseOverlayLine(value)); got != value {
			t.Errorf("round trip %q = %q", value, got)
		}
		w, _ := ModalCellSize(value)
		if w != lipgloss.Width(value) {
			t.Errorf("width %q = %d", value, w)
		}
	}
}

func TestOverlayPreservesBackgroundZoneEndpoints(t *testing.T) {
	const marker = "\x1b[1001z"
	for _, tc := range []struct {
		bg, fg, want string
		width, left  int
	}{
		{marker + "abcdef" + marker, "XY", marker + "XYcdef" + marker, 6, 0},
		{marker + "abcdef" + marker, "XY", marker + "abcdXY" + marker, 6, 4},
		{marker + "abcdef" + marker, "XY", marker + "abXY" + marker, 4, 2},
		{marker + "👩‍💻ab" + marker, "X", marker + " Xab" + marker, 4, 1},
		{marker + "ab👩‍💻" + marker, "X", marker + "abX " + marker, 4, 2},
	} {
		got := OverlayView(tc.bg, tc.fg, tc.width, 1, 0, tc.left)
		if got != tc.want {
			t.Errorf("overlay %q at %d: got %q, want %q", tc.bg, tc.left, got, tc.want)
		}
	}
}
