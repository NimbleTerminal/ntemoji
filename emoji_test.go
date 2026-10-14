package ntemoji

import (
	"charm.land/lipgloss/v2"
	"testing"
)

func TestToolConstantsAndSearch(t *testing.T) {
	if EmojiGear != "⚙️" || EmojiWrench != "🔧" || EmojiDesktop != "🖥️" {
		t.Fatal("incorrect tool or desktop glyph")
	}
	if !MatchEmoji(EmojiGear, "GEAR") || !MatchEmoji(EmojiWrench, "WRENCH") {
		t.Fatal("tool searches do not match corrected glyphs")
	}
	if MatchEmoji(EmojiWrench, "gear") {
		t.Fatal("wrench still has gear tags")
	}
}

func TestToolConstantsWidthAndRoundTrip(t *testing.T) {
	for _, emoji := range []string{EmojiGear, EmojiDesktop} {
		if width := lipgloss.Width(emoji); width != 2 {
			t.Fatalf("%q is %d columns", emoji, width)
		}
		text, err := DemojiString(emoji, WithDelimiters(":", ":"))
		if err != nil || text == emoji {
			t.Fatalf("%q did not encode: %q %v", emoji, text, err)
		}
		decoded, err := RemojiString(text)
		if err != nil || decoded != emoji {
			t.Fatalf("%q round trip: %q %v", emoji, decoded, err)
		}
	}
}
