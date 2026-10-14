package ntemoji

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

func TestRenderedRowWidths(t *testing.T) {
	picker := New(
		WithInitialEmoji("🚀"),
		WithPresets([]string{"😀", "👍", "🔥", "✅", "❌", "🚨"}),
		WithShowSearch(true),
	)

	view := picker.View().Content
	lines := strings.Split(view, "\n")

	want := lipgloss.Width(view)
	for i, line := range lines {
		if got := lipgloss.Width(line); got != want {
			t.Errorf("line %d width = %d, want %d", i, got, want)
		}
	}
}

func TestOverlayWithAnsi(t *testing.T) {
	bgStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
	bgLine := bgStyle.Render("012345678901234567890123456789")
	bgView := bgLine + "\n" + bgLine + "\n" + bgLine

	modalView := "XXXX\nXXXX"

	result := OverlayView(bgView, modalView, 30, 3, 1, 5)
	resultLines := strings.Split(result, "\n")
	want := "012345678901234567890123456789\n01234XXXX901234567890123456789\n01234XXXX901234567890123456789"
	if got := ansi.Strip(result); got != want {
		t.Fatalf("overlay content: %q, want %q", got, want)
	}

	for i, line := range resultLines {
		w := lipgloss.Width(line)
		if w != 30 {
			t.Errorf("Expected line %d to have visual width 30, got %d", i, w)
		}
	}
}

func TestAllDefaultEmojisWidths(t *testing.T) {
	for _, p := range DefaultPalettes {
		for _, e := range p.Emojis {
			lw := lipgloss.Width(e)
			if lw != 2 {
				t.Errorf("Palette %s emoji %s: width=%d, want 2", p.Name, e, lw)
			}
		}
	}
}
