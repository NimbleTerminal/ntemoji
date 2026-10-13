package ntemoji

import "testing"

func TestToolConstantsAndSearch(t *testing.T) {
	if EmojiGear != "⚙" || EmojiWrench != "🔧" || EmojiDesktop != "🖥" {
		t.Fatal("incorrect tool or desktop glyph")
	}
	if !MatchEmoji(EmojiGear, "GEAR") || !MatchEmoji(EmojiWrench, "WRENCH") {
		t.Fatal("tool searches do not match corrected glyphs")
	}
	if MatchEmoji(EmojiWrench, "gear") {
		t.Fatal("wrench still has gear tags")
	}
}
