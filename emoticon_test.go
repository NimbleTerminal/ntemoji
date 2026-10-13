package ntemoji

import "testing"

func TestEmoticonConversions(t *testing.T) {
	for _, tc := range []struct{ input, emoji, preferred string }{
		{":)", "🙂", ":)"},
		{":-)", "🙂", ":)"},
		{":-(", "🙁", ":("},
		{":-D", "😃", ":D"},
		{";-)", "😉", ";)"},
		{"xD", "😆", "XD"},
		{":-p", "😛", ":P"},
		{":o", "😮", ":O"},
		{":-\\", "😕", ":/"},
		{":-|", "😐", ":|"},
		{":'-(", "😢", ":'("},
		{">:-(", "😠", ">:("},
		{":-*", "😘", ":*"},
		{"8-)", "😎", "B)"},
		{"<3", "❤️", "<3"},
		{"</3", "💔", "</3"},
	} {
		if got := EmoticonToEmoji(tc.input); got != tc.emoji {
			t.Errorf("EmoticonToEmoji(%q) = %q, want %q", tc.input, got, tc.emoji)
		}
		if got := EmojiToEmoticon(tc.emoji); got != tc.preferred {
			t.Errorf("EmojiToEmoticon(%q) = %q, want %q", tc.emoji, got, tc.preferred)
		}
	}
}

func TestEmoticonUnknownInputsPassThrough(t *testing.T) {
	for _, value := range []string{"", "hello", "hi :)", " :) ", "🙂🙂", "🚀", "https://example.com"} {
		if got := EmoticonToEmoji(value); got != value {
			t.Errorf("EmoticonToEmoji(%q) unexpectedly returned %q", value, got)
		}
		if got := EmojiToEmoticon(value); got != value {
			t.Errorf("EmojiToEmoticon(%q) unexpectedly returned %q", value, got)
		}
	}
}

func TestEmoticonTableConsistency(t *testing.T) {
	seenEmoji, seenSpelling := map[string]bool{}, map[string]bool{}
	for _, entry := range emoticons {
		if entry.emoji == "" || len(entry.spellings) == 0 || seenEmoji[entry.emoji] {
			t.Fatalf("invalid or duplicate entry for %q", entry.emoji)
		}
		seenEmoji[entry.emoji] = true
		if EmojiName(entry.emoji) == entry.emoji || len(Tags(entry.emoji)) == 0 {
			t.Errorf("conversion result %q lacks descriptive metadata", entry.emoji)
		}
		for _, spelling := range entry.spellings {
			if spelling == "" || seenSpelling[spelling] {
				t.Fatalf("empty or duplicate emoticon %q", spelling)
			}
			seenSpelling[spelling] = true
			for _, r := range spelling {
				if r < 33 || r > 126 {
					t.Errorf("emoticon %q is not printable ASCII", spelling)
				}
			}
			if got := EmoticonToEmoji(EmojiToEmoticon(EmoticonToEmoji(spelling))); got != entry.emoji {
				t.Errorf("round trip for %q returned %q", spelling, got)
			}
		}
	}
}
