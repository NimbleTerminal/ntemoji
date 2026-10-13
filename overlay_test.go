package ntemoji

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

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
